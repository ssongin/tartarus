package api

import (
	"encoding/json"
	"net/http"

	"github.com/ssongin/tartarus/cmd/compressor"
)

type CompressRequest struct {
	Src      string `json:"src"`
	Dst      string `json:"dst"`
	Separate bool   `json:"separate"`
	Level    int    `json:"level"`
}

type DecompressRequest struct {
	Src      string `json:"src"`
	Dst      string `json:"dst"`
	Separate bool   `json:"separate"`
}

func GetCompressorRoutes() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("POST /", HandleCompress)
	return router
}

func GetDecompressorRoutes() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("POST /", HandleDecompress)
	return router
}

func HandleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := compressor.Compress(req.Src, req.Dst, req.Separate, req.Level); err != nil {
		http.Error(w, "Compression failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Compression completed successfully."))
}

func HandleDecompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DecompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := compressor.Decompress(req.Src, req.Dst, req.Separate); err != nil {
		http.Error(w, "Decompression failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Decompression completed successfully."))
}
