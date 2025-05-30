package archive

import (
	"archive/tar"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func TarFolderFiltered(src string, w io.Writer, filter func(string) bool) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil // skip the root directory itself
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath) // Normalize to forward slashes

		if filter != nil && !filter(relPath) {
			if info.Mode().IsRegular() {
				slog.Info("Skipping file: " + relPath)
				return nil // skip file
			}
			slog.Info("Skipping dir (but still walk): " + relPath)
			return nil // don't skip dirs entirely
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			hdr.Typeflag = tar.TypeDir
			hdr.Name += "/" // ensure directories are recognized
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
}

func UntarStream(input io.Reader, destDir string) error {
	tr := tar.NewReader(input)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			err = os.MkdirAll(filepath.Dir(target), 0755)
			if err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, tr)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
