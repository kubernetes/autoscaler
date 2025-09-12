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
	"net/netip"

	"sigs.k8s.io/cloud-provider-azure/pkg/provider/loadbalancer/fnutil"
)

type Family string

const (
	IPv4 Family = "IPv4"
	IPv6 Family = "IPv6"
)

func FamilyOfAddr(addr netip.Addr) Family {
	if addr.Is4() {
		return IPv4
	}
	return IPv6
}

func ArePrefixesFromSameFamily(prefixes []netip.Prefix) bool {
	if len(prefixes) <= 1 {
		return true
	}
	return fnutil.IsAll(func(p netip.Prefix) bool {
		return p.Addr().Is4() == prefixes[0].Addr().Is4()
	}, prefixes)
}

func AreAddressesFromSameFamily(addresses []netip.Addr) bool {
	if len(addresses) <= 1 {
		return true
	}
	return fnutil.IsAll(func(p netip.Addr) bool {
		return p.Is4() == addresses[0].Is4()
	}, addresses)
}
