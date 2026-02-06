package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholaskarlson/proof-first-normalizer/internal/normalizer"
)

func TestGoldenCase11HeaderExtraColumn(t *testing.T) {
	root := projectRoot(t)

	caseName := "case11_header_extra_column"
	inCSV := filepath.Join(root, "fixtures", "input", caseName, "raw.csv")
	schemaFile := filepath.Join(root, "fixtures", "input", caseName, "schema.json")
	expDir := filepath.Join(root, "fixtures", "expected", caseName)

	outDir := t.TempDir()

	expB, err := os.ReadFile(filepath.Join(expDir, "error.txt"))
	if err != nil {
		t.Fatalf("read expected error: %v", err)
	}
	exp := strings.TrimSpace(string(expB))

	opt := normalizer.Options{
		Tool:    "proof-first-normalizer",
		Version: "dev",
		Label:   caseName,
		Schema:  "fixtures/input/" + caseName + "/schema.json",
		Input:   "fixtures/input/" + caseName + "/raw.csv",
	}

	_, gotErr := normalizer.NormalizeCSV(inCSV, schemaFile, outDir, opt)
	if gotErr == nil {
		t.Fatalf("expected error, got success")
	}
	got := strings.TrimSpace(gotErr.Error())
	if got != exp {
		t.Fatalf("error mismatch\n got: %s\n exp: %s", got, exp)
	}
}
