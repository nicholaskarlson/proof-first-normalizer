package tests

import (
	"path/filepath"
	"testing"

	"github.com/nicholaskarlson/proof-first-normalizer/internal/normalizer"
)

func TestGoldenCase12HeaderWhitespace(t *testing.T) {
	root := projectRoot(t)

	caseName := "case12_header_whitespace"
	inCSV := filepath.Join(root, "fixtures", "input", caseName, "raw.csv")
	schemaFile := filepath.Join(root, "fixtures", "input", caseName, "schema.json")
	expDir := filepath.Join(root, "fixtures", "expected", caseName)

	outDir := t.TempDir()

	opt := normalizer.Options{
		Tool:    "proof-first-normalizer",
		Version: "dev",
		Label:   caseName,
		// record stable, repo-relative strings in report.json
		Schema: "fixtures/input/" + caseName + "/schema.json",
		Input:  "fixtures/input/" + caseName + "/raw.csv",
	}

	res, err := normalizer.NormalizeCSV(inCSV, schemaFile, outDir, opt)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if res.RowsError != 0 {
		t.Fatalf("expected 0 errors, got %d", res.RowsError)
	}

	assertFileEqual(t, filepath.Join(expDir, "normalized.csv"), filepath.Join(outDir, "normalized.csv"))
	assertFileEqual(t, filepath.Join(expDir, "errors.csv"), filepath.Join(outDir, "errors.csv"))
	assertFileEqual(t, filepath.Join(expDir, "report.json"), filepath.Join(outDir, "report.json"))
}
