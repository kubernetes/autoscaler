package deprecationutil

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/exp/kit/sliceutil"
)

// ServerTypeMessage return a deprecation message when the given Server Type is
// deprecated and whether the given Server Type is unavailable.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func ServerTypeMessage(serverType *hcloud.ServerType, locationName string) (string, bool) {
	if serverType.IsDeprecated() {
		if time.Now().After(serverType.UnavailableAfter()) {
			return fmt.Sprintf(
				"Server Type %q is unavailable in all locations and can no longer be ordered",
				serverType.Name,
			), true
		}
		return fmt.Sprintf(
			"Server Type %q is deprecated in all locations and will no longer be available for order as of %s",
			serverType.Name,
			serverType.UnavailableAfter().Format(time.DateOnly),
		), false
	}

	deprecated := make([]hcloud.ServerTypeLocation, 0, len(serverType.Locations))
	unavailable := make([]hcloud.ServerTypeLocation, 0, len(serverType.Locations))

	for _, o := range serverType.Locations {
		if o.IsDeprecated() {
			deprecated = append(deprecated, o)
			if time.Now().After(o.UnavailableAfter()) {
				unavailable = append(unavailable, o)
			}
		}
	}

	if len(deprecated) == 0 {
		return "", false
	}

	// A location or a datacenter was provided
	if locationName != "" {
		locationIndex := slices.IndexFunc(deprecated, func(o hcloud.ServerTypeLocation) bool {
			return o.Location.Name == locationName
		})

		// No deprecation for this location
		if locationIndex < 0 {
			return "", false
		}

		if time.Now().After(deprecated[locationIndex].UnavailableAfter()) {
			return fmt.Sprintf(
				"Server Type %q is unavailable in %q and can no longer be ordered",
				serverType.Name,
				deprecated[locationIndex].Location.Name,
			), true
		}
		return fmt.Sprintf(
			"Server Type %q is deprecated in %q and will no longer be available for order as of %s",
			serverType.Name,
			deprecated[locationIndex].Location.Name,
			deprecated[locationIndex].UnavailableAfter().Format(time.DateOnly),
		), false
	}

	// Only print a warning when all locations are deprecated
	if len(serverType.Locations) != len(deprecated) {
		return "", false
	}

	deprecatedNames := sliceutil.Transform(deprecated, func(e hcloud.ServerTypeLocation) string { return e.Location.Name })
	unavailableNames := sliceutil.Transform(unavailable, func(e hcloud.ServerTypeLocation) string { return e.Location.Name })

	if len(unavailable) > 0 {
		if len(deprecated) == len(unavailable) {
			// All are deprecated and all are unavailable
			return fmt.Sprintf(
				"Server Type %q is unavailable in all locations (%s) and can no longer be ordered",
				serverType.Name,
				strings.Join(unavailableNames, ","),
			), true
		}
		// All are deprecated and some are unavailable
		return fmt.Sprintf(
			"Server Type %q is deprecated in all locations (%s) and can no longer be ordered some locations (%s)",
			serverType.Name,
			strings.Join(deprecatedNames, ","),
			strings.Join(unavailableNames, ","),
		), false
	}
	// All are deprecated and none are unavailable
	return fmt.Sprintf(
		"Server Type %q is deprecated in all locations (%s) and will no longer be available for order",
		serverType.Name,
		strings.Join(deprecatedNames, ","),
	), false
}
