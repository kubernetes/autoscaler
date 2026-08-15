/*
Copyright 2021-2024 Oracle and/or its affiliates.
*/

package common

import (
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
)

// TagsGetter returns the oci tags for the pool.
type TagsGetter interface {
	GetNodePoolFreeformTags(*oke.NodePool) (map[string]string, error)
}

// TagsGetterImpl is the implementation to fetch the oci tags for the pool.
type TagsGetterImpl struct{}

// CreateTagsGetter creates a new oci tags getter.
func CreateTagsGetter() TagsGetter {
	return &TagsGetterImpl{}
}

// GetNodePoolFreeformTags returns the FreeformTags for the nodepool
func (tgi *TagsGetterImpl) GetNodePoolFreeformTags(np *oke.NodePool) (map[string]string, error) {
	return np.FreeformTags, nil
}
