package testing

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
)

// Enumeration values for supported compress Algorithms.
const (
	GZIP = "gzip"
)

type compareCompressFunc func([]byte, io.Reader) error

var allowedAlgorithms = map[string]compareCompressFunc{
	GZIP: GzipCompareCompressBytes,
}

// CompareReaderEmpty checks if the reader is nil, or contains no bytes.
// Returns an error if not empty.
func CompareReaderEmpty(r io.Reader) error {
	if r == nil {
		return nil
	}
	b, err := io.ReadAll(r)
	if err != nil && err != io.EOF {
		return fmt.Errorf("unable to read from reader, %v", err)
	}
	if len(b) != 0 {
		return fmt.Errorf("reader not empty, got\n%v", hex.Dump(b))
	}
	return nil
}

// CompareReaderBytes compares the reader with the expected bytes. Returns an
// error if the two bytes are not equal.
func CompareReaderBytes(r io.Reader, expect []byte) error {
	if r == nil {
		return fmt.Errorf("missing body")
	}
	actual, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read, %v", err)
	}

	if !bytes.Equal(expect, actual) {
		return fmt.Errorf("bytes not equal\nexpect:\n%v\nactual:\n%v",
			hex.Dump(expect), hex.Dump(actual))
	}
	return nil
}

// CompareJSONReaderBytes compares the reader containing serialized JSON
// document. Deserializes the JSON documents to determine if they are equal.
// Return an error if the two JSON documents are not equal.
func CompareJSONReaderBytes(r io.Reader, expect []byte) error {
	if r == nil {
		return fmt.Errorf("missing body")
	}
	actual, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read, %v", err)
	}

	if err := JSONEqual(expect, actual); err != nil {
		return fmt.Errorf("JSON documents not equal, %v", err)
	}
	return nil
}

// CompareXMLReaderBytes compares the reader with expected xml byte
func CompareXMLReaderBytes(r io.Reader, expect []byte) error {
	if r == nil {
		return fmt.Errorf("missing body")
	}

	actual, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err := XMLEqual(expect, actual); err != nil {
		return fmt.Errorf("XML documents not equal, %w", err)
	}
	return nil
}

// CompareURLFormReaderBytes compares the reader containing serialized URLForm
// document. Deserializes the URLForm documents to determine if they are equal.
// Return an error if the two URLForm documents are not equal.
func CompareURLFormReaderBytes(r io.Reader, expect []byte) error {
	if r == nil {
		return fmt.Errorf("missing body")
	}
	actual, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read, %v", err)
	}

	if err := URLFormEqual(expect, actual); err != nil {
		return fmt.Errorf("URL query forms not equal, %v", err)
	}
	return nil
}

// CompareCompressedBytes compares the request stream before and after possible request compression
func CompareCompressedBytes(expect *bytes.Buffer, actual io.Reader, disable bool, min int64, algorithm string) error {
	expectBytes := expect.Bytes()
	if disable || int64(len(expectBytes)) < min {
		actualBytes, err := io.ReadAll(actual)
		if err != nil {
			return fmt.Errorf("error while reading request: %q", err)
		}
		if e, a := expectBytes, actualBytes; !bytes.Equal(e, a) {
			return fmt.Errorf("expect content to be %s, got %s", e, a)
		}
	} else {
		compareFn := allowedAlgorithms[algorithm]
		if compareFn == nil {
			return fmt.Errorf("compress algorithm %s is not allowed", algorithm)
		}
		if err := compareFn(expectBytes, actual); err != nil {
			return fmt.Errorf("error while comparing unzipped content: %q", err)
		}
	}

	return nil
}
