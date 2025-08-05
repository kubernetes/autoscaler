/*
Copyright 2025 The Kubernetes Authors.

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

package utho

import (
	"encoding/json"
	"errors"
)

// CloudInstancesService provides methods for managing cloud instances via the Utho API.
type CloudInstancesService service

// CloudInstances represents a list of cloud instances returned by the API.
type CloudInstances struct {
	CloudInstance []CloudInstance `json:"cloud"`
	Meta          Meta            `json:"meta"`
	Status        string          `json:"status,omitempty" faker:"oneof:success,error"`
	Message       string          `json:"message,omitempty" faker:"sentence"`
}

// CloudInstance represents a single cloud instance.
type CloudInstance struct {
	ID                string                   `json:"cloudid" faker:"oneof: 00000,11111,22222,33333"`
	Hostname          string                   `json:"hostname"`
	CPU               string                   `json:"cpu" faker:"oneof:1,2,4,8,16"`
	RAM               string                   `json:"ram" faker:"oneof:1024,2048,4096,8192,16384"`
	DiscountType      string                   `json:"discount_type" faker:"oneof:Percentage,Flat"`
	DiscountValue     string                   `json:"discount_value" faker:"oneof:,10,20,30"`
	ManagedOs         string                   `json:"managed_os,omitempty"`
	ManagedFull       string                   `json:"managed_full,omitempty"`
	ManagedOnetime    string                   `json:"managed_onetime,omitempty"`
	PlanDisksize      int                      `json:"plan_disksize"`
	Disksize          int                      `json:"disksize"`
	Ha                string                   `json:"ha" faker:"oneof:0,1"`
	Status            string                   `json:"status" faker:"oneof:Active,Stopped,Pending"`
	Iso               string                   `json:"iso,omitempty"`
	IP                string                   `json:"ip" faker:"ipv4"`
	Billingcycle      string                   `json:"billingcycle" faker:"oneof:hourly,monthly"`
	Cost              float64                  `json:"cost"`
	Vmcost            float64                  `json:"vmcost"`
	Imagecost         int                      `json:"imagecost"`
	Backupcost        float64                  `json:"backupcost"`
	Hourlycost        float64                  `json:"hourlycost"`
	Cloudhourlycost   float64                  `json:"cloudhourlycost"`
	Imagehourlycost   int                      `json:"imagehourlycost"`
	Backuphourlycost  float64                  `json:"backuphourlycost"`
	Creditrequired    float64                  `json:"creditrequired"`
	Creditreserved    int                      `json:"creditreserved"`
	Nextinvoiceamount float64                  `json:"nextinvoiceamount"`
	Nextinvoicehours  string                   `json:"nextinvoicehours" faker:"oneof:1,2,3,4,5,6,7,8,9,10"`
	Consolepassword   string                   `json:"consolepassword" faker:"password"`
	Powerstatus       string                   `json:"powerstatus" faker:"oneof:Running,Stopped"`
	CreatedAt         string                   `json:"created_at" faker:"date"`
	UpdatedAt         string                   `json:"updated_at" faker:"date"`
	Nextduedate       string                   `json:"nextduedate" faker:"date"`
	Bandwidth         string                   `json:"bandwidth" faker:"oneof:100,500,1000,2000"`
	BandwidthUsed     int                      `json:"bandwidth_used"`
	BandwidthFree     int                      `json:"bandwidth_free"`
	Features          Features                 `json:"features"`
	Image             Image                    `json:"image"`
	Dclocation        Dclocation               `json:"dclocation"`
	V4                V4Public                 `json:"v4"`
	Networks          Networks                 `json:"networks"`
	V4Private         V4Private                `json:"v4private"`
	Storages          []Storages               `json:"storages,omitempty"`
	Storage           Storage                  `json:"storage"`
	DiskUsed          int                      `json:"disk_used"`
	DiskFree          int                      `json:"disk_free"`
	DiskUsedp         int                      `json:"disk_usedp"`
	Backups           []any                    `json:"backups,omitempty"`
	Snapshots         []Snapshots              `json:"snapshots,omitempty"`
	Firewalls         []CloudInstanceFirewalls `json:"firewalls,omitempty"`
	GpuAvailable      string                   `json:"gpu_available,omitempty" faker:"oneof:0,1"`
	Gpus              []any                    `json:"gpus,omitempty"`
	Rescue            int                      `json:"rescue"`
}

// Features represents the features available for a cloud instance.
type Features struct {
	Backups string `json:"backups" faker:"oneof:0,1"`
}

// Image represents an image available for a cloud instance.
type Image struct {
	Name         string `json:"name"`
	Distribution string `json:"distribution"`
	Version      string `json:"version" faker:"semver"`
	Image        string `json:"image"`
	Cost         string `json:"cost"`
}

// Networks represents the network configuration for a cloud instance.
type Networks struct {
	Public  Public  `json:"public"`
	Private Private `json:"private"`
}

// Public represents the public network configuration for a cloud instance.
type Public struct {
	V4 V4PublicArray `json:"v4"`
}

// V4Public represents a public IPv4 address for a cloud instance.
type V4Public struct {
	IPAddress string `json:"ip_address,omitempty" faker:"ipv4"`
	Netmask   string `json:"netmask,omitempty" faker:"ipv4_netmask"`
	Gateway   string `json:"gateway,omitempty" faker:"ipv4"`
	Type      string `json:"type,omitempty" faker:"oneof:public"`
	Nat       bool   `json:"nat,omitempty" faker:"bool"`
	Primary   string `json:"primary,omitempty" faker:"oneof:1,0"`
	Rdns      string `json:"rdns,omitempty" faker:"domain_name"`
	Mac       string `json:"mac,omitempty"`
}

// Private represents the private network configuration for a cloud instance.
type Private struct {
	V4 []V4Private `json:"v4"`
}

// V4Private represents a private IPv4 address for a cloud instance.
type V4Private struct {
	Noip        int    `json:"noip"`
	IPAddress   string `json:"ip_address,omitempty" faker:"ipv4"`
	NatPublicIP string `json:"nat_publicip,omitempty" faker:"ipv4"`
	VpcName     string `json:"vpc_name,omitempty"`
	Network     string `json:"network,omitempty" faker:"ipv4"`
	VpcID       string `json:"vpc_id,omitempty" faker:"uuid_digit"`
	Netmask     string `json:"netmask,omitempty" faker:"ipv4_netmask"`
	Gateway     string `json:"gateway,omitempty" faker:"ipv4"`
	Type        string `json:"type,omitempty" faker:"oneof:private"`
	Mac         string `json:"mac,omitempty"`
	Primary     string `json:"primary,omitempty" faker:"oneof:1,0"`
}

// Storages represents a collection of storage devices attached to a cloud instance.
type Storages struct {
	ID        string `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Size      int    `json:"size"`
	DiskUsed  string `json:"disk_used" faker:"oneof:1GB,50GB,100GB"`
	DiskFree  string `json:"disk_free" faker:"oneof:1GB,50GB,100GB"`
	DiskUsedp string `json:"disk_usedp"`
	CreatedAt string `json:"created_at" faker:"date"`
	Bus       string `json:"bus" faker:"oneof:virtio,sata"`
	Type      string `json:"type" faker:"oneof:ssd,hdd"`
}

// Storage represents a single storage device attached to a cloud instance.
type Storage struct {
	ID        string `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Size      int    `json:"size"`
	DiskUsed  string `json:"disk_used" faker:"oneof:1GB,50GB,100GB"`
	DiskFree  string `json:"disk_free" faker:"oneof:1GB,50GB,100GB"`
	DiskUsedp string `json:"disk_usedp"`
	CreatedAt string `json:"created_at" faker:"date"`
	Bus       string `json:"bus" faker:"oneof:virtio,sata"`
	Type      string `json:"type" faker:"oneof:ssd,hdd"`
}

// Snapshot represents a snapshot of a cloud instance's storage.
type Snapshot struct {
	ID        string `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Size      string `json:"size" faker:"oneof:1GB,50GB,100GB"`
	CreatedAt string `json:"created_at" faker:"date"`
	Note      string `json:"note" faker:"sentence"`
	Name      string `json:"name"`
}

// CloudInstanceFirewall represents a firewall attached to a cloud instance.
type CloudInstanceFirewall struct {
	ID        string `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" faker:"date"`
}

// Meta contains metadata about a list of resources.
type Meta struct {
	Total       int `json:"total"`
	Totalpages  int `json:"totalpages"`
	Currentpage int `json:"currentpage"`
}

// Snapshots represents a collection of snapshots for a cloud instance.
type Snapshots struct {
	ID        string `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Size      string `json:"size" faker:"oneof:1GB,50GB,100GB"`
	CreatedAt string `json:"created_at" faker:"date"`
	Note      string `json:"note" faker:"sentence"`
	Name      string `json:"name"`
}

// CloudInstanceFirewalls represents a collection of firewalls attached to a cloud instance.
type CloudInstanceFirewalls struct {
	ID        string `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" faker:"date"`
}

// OsImages represents a list of OS images available for cloud instances.
type OsImages struct {
	OsImages []OsImage `json:"images"`
	Status   string    `json:"status,omitempty" faker:"oneof:success,error"`
	Message  string    `json:"message,omitempty" faker:"sentence"`
}

// OsImage represents a single OS image available for cloud instances.
type OsImage struct {
	Distro       string  `json:"distro,omitempty"`
	Distribution string  `json:"distribution"`
	Version      string  `json:"version" faker:"semver"`
	Image        string  `json:"image"`
	Cost         float64 `json:"cost"`
}

// Plans represents a list of available plans for cloud instances.
type Plans struct {
	Plans   []Plan `json:"plans"`
	Status  string `json:"status,omitempty" faker:"oneof:success,error"`
	Message string `json:"message,omitempty" faker:"sentence"`
}

// Plan represents a single plan for a cloud instance.
type Plan struct {
	ID        string  `json:"id" faker:"oneof: 00000,11111,22222,33333"`
	Type      string  `json:"type" faker:"oneof:ramcpu,disk"`
	Disk      string  `json:"disk" faker:"oneof:20GB,50GB,100GB"`
	RAM       string  `json:"ram" faker:"oneof:1024,2048,4096,8192,16384"`
	CPU       string  `json:"cpu" faker:"oneof:1,2,4,8,16"`
	Bandwidth string  `json:"bandwidth" faker:"oneof:100,500,1000,2000"`
	Slug      string  `json:"slug"`
	Price     float64 `json:"price"`
	Monthly   float64 `json:"monthly"`
	Plantype  string  `json:"plantype" faker:"oneof:cloud,dedicated"`
}

// CreateCloudInstanceParams contains parameters for creating a cloud instance.
type CreateCloudInstanceParams struct {
	Dcslug         string          `json:"dcslug"`
	Image          string          `json:"image"`
	Planid         string          `json:"planid"`
	VpcId          string          `json:"vpc"`
	EnablePublicip string          `json:"enable_publicip"`
	SubnetRequired string          `json:"subnetRequired"`
	Cpumodel       string          `json:"cpumodel"`
	Auth           string          `json:"auth,omitempty"`
	RootPassword   string          `json:"root_password,omitempty"`
	Firewall       string          `json:"firewall"`
	Enablebackup   string          `json:"enablebackup,omitempty"`
	Support        string          `json:"support,omitempty"`
	Management     string          `json:"management,omitempty"`
	Billingcycle   string          `json:"billingcycle,omitempty"`
	Backupid       string          `json:"backupid,omitempty"`
	Snapshotid     string          `json:"snapshotid,omitempty"`
	Sshkeys        string          `json:"sshkeys,omitempty"`
	Cloud          []CloudHostname `json:"cloud"`
}

// CloudHostname represents the hostname for a cloud instance.
type CloudHostname struct {
	Hostname string `json:"hostname"`
}

// CreateCloudInstanceResponse represents the response after creating a cloud instance.
type CreateCloudInstanceResponse struct {
	ID       string `json:"cloudid" faker:"oneof: 00000,11111,22222,33333"`
	Password string `json:"password" faker:"password"`
	Ipv4     string `json:"ipv4" faker:"ipv4"`
	Status   string `json:"status,omitempty" faker:"oneof:success,error"`
	Message  string `json:"message,omitempty" faker:"sentence"`
}

// Create creates a new cloud instance with the provided parameters.
func (s *CloudInstancesService) Create(params CreateCloudInstanceParams) (*CreateCloudInstanceResponse, error) {
	reqUrl := "cloud/deploy"
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var cloudInstances CreateCloudInstanceResponse
	_, err := s.client.Do(req, &cloudInstances)
	if err != nil {
		return nil, err
	}
	if cloudInstances.Status != "success" && cloudInstances.Status != "" {
		return nil, errors.New(cloudInstances.Message)
	}

	return &cloudInstances, nil
}

// Read retrieves information about a specific cloud instance by its ID.
func (s *CloudInstancesService) Read(instanceId string) (*CloudInstance, error) {
	reqUrl := "cloud/" + instanceId
	req, _ := s.client.NewRequest("GET", reqUrl)

	var cloudInstances CloudInstances
	_, err := s.client.Do(req, &cloudInstances)
	if err != nil {
		return nil, err
	}
	if cloudInstances.Status != "success" && cloudInstances.Status != "" {
		return nil, errors.New(cloudInstances.Message)
	}
	if len(cloudInstances.CloudInstance) == 0 {
		return nil, errors.New("NotFound")
	}

	return &cloudInstances.CloudInstance[0], nil
}

// List retrieves all cloud instances.
func (s *CloudInstancesService) List() ([]CloudInstance, error) {
	reqUrl := "cloud"
	req, _ := s.client.NewRequest("GET", reqUrl)

	var cloudInstances CloudInstances
	_, err := s.client.Do(req, &cloudInstances)
	if err != nil {
		return nil, err
	}
	if cloudInstances.Status != "success" && cloudInstances.Status != "" {
		return nil, errors.New(cloudInstances.Message)
	}

	return cloudInstances.CloudInstance, nil
}

// DeleteCloudInstanceParams contains parameters for deleting a cloud instance.
type DeleteCloudInstanceParams struct {
	// Please provide confirm string as follow: "I am aware this action will delete data and server permanently"
	Confirm string `json:"confirm"`
}

// Delete deletes a cloud instance specified by the given ID and parameters.
func (s *CloudInstancesService) Delete(cloudInstancesId string, deleteCloudInstanceParams DeleteCloudInstanceParams) (*DeleteResponse, error) {
	reqUrl := "cloud/" + cloudInstancesId + "/destroy"

	req, _ := s.client.NewRequest("DELETE", reqUrl, deleteCloudInstanceParams)

	var delResponse DeleteResponse
	if _, err := s.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}

// ListOsImages retrieves all OS images available for cloud instances.
func (s *CloudInstancesService) ListOsImages() ([]OsImage, error) {
	reqUrl := "cloud/images"
	req, _ := s.client.NewRequest("GET", reqUrl)

	var osImages OsImages
	_, err := s.client.Do(req, &osImages)
	if err != nil {
		return nil, err
	}
	if osImages.Status != "success" && osImages.Status != "" {
		return nil, errors.New(osImages.Message)
	}

	return osImages.OsImages, nil
}

// ListResizePlans retrieves all available resize plans for a specific cloud instance.
func (s *CloudInstancesService) ListResizePlans(instanceId string) ([]Plan, error) {
	reqUrl := "cloud/" + instanceId + "/resizeplans"
	req, _ := s.client.NewRequest("GET", reqUrl)

	var plans Plans
	_, err := s.client.Do(req, &plans)
	if err != nil {
		return nil, err
	}
	if plans.Status != "success" && plans.Status != "" {
		return nil, errors.New(plans.Message)
	}

	return plans.Plans, nil
}

// CreateSnapshotParams contains parameters for creating a snapshot.
type CreateSnapshotParams struct {
	Name string `json:"name"`
}

// CreateSnapshot creates a snapshot for the specified cloud instance.
func (s *CloudInstancesService) CreateSnapshot(instanceId string, params CreateSnapshotParams) (*CreateBasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/snapshot/create"
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var snapshot CreateBasicResponse
	_, err := s.client.Do(req, &snapshot)
	if err != nil {
		return nil, err
	}
	if snapshot.Status != "success" && snapshot.Status != "" {
		return nil, errors.New(snapshot.Message)
	}

	return &snapshot, nil
}

// DeleteSnapshot deletes a snapshot from the specified cloud instance.
func (s *CloudInstancesService) DeleteSnapshot(cloudInstanceId, snapshotId string) (*DeleteResponse, error) {
	reqUrl := "cloud/" + cloudInstanceId + "/snapshot/" + snapshotId + "/delete"
	req, _ := s.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := s.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}

// EnableBackup enables backups for the specified cloud instance.
func (s *CloudInstancesService) EnableBackup(instanceId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/backups/enable"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// DisableBackup disables backups for the specified cloud instance.
func (s *CloudInstancesService) DisableBackup(instanceId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/backups/disable"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// UpdateBillingCycleParams contains parameters for updating the billing cycle of a cloud instance.
type UpdateBillingCycleParams struct {
	Billingcycle string `json:"billingcycle"`
}

// UpdateBillingCycle updates the billing cycle for the specified cloud instance.
func (s *CloudInstancesService) UpdateBillingCycle(cloudid string, params UpdateBillingCycleParams) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/billingcycle"
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// HardReboot performs a hard reboot on the specified cloud instance.
func (s *CloudInstancesService) HardReboot(instanceId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/hardreboot"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// PowerCycle performs a power cycle on the specified cloud instance.
func (s *CloudInstancesService) PowerCycle(instanceId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/powercycle"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// PowerOff powers off the specified cloud instance.
func (s *CloudInstancesService) PowerOff(instanceId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/poweroff"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// PowerOn powers on the specified cloud instance.
func (s *CloudInstancesService) PowerOn(instanceId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/poweron"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// RebuildCloudInstanceParams contains parameters for rebuilding a cloud instance.
type RebuildCloudInstanceParams struct {
	Image string `json:"image"`
	// Please provide confirm string as follow: "I am aware this action will delete data permanently and build a fresh server"
	Confirm string `json:"confirm"`
}

// Rebuild rebuilds the specified cloud instance with a new image.
func (s *CloudInstancesService) Rebuild(instanceId string, rebuildCloudInstanceParams RebuildCloudInstanceParams) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/rebuild"
	req, _ := s.client.NewRequest("POST", reqUrl, rebuildCloudInstanceParams)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// ResetPasswordResponse represents the response after resetting a cloud instance password.
type ResetPasswordResponse struct {
	Password string `json:"password" faker:"password"`
	Status   string `json:"status,omitempty" faker:"oneof:success,error"`
	Message  string `json:"message,omitempty" faker:"sentence"`
}

// ResetPassword resets the password for the specified cloud instance.
func (s *CloudInstancesService) ResetPassword(instanceId string) (*ResetPasswordResponse, error) {
	reqUrl := "cloud/" + instanceId + "/resetpassword"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var resetPasswordResponse ResetPasswordResponse
	_, err := s.client.Do(req, &resetPasswordResponse)
	if err != nil {
		return nil, err
	}
	if resetPasswordResponse.Status != "success" && resetPasswordResponse.Status != "" {
		return nil, errors.New(resetPasswordResponse.Message)
	}

	return &resetPasswordResponse, nil
}

// ResizeCloudInstanceParams contains parameters for resizing a cloud instance.
type ResizeCloudInstanceParams struct {
	Type string `json:"type"`
	Plan string `json:"plan"`
}

// Resize resizes the specified cloud instance with the given parameters.
func (s *CloudInstancesService) Resize(instanceId string, resizeCloudInstanceParams ResizeCloudInstanceParams) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/resize"
	req, _ := s.client.NewRequest("POST", reqUrl, resizeCloudInstanceParams)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// RestoreSnapshot restores a snapshot for the specified cloud instance.
func (s *CloudInstancesService) RestoreSnapshot(instanceId, snapshotId string) (*BasicResponse, error) {
	reqUrl := "cloud/" + instanceId + "/snapshot/" + snapshotId + "/restore"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// UpdateStorageParams contains parameters for updating a storage device attached to a cloud instance.
type UpdateStorageParams struct {
	Bus  string `json:"bus"`
	Type string `json:"type"`
}

// UpdateStorage updates the storage device attached to the specified cloud instance.
func (s *CloudInstancesService) UpdateStorage(cloudid, storageid string, params UpdateStorageParams) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/storage/" + storageid + "/update"
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// AssignPublicIP assigns a public IP to the specified cloud instance.
func (s *CloudInstancesService) AssignPublicIP(cloudid string) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/assignpublicip"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// UpdateRDNSParams contains parameters for updating the RDNS of a cloud instance.
type UpdateRDNSParams struct {
	Rdns string `json:"rdns"`
}

// UpdateRDNS updates the RDNS for the specified IP address of a cloud instance.
func (s *CloudInstancesService) UpdateRDNS(cloudId, ipAddress string, params UpdateRDNSParams) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudId + "/updaterdns/" + ipAddress
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// EnableRescue enables rescue mode for the specified cloud instance.
func (s *CloudInstancesService) EnableRescue(cloudid string) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/enablerescue"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// DisableRescue disables rescue mode for the specified cloud instance.
func (s *CloudInstancesService) DisableRescue(cloudid string) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/disablerescue"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// MountISOParams contains parameters for mounting an ISO to a cloud instance.
type MountISOParams struct {
	Iso string `json:"iso"`
}

// MountISO mounts an ISO to the specified cloud instance.
func (s *CloudInstancesService) MountISO(cloudid string, params MountISOParams) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/mountiso"
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// UnmountISO unmounts an ISO from the specified cloud instance.
func (s *CloudInstancesService) UnmountISO(cloudid string) (*BasicResponse, error) {
	reqUrl := "cloud/" + cloudid + "/umountiso"
	req, _ := s.client.NewRequest("POST", reqUrl, nil)

	var basicResponse BasicResponse
	_, err := s.client.Do(req, &basicResponse)
	if err != nil {
		return nil, err
	}
	if basicResponse.Status != "success" && basicResponse.Status != "" {
		return nil, errors.New(basicResponse.Message)
	}

	return &basicResponse, nil
}

// Custom type to handle unmarshaling of V4Public
// V4PublicArray is a custom type to handle unmarshaling of V4Public.
type V4PublicArray []V4Public

// UnmarshalJSON implements the json.Unmarshaler interface for V4PublicArray.
func (v *V4PublicArray) UnmarshalJSON(data []byte) error {
	var single []V4Public
	var nested [][]V4Public

	// Try unmarshaling as a single array
	if err := json.Unmarshal(data, &single); err == nil {
		*v = single
		return nil
	}

	// Try unmarshaling as a nested array
	if err := json.Unmarshal(data, &nested); err == nil {
		for _, inner := range nested {
			*v = append(*v, inner...)
		}
		return nil
	}

	return errors.New("invalid format for V4PublicArray")
}
