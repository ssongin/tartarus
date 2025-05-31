package api

import (
	"encoding/json"
	"net/http"

	"github.com/ssongin/tartarus/cmd/explorer"
)

type ExplorerRestHandler struct{}

func (h *ExplorerRestHandler) GetExplorerRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/list", h.List)
	mux.HandleFunc("/rename", h.Rename)
	mux.HandleFunc("/delete", h.Delete)
	mux.HandleFunc("/move", h.Move)
	mux.HandleFunc("/copy", h.Copy)

	return mux
}

func (h *ExplorerRestHandler) List(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}
	entries, err := explorer.ListDirectory(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entries)
}

func (h *ExplorerRestHandler) Rename(w http.ResponseWriter, r *http.Request) {
	old := r.URL.Query().Get("old")
	new := r.URL.Query().Get("new")
	if old == "" || new == "" {
		http.Error(w, "old and new paths required", http.StatusBadRequest)
		return
	}
	if err := explorer.Rename(old, new); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ExplorerRestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path required", http.StatusBadRequest)
		return
	}
	if err := explorer.Delete(path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ExplorerRestHandler) Move(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("src")
	dst := r.URL.Query().Get("dst")
	if src == "" || dst == "" {
		http.Error(w, "src and dst paths required", http.StatusBadRequest)
		return
	}
	if err := explorer.Move(src, dst); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ExplorerRestHandler) Copy(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("src")
	dst := r.URL.Query().Get("dst")
	if src == "" || dst == "" {
		http.Error(w, "src and dst paths required", http.StatusBadRequest)
		return
	}
	if err := explorer.Copy(src, dst); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
