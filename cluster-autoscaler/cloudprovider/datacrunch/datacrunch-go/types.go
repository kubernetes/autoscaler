package datacrunchclient

// APIError represents an error response from the DataCrunch API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Instance represents a DataCrunch instance.
type Instance struct {
	ID            string   `json:"id"`
	IP            string   `json:"ip"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"created_at"`
	CPU           CPU      `json:"cpu"`
	GPU           GPU      `json:"gpu"`
	GPUMemory     GPUMem   `json:"gpu_memory"`
	Memory        Memory   `json:"memory"`
	Storage       Storage  `json:"storage"`
	Hostname      string   `json:"hostname"`
	Description   string   `json:"description"`
	Location      string   `json:"location"`
	PricePerHour  float64  `json:"price_per_hour"`
	IsSpot        bool     `json:"is_spot"`
	InstanceType  string   `json:"instance_type"`
	Image         string   `json:"image"`
	OSName        string   `json:"os_name"`
	StartupScript string   `json:"startup_script_id"`
	SSHKeyIDs     []string `json:"ssh_key_ids"`
	OSVolumeID    string   `json:"os_volume_id"`
	JupyterToken  string   `json:"jupyter_token"`
	Contract      string   `json:"contract"`
	Pricing       string   `json:"pricing"`
}

// CPU represents a CPU in DataCrunch.
type CPU struct {
	Description   string `json:"description"`
	NumberOfCores int    `json:"number_of_cores"`
}

// GPU represents a GPU in DataCrunch.
type GPU struct {
	Description  string `json:"description"`
	NumberOfGPUs int    `json:"number_of_gpus"`
}

// GPUMem represents a GPU memory in DataCrunch.
type GPUMem struct {
	Description     string `json:"description"`
	SizeInGigabytes int    `json:"size_in_gigabytes"`
}

// Memory represents a memory in DataCrunch.
type Memory struct {
	Description     string `json:"description"`
	SizeInGigabytes int    `json:"size_in_gigabytes"`
}

// Storage represents a storage in DataCrunch.
type Storage struct {
	Description string `json:"description"`
}

// InstanceList is a list of instances.
type InstanceList []Instance

// DeployInstanceRequest is the request body for creating a new instance.
type DeployInstanceRequest struct {
	InstanceType    string         `json:"instance_type"`
	Image           string         `json:"image"`
	SSHKeyIDs       []string       `json:"ssh_key_ids,omitempty"`
	StartupScriptID string         `json:"startup_script_id,omitempty"`
	Hostname        string         `json:"hostname"`
	Description     string         `json:"description"`
	LocationCode    string         `json:"location_code,omitempty"`
	OSVolume        *OSVolume      `json:"os_volume,omitempty"`
	IsSpot          bool           `json:"is_spot"`
	Coupon          string         `json:"coupon,omitempty"`
	Volumes         []DeployVolume `json:"volumes,omitempty"`
	ExistingVolumes []string       `json:"existing_volumes,omitempty"`
	Contract        string         `json:"contract,omitempty"`
	Pricing         string         `json:"pricing,omitempty"`
}

// OSVolume represents an OS volume in DataCrunch when creating a new instance.
type OSVolume struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// DeployVolume represents a deploy volume in DataCrunch when creating a new instance.
type DeployVolume struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Type string `json:"type"`
}

// DeployInstanceResponse is the response for creating a new instance (instance ID).
type DeployInstanceResponse string

// InstanceActionRequest is the request body for performing an action on an instance.
type InstanceActionRequest struct {
	Action    string   `json:"action"`
	ID        string   `json:"id"`
	VolumeIDs []string `json:"volume_ids,omitempty"`
}

// InstanceType represents an instance type.
type InstanceType struct {
	BestFor             []string `json:"best_for"`
	CPU                 CPU      `json:"cpu"`
	DeployWarning       string   `json:"deploy_warning"`
	Description         string   `json:"description"`
	GPU                 GPU      `json:"gpu"`
	GPUMemory           GPUMem   `json:"gpu_memory"`
	Memory              Memory   `json:"memory"`
	Model               string   `json:"model"`
	ID                  string   `json:"id"`
	InstanceType        string   `json:"instance_type"`
	Name                string   `json:"name"`
	P2P                 string   `json:"p2p"`
	PricePerHour        string   `json:"price_per_hour"`
	SpotPrice           string   `json:"spot_price"`
	DynamicPrice        string   `json:"dynamic_price"`
	MaxDynamicPrice     string   `json:"max_dynamic_price"`
	ServerlessPrice     string   `json:"serverless_price"`
	ServerlessSpotPrice string   `json:"serverless_spot_price"`
	Storage             Storage  `json:"storage"`
	Currency            string   `json:"currency"`
	Manufacturer        string   `json:"manufacturer"`
	DisplayName         string   `json:"display_name"`
}

// InstanceTypeList is a list of instance types.
type InstanceTypeList []InstanceType

// PriceHistoryEntry represents a price history entry for an instance type.
type PriceHistoryEntry struct {
	Date                string  `json:"date"`
	FixedPricePerHour   float64 `json:"fixed_price_per_hour"`
	DynamicPricePerHour float64 `json:"dynamic_price_per_hour"`
	Currency            string  `json:"currency"`
}

// PriceHistory is a map of instance types to price history entries.
type PriceHistory map[string][]PriceHistoryEntry

// InstanceAvailability represents instance availability for a location.
type InstanceAvailability struct {
	LocationCode   string   `json:"location_code"`
	Availabilities []string `json:"availabilities"`
}

// InstanceAvailabilityList is a list of instance availabilities.
type InstanceAvailabilityList []InstanceAvailability

// StartupScript represents a startup script in DataCrunch.
type StartupScript struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Script string `json:"script"`
}
