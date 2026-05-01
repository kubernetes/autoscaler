package errutil

import (
	"errors"
	"fmt"
	"log/slog"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
)

// LogValue returns a [slog.Value] for a [hcloud.Error].
func LogValue(err error) slog.Value {
	var herr *hcloud.Error
	if !errors.As(err, &herr) {
		return slog.AnyValue(err)
	}

	attrs := []slog.Attr{
		slog.String("msg", herr.Message),
		slog.String("code", string(herr.Code)),
	}

	// Rework once moved from the exp package namespace to the hcloud package, because
	// some hcloud private APIs are not available from here.
	if resp := herr.Response(); resp != nil {
		if value := resp.Header.Get("X-Correlation-Id"); value != "" {
			attrs = append(attrs,
				slog.String("correlation-id", value),
			)
		}
	}

	if herr.Details != nil {
		switch details := herr.Details.(type) {
		case hcloud.ErrorDetailsInvalidInput:
			attrs = append(attrs,
				slog.String("details", fmt.Sprintf("%v", details.Fields)),
			)
		case hcloud.ErrorDetailsDeprecatedAPIEndpoint:
			attrs = append(attrs,
				slog.String("details", fmt.Sprintf("%v", details)),
			)
		default:
			attrs = append(attrs,
				slog.String("details", fmt.Sprintf("%v", details)),
			)
		}
	}

	return slog.GroupValue(attrs...)
}
