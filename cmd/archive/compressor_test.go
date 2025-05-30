package archive

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func createTestFiles(t *testing.T, base string, files map[string]string) {
	for path, content := range files {
		fullPath := filepath.Join(base, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
}

func readFile(t *testing.T, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(data)
}

func TestPipelineFull(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	files := map[string]string{
		"a.txt":     "hello",
		"b/b.txt":   "world",
		"c/d/e.txt": "nested",
		"skip.tmp":  "should skip",
	}
	createTestFiles(t, inputDir, files)

	var buf bytes.Buffer
	pass := []byte("test-pass")
	filters := []string{"*.txt", "b/*", "c/d/*"}
	filterFunc := FilterFunc(filters)

	if err := ArchiveAndCompressEncrypt(inputDir, &buf, 5, pass, filterFunc); err != nil {
		t.Fatalf("ArchiveAndCompressEncrypt: %v", err)
	}

	if err := DecryptDecompressExtract(&buf, outputDir, pass); err != nil {
		t.Fatalf("DecryptDecompressExtract: %v", err)
	}

	expect := map[string]string{
		"a.txt":     "hello",
		"b/b.txt":   "world",
		"c/d/e.txt": "nested",
	}
	for path, content := range expect {
		out := filepath.Join(outputDir, path)
		got := readFile(t, out)
		if got != content {
			t.Errorf("file %s mismatch: got %q want %q", path, got, content)
		}
	}

	if _, err := os.Stat(filepath.Join(outputDir, "skip.tmp")); err == nil {
		t.Errorf("skip.tmp should not be present")
	}
}

func TestTarUntar(t *testing.T) {
	dir := t.TempDir()
	outDir := t.TempDir()
	files := map[string]string{
		"x.txt":      "test",
		"deep/y.txt": "value",
	}
	createTestFiles(t, dir, files)

	var buf bytes.Buffer
	if err := TarFolderFiltered(dir, &buf, func(string) bool { return true }); err != nil {
		t.Fatal(err)
	}
	if err := UntarStream(&buf, outDir); err != nil {
		t.Fatal(err)
	}

	for path, content := range files {
		got := readFile(t, filepath.Join(outDir, path))
		if got != content {
			t.Errorf("untar %s mismatch: got %q want %q", path, got, content)
		}
	}
}

func TestDeflateRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	original := "deflate compression test string with symbols !@#$%^&*()_+"

	w, err := CompressWriter(&buf, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(original)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := DecompressReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if got != original {
		t.Errorf("deflate roundtrip mismatch: got %q want %q", got, original)
	}
}

func TestAESEncryptionRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	plaintext := []byte("sensitive data string")
	pass := []byte("testpass")

	enc, err := EncryptWriterCTR_HMAC(&buf, pass)
	if err != nil {
		t.Fatal(err)
	}
	enc.Write(plaintext)
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}

	dec, err := DecryptReaderCTR_HMAC(&buf, pass)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := io.ReadAll(dec)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestLargeFilePipeline(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create 10MB content safely
	oneMB := bytes.Repeat([]byte("0123456789ABCDEF"), 65536) // ~1MB
	tenMB := bytes.Repeat(oneMB, 10)

	err := os.WriteFile(filepath.Join(inputDir, "large.txt"), tenMB, 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	pass := []byte("securepass")
	if err := ArchiveAndCompressEncrypt(inputDir, &buf, 6, pass, FilterFunc([]string{"*.txt"})); err != nil {
		t.Fatal(err)
	}

	if err := DecryptDecompressExtract(&buf, outputDir, pass); err != nil {
		t.Fatal(err)
	}

	got := readFile(t, filepath.Join(outputDir, "large.txt"))
	if len(got) != 10*1024*1024 {
		t.Fatalf("expected 10MB file, got %d bytes", len(got))
	}
}

func TestFilterMatching(t *testing.T) {
	tests := []struct {
		path    string
		filters []string
		match   bool
	}{
		{"a.txt", []string{"*.txt"}, true},
		{"logs/app.log", []string{"logs/*"}, true},
		{"src/main.go", []string{"*.txt"}, false},
	}

	for _, tt := range tests {
		got := FilterFunc(tt.filters)(tt.path)
		if got != tt.match {
			t.Errorf("ShouldIncludeFile(%q, %v) = %v; want %v", tt.path, tt.filters, got, tt.match)
		}
	}
}

func BenchmarkCompressDecompress(b *testing.B) {
	input := bytes.Repeat([]byte("data "), 1000)
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		buf.Reset()
		w, _ := CompressWriter(&buf, 6)
		_, _ = w.Write(input)
		_ = w.Close()

		r, _ := DecompressReader(&buf)
		_, _ = io.Copy(io.Discard, r)
	}
}
