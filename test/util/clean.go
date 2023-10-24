package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustCleanFiles(t *testing.T, baseDir string, files ...string) {
	for _, file := range files {
		path := filepath.Join(baseDir, file)
		require.NoErrorf(
			t,
			os.Remove(path),
			"It should remove file %s",
			path,
		)
	}
}
