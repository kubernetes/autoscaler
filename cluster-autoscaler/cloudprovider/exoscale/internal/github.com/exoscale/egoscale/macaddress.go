/*
Copyright 2020 The Kubernetes Authors.

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

package egoscale

import (
	"encoding/json"
	"fmt"
	"net"
)

// MACAddress is a nicely JSON serializable net.HardwareAddr
type MACAddress net.HardwareAddr

// String returns the MAC address in standard format
func (mac MACAddress) String() string {
	return (net.HardwareAddr)(mac).String()
}

// MAC48 builds a MAC-48 MACAddress
func MAC48(a, b, c, d, e, f byte) MACAddress {
	m := make(MACAddress, 6)
	m[0] = a
	m[1] = b
	m[2] = c
	m[3] = d
	m[4] = e
	m[5] = f
	return m
}

// UnmarshalJSON unmarshals the raw JSON into the MAC address
func (mac *MACAddress) UnmarshalJSON(b []byte) error {
	var addr string
	if err := json.Unmarshal(b, &addr); err != nil {
		return err
	}
	hw, err := ParseMAC(addr)
	if err != nil {
		return err
	}

	*mac = make(MACAddress, 6)
	copy(*mac, hw)
	return nil
}

// MarshalJSON converts the MAC Address to a string representation
func (mac MACAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", mac.String())), nil
}

// ParseMAC converts a string into a MACAddress
func ParseMAC(s string) (MACAddress, error) {
	hw, err := net.ParseMAC(s)
	return (MACAddress)(hw), err
}

// MustParseMAC acts like ParseMAC but panics if in case of an error
func MustParseMAC(s string) MACAddress {
	mac, err := ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return mac
}
