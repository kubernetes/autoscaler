package testing

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/document"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
)

// CompareValues compares two values to determine if they are equal,
// specialized for comparison of SDK operation output types.
//
// CompareValues expects the two values to be of the same underlying type.
// Doing otherwise will result in undefined behavior.
//
// The third variadic argument is vestigial from a previous implementation that
// depended on go-cmp. Values passed therein have no effect.
func CompareValues(expect, actual interface{}, _ ...interface{}) error {
	return deepEqual(reflect.ValueOf(expect), reflect.ValueOf(actual), "<root>")
}

func deepEqual(expect, actual reflect.Value, path string) error {
	if et, at := expect.Kind(), actual.Kind(); et != at {
		return fmt.Errorf("%s: kind %s != %s", path, et, at)
	}

	// there are a handful of short-circuit cases here within the context of
	// operation responses:
	//   - ResultMetadata     (we don't care)
	//   - document.Interface (check for marshaled []byte equality)
	//   - io.Reader          (check for Read() []byte equality)
	ei, ai := expect.Interface(), actual.Interface()
	if _, _, ok := asMetadatas(ei, ai); ok {
		return nil
	}
	if e, a, ok := asDocuments(ei, ai); ok {
		if !compareDocumentTypes(e, a) {
			return fmt.Errorf("%s: document values unequal", path)
		}
		return nil
	}
	if e, a, ok := asReaders(ei, ai); ok {
		if err := CompareReaders(e, a); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}

	switch expect.Kind() {
	case reflect.Pointer:
		if expect.Type() != actual.Type() {
			return fmt.Errorf("%s: type mismatch", path)
		}

		expect = deref(expect)
		actual = deref(actual)
		ek, ak := expect.Kind(), actual.Kind()
		if ek == reflect.Invalid || ak == reflect.Invalid {
			// one was a nil pointer, so they both must be nil
			if ek == ak {
				return nil
			}
			return fmt.Errorf("%s: %s != %s", path, fmtNil(ek), fmtNil(ak))
		}
		if err := deepEqual(expect, actual, path); err != nil {
			return err
		}
		return nil
	case reflect.Slice:
		if expect.Len() != actual.Len() {
			return fmt.Errorf("%s: slice length unequal", path)
		}
		for i := 0; i < expect.Len(); i++ {
			ipath := fmt.Sprintf("%s[%d]", path, i)
			if err := deepEqual(expect.Index(i), actual.Index(i), ipath); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		if expect.Len() != actual.Len() {
			return fmt.Errorf("%s: map length unequal", path)
		}
		for _, k := range expect.MapKeys() {
			kpath := fmt.Sprintf("%s[%q]", path, k.String())
			if err := deepEqual(expect.MapIndex(k), actual.MapIndex(k), kpath); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		for i := 0; i < expect.NumField(); i++ {
			if !expect.Field(i).CanInterface() {
				continue // unexported
			}
			fpath := fmt.Sprintf("%s.%s", path, expect.Type().Field(i).Name)
			if err := deepEqual(expect.Field(i), actual.Field(i), fpath); err != nil {
				return err
			}
		}
		return nil
	case reflect.Float32, reflect.Float64:
		ef, af := expect.Float(), actual.Float()
		ebits, abits := math.Float64bits(ef), math.Float64bits(af)
		if enan, anan := math.IsNaN(ef), math.IsNaN(af); enan || anan {
			if enan != anan {
				return fmt.Errorf("%s: NaN: float64(0x%x) != float64(0x%x)", path, ebits, abits)
			}
			return nil
		}
		if ebits != abits {
			return fmt.Errorf("%s: float64(0x%x) != float64(0x%x)", path, ebits, abits)
		}
		return nil
	default:
		// everything else is just scalars and can be delegated
		if !reflect.DeepEqual(ei, ai) {
			return fmt.Errorf("%s: %v != %v", path, ei, ai)
		}
		return nil
	}
}

func asMetadatas(i, j interface{}) (ii, jj middleware.Metadata, ok bool) {
	ii, iok := i.(middleware.Metadata)
	jj, jok := j.(middleware.Metadata)
	return ii, jj, iok || jok
}

func asDocuments(i, j interface{}) (ii, jj documentInterface, ok bool) {
	ii, iok := i.(documentInterface)
	jj, jok := j.(documentInterface)
	return ii, jj, iok || jok
}

func asReaders(i, j interface{}) (ii, jj io.Reader, ok bool) {
	ii, iok := i.(io.Reader)
	jj, jok := j.(io.Reader)
	return ii, jj, iok || jok
}

func deref(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}
	return v
}

type documentInterface interface {
	document.Marshaler
	document.Unmarshaler
}

func compareDocumentTypes(x documentInterface, y documentInterface) bool {
	if x == nil {
		x = nopMarshaler{}
	}
	if y == nil {
		y = nopMarshaler{}
	}

	xBytes, err := x.MarshalSmithyDocument()
	if err != nil {
		panic(fmt.Sprintf("MarshalSmithyDocument error: %v", err))
	}
	yBytes, err := y.MarshalSmithyDocument()
	if err != nil {
		panic(fmt.Sprintf("MarshalSmithyDocument error: %v", err))
	}
	return JSONEqual(xBytes, yBytes) == nil
}

// CompareReaders two io.Reader values together to determine if they are equal.
// Will read the contents of the readers until they are empty.
func CompareReaders(expect, actual io.Reader) error {
	if expect == nil {
		expect = nopReader{}
	}
	if actual == nil {
		actual = nopReader{}
	}

	e, err := io.ReadAll(expect)
	if err != nil {
		return fmt.Errorf("failed to read expect body, %w", err)
	}

	a, err := io.ReadAll(actual)
	if err != nil {
		return fmt.Errorf("failed to read actual body, %w", err)
	}

	if !bytes.Equal(e, a) {
		return fmt.Errorf("bytes do not match\nexpect:\n%s\nactual:\n%s",
			hex.Dump(e), hex.Dump(a))
	}

	return nil
}

func fmtNil(k reflect.Kind) string {
	if k == reflect.Invalid {
		return "nil"
	}
	return "non-nil"
}

type nopReader struct{}

func (nopReader) Read(p []byte) (int, error) { return 0, io.EOF }

type nopMarshaler struct{}

func (nopMarshaler) MarshalSmithyDocument() ([]byte, error)      { return nil, nil }
func (nopMarshaler) UnmarshalSmithyDocument(v interface{}) error { return nil }
