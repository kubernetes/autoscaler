package nodegroups

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/pagination"
)

type commonResult struct {
	gophercloud.Result
}

func (r commonResult) Extract() (*NodeGroup, error) {
	var s NodeGroup
	err := r.ExtractInto(&s)
	return &s, err
}

// GetResult is the response from a Get request.
// Use the Extract method to retrieve the NodeGroup itself.
type GetResult struct {
	commonResult
}

// CreateResult is the response from a Create request.
// Use the Extract method to retrieve the created node group.
type CreateResult struct {
	commonResult
}

// UpdateResult is the response from an Update request.
// Use the Extract method to retrieve the updated node group.
type UpdateResult struct {
	commonResult
}

// DeleteResult is the response from a Delete request.
// Use the ExtractErr method to extract the error from the result.
type DeleteResult struct {
	gophercloud.ErrResult
}

// NodeGroup is the API representation of a Magnum node group.
type NodeGroup struct {
	ID               int                `json:"id"`
	UUID             string             `json:"uuid"`
	Name             string             `json:"name"`
	ClusterID        string             `json:"cluster_id"`
	ProjectID        string             `json:"project_id"`
	DockerVolumeSize *int               `json:"docker_volume_size"`
	Labels           map[string]string  `json:"labels"`
	Links            []gophercloud.Link `json:"links"`
	FlavorID         string             `json:"flavor_id"`
	ImageID          string             `json:"image_id"`
	NodeAddresses    []string           `json:"node_addresses"`
	NodeCount        int                `json:"node_count"`
	Role             string             `json:"role"`
	MinNodeCount     int                `json:"min_node_count"`
	MaxNodeCount     *int               `json:"max_node_count"`
	IsDefault        bool               `json:"is_default"`
	StackID          string             `json:"stack_id"`
	Status           string             `json:"status"`
	StatusReason     string             `json:"status_reason"`
	Version          string             `json:"version"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type NodeGroupPage struct {
	pagination.LinkedPageBase
}

func (r NodeGroupPage) NextPageURL() (string, error) {
	var s struct {
		Next string `json:"next"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return s.Next, nil
}

func (r NodeGroupPage) IsEmpty() (bool, error) {
	s, err := ExtractNodeGroups(r)
	return len(s) == 0, err
}

// ExtractNodeGroups takes a Page of node groups as returned from List
// or from AllPages and extracts it as a slice of NodeGroups.
func ExtractNodeGroups(r pagination.Page) ([]NodeGroup, error) {
	var s struct {
		NodeGroups []NodeGroup `json:"nodegroups"`
	}
	err := (r.(NodeGroupPage)).ExtractInto(&s)
	return s.NodeGroups, err
}
