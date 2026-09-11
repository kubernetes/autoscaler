package upcloud

import "encoding/json"

type IPAddressReleasePolicy string

// Constants
const (
	IPAddressReleasePolicyKeep    IPAddressReleasePolicy = "keep"
	IPAddressReleasePolicyRelease IPAddressReleasePolicy = "release"

	IPAddressFamilyIPv4 = "IPv4"
	IPAddressFamilyIPv6 = "IPv6"

	IPAddressAccessPrivate = "private"
	IPAddressAccessPublic  = "public"
	IPAddressAccessUtility = "utility"
)

// IPAddresses represents a /ip_address response
type IPAddresses struct {
	IPAddresses []IPAddress `json:"ip_addresses"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *IPAddresses) UnmarshalJSON(b []byte) error {
	type localIPAddress IPAddress
	type ipAddressWrapper struct {
		IPAddresses []localIPAddress `json:"ip_address"`
	}

	v := struct {
		IPAddresses ipAddressWrapper `json:"ip_addresses"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ip := range v.IPAddresses.IPAddresses {
		s.IPAddresses = append(s.IPAddresses, IPAddress(ip))
	}

	return nil
}

// IPAddressSlice is a slice of IPAddress.
// It exists to allow for a custom JSON unmarshaller.
type IPAddressSlice []IPAddress

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (i *IPAddressSlice) UnmarshalJSON(b []byte) error {
	type localIPAddress IPAddress
	v := struct {
		IPAddresses []localIPAddress `json:"ip_address"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ip := range v.IPAddresses {
		(*i) = append((*i), IPAddress(ip))
	}

	return nil
}

// IPAddress represents an IP address
type IPAddress struct {
	Access        string                 `json:"access"`
	Address       string                 `json:"address"`
	DHCPProvided  Boolean                `json:"dhcp_provided"`
	Family        string                 `json:"family"`
	PartOfPlan    Boolean                `json:"part_of_plan"`
	PTRRecord     string                 `json:"ptr_record"`
	ReleasePolicy IPAddressReleasePolicy `json:"release_policy"`
	ServerUUID    string                 `json:"server"`
	MAC           string                 `json:"mac"`
	Floating      Boolean                `json:"floating"`
	Zone          string                 `json:"zone"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *IPAddress) UnmarshalJSON(b []byte) error {
	type localIPAddress IPAddress

	v := struct {
		IPAddress localIPAddress `json:"ip_address"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = IPAddress(v.IPAddress)

	return nil
}
