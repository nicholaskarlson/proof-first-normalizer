package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nicholaskarlson/proof-first-normalizer/internal/normalizer"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("proof-first-normalizer %s\n", version)
		return

	case "help", "-h", "--help":
		usage()
		return

	case "validate":
		cmdValidate(os.Args[2:])

	case "normalize":
		cmdNormalize(os.Args[2:])

	case "demo":
		cmdDemo(os.Args[2:])

	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println()
		usage()
		os.Exit(2)
	}
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	in := fs.String("in", "", "input CSV path")
	schema := fs.String("schema", "", "schema JSON path")
	label := fs.String("label", "", "stable label for report/logging")
	_ = fs.Parse(args)

	if *in == "" || *schema == "" {
		fmt.Println("validate: --in and --schema are required")
		os.Exit(2)
	}

	res, errs, err := normalizer.ValidateCSV(*in, *schema, *label)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}

	if len(errs) == 0 {
		fmt.Printf("OK: %d rows, %d cols\n", res.RowsTotal, res.Cols)
		os.Exit(0)
	}
	fmt.Printf("FAIL: %d row(s) with errors\n", res.RowsError)
	os.Exit(1)
}

func cmdNormalize(args []string) {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	in := fs.String("in", "", "input CSV path")
	schema := fs.String("schema", "", "schema JSON path")
	out := fs.String("out", "", "output directory")
	label := fs.String("label", "", "stable label recorded in report.json")
	_ = fs.Parse(args)

	if *in == "" || *schema == "" || *out == "" {
		fmt.Println("normalize: --in, --schema, and --out are required")
		os.Exit(2)
	}

	opt := normalizer.Options{
		Tool:    "proof-first-normalizer",
		Version: version,
		Label:   *label,
		Schema:  *schema,
		Input:   *in,
	}

	res, err := normalizer.NormalizeCSV(*in, *schema, *out, opt)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}

	fmt.Printf("Wrote: %s (ok=%d, errors=%d)\n", *out, res.RowsOK, res.RowsError)
	if res.RowsError > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func cmdDemo(args []string) {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	outRoot := fs.String("out", "", "output root directory")
	_ = fs.Parse(args)

	if *outRoot == "" {
		fmt.Println("demo: --out is required")
		os.Exit(2)
	}

	root, err := findRepoRoot()
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}

	// Keep this list small; add more cases incrementally via separate PRs.
	cases := []string{

		"case01",
		"case02_errors",
		"case03_bom_crlf",
		"case04_quoted_fields",
		"case05_optional_blanks",
		"case06_dup_headers",
		"case07_schema_dup_columns",
		"case08_blank_lines",
		"case09_ragged_rows",
		"case10_header_reordered",
		"case11_header_extra_column",
		"case12_header_whitespace",
	}

	for _, c := range cases {
		inCSV := filepath.Join(root, "fixtures", "input", c, "raw.csv")
		schemaFile := filepath.Join(root, "fixtures", "input", c, "schema.json")
		expDir := filepath.Join(root, "fixtures", "expected", c)
		outDir := filepath.Join(*outRoot, c)

		opt := normalizer.Options{
			Tool:    "proof-first-normalizer",
			Version: version,
			Label:   c,
			// record stable, repo-relative strings in report.json
			Schema: filepath.ToSlash(filepath.Join("fixtures", "input", c, "schema.json")),
			Input:  filepath.ToSlash(filepath.Join("fixtures", "input", c, "raw.csv")),
		}

		_ = os.RemoveAll(outDir)

		expErrPath := filepath.Join(expDir, "error.txt")
		if b, err := os.ReadFile(expErrPath); err == nil {
			// Expected failure case: NormalizeCSV must return an error matching fixtures/expected/<case>/error.txt.
			_, gotErr := normalizer.NormalizeCSV(inCSV, schemaFile, outDir, opt)
			if gotErr == nil {
				fmt.Printf("MISMATCH: %s expected failure but got success\n", c)
				os.Exit(1)
			}
			exp := strings.TrimSpace(string(b))
			got := strings.TrimSpace(gotErr.Error())
			if got != exp {
				fmt.Printf("MISMATCH: %s (error)\n  got: %s\n  exp: %s\n", c, got, exp)
				os.Exit(1)
			}
			continue
		} else if err != nil && !os.IsNotExist(err) {
			fmt.Printf("ERROR: %s: %v\n", c, err)
			os.Exit(2)
		}

		if _, err := normalizer.NormalizeCSV(inCSV, schemaFile, outDir, opt); err != nil {
			fmt.Printf("ERROR: %s: %v\n", c, err)
			os.Exit(2)
		}

		// Compare normalized/errors byte-for-byte.
		for _, name := range []string{"normalized.csv", "errors.csv"} {
			exp := filepath.Join(expDir, name)
			got := filepath.Join(outDir, name)
			eq, err := filesEqual(exp, got)
			if err != nil {
				fmt.Printf("ERROR: %s: %v\n", name, err)
				os.Exit(2)
			}
			if !eq {
				fmt.Printf("MISMATCH: %s (%s)\n", c, name)
				os.Exit(1)
			}
		}

		// Compare report.json structurally (ignore version so demo works on released binaries too).
		expRep, err := readReport(filepath.Join(expDir, "report.json"))
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(2)
		}
		gotRep, err := readReport(filepath.Join(outDir, "report.json"))
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(2)
		}
		if !reportsEqualIgnoringVersion(expRep, gotRep) {
			fmt.Printf("MISMATCH: %s (report.json)\n", c)
			os.Exit(1)
		}
	}

	fmt.Printf("OK: demo outputs match fixtures (%d case(s))\n", len(cases))
	os.Exit(0)
}

func filesEqual(a, b string) (bool, error) {
	ab, err := os.ReadFile(a)
	if err != nil {
		return false, err
	}
	bb, err := os.ReadFile(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(ab, bb), nil
}

func readReport(path string) (normalizer.Report, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return normalizer.Report{}, err
	}
	var r normalizer.Report
	if err := json.Unmarshal(b, &r); err != nil {
		return normalizer.Report{}, err
	}
	return r, nil
}

func reportsEqualIgnoringVersion(a, b normalizer.Report) bool {
	a.Version = ""
	b.Version = ""
	return reflect.DeepEqual(a, b)
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not locate repo root from %s", wd)
}

func usage() {
	fmt.Println("proof-first-normalizer")
	fmt.Println()
	fmt.Println("Commands (v0.1.0):")
	fmt.Println("  normalizer normalize --in <raw.csv> --schema <schema.json> --out <dir> [--label <string>]")
	fmt.Println("  normalizer validate  --in <raw.csv> --schema <schema.json> [--label <string>]")
	fmt.Println("  normalizer demo      --out <dir>")
	fmt.Println("  normalizer version   (--version, -v)")

	fmt.Println()
	fmt.Println("Demo cases:")
	fmt.Println("  " + strings.Join([]string{"case01", "case02_errors"}, ", "))
}
