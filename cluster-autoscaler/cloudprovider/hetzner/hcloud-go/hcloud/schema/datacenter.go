package schema

// Datacenter defines the schema of a datacenter.
type Datacenter struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Location    Location `json:"location"`
	ServerTypes struct {
		Supported []int64 `json:"supported"`
		Available []int64 `json:"available"`
	} `json:"server_types"`
}

// DatacenterGetResponse defines the schema of the response when retrieving a single datacenter.
type DatacenterGetResponse struct {
	Datacenter Datacenter `json:"datacenter"`
}

// DatacenterListResponse defines the schema of the response when listing datacenters.
type DatacenterListResponse struct {
	Datacenters []Datacenter `json:"datacenters"`
}
