package testing

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/encoding/cbor"
)

// CompareCBOR checks whether two CBOR values are equivalent.
//
// The function signature is tailored for use in smithy protocol tests, where
// the expected encoding is given in base64, and the actual value to check is
// passed from the mock HTTP request body.
func CompareCBOR(actual io.Reader, expect64 string) error {
	ap, err := io.ReadAll(actual)
	if err != nil {
		return fmt.Errorf("read actual: %w", err)
	}

	av, err := cbor.Decode(ap)
	if err != nil {
		return fmt.Errorf("decode actual: %w", err)
	}

	ep, err := base64.StdEncoding.DecodeString(expect64)
	if err != nil {
		return fmt.Errorf("decode expect64: %w", err)
	}

	ev, err := cbor.Decode(ep)
	if err != nil {
		return fmt.Errorf("decode expect: %w", err)
	}

	return cmpCBOR(ev, av, "<root>")
}

func cmpCBOR(e, a cbor.Value, path string) error {
	switch v := e.(type) {
	case cbor.Uint, cbor.NegInt, cbor.Slice, cbor.String, cbor.Bool, *cbor.Nil, *cbor.Undefined:
		if !reflect.DeepEqual(e, a) {
			return fmt.Errorf("%s: %v != %v", path, e, a)
		}
		return nil
	case cbor.List:
		return cmpList(v, a, path)
	case cbor.Map:
		return cmpMap(v, a, path)
	case *cbor.Tag:
		return cmpTag(v, a, path)
	case cbor.Float32:
		return cmpF32(v, a, path)
	case cbor.Float64:
		return cmpF64(v, a, path)
	default:
		return fmt.Errorf("%s: unrecognized variant %T", path, e)
	}
}

func cmpList(e cbor.List, a cbor.Value, path string) error {
	av, ok := a.(cbor.List)
	if !ok {
		return fmt.Errorf("%s: %T != %T", path, e, a)
	}

	if len(e) != len(av) {
		return fmt.Errorf("%s: length %d != %d", path, len(e), len(av))
	}

	for i := 0; i < len(e); i++ {
		ipath := fmt.Sprintf("%s[%d]", path, i)
		if err := cmpCBOR(e[i], av[i], ipath); err != nil {
			return err
		}
	}
	return nil
}

func cmpMap(e cbor.Map, a cbor.Value, path string) error {
	av, ok := a.(cbor.Map)
	if !ok {
		return fmt.Errorf("%s: %T != %T", path, e, a)
	}

	if len(e) != len(av) {
		return fmt.Errorf("%s: length %d != %d", path, len(e), len(av))
	}

	for k, ev := range e {
		avv, ok := av[k]
		if !ok {
			return fmt.Errorf("%s: missing key %s", path, k)
		}

		kpath := fmt.Sprintf("%s[%q]", path, k)
		if err := cmpCBOR(ev, avv, kpath); err != nil {
			return err
		}
	}
	return nil
}

func cmpTag(e *cbor.Tag, a cbor.Value, path string) error {
	av, ok := a.(*cbor.Tag)
	if !ok {
		return fmt.Errorf("%s: %T != %T", path, e, a)
	}

	if e.ID != av.ID {
		return fmt.Errorf("%s: tag ID %d != %d", path, e.ID, av.ID)
	}
	return cmpCBOR(e.Value, av.Value, path)
}

func cmpF32(e cbor.Float32, a cbor.Value, path string) error {
	av, ok := a.(cbor.Float32)
	if !ok {
		return fmt.Errorf("%s: %T != %T", path, e, a)
	}

	ebits, abits := math.Float32bits(float32(e)), math.Float32bits(float32(av))
	if enan, anan := isNaN32(ebits), isNaN32(abits); enan || anan {
		if enan != anan {
			return fmt.Errorf("%s: NaN: float32(%x) != float32(%x)", path, ebits, abits)
		}
		return nil
	}

	if ebits != abits {
		return fmt.Errorf("%s: float32(%x) != float32(%x)", path, ebits, abits)
	}
	return nil
}

func cmpF64(e cbor.Float64, a cbor.Value, path string) error {
	av, ok := a.(cbor.Float64)
	if !ok {
		return fmt.Errorf("%s: %T != %T", path, e, a)
	}

	ebits, abits := math.Float64bits(float64(e)), math.Float64bits(float64(av))
	if enan, anan := isNaN64(ebits), isNaN64(abits); enan || anan {
		if enan != anan {
			return fmt.Errorf("%s: NaN: float64(%x) != float64(%x)", path, ebits, abits)
		}
		return nil
	}

	if math.Float64bits(float64(e)) != math.Float64bits(float64(av)) {
		return fmt.Errorf("%s: float64(%x) != float64(%x)", path, ebits, abits)
	}
	return nil
}

func isNaN32(f uint32) bool {
	const infmask = 0x7f800000

	return f&infmask == infmask && f != infmask && f != (1<<31)|infmask
}

func isNaN64(f uint64) bool {
	const infmask = 0x7ff00000_00000000

	return f&infmask == infmask && f != infmask && f != (1<<63)|infmask
}
