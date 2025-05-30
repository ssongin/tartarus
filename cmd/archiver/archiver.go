package archiver

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

// TarFolder creates a .tar archive from the srcDir and writes it to destTarFile.
func TarFolder(srcDir, destTarFile string) error {
	tarFile, err := os.Create(destTarFile)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Generate relative path
		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		// Get header
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If not a regular file, skip writing content
		if !fi.Mode().IsRegular() {
			return nil
		}

		// Open file and copy contents
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

// UntarFile extracts a .tar file into the target directory.
func UntarFile(tarFile, targetDir string) error {
	f, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(f)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// Construct full path
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(targetPath, os.FileMode(header.Mode))
		case tar.TypeReg:
			err = os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				return err
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			_, err = io.Copy(outFile, tr)
			outFile.Close()
		default:
			// Skip unsupported types (symlinks, etc.)
			continue
		}

		if err != nil {
			return err
		}
	}
	return nil
}
