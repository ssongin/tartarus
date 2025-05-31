package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExplorerAPI(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "test.txt"), []byte("hello"), 0644)

	handler := &ExplorerRestHandler{}
	server := httptest.NewServer(handler.GetExplorerRouter())
	defer server.Close()

	t.Run("List", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/list?path=" + tmp)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
		var items []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 || items[0]["name"] != "test.txt" {
			t.Fatalf("unexpected list result: %v", items)
		}
	})

	t.Run("Rename", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/rename?old=" + filepath.Join(tmp, "test.txt") + "&new=" + filepath.Join(tmp, "renamed.txt"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
	})

	t.Run("Copy", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/copy?src=" + filepath.Join(tmp, "renamed.txt") + "&dst=" + filepath.Join(tmp, "copied.txt"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
	})

	t.Run("Move", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/move?src=" + filepath.Join(tmp, "copied.txt") + "&dst=" + filepath.Join(tmp, "moved.txt"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/delete?path=" + filepath.Join(tmp, "moved.txt"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
	})
}
