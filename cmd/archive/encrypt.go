package archive

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"hash"
	"io"
)

const (
	aesKeySize = 32 // AES-256
	nonceSize  = 16
	hmacSize   = 32
	bufferSize = 32 * 1024
)

func deriveKeys(passphrase []byte) (encKey, hmacKey []byte) {
	hash := sha256.Sum256(passphrase)
	return hash[:16], hash[16:]
}

// EncryptWriterCTR_HMAC returns a WriteCloser that encrypts and HMACs the data
func EncryptWriterCTR_HMAC(w io.Writer, passphrase []byte) (io.WriteCloser, error) {
	encKey, hmacKey := deriveKeys(passphrase)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Write nonce first
	if _, err := w.Write(nonce); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, nonce)
	h := hmac.New(sha256.New, hmacKey)

	writer := &ctrHMACWriter{
		dst:    w,
		stream: stream,
		hmac:   h,
		buf:    make([]byte, bufferSize),
	}
	return writer, nil
}

type ctrHMACWriter struct {
	dst    io.Writer
	stream cipher.Stream
	hmac   hash.Hash
	buf    []byte
}

func (w *ctrHMACWriter) Write(p []byte) (int, error) {
	n := len(p)
	out := make([]byte, n)
	w.stream.XORKeyStream(out, p)

	// Write to HMAC and dst
	if _, err := w.hmac.Write(out); err != nil {
		return 0, err
	}
	_, err := w.dst.Write(out)
	return n, err
}

func (w *ctrHMACWriter) Close() error {
	_, err := w.dst.Write(w.hmac.Sum(nil))
	return err
}

func DecryptReaderCTR_HMAC(r io.Reader, passphrase []byte) (io.Reader, error) {
	encKey, hmacKey := deriveKeys(passphrase)

	// Read nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, err
	}

	// Buffer the rest to verify HMAC at the end
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	encryptedData, err := io.ReadAll(tee)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < hmacSize {
		return nil, errors.New("data too short for HMAC")
	}

	data := encryptedData[:len(encryptedData)-hmacSize]
	expectedMAC := encryptedData[len(encryptedData)-hmacSize:]

	// Verify HMAC
	h := hmac.New(sha256.New, hmacKey)
	h.Write(data)
	if !hmac.Equal(h.Sum(nil), expectedMAC) {
		return nil, errors.New("HMAC verification failed")
	}

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, nonce)

	streamReader := &cipher.StreamReader{
		S: stream,
		R: bytes.NewReader(data),
	}
	return streamReader, nil
}
