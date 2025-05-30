package archiver

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTarAndUntar(t *testing.T) {
	// Setup temp dirs
	srcDir := t.TempDir()
	destDir := t.TempDir()
	tarFile := filepath.Join(t.TempDir(), "test.tar")

	// Create nested file structure
	files := map[string]string{
		"file1.txt":                         "hello",
		"subdir/file2.txt":                  "world",
		"subdir/inner/file3.txt":            "nested",
		"subdir/inner/deep/file4.txt":       "deeper",
		"subdir/inner/deep/file5-large.txt": string(bytes.Repeat([]byte("x"), 1024*1024)), // 1MB
	}
	for name, content := range files {
		fullPath := filepath.Join(srcDir, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("mkdir error: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	// Archive
	if err := TarFolder(srcDir, tarFile); err != nil {
		t.Fatalf("TarFolder error: %v", err)
	}

	// Extract
	if err := UntarFile(tarFile, destDir); err != nil {
		t.Fatalf("UntarFile error: %v", err)
	}

	// Validate extracted files
	for name, expected := range files {
		extracted := filepath.Join(destDir, name)
		data, err := os.ReadFile(extracted)
		if err != nil {
			t.Errorf("failed to read extracted file %s: %v", name, err)
			continue
		}
		if string(data) != expected {
			t.Errorf("file content mismatch for %s: got %d bytes, expected %d", name, len(data), len(expected))
		}
	}
}
