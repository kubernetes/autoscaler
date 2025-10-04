package cbor

import (
	"fmt"
	"math/big"
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/document"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/document/internal/serde"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/encoding/cbor"
)

// encoderOptions is the set of options that can be configured for an Encoder.
//
// FUTURE(rpc2cbor): document support is currently disabled. This API is
// unexported until that changes.
type encoderOptions struct{}

// encoder is a Smithy document encoder for CBOR-based protocols.
//
// FUTURE(rpc2cbor): document support is currently disabled. This API is
// unexported until that changes.
type encoder struct {
	options encoderOptions
}

// newEncoder returns an Encoder for serializing Smithy documents.
//
// FUTURE(rpc2cbor): document support is currently disabled. This API is
// unexported until that changes.
func newEncoder(optFns ...func(options *encoderOptions)) *encoder {
	o := encoderOptions{}

	for _, fn := range optFns {
		fn(&o)
	}

	return &encoder{
		options: o,
	}
}

// Encode returns the CBOR encoding of v.
func (e *encoder) Encode(v interface{}) ([]byte, error) {
	cv, err := e.encode(reflect.ValueOf(v), serde.Tag{})
	if err != nil {
		return nil, err
	}

	return cbor.Encode(cv), nil
}

func (e *encoder) encode(rv reflect.Value, tag serde.Tag) (cbor.Value, error) {
	if serde.IsZeroValue(rv) {
		if tag.OmitEmpty {
			return nil, nil
		}
		return e.encodeZeroValue(rv)
	}

	rv = serde.ValueElem(rv)
	switch rv.Kind() {
	case reflect.Struct:
		return e.encodeStruct(rv)
	case reflect.Map:
		return e.encodeMap(rv)
	case reflect.Slice, reflect.Array:
		return e.encodeSlice(rv)
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil, nil
	default:
		return e.encodeScalar(rv)
	}
}

func (e *encoder) encodeZeroValue(rv reflect.Value) (cbor.Value, error) {
	switch rv.Kind() {
	case reflect.Array:
		return cbor.List{}, nil
	case reflect.String:
		return cbor.String(""), nil
	case reflect.Bool:
		return cbor.Bool(false), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cbor.EncodeFixedUint(0), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return cbor.EncodeFixedUint(0), nil
	case reflect.Float32, reflect.Float64:
		return cbor.Float64(0), nil
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
		return &cbor.Nil{}, nil
	default:
		return nil, &document.InvalidMarshalError{Message: fmt.Sprintf("unknown value type: %s", rv.String())}
	}
}

func (e *encoder) encodeStruct(rv reflect.Value) (cbor.Value, error) {
	if rv.CanInterface() && document.IsNoSerde(rv.Interface()) {
		return nil, &document.UnmarshalTypeError{
			Value: fmt.Sprintf("unsupported type"),
			Type:  rv.Type(),
		}
	}

	switch {
	case rv.Type().ConvertibleTo(serde.ReflectTypeOf.Time):
		return nil, &document.InvalidMarshalError{
			Message: fmt.Sprintf("unsupported type %s", rv.Type().String()),
		}
	case rv.Type().ConvertibleTo(serde.ReflectTypeOf.BigFloat):
		fallthrough
	case rv.Type().ConvertibleTo(serde.ReflectTypeOf.BigInt):
		return e.encodeNumber(rv)
	}

	fields := serde.GetStructFields(rv.Type())

	mv := cbor.Map{}
	for _, f := range fields.All() {
		if f.Name == "" {
			return nil, &document.InvalidMarshalError{Message: "map key cannot be empty"}
		}

		fv, found := serde.EncoderFieldByIndex(rv, f.Index)
		if !found {
			continue
		}

		cv, err := e.encode(fv, f.Tag)
		if err != nil {
			return nil, err
		}
		if cv == nil { // from omitEmpty
			continue
		}

		mv[f.Name] = cv
	}

	return mv, nil
}

func (e *encoder) encodeMap(rv reflect.Value) (cbor.Map, error) {
	mv := cbor.Map{}
	for _, key := range rv.MapKeys() {
		keyName := fmt.Sprint(key.Interface())
		if keyName == "" {
			return nil, &document.InvalidMarshalError{Message: "map key cannot be empty"}
		}

		cv, err := e.encode(rv.MapIndex(key), serde.Tag{})
		if err != nil {
			return nil, err
		}

		mv[keyName] = cv
	}
	return mv, nil
}

func (e *encoder) encodeSlice(rv reflect.Value) (cbor.List, error) {
	lv := cbor.List{}
	for i := 0; i < rv.Len(); i++ {
		cv, err := e.encode(rv.Index(i), serde.Tag{})
		if err != nil {
			return nil, err
		}

		lv = append(lv, cv)
	}

	return lv, nil
}

func (e *encoder) encodeScalar(rv reflect.Value) (cbor.Value, error) {
	switch rv.Kind() {
	case reflect.Bool:
		return cbor.Bool(rv.Bool()), nil
	case reflect.String:
		return cbor.String(rv.String()), nil
	default:
		return e.encodeNumber(rv)
	}
}

func (e *encoder) encodeNumber(rv reflect.Value) (cbor.Value, error) {
	const tagbigpos = 2
	const tagbigneg = 3
	const tagbigfloat = 4

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv := rv.Int()
		if iv >= 0 {
			return cbor.EncodeFixedUint(uint64(iv)), nil
		}
		return cbor.EncodeFixedNegInt(uint64(-iv)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cbor.EncodeFixedUint(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return cbor.Float64(rv.Float()), nil
	default:
		rvt := rv.Type()
		switch {
		case rvt.ConvertibleTo(serde.ReflectTypeOf.BigInt):
			i := rv.Convert(serde.ReflectTypeOf.BigInt).Interface().(big.Int)
			if i.Sign() > -1 {
				return &cbor.Tag{
					ID:    tagbigpos,
					Value: cbor.Slice(i.Bytes()),
				}, nil
			}

			biased := new(big.Int).Add(&i, big.NewInt(1))
			return &cbor.Tag{
				ID:    tagbigneg,
				Value: cbor.Slice(biased.Bytes()),
			}, nil
		case rvt.ConvertibleTo(serde.ReflectTypeOf.BigFloat):
			// FUTURE(rpc2cbor): when document support is enabled, complete this logic
			return &cbor.Tag{}, nil
		default:
			return nil, &document.InvalidMarshalError{
				Message: fmt.Sprintf("incompatible type: %s", rvt.String()),
			}
		}
	}
}
