package main

import (
	"flag"
	"fmt"
	"os"
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

	case "demo":
		fs := flag.NewFlagSet("demo", flag.ContinueOnError)
		out := fs.String("out", "./out/demo", "output directory")
		_ = fs.Parse(os.Args[2:])
		_ = out
		// v0.1.0 will implement demo runner over fixtures
		fmt.Println("demo: not implemented yet")
		os.Exit(1)

	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println()
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("proof-first-normalizer")
	fmt.Println()
	fmt.Println("Commands (planned v0.1.0):")
	fmt.Println("  normalizer normalize --in <raw.csv> --schema <schema.json> --out <dir>")
	fmt.Println("  normalizer validate  --in <raw.csv> --schema <schema.json>")
	fmt.Println("  normalizer demo      --out <dir>")
	fmt.Println("  normalizer version   (--version, -v)")
}
