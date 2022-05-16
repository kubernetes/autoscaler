// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	serverBasePath   = "/servers"
	flavorPath       = "/flavors"
	osImagePath      = "/images"
	taskPath         = "/tasks"
	customImagesPath = "/user/images"
	customImagePath  = "/user/image"
)

var _ ServerService = (*server)(nil)

// ServerSecurityGroup contains information of security group of a server.
type ServerSecurityGroup struct {
	Name string `json:"name"`
}

// AttachedVolume contains attached volumes of a server.
type AttachedVolume struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Size         int    `json:"size"`
	AttachedType string `json:"attached_type"`
	Type         string `json:"type"`
	Category     string `json:"category"`
}

// IP represents the IP address, version and mac address of a port
type IP struct {
	Version    int    `json:"version"`
	Address    string `json:"addr"`
	Type       string `json:"OS-EXT-IPS:type"`
	MacAddress string `json:"0S-EXT-IPS-MAC:mac_addr"`
}

// IPAddresses contains LAN & WAN Ip address of a Cloud Server
type IPAddress struct {
	LanAddresses   []IP `json:"LAN"`
	WanV4Addresses []IP `json:"WAN_V4"`
	WanV6Addresses []IP `json:"WAN_V6"`
}

// Flavor contains flavor information.
type Flavor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Ram  int    `json:"ram"`
	VCPU int    `json:"vcpu"`
}

// Server contains server information.
type Server struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	KeyName          string                 `json:"key_name"`
	UserID           string                 `json:"user_id"`
	ProjectID        string                 `json:"tenant_id"`
	CreatedAt        string                 `json:"created"`
	UpdatedAt        string                 `json:"updated"`
	Status           string                 `json:"status"`
	IPv6             bool                   `json:"ipv6"`
	SecurityGroup    []ServerSecurityGroup  `json:"security_group"`
	Addresses        map[string]interface{} `json:"addresses"` // Deprecated: This field will be removed in the near future
	Metadata         map[string]string      `json:"metadata"`
	Flavor           Flavor                 `json:"flavor"`
	Progress         int                    `json:"progress"`
	AttachedVolumes  []AttachedVolume       `json:"os-extended-volumes:volumes_attached"`
	AvailabilityZone string                 `json:"OS-EXT-AZ:availability_zone"`
	Category         string                 `json:"category"`
	IPAddresses      IPAddress              `json:"ip_addresses"`
	RegionName       string                 `json:"region_name"`
}

type CreateCustomImagePayload struct {
	Name        string `json:"name"`
	DiskFormat  string `json:"disk_format"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

type Location struct {
	URL      string            `json:"url"`
	Metadata map[string]string `json:"metadata,"`
}

type CustomImage struct {
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	DiskFormat      string     `json:"disk_format"`
	ContainerFormat string     `json:"container_format"`
	Visibility      string     `json:"visibility"`
	Size            int        `json:"size"`
	VirtualSize     int        `json:"virtual_size"`
	Status          string     `json:"status"`
	Checksum        string     `json:"checksum"`
	Protected       bool       `json:"protected"`
	MinRam          int        `json:"min_ram"`
	MinDisk         int        `json:"min_disk"`
	Owner           string     `json:"owner"`
	OSHidden        bool       `json:"os_hidden"`
	OSHashAlgo      string     `json:"os_hash_algo"`
	OSHashValue     string     `json:"os_hash_value"`
	ID              string     `json:"id"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
	Locations       []Location `json:"locations"`
	DirectURL       string     `json:"direct_url"`
	Tags            []string   `json:"tags"`
	File            string     `json:"file"`
	Schema          string     `json:"schema"`
}

type CreateCustomImageResp struct {
	Image     CustomImage `json:"image"`
	Success   bool        `json:"success"`
	Token     string      `json:"token,omitempty"`
	UploadURI string      `json:"upload_uri,omitempty"`
}

type CustomImageGetResp struct {
	Image CustomImage `json:"image"`
	Token string      `json:"token"`
}

type server struct {
	client *Client
}

// ServerService is an interface to interact with BizFly Cloud API.
type ServerService interface {
	List(ctx context.Context, opts *ListOptions) ([]*Server, error)
	Create(ctx context.Context, scr *ServerCreateRequest) (*ServerCreateResponse, error)
	Get(ctx context.Context, id string) (*Server, error)
	Delete(ctx context.Context, id string) error
	Resize(ctx context.Context, id string, newFlavor string) (*ServerTask, error)
	Start(ctx context.Context, id string) (*Server, error)
	Stop(ctx context.Context, id string) (*Server, error)
	SoftReboot(ctx context.Context, id string) (*ServerMessageResponse, error)
	HardReboot(ctx context.Context, id string) (*ServerMessageResponse, error)
	Rebuild(ctx context.Context, id string, imageID string) (*ServerTask, error)
	GetVNC(ctx context.Context, id string) (*ServerConsoleResponse, error)
	ListFlavors(ctx context.Context) ([]*serverFlavorResponse, error)
	ListOSImages(ctx context.Context) ([]osImageResponse, error)
	GetTask(ctx context.Context, id string) (*ServerTaskResponse, error)
	ChangeCategory(ctx context.Context, id string, newCategory string) (*ServerTask, error)
	AddVPC(ctx context.Context, id string, vpcs []string) (*Server, error)
	RemoveVPC(ctx context.Context, id string, vpcs []string) (*Server, error)
	ListCustomImages(ctx context.Context) ([]*CustomImage, error)
	CreateCustomImage(ctx context.Context, cipl *CreateCustomImagePayload) (*CreateCustomImageResp, error)
	DeleteCustomImage(ctx context.Context, imageID string) error
	GetCustomImage(ctx context.Context, imageID string) (*CustomImageGetResp, error)
}

// ServerConsoleResponse contains information of server console url.
type ServerConsoleResponse struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

// ServerMessageResponse contains message response from Cloud Server API.
type ServerMessageResponse struct {
	Message string `json:"message"`
}

// ServerAction represents server action request payload.
type ServerAction struct {
	Action        string   `json:"action"`
	ImageID       string   `json:"image,omitempty"`
	FlavorName    string   `json:"flavor_name,omitempty"`
	ConsoleType   string   `json:"type,omitempty"`
	FirewallIDs   []string `json:"firewall_ids,omitempty"`
	NewType       string   `json:"new_type,omitempty"`
	VPCNetworkIDs []string `json:"vpc_network_ids,omitempty"`
}

// ServerTask contains task information.
type ServerTask struct {
	TaskID string `json:"task_id"`
}

// ServerCreateResponse contains response tasks when create server
type ServerCreateResponse struct {
	Task []string `json:"task_id"`
}

// ServerDisk contains server's disk information.
type ServerDisk struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

// ServerOS contains OS information of server.
type ServerOS struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ServerCreateRequest represents create a new server payload.
type ServerCreateRequest struct {
	Name             string        `json:"name"`
	FlavorName       string        `json:"flavor"`
	SSHKey           string        `json:"sshkey,omitempty"`
	Password         bool          `json:"password"`
	RootDisk         *ServerDisk   `json:"rootdisk"`
	DataDisks        []*ServerDisk `json:"datadisks,omitempty"`
	Type             string        `json:"type"`
	AvailabilityZone string        `json:"availability_zone"`
	OS               *ServerOS     `json:"os"`
	Quantity         int           `json:"quantity,omitempty"`
}

// itemActionPath return http path of server action
func (s *server) itemActionPath(id string) string {
	return strings.Join([]string{serverBasePath, id, "action"}, "/")
}

func (s *server) itemCustomImagePath(id string) string {
	return strings.Join([]string{customImagePath, id}, "/")
}

// List lists all servers.
func (s *server) List(ctx context.Context, opts *ListOptions) ([]*Server, error) {

	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, serverBasePath, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var servers []*Server

	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, err
	}

	return servers, nil
}

// Create creates a new server.
func (s *server) Create(ctx context.Context, scr *ServerCreateRequest) (*ServerCreateResponse, error) {
	payload := []*ServerCreateRequest{scr}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, serverBasePath, payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var task *ServerCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return task, nil
}

// Get gets a server.
func (s *server) Get(ctx context.Context, id string) (*Server, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, serverBasePath+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var svr *Server
	if err := json.NewDecoder(resp.Body).Decode(&svr); err != nil {
		return nil, err
	}
	return svr, nil
}

// Delete deletes a server.
func (s *server) Delete(ctx context.Context, id string) error {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, serverServiceName, serverBasePath+"/"+id, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

// Resize resizes a server.
func (s *server) Resize(ctx context.Context, id string, newFlavor string) (*ServerTask, error) {
	var payload = &ServerAction{
		Action:     "resize",
		FlavorName: newFlavor}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var task *ServerTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return task, nil
}

// Start starts a server.
func (s *server) Start(ctx context.Context, id string) (*Server, error) {
	payload := &ServerAction{Action: "start"}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var svr *Server
	if err := json.NewDecoder(resp.Body).Decode(&svr); err != nil {
		return nil, err
	}
	return svr, nil
}

// Stop stops a server
func (s *server) Stop(ctx context.Context, id string) (*Server, error) {
	payload := &ServerAction{Action: "stop"}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var svr *Server
	if err := json.NewDecoder(resp.Body).Decode(&svr); err != nil {
		return nil, err
	}
	return svr, nil
}

// SoftReboot soft reboots a server.
func (s *server) SoftReboot(ctx context.Context, id string) (*ServerMessageResponse, error) {
	payload := &ServerAction{Action: "soft_reboot"}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var srm *ServerMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&srm); err != nil {
		return nil, err
	}
	return srm, nil
}

// HardReboot hard reboots a server.
func (s *server) HardReboot(ctx context.Context, id string) (*ServerMessageResponse, error) {
	payload := &ServerAction{Action: "hard_reboot"}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var smr *ServerMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&smr); err != nil {
		return nil, err
	}
	return smr, nil
}

// AddVPC add VPC to the server

// Rebuild rebuilds a server.
func (s *server) Rebuild(ctx context.Context, id string, imageID string) (*ServerTask, error) {
	var payload = &ServerAction{
		Action:  "rebuild",
		ImageID: imageID}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var task *ServerTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return task, nil
}

// GetVNC gets vnc console of a server.
func (s *server) GetVNC(ctx context.Context, id string) (*ServerConsoleResponse, error) {
	payload := &ServerAction{
		Action:      "get_vnc",
		ConsoleType: "novnc"}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var respPayload struct {
		Console *ServerConsoleResponse `json:"console"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, err
	}
	return respPayload.Console, nil
}

type serverFlavorResponse struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// ListFlavors lists server flavors
func (s *server) ListFlavors(ctx context.Context) ([]*serverFlavorResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, flavorPath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	var flavors []*serverFlavorResponse

	if err := json.NewDecoder(resp.Body).Decode(&flavors); err != nil {
		return nil, err
	}
	return flavors, nil
}

type osDistributionVersion struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type osImageResponse struct {
	OSDistribution string                  `json:"os"`
	Version        []osDistributionVersion `json:"versions"`
}

// ListOSImage list server os images
func (s *server) ListOSImages(ctx context.Context) ([]osImageResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, osImagePath, nil)

	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("os_images", "True")
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	var respPayload struct {
		OSImages []osImageResponse `json:"os_images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, err
	}
	return respPayload.OSImages, nil
}

type ServerTaskResult struct {
	Action   string `json:"action"`
	Progress int    `json:"progress"`
	Success  bool   `json:"success"`
	Server
}

type ServerTaskResponse struct {
	Ready  bool             `json:"ready"`
	Result ServerTaskResult `json:"result"`
}

// GetTask get tasks result from Server API
func (s *server) GetTask(ctx context.Context, id string) (*ServerTaskResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, taskPath+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	var str *ServerTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&str); err != nil {
		return nil, err
	}
	return str, nil
}

func (s server) ChangeCategory(ctx context.Context, id string, newCategory string) (*ServerTask, error) {
	payload := &ServerAction{
		Action:  "change_type",
		NewType: newCategory}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var svt *ServerTask
	if err := json.NewDecoder(resp.Body).Decode(&svt); err != nil {
		return nil, err
	}
	return svt, nil
}

func (s server) AddVPC(ctx context.Context, id string, vpcs []string) (*Server, error) {
	payload := &ServerAction{
		Action:        "add_vpc",
		VPCNetworkIDs: vpcs,
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var server *Server
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, err
	}
	return server, nil
}

func (s server) RemoveVPC(ctx context.Context, id string, vpcs []string) (*Server, error) {
	payload := &ServerAction{
		Action:        "remove_vpc",
		VPCNetworkIDs: vpcs,
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, s.itemActionPath(id), payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var server *Server
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *server) ListCustomImages(ctx context.Context) ([]*CustomImage, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, customImagesPath, nil)
	if err != nil {
		return nil, err
	}
	var data struct {
		Images []*CustomImage `json:"images"`
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Images, nil
}

func (s *server) CreateCustomImage(ctx context.Context, cipl *CreateCustomImagePayload) (*CreateCustomImageResp, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, strings.Join([]string{customImagePath, "upload"}, "/"), cipl)
	if err != nil {
		return nil, err
	}
	var data *CreateCustomImageResp
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *server) DeleteCustomImage(ctx context.Context, imageID string) error {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, serverServiceName, s.itemCustomImagePath(imageID), nil)

	if err != nil {
		return err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (s *server) GetCustomImage(ctx context.Context, imageID string) (*CustomImageGetResp, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, strings.Join([]string{customImagePath, imageID}, "/"), nil)
	if err != nil {
		return nil, err
	}
	var data *CustomImageGetResp
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
