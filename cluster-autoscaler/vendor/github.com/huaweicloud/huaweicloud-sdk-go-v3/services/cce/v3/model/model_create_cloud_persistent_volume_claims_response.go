package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateCloudPersistentVolumeClaimsResponse struct {
	// API版本，固定值**v1**

	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值**PersistentVolumeClaim**

	Kind *string `json:"kind,omitempty"`

	Metadata *PersistentVolumeClaimMetadata `json:"metadata,omitempty"`

	Spec *PersistentVolumeClaimSpec `json:"spec,omitempty"`

	Status         *PersistentVolumeClaimStatus `json:"status,omitempty"`
	HttpStatusCode int                          `json:"-"`
}

func (o CreateCloudPersistentVolumeClaimsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateCloudPersistentVolumeClaimsResponse struct{}"
	}

	return strings.Join([]string{"CreateCloudPersistentVolumeClaimsResponse", string(data)}, " ")
}
