package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
)

func EncryptFileStream(inputPath, outputPath string, key []byte) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	// Write nonce first
	if _, err := outFile.Write(nonce); err != nil {
		return err
	}

	buf := make([]byte, 32*1024) // 32 KB buffer
	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			encrypted := gcm.Seal(nil, nonce, chunk, nil)
			if _, err := outFile.Write(encrypted); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func DecryptFileStream(inputPath, outputPath string, key []byte) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)

	if _, err := io.ReadFull(inFile, nonce); err != nil {
		return err
	}

	chunkSize := 32*1024 + gcm.Overhead() // account for GCM overhead
	buf := make([]byte, chunkSize)

	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			decrypted, err := gcm.Open(nil, nonce, chunk, nil)
			if err != nil {
				return err
			}
			if _, err := outFile.Write(decrypted); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}
