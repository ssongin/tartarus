package archive

import (
	"io"
)

// PipelineWriter defines a generic step in the pipeline.
type PipelineWriter func(io.Writer) (io.WriteCloser, error)

// PipelineReader defines a generic reader step.
type PipelineReader func(io.Reader) (io.Reader, error)

func ArchiveAndCompressEncrypt(inputDir string, output io.Writer, compressLevel int, passphrase []byte, filterFunc func(string) bool) error {
	encWriter, err := EncryptWriterCTR_HMAC(output, passphrase)
	if err != nil {
		return err
	}
	defer encWriter.Close()

	compWriter, err := CompressWriter(encWriter, compressLevel)
	if err != nil {
		return err
	}
	defer compWriter.Close()

	return TarFolderFiltered(inputDir, compWriter, filterFunc)
}

func DecryptDecompressExtract(input io.Reader, outputDir string, passphrase []byte) error {
	decReader, err := DecryptReaderCTR_HMAC(input, passphrase)
	if err != nil {
		return err
	}

	decompReader, err := DecompressReader(decReader)
	if err != nil {
		return err
	}

	return UntarStream(decompReader, outputDir)
}
