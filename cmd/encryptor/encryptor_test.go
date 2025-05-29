package encryptor

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptFileStream(t *testing.T) {
	// Prepare test data
	originalContent := []byte("This is a test content to be encrypted and decrypted using AES-GCM stream.")

	// Create temp dir
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "input.txt")
	encryptedPath := filepath.Join(tmpDir, "encrypted.bin")
	decryptedPath := filepath.Join(tmpDir, "decrypted.txt")

	// Write original content to file
	if err := os.WriteFile(inputPath, originalContent, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Generate 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Encrypt
	if err := EncryptFileStream(inputPath, encryptedPath, key); err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt
	if err := DecryptFileStream(encryptedPath, decryptedPath, key); err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Read decrypted content
	decryptedContent, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	// Compare
	if !bytes.Equal(originalContent, decryptedContent) {
		t.Errorf("decrypted content does not match original.\nExpected: %q\nGot: %q", originalContent, decryptedContent)
	}
}
