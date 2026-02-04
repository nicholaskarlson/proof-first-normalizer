package normalizer

import (
	"bytes"
)

func canonicalizeBytes(b []byte) []byte {
	// Strip UTF-8 BOM if present.
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	// Normalize CRLF/CR -> LF.
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))
	return b
}
