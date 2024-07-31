/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iputil

import (
	"fmt"
	"net/netip"
)

func IsPrefixesAllowAll(prefixes []netip.Prefix) bool {
	for _, p := range prefixes {
		if p.Bits() == 0 {
			return true
		}
	}
	return false
}

func ParsePrefixes(vs []string) ([]netip.Prefix, error) {
	var rv []netip.Prefix
	for _, v := range vs {
		prefix, err := netip.ParsePrefix(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR `%s`: %w", v, err)
		}
		rv = append(rv, prefix)
	}
	return rv, nil
}
