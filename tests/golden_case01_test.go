package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholaskarlson/proof-first-normalizer/internal/normalizer"
)

func TestGoldenCase01(t *testing.T) {
	root := projectRoot(t)

	inCSV := filepath.Join(root, "fixtures", "input", "case01", "raw.csv")
	schemaFile := filepath.Join(root, "fixtures", "input", "case01", "schema.json")
	expDir := filepath.Join(root, "fixtures", "expected", "case01")

	outDir := t.TempDir()

	opt := normalizer.Options{
		Tool:    "proof-first-normalizer",
		Version: "dev",
		Label:   "case01",
		// record stable, repo-relative strings in report.json
		Schema: "fixtures/input/case01/schema.json",
		Input:  "fixtures/input/case01/raw.csv",
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

func assertFileEqual(t *testing.T, a, b string) {
	t.Helper()
	ab, err := os.ReadFile(a)
	if err != nil {
		t.Fatalf("read %s: %v", a, err)
	}
	bb, err := os.ReadFile(b)
	if err != nil {
		t.Fatalf("read %s: %v", b, err)
	}
	if string(ab) != string(bb) {
		t.Fatalf("files differ:\nA=%s\nB=%s", a, b)
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate repo root from %s", wd)
	return ""
}
