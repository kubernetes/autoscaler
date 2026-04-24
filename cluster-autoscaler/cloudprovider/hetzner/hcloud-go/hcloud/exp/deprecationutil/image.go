package deprecationutil

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
)

// ImageMessage return a deprecation message when the given Image is
// deprecated and whether the given Image is unavailable.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func ImageMessage(image *hcloud.Image) (string, bool) {
	if image.IsDeprecated() {
		// Images are unavailable 3 months after the announcement
		unavailableAfter := image.Deprecated.AddDate(0, 3, 0)

		if time.Now().After(unavailableAfter) {
			return fmt.Sprintf(
				"Image %q is unavailable and can no longer be ordered",
				image.Name,
			), true
		}
		return fmt.Sprintf(
			"Image %q is deprecated and will no longer be available for order as of %s",
			image.Name,
			unavailableAfter.Format(time.DateOnly),
		), false
	}

	return "", false
}
