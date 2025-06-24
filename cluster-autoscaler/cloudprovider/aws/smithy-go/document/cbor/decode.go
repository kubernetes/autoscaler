package cbor

import (
	"fmt"
	"math/big"
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/document"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/document/internal/serde"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/encoding/cbor"
)

// decoderOptions is the set of options that can be configured for a Decoder.
//
// FUTURE(rpc2cbor): document support is currently disabled. This API is
// unexported until that changes.
type decoderOptions struct{}

// decoder is a Smithy document decoder for CBOR-based protocols.
//
// FUTURE(rpc2cbor): document support is currently disabled. This API is
// unexported until that changes.
type decoder struct {
	options decoderOptions
}

// newDecoder returns a Decoder for deserializing Smithy documents.
//
// FUTURE(rpc2cbor): document support is currently disabled. This API is
// unexported until that changes.
func newDecoder(optFns ...func(options *decoderOptions)) *decoder {
	o := decoderOptions{}

	for _, fn := range optFns {
		fn(&o)
	}

	return &decoder{
		options: o,
	}
}

// Decode unmarshals a CBOR Value into the target.
func (d *decoder) Decode(v cbor.Value, to interface{}) error {
	if document.IsNoSerde(to) {
		return fmt.Errorf("unsupported type: %T", to)
	}

	rv := reflect.ValueOf(to)
	if rv.Kind() != reflect.Ptr || rv.IsNil() || !rv.IsValid() {
		return &document.InvalidUnmarshalError{Type: reflect.TypeOf(to)}
	}

	return d.decode(v, rv, serde.Tag{})
}

func (d *decoder) decode(cv cbor.Value, rv reflect.Value, tag serde.Tag) error {
	if _, ok := cv.(*cbor.Nil); ok {
		return d.decodeNil(serde.Indirect(rv, true))
	}

	rv = serde.Indirect(rv, false)
	if err := d.unsupportedType(rv); err != nil {
		return err
	}

	switch v := cv.(type) {
	case cbor.Uint, cbor.NegInt:
		return d.decodeInt(v, rv)
	case cbor.Float64:
		return d.decodeFloat(float64(v), rv)
	case cbor.String:
		return d.decodeString(string(v), rv)
	case cbor.Bool:
		return d.decodeBool(bool(v), rv)
	case cbor.List:
		return d.decodeList(v, rv)
	case cbor.Map:
		return d.decodeMap(v, rv)
	case *cbor.Tag:
		return d.decodeTag(v, rv)
	default:
		return fmt.Errorf("unsupported cbor document type %T", v)
	}
}

func (d *decoder) decodeInt(v cbor.Value, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := cbor.AsInt64(v)
		if err != nil {
			return err
		}
		if rv.OverflowInt(i) {
			return &document.UnmarshalTypeError{
				Value: fmt.Sprintf("number overflow, %d", i),
				Type:  rv.Type(),
			}
		}
		rv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, ok := v.(cbor.Uint)
		if !ok {
			return &document.UnmarshalTypeError{Value: "number", Type: rv.Type()}
		}
		if rv.OverflowUint(uint64(u)) {
			return &document.UnmarshalTypeError{
				Value: fmt.Sprintf("number overflow, %d", u),
				Type:  rv.Type(),
			}
		}
		rv.SetUint(uint64(u))
	default:
		return &document.UnmarshalTypeError{Value: "number", Type: rv.Type()}
	}
	return nil
}

func (d *decoder) decodeNil(rv reflect.Value) error {
	if rv.IsValid() && rv.CanSet() {
		rv.Set(reflect.Zero(rv.Type()))
	}
	return nil
}

func (d *decoder) decodeBool(v bool, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Bool, reflect.Interface:
		rv.Set(reflect.ValueOf(v).Convert(rv.Type()))
	default:
		return &document.UnmarshalTypeError{Value: "bool", Type: rv.Type()}
	}
	return nil
}

func (d *decoder) decodeFloat(v float64, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Interface:
		rv.Set(reflect.ValueOf(v))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, accuracy := big.NewFloat(v).Int64()
		if accuracy != big.Exact || rv.OverflowInt(i) {
			return &document.UnmarshalTypeError{
				Value: fmt.Sprintf("int overflow, %e", v),
				Type:  rv.Type(),
			}
		}
		rv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, accuracy := big.NewFloat(v).Uint64()
		if accuracy != big.Exact || rv.OverflowUint(u) {
			return &document.UnmarshalTypeError{
				Value: fmt.Sprintf("uint overflow, %e", v),
				Type:  rv.Type(),
			}
		}
		rv.SetUint(u)
	case reflect.Float32, reflect.Float64:
		if rv.OverflowFloat(v) {
			return &document.UnmarshalTypeError{
				Value: fmt.Sprintf("float overflow, %e", v),
				Type:  rv.Type(),
			}
		}
		rv.SetFloat(v)
	default:
		return &document.UnmarshalTypeError{Value: "number", Type: rv.Type()}
	}
	return nil
}

func (d *decoder) decodeList(v cbor.List, rv reflect.Value) error {
	var isArray bool

	switch rv.Kind() {
	case reflect.Slice:
		// Make room for the slice elements if needed
		if rv.IsNil() || rv.Cap() < len(v) {
			rv.Set(reflect.MakeSlice(rv.Type(), 0, len(v)))
		}
	case reflect.Array:
		// Limited to capacity of existing array.
		isArray = true
	case reflect.Interface:
		s := make([]interface{}, len(v))
		for i, av := range v {
			if err := d.decode(av, reflect.ValueOf(&s[i]).Elem(), serde.Tag{}); err != nil {
				return err
			}
		}
		rv.Set(reflect.ValueOf(s))
		return nil
	default:
		return &document.UnmarshalTypeError{Value: "list", Type: rv.Type()}
	}

	// If rv is not a slice, array
	for i := 0; i < rv.Cap() && i < len(v); i++ {
		if !isArray {
			rv.SetLen(i + 1)
		}
		if err := d.decode(v[i], rv.Index(i), serde.Tag{}); err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) decodeString(v string, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.String:
		rv.SetString(v)
	case reflect.Interface:
		rv.Set(reflect.ValueOf(v).Convert(rv.Type()))
	default:
		return &document.UnmarshalTypeError{Value: "string", Type: rv.Type()}
	}
	return nil
}

func (d *decoder) decodeMap(tv cbor.Map, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Map:
		t := rv.Type()
		if t.Key().Kind() != reflect.String {
			return &document.UnmarshalTypeError{Value: "map string key", Type: t.Key()}
		}
		if rv.IsNil() {
			rv.Set(reflect.MakeMap(t))
		}
	case reflect.Struct:
		if rv.CanInterface() && document.IsNoSerde(rv.Interface()) {
			return &document.UnmarshalTypeError{
				Value: fmt.Sprintf("unsupported type"),
				Type:  rv.Type(),
			}
		}
	case reflect.Interface:
		rv.Set(reflect.MakeMap(serde.ReflectTypeOf.MapStringToInterface))
		rv = rv.Elem()
	default:
		return &document.UnmarshalTypeError{Value: "map", Type: rv.Type()}
	}

	if rv.Kind() == reflect.Map {
		for k, kv := range tv {
			key := reflect.New(rv.Type().Key()).Elem()
			key.SetString(k)
			elem := reflect.New(rv.Type().Elem()).Elem()
			if err := d.decode(kv, elem, serde.Tag{}); err != nil {
				return err
			}
			rv.SetMapIndex(key, elem)
		}
	} else if rv.Kind() == reflect.Struct {
		fields := serde.GetStructFields(rv.Type())
		for k, kv := range tv {
			if f, ok := fields.FieldByName(k); ok {
				fv := serde.DecoderFieldByIndex(rv, f.Index)
				if err := d.decode(kv, fv, f.Tag); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (d *decoder) decodeTag(tv *cbor.Tag, rv reflect.Value) error {
	rvt := rv.Type()
	switch {
	case rvt.ConvertibleTo(serde.ReflectTypeOf.BigInt):
		i, err := cbor.AsBigInt(tv)
		if err != nil {
			return &document.UnmarshalTypeError{Value: "tag", Type: rv.Type()}
		}

		rv.Set(reflect.ValueOf(*i).Convert(rvt))
		return nil
	case rvt.ConvertibleTo(serde.ReflectTypeOf.BigFloat):
		i, err := asBigFloat(tv)
		if err != nil {
			return &document.UnmarshalTypeError{Value: "tag", Type: rv.Type()}
		}

		rv.Set(reflect.ValueOf(*i).Convert(rvt))
		return nil
	default:
		return &document.UnmarshalTypeError{Value: "tag", Type: rv.Type()}
	}
}

func (d *decoder) unsupportedType(rv reflect.Value) error {
	if rv.Kind() == reflect.Interface && rv.NumMethod() != 0 {
		return &document.UnmarshalTypeError{Value: "non-empty interface", Type: rv.Type()}
	}

	if rv.Type().ConvertibleTo(serde.ReflectTypeOf.Time) {
		return &document.UnmarshalTypeError{
			Type:  rv.Type(),
			Value: fmt.Sprintf("time value"),
		}
	}
	return nil
}

func asBigFloat(tv *cbor.Tag) (*big.Float, error) {
	const tagbase10 = 4

	if tv.ID != tagbase10 {
		return nil, fmt.Errorf("invalid tag: %d", tv.ID)
	}

	pcs, ok := tv.Value.(cbor.List)
	if !ok {
		return nil, fmt.Errorf("invalid tagged type: %T", tv.Value)
	}

	if len(pcs) != 2 {
		return nil, fmt.Errorf("invalid tagged list len: %d", len(pcs))
	}

	eval, mval := pcs[0], pcs[1]
	exp, err := cbor.AsBigInt(eval)
	if err != nil {
		return nil, fmt.Errorf("invalid exp: %w", err)
	}

	mant, err := cbor.AsBigInt(mval)
	if !ok {
		return nil, fmt.Errorf("invalid mant: %w", err)
	}

	// We literally re-express this as <mant>e<exp> and send it through
	// bigfloat parse. Not mathematically amazing, but ensures that
	// string-borne bignums and this are computed identically.
	str := fmt.Sprintf("%se%s", mant.String(), exp.String())
	x, _, err := new(big.Float).Parse(str, 0)
	return x, err
}
