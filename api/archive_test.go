package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) (string, string) {
	tmpDir := t.TempDir()

	// Input structure
	inputDir := filepath.Join(tmpDir, "input")
	err := os.MkdirAll(filepath.Join(inputDir, "nested"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(inputDir, "root.txt"), []byte("root content"), 0644)
	os.WriteFile(filepath.Join(inputDir, "nested", "nested.txt"), []byte("nested content"), 0644)

	// Output base
	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	return inputDir, outputDir
}

func postJSON(t *testing.T, handler http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestArchiveAndExtract(t *testing.T) {
	inputDir, outputDir := setupTestDir(t)
	archivePath := filepath.Join(outputDir, "archive.tar")
	extractDir := filepath.Join(outputDir, "extracted")

	// Archive
	rr := postJSON(t, HandleArchive, "/archive", Request{
		InputPath:  inputDir,
		OutputPath: archivePath,
	})
	if rr.Code != 200 {
		t.Fatalf("Archive failed: %s", rr.Body.String())
	}

	// Extract
	rr = postJSON(t, HandleExtract, "/extract", Request{
		InputPath:  archivePath,
		OutputPath: extractDir,
	})
	if rr.Code != 200 {
		t.Fatalf("Extract failed: %s", rr.Body.String())
	}

	data, _ := os.ReadFile(filepath.Join(extractDir, "root.txt"))
	if string(data) != "root content" {
		t.Fatalf("Expected root content, got %s", string(data))
	}
}

func TestCompressAndDecompress(t *testing.T) {
	inputDir, outputDir := setupTestDir(t)
	in := filepath.Join(inputDir, "root.txt")
	compressed := filepath.Join(outputDir, "root.deflate")
	decompressed := filepath.Join(outputDir, "root.txt")

	// Compress
	rr := postJSON(t, HandleCompress, "/compress", Request{
		InputPath:     in,
		OutputPath:    compressed,
		CompressLevel: 5,
	})
	if rr.Code != 200 {
		t.Fatalf("Compress failed: %s", rr.Body.String())
	}

	// Decompress
	rr = postJSON(t, HandleDecompress, "/decompress", Request{
		InputPath:  compressed,
		OutputPath: decompressed,
	})
	if rr.Code != 200 {
		t.Fatalf("Decompress failed: %s", rr.Body.String())
	}

	data, _ := os.ReadFile(decompressed)
	if string(data) != "root content" {
		t.Fatalf("Expected root content, got %s", string(data))
	}
}

func TestEncryptAndDecrypt(t *testing.T) {
	inputDir, outputDir := setupTestDir(t)
	in := filepath.Join(inputDir, "root.txt")
	encrypted := filepath.Join(outputDir, "enc.aes")
	decrypted := filepath.Join(outputDir, "dec.txt")
	pass := "supersecret"

	rr := postJSON(t, HandleEncrypt, "/encrypt", Request{
		InputPath:  in,
		OutputPath: encrypted,
		Passphrase: pass,
	})
	if rr.Code != 200 {
		t.Fatalf("Encrypt failed: %s", rr.Body.String())
	}

	rr = postJSON(t, HandleDecrypt, "/decrypt", Request{
		InputPath:  encrypted,
		OutputPath: decrypted,
		Passphrase: pass,
	})
	if rr.Code != 200 {
		t.Fatalf("Decrypt failed: %s", rr.Body.String())
	}

	data, _ := os.ReadFile(decrypted)
	if string(data) != "root content" {
		t.Fatalf("Expected root content, got %s", string(data))
	}
}

func TestPipeline(t *testing.T) {
	inputDir, outputDir := setupTestDir(t)
	outputPath := filepath.Join(outputDir, "full.pipeline")
	pass := "p@ss"

	rr := postJSON(t, HandlePipeline, "/pipeline", Request{
		InputPath:     inputDir,
		OutputPath:    outputPath,
		Passphrase:    pass,
		CompressLevel: 6,
	})
	if rr.Code != 200 {
		t.Fatalf("Pipeline failed: %s", rr.Body.String())
	}

	fi, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output not created: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatal("Output file is empty")
	}
}

func TestDecryptWithWrongPassphrase(t *testing.T) {
	inputDir, outputDir := setupTestDir(t)
	in := filepath.Join(inputDir, "root.txt")
	encrypted := filepath.Join(outputDir, "enc.aes")
	decrypted := filepath.Join(outputDir, "dec.txt")

	// Encrypt with one passphrase
	_ = postJSON(t, HandleEncrypt, "/encrypt", Request{
		InputPath:  in,
		OutputPath: encrypted,
		Passphrase: "correct",
	})

	// Try to decrypt with the wrong one
	rr := postJSON(t, HandleDecrypt, "/decrypt", Request{
		InputPath:  encrypted,
		OutputPath: decrypted,
		Passphrase: "wrong",
	})

	if rr.Code == http.StatusOK {
		t.Fatal("Expected failure with wrong passphrase")
	}
}
