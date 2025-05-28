package compressor

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFiles(base string) error {
	files := map[string]string{
		"file1.txt":              "Hello, world!",
		"subdir/file2.txt":       "Another file inside subdir.",
		"subdir/nested/file3.md": "This is a deeply nested file.",
	}
	for name, content := range files {
		fullPath := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func TestCompressDecompress_Separate(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	comp := filepath.Join(base, "compressed")
	decomp := filepath.Join(base, "decompressed")

	if err := writeTestFiles(src); err != nil {
		t.Fatalf("Failed to write test files: %v", err)
	}

	if err := Compress(src, comp, true, -2); err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	if err := Decompress(comp, decomp, true); err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	checkRestoredFiles(t, src, decomp)
}

func TestCompressDecompress_Combined(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	archive := filepath.Join(base, "archive.deflate")
	decomp := filepath.Join(base, "decompressed")

	if err := writeTestFiles(src); err != nil {
		t.Fatalf("Failed to write test files: %v", err)
	}

	if err := Compress(src, archive, false, -2); err != nil {
		t.Fatalf("Combined compression failed: %v", err)
	}

	if err := Decompress(archive, decomp, false); err != nil {
		t.Fatalf("Combined decompression failed: %v", err)
	}

	checkRestoredFiles(t, src, decomp)
}

func checkRestoredFiles(t *testing.T, original, restored string) {
	err := filepath.Walk(original, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(original, path)
		if err != nil {
			t.Fatalf("failed to get relative path: %v", err)
		}
		restoredPath := filepath.Join(restored, rel)

		origContent, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read original file %s: %v", path, err)
		}
		newContent, err := os.ReadFile(restoredPath)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", restoredPath, err)
		}

		if string(origContent) != string(newContent) {
			t.Errorf("Content mismatch: %s != %s", path, restoredPath)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Walk error: %v", err)
	}
}
