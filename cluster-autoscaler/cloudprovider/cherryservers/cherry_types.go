/*
Copyright 2022 The Kubernetes Authors.

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

package cherryservers

// BGPRoute single server BGP route
type BGPRoute struct {
	Subnet  string `json:"subnet,omitempty"`
	Active  bool   `json:"active,omitempty"`
	Router  string `json:"router,omitempty"`
	Age     string `json:"age,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// ServerBGP status of BGP on a server
type ServerBGP struct {
	Enabled   bool       `json:"enabled"`
	Available bool       `json:"available,omitempty"`
	Status    string     `json:"status,omitempty"`
	Routers   int        `json:"routers,omitempty"`
	Connected int        `json:"connected,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Active    int        `json:"active,omitempty"`
	Routes    []BGPRoute `json:"routes,omitempty"`
	Updated   string     `json:"updated,omitempty"`
}

// Project a CherryServers project
type Project struct {
	ID   int        `json:"id,omitempty"`
	Name string     `json:"name,omitempty"`
	Bgp  ProjectBGP `json:"bgp,omitempty"`
	Href string     `json:"href,omitempty"`
}

// Region a CherryServers region
type Region struct {
	ID         int       `json:"id,omitempty"`
	Slug       string    `json:"slug,omitempty"`
	Name       string    `json:"name,omitempty"`
	RegionIso2 string    `json:"region_iso_2,omitempty"`
	BGP        RegionBGP `json:"bgp,omitempty"`
	Href       string    `json:"href,omitempty"`
}

// RegionBGP information about BGP in a region
type RegionBGP struct {
	Hosts []string `json:"hosts,omitempty"`
	Asn   int      `json:"asn,omitempty"`
}

// ProjectBGP information about BGP on an individual project
type ProjectBGP struct {
	Enabled  bool `json:"enabled,omitempty"`
	LocalASN int  `json:"local_asn,omitempty"`
}

// Plan a server plan
type Plan struct {
	ID               int                `json:"id,omitempty"`
	Slug             string             `json:"slug,omitempty"`
	Name             string             `json:"name,omitempty"`
	Custom           bool               `json:"custom,omitempty"`
	Specs            Specs              `json:"specs,omitempty"`
	Pricing          []Pricing          `json:"pricing,omitempty"`
	AvailableRegions []AvailableRegions `json:"available_regions,omitempty"`
}

// Plans represents a list of Cherry Servers plans
type Plans []Plan

// Pricing price for a specific plan
type Pricing struct {
	Price    float32 `json:"price,omitempty"`
	Taxed    bool    `json:"taxed,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Unit     string  `json:"unit,omitempty"`
}

// AvailableRegions regions that are available to the user
type AvailableRegions struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	RegionIso2 string `json:"region_iso_2,omitempty"`
	StockQty   int    `json:"stock_qty,omitempty"`
}

// AttachedTo what a resource is attached to
type AttachedTo struct {
	Href string `json:"href"`
}

// BlockStorage cloud block storage
type BlockStorage struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Href          string     `json:"href"`
	Size          int        `json:"size"`
	AllowEditSize bool       `json:"allow_edit_size"`
	Unit          string     `json:"unit"`
	Description   string     `json:"description,omitempty"`
	AttachedTo    AttachedTo `json:"attached_to,omitempty"`
	VlanID        string     `json:"vlan_id"`
	VlanIP        string     `json:"vlan_ip"`
	Initiator     string     `json:"initiator"`
	DiscoveryIP   string     `json:"discovery_ip"`
}

// AssignedTo assignment of a network floating IP to a server
type AssignedTo struct {
	ID       int     `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	Href     string  `json:"href,omitempty"`
	Hostname string  `json:"hostname,omitempty"`
	Image    string  `json:"image,omitempty"`
	Region   Region  `json:"region,omitempty"`
	State    string  `json:"state,omitempty"`
	Pricing  Pricing `json:"pricing,omitempty"`
}

// RoutedTo routing of a floating IP to an underlying IP
type RoutedTo struct {
	ID            string `json:"id,omitempty"`
	Address       string `json:"address,omitempty"`
	AddressFamily int    `json:"address_family,omitempty"`
	Cidr          string `json:"cidr,omitempty"`
	Gateway       string `json:"gateway,omitempty"`
	Type          string `json:"type,omitempty"`
	Region        Region `json:"region,omitempty"`
}

// IPAddresses individual IP address
type IPAddresses struct {
	ID            string            `json:"id,omitempty"`
	Address       string            `json:"address,omitempty"`
	AddressFamily int               `json:"address_family,omitempty"`
	Cidr          string            `json:"cidr,omitempty"`
	Gateway       string            `json:"gateway,omitempty"`
	Type          string            `json:"type,omitempty"`
	Region        Region            `json:"region,omitempty"`
	RoutedTo      RoutedTo          `json:"routed_to,omitempty"`
	AssignedTo    AssignedTo        `json:"assigned_to,omitempty"`
	TargetedTo    AssignedTo        `json:"targeted_to,omitempty"`
	Project       Project           `json:"project,omitempty"`
	PtrRecord     string            `json:"ptr_record,omitempty"`
	ARecord       string            `json:"a_record,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	Href          string            `json:"href,omitempty"`
}

// Server represents a Cherry Servers server
type Server struct {
	ID               int               `json:"id,omitempty"`
	Name             string            `json:"name,omitempty"`
	Href             string            `json:"href,omitempty"`
	Hostname         string            `json:"hostname,omitempty"`
	Image            string            `json:"image,omitempty"`
	SpotInstance     bool              `json:"spot_instance"`
	BGP              ServerBGP         `json:"bgp,omitempty"`
	Project          Project           `json:"project,omitempty"`
	Region           Region            `json:"region,omitempty"`
	State            string            `json:"state,omitempty"`
	Plan             Plan              `json:"plan,omitempty"`
	AvailableRegions AvailableRegions  `json:"availableregions,omitempty"`
	Pricing          Pricing           `json:"pricing,omitempty"`
	IPAddresses      []IPAddresses     `json:"ip_addresses,omitempty"`
	SSHKeys          []SSHKeys         `json:"ssh_keys,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
	Storage          BlockStorage      `json:"storage,omitempty"`
	Created          string            `json:"created_at,omitempty"`
	TerminationDate  string            `json:"termination_date,omitempty"`
}

// SSHKeys an ssh key
type SSHKeys struct {
	ID          int    `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Key         string `json:"key,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Updated     string `json:"updated,omitempty"`
	Created     string `json:"created,omitempty"`
	Href        string `json:"href,omitempty"`
}

// Cpus cpu information for a server
type Cpus struct {
	Count     int     `json:"count,omitempty"`
	Name      string  `json:"name,omitempty"`
	Cores     int     `json:"cores,omitempty"`
	Frequency float32 `json:"frequency,omitempty"`
	Unit      string  `json:"unit,omitempty"`
}

// Memory cpu information for a server
type Memory struct {
	Count int    `json:"count,omitempty"`
	Total int    `json:"total,omitempty"`
	Unit  string `json:"unit,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Nics network interface information for a server
type Nics struct {
	Name string `json:"name,omitempty"`
}

// Raid raid for block storage on a server
type Raid struct {
	Name string `json:"name,omitempty"`
}

// Storage amount of storage
type Storage struct {
	Count int     `json:"count,omitempty"`
	Name  string  `json:"name,omitempty"`
	Size  float32 `json:"size,omitempty"`
	Unit  string  `json:"unit,omitempty"`
}

// Bandwidth total bandwidth available
type Bandwidth struct {
	Name string `json:"name,omitempty"`
}

// Specs aggregated specs for a server
type Specs struct {
	Cpus      Cpus      `json:"cpus,omitempty"`
	Memory    Memory    `json:"memory,omitempty"`
	Storage   []Storage `json:"storage,omitempty"`
	Raid      Raid      `json:"raid,omitempty"`
	Nics      Nics      `json:"nics,omitempty"`
	Bandwidth Bandwidth `json:"bandwidth,omitempty"`
}

// IPAddressCreateRequest represents a request to create a new IP address within a CreateServer request
type IPAddressCreateRequest struct {
	AddressFamily int  `json:"address_family"`
	Public        bool `json:"public"`
}

// CreateServer represents a request to create a new Cherry Servers server. Used by createNodes
type CreateServer struct {
	ProjectID       int                `json:"project_id,omitempty"`
	Plan            string             `json:"plan,omitempty"`
	Hostname        string             `json:"hostname,omitempty"`
	Image           string             `json:"image,omitempty"`
	Region          string             `json:"region,omitempty"`
	SSHKeys         []int              `json:"ssh_keys"`
	IPAddresses     []string           `json:"ip_addresses"`
	UserData        string             `json:"user_data,omitempty"`
	Tags            *map[string]string `json:"tags,omitempty"`
	SpotInstance    int                `json:"spot_market,omitempty"`
	OSPartitionSize int                `json:"os_partition_size,omitempty"`
}
