package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/ssongin/tartarus/cmd/archive"
)

type Request struct {
	InputPath     string   `json:"input"`
	OutputPath    string   `json:"output"`
	Passphrase    string   `json:"passphrase,omitempty"`
	CompressLevel int      `json:"compression_level,omitempty"`
	Filters       []string `json:"filters,omitempty"`
}

func GetArchiveRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/pipeline", HandlePipeline)
	mux.HandleFunc("/compress", HandleCompress)
	mux.HandleFunc("/decompress", HandleDecompress)
	mux.HandleFunc("/encrypt", HandleEncrypt)
	mux.HandleFunc("/decrypt", HandleDecrypt)
	mux.HandleFunc("/archive", HandleArchive)
	mux.HandleFunc("/extract", HandleExtract)

	return mux
}

// writeError wraps errors into a JSON response
func writeError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}

func HandlePipeline(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	in := req.InputPath
	out := req.OutputPath
	pass := []byte(req.Passphrase)
	filter := archive.FilterFunc(req.Filters)

	outFile, err := os.Create(out)
	if err != nil {
		writeError(w, "Failed to create output file", 500)
		return
	}
	defer outFile.Close()

	if err := archive.ArchiveAndCompressEncrypt(in, outFile, req.CompressLevel, pass, filter); err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleCompress(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", 400)
		return
	}
	inFile, err := os.Open(req.InputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer inFile.Close()

	outFile, err := os.Create(req.OutputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer outFile.Close()

	writer, err := archive.CompressWriter(outFile, req.CompressLevel)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer writer.Close()

	if _, err := io.Copy(writer, inFile); err != nil {
		writeError(w, err.Error(), 500)
	}
}

func HandleDecompress(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", 400)
		return
	}
	inFile, err := os.Open(req.InputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer inFile.Close()

	reader, err := archive.DecompressReader(inFile)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}

	outFile, err := os.Create(req.OutputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		writeError(w, err.Error(), 500)
	}
}

func HandleEncrypt(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", 400)
		return
	}
	inFile, err := os.Open(req.InputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer inFile.Close()

	outFile, err := os.Create(req.OutputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer outFile.Close()

	writer, err := archive.EncryptWriterCTR_HMAC(outFile, []byte(req.Passphrase))
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer writer.Close()

	if _, err := io.Copy(writer, inFile); err != nil {
		writeError(w, err.Error(), 500)
	}
}

func HandleDecrypt(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", 400)
		return
	}
	inFile, err := os.Open(req.InputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer inFile.Close()

	reader, err := archive.DecryptReaderCTR_HMAC(inFile, []byte(req.Passphrase))
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}

	outFile, err := os.Create(req.OutputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		writeError(w, err.Error(), 500)
	}
}

func HandleArchive(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", 400)
		return
	}
	outFile, err := os.Create(req.OutputPath)
	filter := archive.FilterFunc(req.Filters)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer outFile.Close()

	if err := archive.TarFolderFiltered(req.InputPath, outFile, filter); err != nil {
		writeError(w, err.Error(), 500)
	}
}

func HandleExtract(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", 400)
		return
	}
	inFile, err := os.Open(req.InputPath)
	if err != nil {
		writeError(w, err.Error(), 500)
		return
	}
	defer inFile.Close()

	if err := archive.UntarStream(inFile, req.OutputPath); err != nil {
		writeError(w, err.Error(), 500)
	}
}
