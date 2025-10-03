package labelutil

import (
	"fmt"
	"sort"
	"strings"
)

// Selector combines the label set into a [label selector](https://docs.hetzner.cloud/reference/cloud#label-selector) that only selects
// resources have all specified labels set.
//
// The selector string can be used to filter resources when listing, for example with [hcloud.ServerClient.AllWithOpts()].
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func Selector(labels map[string]string) string {
	selectors := make([]string, 0, len(labels))

	for k, v := range labels {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}

	// Reproducible result for tests
	sort.Strings(selectors)

	return strings.Join(selectors, ",")
}
