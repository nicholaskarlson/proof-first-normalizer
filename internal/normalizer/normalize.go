package normalizer

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type Options struct {
	Tool    string
	Version string
	Label   string // stable input label (recommended)
	Schema  string // schema path (as provided)
	Input   string // input path (as provided)
}

type Result struct {
	RowsTotal int
	RowsOK    int
	RowsError int
	Cols      int
}

type rowErr struct {
	Row     int
	Field   string
	Code    string
	Message string
	Value   string
}


func isBlankRecord(rec []string) bool {
	if len(rec) == 0 {
		return true
	}
	for _, f := range rec {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}


// Report is a struct (not a map) to guarantee stable JSON field ordering.
type Report struct {
	Tool             string   `json:"tool"`
	Version          string   `json:"version"`
	Input            string   `json:"input"`
	Schema           string   `json:"schema"`
	RowsTotal        int      `json:"rows_total"`
	RowsOK           int      `json:"rows_ok"`
	RowsError        int      `json:"rows_error"`
	Cols             int      `json:"cols"`
	Sha256Input      string   `json:"sha256_input"`
	Sha256Schema     string   `json:"sha256_schema"`
	Sha256Normalized string   `json:"sha256_normalized"`
	Sha256Errors     string   `json:"sha256_errors"`
	GeneratedFiles   []string `json:"generated_files"`
}

func ValidateCSV(inPath, schemaPath, _ string) (Result, []rowErr, error) {
	schema, _, err := LoadSchema(schemaPath)
	if err != nil {
		return Result{}, nil, err
	}

	raw, err := os.ReadFile(inPath)
	if err != nil {
		return Result{}, nil, err
	}
	raw = canonicalizeBytes(raw)

	r := csv.NewReader(bytes.NewReader(raw))
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return Result{}, nil, fmt.Errorf("read header: %w", err)
	}
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}

	hmap := make(map[string]int, len(header))
	for i, h := range header {
		if h == "" {
			return Result{}, nil, fmt.Errorf("header has empty column name")
		}
		if _, ok := hmap[h]; ok {
			return Result{}, nil, fmt.Errorf("header has duplicate column %q", h)
		}
		hmap[h] = i
	}

	// v0.1.0: header must match schema exactly (no extras, no silent drops).
	for _, c := range schema.Columns {
		if _, ok := hmap[c.Name]; !ok {
			return Result{}, nil, fmt.Errorf("header missing required column %q", c.Name)
		}
	}
	if len(header) != len(schema.Columns) {
		return Result{}, nil, fmt.Errorf("header must match schema columns exactly (got %d, want %d)", len(header), len(schema.Columns))
	}

	colOrder := make([]int, len(schema.Columns))
	for i, c := range schema.Columns {
		colOrder[i] = hmap[c.Name]
	}

	var errs []rowErr
	rowsTotal, rowsOK, rowsErr := 0, 0, 0

	rowNum := 1 // header row is 1
	for {
		rec, e := r.Read()
		if e != nil {
			if errors.Is(e, io.EOF) {
				break
			}
			return Result{}, nil, fmt.Errorf("read row: %w", e)
		}
		rowNum++
		if isBlankRecord(rec) {
			continue
		}
		rowsTotal++

		if len(rec) != len(header) {
			rowsErr++
			errs = append(errs, rowErr{
				Row:     rowNum,
				Field:   "",
				Code:    "ERR_COLUMNS",
				Message: "wrong number of columns",
				Value:   fmt.Sprintf("%d", len(rec)),
			})
			continue
		}

		rowHasErr := false
		for i, c := range schema.Columns {
			v := strings.TrimSpace(rec[colOrder[i]])

			if c.Required && v == "" {
				rowHasErr = true
				errs = append(errs, rowErr{
					Row:     rowNum,
					Field:   c.Name,
					Code:    "ERR_REQUIRED",
					Message: "required value missing",
					Value:   v,
				})
				continue
			}
			if v == "" {
				continue
			}

			switch c.Type {
			case "string":
				// ok
			case "date":
				if _, pe := time.Parse("2006-01-02", v); pe != nil {
					rowHasErr = true
					errs = append(errs, rowErr{
						Row:     rowNum,
						Field:   c.Name,
						Code:    "ERR_DATE",
						Message: "invalid date (want YYYY-MM-DD)",
						Value:   v,
					})
				}
			case "decimal":
				if !looksDecimal(v) {
					rowHasErr = true
					errs = append(errs, rowErr{
						Row:     rowNum,
						Field:   c.Name,
						Code:    "ERR_DECIMAL",
						Message: "invalid decimal",
						Value:   v,
					})
				}
			}
		}

		if rowHasErr {
			rowsErr++
		} else {
			rowsOK++
		}
	}

	sort.Slice(errs, func(i, j int) bool {
		if errs[i].Row != errs[j].Row {
			return errs[i].Row < errs[j].Row
		}
		if errs[i].Field != errs[j].Field {
			return errs[i].Field < errs[j].Field
		}
		return errs[i].Code < errs[j].Code
	})

	return Result{RowsTotal: rowsTotal, RowsOK: rowsOK, RowsError: rowsErr, Cols: len(schema.Columns)}, errs, nil
}

func NormalizeCSV(inPath, schemaPath, outDir string, opt Options) (Result, error) {
	schema, schemaBytes, err := LoadSchema(schemaPath)
	if err != nil {
		return Result{}, err
	}

	raw, err := os.ReadFile(inPath)
	if err != nil {
		return Result{}, err
	}
	raw = canonicalizeBytes(raw)

	// Validate first (gives deterministic error ordering).
	res, errs, err := ValidateCSV(inPath, schemaPath, opt.Label)
	if err != nil {
		return Result{}, err
	}

	// Read header + build column map/order.
	r := csv.NewReader(bytes.NewReader(raw))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return Result{}, fmt.Errorf("read header: %w", err)
	}
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}
	hmap := make(map[string]int, len(header))
	for i, h := range header {
		if h == "" {
			return Result{}, fmt.Errorf("header has empty column name")
		}
		if _, ok := hmap[h]; ok {
			return Result{}, fmt.Errorf("header has duplicate column %q", h)
		}
		hmap[h] = i
	}

	colOrder := make([]int, len(schema.Columns))
	for i, c := range schema.Columns {
		colOrder[i] = hmap[c.Name]
	}

	// Mark bad rows from errs.
	bad := make(map[int]bool)
	for _, e := range errs {
		bad[e.Row] = true
	}

	// normalized.csv
	var normBuf bytes.Buffer
	w := csv.NewWriter(&normBuf)

	outHeader := make([]string, len(schema.Columns))
	for i, c := range schema.Columns {
		outHeader[i] = c.Name
	}
	if err := w.Write(outHeader); err != nil {
		return Result{}, err
	}

	// Re-parse rows to emit normalized output for OK rows only.
	r2 := csv.NewReader(bytes.NewReader(raw))
	r2.FieldsPerRecord = -1
	_, _ = r2.Read() // header

	rowNum := 1
	for {
		rec, e := r2.Read()
		if e != nil {
			if errors.Is(e, io.EOF) {
				break
			}
			return Result{}, fmt.Errorf("read row: %w", e)
		}
		rowNum++
		if isBlankRecord(rec) {
			continue
		}
		if bad[rowNum] {
			continue
		}

		outRec := make([]string, len(schema.Columns))
		for i, c := range schema.Columns {
			v := strings.TrimSpace(rec[colOrder[i]])
			if v == "" {
				outRec[i] = ""
				continue
			}
			switch c.Type {
			case "string":
				outRec[i] = v
			case "date":
				t, _ := time.Parse("2006-01-02", v)
				outRec[i] = t.Format("2006-01-02")
			case "decimal":
				outRec[i] = canonicalDecimal2(v)
			default:
				outRec[i] = v
			}
		}
		if err := w.Write(outRec); err != nil {
			return Result{}, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return Result{}, err
	}
	normalizedBytes := normBuf.Bytes()

	// errors.csv (always emitted)
	var errBuf bytes.Buffer
	ew := csv.NewWriter(&errBuf)
	_ = ew.Write([]string{"row", "field", "code", "message", "value"})
	for _, e := range errs {
		_ = ew.Write([]string{
			fmt.Sprintf("%d", e.Row),
			e.Field,
			e.Code,
			e.Message,
			e.Value,
		})
	}
	ew.Flush()
	if err := ew.Error(); err != nil {
		return Result{}, err
	}
	errorsBytes := errBuf.Bytes()

	// report.json (stable ordering via struct)
	inputLabel := opt.Label
	if inputLabel == "" {
		inputLabel = opt.Input
	}
	rep := Report{
		Tool:             opt.Tool,
		Version:          opt.Version,
		Input:            inputLabel,
		Schema:           opt.Schema,
		RowsTotal:        res.RowsTotal,
		RowsOK:           res.RowsOK,
		RowsError:        res.RowsError,
		Cols:             res.Cols,
		Sha256Input:      sha256Hex(raw),
		Sha256Schema:     sha256Hex(schemaBytes),
		Sha256Normalized: sha256Hex(normalizedBytes),
		Sha256Errors:     sha256Hex(errorsBytes),
		GeneratedFiles:   []string{"normalized.csv", "errors.csv", "report.json"},
	}

	repBytes, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return Result{}, err
	}
	repBytes = append(repBytes, '\n')

	// Write outputs (atomic).
	if err := writeFileAtomic(outDir, "normalized.csv", normalizedBytes); err != nil {
		return Result{}, err
	}
	if err := writeFileAtomic(outDir, "errors.csv", errorsBytes); err != nil {
		return Result{}, err
	}
	if err := writeFileAtomic(outDir, "report.json", repBytes); err != nil {
		return Result{}, err
	}

	return res, nil
}

func looksDecimal(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}
	dot := false
	digits := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
		digits++
	}
	return digits > 0
}

// v0.1.0: deterministic formatting to 2 decimals (pad/truncate, no rounding).
func canonicalDecimal2(s string) string {
	s = strings.TrimSpace(s)
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = strings.TrimPrefix(s, "-")
	}
	parts := strings.SplitN(s, ".", 2)
	intp := parts[0]
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if intp == "" {
		intp = "0"
	}
	intp = strings.TrimLeft(intp, "0")
	if intp == "" {
		intp = "0"
	}
	if len(frac) >= 2 {
		frac = frac[:2]
	} else if len(frac) == 1 {
		frac = frac + "0"
	} else {
		frac = "00"
	}
	out := intp + "." + frac
	if neg && out != "0.00" {
		out = "-" + out
	}
	return out
}
