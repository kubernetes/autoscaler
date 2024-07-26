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

func ParsePrefix(v string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(v)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("invalid CIDR `%s`: %w", v, err)
	}
	masked := prefix.Masked()
	if prefix.Addr().Compare(masked.Addr()) != 0 {
		return netip.Prefix{}, fmt.Errorf("invalid CIDR `%s`: not a valid network prefix, should be properly masked like %s", v, masked)
	}
	return prefix, nil
}
