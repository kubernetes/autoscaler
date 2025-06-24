package smithyoteltracing

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go"
	otelattribute "go.opentelemetry.io/otel/attribute"
)

// IMPORTANT: The contents of this file are mirrored in
// smithyotelmetrics/attribute.go. Any changes made here must be replicated in
// that module's copy of the file, although that will probably never happen, as
// the set of attribute types supported by the OTEL API cannot reasonably
// expand to include anything else that would be useful.
//
// This is done in order to avoid the one-way door of exposing an internal-only
// module for what is effectively a simple value mapper (that will likely never
// change).
//
// While the contents of the file are mirrored, the tests are only present
// here.

func toOTELKeyValue(k, v any) otelattribute.KeyValue {
	kk := str(k)

	switch vv := v.(type) {
	case bool:
		return otelattribute.Bool(kk, vv)
	case []bool:
		return otelattribute.BoolSlice(kk, vv)
	case int:
		return otelattribute.Int(kk, vv)
	case []int:
		return otelattribute.IntSlice(kk, vv)
	case int64:
		return otelattribute.Int64(kk, vv)
	case []int64:
		return otelattribute.Int64Slice(kk, vv)
	case float64:
		return otelattribute.Float64(kk, vv)
	case []float64:
		return otelattribute.Float64Slice(kk, vv)
	case string:
		return otelattribute.String(kk, vv)
	case []string:
		return otelattribute.StringSlice(kk, vv)
	default:
		return otelattribute.String(kk, str(v))
	}
}

func toOTELKeyValues(props smithy.Properties) []otelattribute.KeyValue {
	var kvs []otelattribute.KeyValue
	for k, v := range props.Values() {
		kvs = append(kvs, toOTELKeyValue(k, v))
	}
	return kvs
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	} else if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%#v", v)
}
