package normalizer

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeFileAtomic(dir, name string, data []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := filepath.Join(dir, fmt.Sprintf(".tmp.%s.%d", name, os.Getpid()))
	final := filepath.Join(dir, name)

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		// On Windows, Rename fails if the destination exists. Best-effort remove+retry.
		_ = os.Remove(final)
		if err2 := os.Rename(tmp, final); err2 != nil {
			_ = os.Remove(tmp)
			return err2
		}
	}
	return nil
}
