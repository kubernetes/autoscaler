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

func ParseAddresses(vs []string) ([]netip.Addr, error) {
	var rv []netip.Addr
	for _, v := range vs {
		addr, err := netip.ParseAddr(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR `%s`: %w", v, err)
		}
		rv = append(rv, addr)
	}
	return rv, nil
}

func GroupAddressesByFamily(vs []netip.Addr) ([]netip.Addr, []netip.Addr) {
	var (
		v4 []netip.Addr
		v6 []netip.Addr
	)
	for _, v := range vs {
		if v.Is4() {
			v4 = append(v4, v)
		} else {
			v6 = append(v6, v)
		}
	}
	return v4, v6
}
