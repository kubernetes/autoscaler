// Package smithyotelmetrics implements a Smithy client metrics adapter for the
// OTEL Go SDK.
//
// # Usage
//
// Callers use the [Adapt] API in this package to wrap a concrete OTEL SDK
// MeterProvider.
//
// The following example uses the AWS SDK for S3:
//
//	import (
//		"github.com/aws/aws-sdk-go-v2/config"
//		"github.com/aws/aws-sdk-go-v2/service/s3"
//		"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/metrics/smithyotelmetrics"
//		"go.opentelemetry.io/otel/sdk/metric"
//		"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
//	)
//
//	func main() {
//		cfg, err := config.LoadDefaultConfig(context.Background())
//		if err != nil {
//			panic(err)
//		}
//
//		// export via OTLP - perhaps to the otel collector, etc.
//		exporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL("http://localhost:4318"))
//		if err != nil {
//			panic(err)
//		}
//
//		// aggressive reader interval for demonstration purposes
//		reader := metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second))
//		provider := metric.NewMeterProvider(metric.WithReader(reader)
//
//		svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
//			o.MeterProvider = smithyotelmetrics.Adapt(provider)
//		})
//		// ...
//	}
package smithyotelmetrics
