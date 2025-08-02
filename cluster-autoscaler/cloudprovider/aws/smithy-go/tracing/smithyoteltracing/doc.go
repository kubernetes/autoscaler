// Package smithyoteltracing implements a Smithy client tracing adapter for the
// OTEL Go SDK.
//
// # Usage
//
// Callers use the [Adapt] API in this package to wrap a concrete OTEL SDK
// TracerProvider.
//
// The following example uses the AWS SDK for S3:
//
//	import (
//		"github.com/aws/aws-sdk-go-v2/config"
//		"github.com/aws/aws-sdk-go-v2/service/s3"
//		"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/tracing/smithyoteltracing"
//		"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
//		"go.opentelemetry.io/otel/sdk/trace"
//	)
//
//	func main() {
//		exporter, err := stdouttrace.New()
//		if err != nil {
//			panic(err)
//		}
//
//		cfg, err := config.LoadDefaultConfig(context.Background())
//		if err != nil {
//			panic(err)
//		}
//
//		provider := trace.NewTracerProvider(trace.WithBatcher(exporter))
//		svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
//			o.TracerProvider = smithyoteltracing.Adapt(provider)
//		})
//		// ...
//	}
//
// # OTEL Attributes
//
// This adapter supports all attribute types used in the OTEL SDK (including
// their slice-of variants):
//   - bool
//   - int
//   - int64
//   - float64
//   - string
//
// A key-value pair set on a [smithy.Properties] container in any of the
// tracing APIs will automatically propagate to the underlying OTEL SDK if its
// key is of type string and its value is one of those supported. Otherwise,
// the adapter will make an effort to stringify the key or value as follows:
//   - if value implements String(), use that value
//   - otherwise, use the Go-string representation of the value (the result of fmt.Sprintf("%#v", ...))
//
// e.g.
//
//	type key struct{}
//
//	func (k key) String() string {
//		return "app.key"
//	}
//
//	type key2 struct{}
//
//	ctx, span := tracing.StartSpan(ctx, "Foo", func(o *tracing.SpanOptions) {
//		o.Properties.Set("app.version", "bar") // -> ("app.version", "bar")
//		o.Properties.Set(key{}, "baz")         // -> ("app.key", "baz")
//		o.Properties.Set(key2{}, "qux")        // -> ("main.key2{}", "qux")
//	})
package smithyoteltracing
