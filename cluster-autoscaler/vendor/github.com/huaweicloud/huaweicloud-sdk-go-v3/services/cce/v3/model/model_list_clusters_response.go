package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListClustersResponse struct {
	// Api type

	Kind *string `json:"kind,omitempty"`
	// API version

	ApiVersion *string `json:"apiVersion,omitempty"`
	// 集群对象列表，包含了当前项目下所有集群的详细信息。您可通过items.metadata.name下的值来找到对应的集群。

	Items          *[]Cluster `json:"items,omitempty"`
	HttpStatusCode int        `json:"-"`
}

func (o ListClustersResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListClustersResponse struct{}"
	}

	return strings.Join([]string{"ListClustersResponse", string(data)}, " ")
}
