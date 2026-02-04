package main

import (
	"flag"
	"fmt"
	"os"

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
		// v0.1.0 PR3: implement demo runner over fixtures
		fmt.Println("demo: not implemented yet")
		os.Exit(1)

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

func usage() {
	fmt.Println("proof-first-normalizer")
	fmt.Println()
	fmt.Println("Commands (v0.1.0):")
	fmt.Println("  normalizer normalize --in <raw.csv> --schema <schema.json> --out <dir> [--label <string>]")
	fmt.Println("  normalizer validate  --in <raw.csv> --schema <schema.json> [--label <string>]")
	fmt.Println("  normalizer demo      --out <dir>")
	fmt.Println("  normalizer version   (--version, -v)")
}
