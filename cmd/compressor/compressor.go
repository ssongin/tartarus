package compressor

import (
	"bufio"
	"compress/flate"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(logWriter{})
}

type logWriter struct{}

func (logWriter) Write(p []byte) (int, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Fprintf(os.Stdout, "%s [INFO] %s", now, p)
}

func Compress(srcPath, dstPath string, separate bool, level int) error {
	if separate {
		return compressSeparate(srcPath, dstPath, level)
	}
	return compressCombined(srcPath, dstPath, level)
}

func compressSeparate(srcPath, dstPath string, level int) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		dstFile := filepath.Join(dstPath, relPath+".deflate")

		if err := os.MkdirAll(filepath.Dir(dstFile), 0755); err != nil {
			return err
		}

		origSize, err := fileSize(path)
		if err != nil {
			return err
		}

		if err := compressFile(path, dstFile, level); err != nil {
			return err
		}

		compSize, err := fileSize(dstFile)
		if err != nil {
			return err
		}

		log.Printf("%s -> %s [%dB/%dB]\n", path, dstFile, origSize, compSize)
		return nil
	})
}

func compressCombined(srcPath, dstFile string, level int) error {
	var paths []string
	var totalSize int64

	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		size, err := fileSize(path)
		if err != nil {
			return err
		}
		totalSize += size
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return err
	}

	tmpFile := dstFile + ".tmp"
	fOut, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer fOut.Close()

	w, err := flate.NewWriter(fOut, level)
	if err != nil {
		return err
	}
	defer w.Close()

	for _, path := range paths {
		relPath, _ := filepath.Rel(srcPath, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// Write header: relative path and size
		fmt.Fprintf(w, "%s\n%d\n", relPath, len(content))
		if _, err := w.Write(content); err != nil {
			return err
		}
	}

	w.Close()
	fOut.Close()

	if err := os.Rename(tmpFile, dstFile); err != nil {
		return err
	}

	compSize, err := fileSize(dstFile)
	if err != nil {
		return err
	}

	log.Printf("%s -> %s [%dB/%dB]\n", srcPath, dstFile, totalSize, compSize)
	return nil
}

func compressFile(src, dst string, level int) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	w, err := flate.NewWriter(out, level)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, in)
	return err
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func Decompress(srcPath, dstPath string, separate bool) error {
	if separate {
		return decompressSeparate(srcPath, dstPath)
	}
	return decompressCombined(srcPath, dstPath)
}

func decompressSeparate(srcPath, dstPath string) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".deflate") {
			return err
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(dstPath, strings.TrimSuffix(relPath, ".deflate"))

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		r := flate.NewReader(in)
		defer r.Close()

		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()

		n, err := io.Copy(out, r)
		if err != nil {
			return err
		}

		log.Printf("%s -> %s [%dB]\n", path, outPath, n)
		return nil
	})
}

func decompressCombined(srcFile, dstPath string) error {
	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	r := flate.NewReader(in)
	defer r.Close()

	bufr := bufio.NewReader(r)

	for {
		relPath, err := bufr.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		relPath = strings.TrimSpace(relPath)

		sizeStr, err := bufr.ReadString('\n')
		if err != nil {
			return fmt.Errorf("expected size after file name %s: %v", relPath, err)
		}
		sizeStr = strings.TrimSpace(sizeStr)

		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return fmt.Errorf("invalid size %q for file %s: %v", sizeStr, relPath, err)
		}

		fullPath := filepath.Join(dstPath, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}

		out, err := os.Create(fullPath)
		if err != nil {
			return err
		}

		n, err := io.CopyN(out, bufr, int64(size))
		out.Close()
		if err != nil {
			return err
		}

		log.Printf("%s -> %s [%dB]\n", srcFile, fullPath, n)
	}
	return nil
}
