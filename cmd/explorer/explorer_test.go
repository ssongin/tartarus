package explorer

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	touch := func(name string) {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		f.Close()
	}
	touch("a.txt")
	touch("b.txt")
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	return dir
}

func TestListDirectory(t *testing.T) {
	dir := setupTestDir(t)
	entries, err := ListDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestRenameAndDelete(t *testing.T) {
	dir := setupTestDir(t)
	oldPath := filepath.Join(dir, "a.txt")
	newPath := filepath.Join(dir, "renamed.txt")
	if err := Rename(oldPath, newPath); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
	if err := Delete(newPath); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted")
	}
}

func TestMoveAndCopy(t *testing.T) {
	dir := setupTestDir(t)
	src := filepath.Join(dir, "b.txt")
	moveDst := filepath.Join(dir, "moved.txt")
	if err := Move(src, moveDst); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	copyDst := filepath.Join(dir, "copy.txt")
	if err := Copy(moveDst, copyDst); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	if _, err := os.Stat(copyDst); err != nil {
		t.Errorf("expected copy to exist: %v", err)
	}
}
