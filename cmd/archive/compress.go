package archive

import (
	"compress/flate"
	"io"
)

func CompressWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return flate.NewWriter(w, level)
}

func DecompressReader(r io.Reader) (io.Reader, error) {
	return flate.NewReader(r), nil
}
