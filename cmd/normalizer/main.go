package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

	inRoot := filepath.Join(root, "fixtures", "input")
	entries, err := os.ReadDir(inRoot)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}

	cases := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			cases = append(cases, e.Name())
		}
	}
	sort.Strings(cases)
	if len(cases) == 0 {
		fmt.Println("ERROR: no fixture cases found under", inRoot)
		os.Exit(2)
	}

	for _, c := range cases {
		inCSV := filepath.Join(inRoot, c, "raw.csv")
		schemaFile := filepath.Join(inRoot, c, "schema.json")
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

		wantErrPath := filepath.Join(expDir, "error.txt")
		if wantErr, errRead := os.ReadFile(wantErrPath); errRead == nil {
			// Expected-fail case: NormalizeCSV must return an error matching fixtures/expected/<case>/error.txt byte-for-byte.
			_, gotErr := normalizer.NormalizeCSV(inCSV, schemaFile, outDir, opt)
			if gotErr == nil {
				fmt.Printf("MISMATCH: %s expected failure but got success\n", c)
				os.Exit(1)
			}
			gotErrB := []byte(gotErr.Error() + "\n")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(2)
			}
			if err := os.WriteFile(filepath.Join(outDir, "error.txt"), gotErrB, 0o644); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(2)
			}
			if !bytes.Equal(gotErrB, wantErr) {
				fmt.Printf("MISMATCH: %s (error.txt)\n", c)
				os.Exit(1)
			}
			continue
		} else if errRead != nil && !os.IsNotExist(errRead) {
			fmt.Printf("ERROR: %s: %v\n", c, errRead)
			os.Exit(2)
		}

		if _, err := normalizer.NormalizeCSV(inCSV, schemaFile, outDir, opt); err != nil {
			fmt.Printf("ERROR: %s: %v\n", c, err)
			os.Exit(2)
		}

		// Compare outputs byte-for-byte.
		for _, name := range []string{"normalized.csv", "errors.csv", "report.json"} {
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
	fmt.Println("Demo:")
	fmt.Println("  Scans fixtures/input/* (sorted) and verifies outputs match fixtures/expected/*.")
}
