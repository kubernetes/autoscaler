package testing

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

func GzipCompareCompressBytes(expect []byte, actual io.Reader) error {
	content, err := gzip.NewReader(actual)
	if err != nil {
		return fmt.Errorf("error while reading request")
	}

	var actualBytes bytes.Buffer
	_, err = actualBytes.ReadFrom(content)
	if err != nil {
		return fmt.Errorf("error while unzipping request payload")
	}

	if e, a := expect, actualBytes.Bytes(); !bytes.Equal(e, a) {
		return fmt.Errorf("expect unzipped content to be %s, got %s", e, a)
	}

	return nil
}
