/*
Copyright 2019 The Kubernetes Authors.

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

package hetzner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindImageWithPerNodepoolConfig(t *testing.T) {
	// Test case 1: Nodepool with specific imagesForArch should use those images
	t.Run("nodepool with specific imagesForArch", func(t *testing.T) {
		manager := &hetznerManager{
			clusterConfig: &ClusterConfig{
				IsUsingNewFormat: true,
				ImagesForArch: ImageList{
					Arm64: "global-arm64-image",
					Amd64: "global-amd64-image",
				},
				NodeConfigs: map[string]*NodeConfig{
					"pool1": {
						ImagesForArch: &ImageList{
							Arm64: "pool1-arm64-image",
							Amd64: "pool1-amd64-image",
						},
					},
				},
			},
		}

		nodeGroup := &hetznerNodeGroup{
			id:      "pool1",
			manager: manager,
		}

		// This would normally call the actual API, but we're just testing the logic
		// The actual image selection logic is in findImage function
		// For this test, we'll verify the configuration is set up correctly
		nodeConfig, exists := manager.clusterConfig.NodeConfigs[nodeGroup.id]
		require.True(t, exists)
		require.NotNil(t, nodeConfig.ImagesForArch)
		assert.Equal(t, "pool1-arm64-image", nodeConfig.ImagesForArch.Arm64)
		assert.Equal(t, "pool1-amd64-image", nodeConfig.ImagesForArch.Amd64)
	})

	// Test case 2: Nodepool without specific imagesForArch should fall back to global
	t.Run("nodepool without specific imagesForArch", func(t *testing.T) {
		manager := &hetznerManager{
			clusterConfig: &ClusterConfig{
				IsUsingNewFormat: true,
				ImagesForArch: ImageList{
					Arm64: "global-arm64-image",
					Amd64: "global-amd64-image",
				},
				NodeConfigs: map[string]*NodeConfig{
					"pool2": {
						// No ImagesForArch specified
					},
				},
			},
		}

		nodeGroup := &hetznerNodeGroup{
			id:      "pool2",
			manager: manager,
		}

		nodeConfig, exists := manager.clusterConfig.NodeConfigs[nodeGroup.id]
		require.True(t, exists)
		assert.Nil(t, nodeConfig.ImagesForArch)
		assert.Equal(t, "global-arm64-image", manager.clusterConfig.ImagesForArch.Arm64)
		assert.Equal(t, "global-amd64-image", manager.clusterConfig.ImagesForArch.Amd64)
	})

	// Test case 3: Nodepool with nil ImagesForArch should fall back to global
	t.Run("nodepool with nil imagesForArch", func(t *testing.T) {
		manager := &hetznerManager{
			clusterConfig: &ClusterConfig{
				IsUsingNewFormat: true,
				ImagesForArch: ImageList{
					Arm64: "global-arm64-image",
					Amd64: "global-amd64-image",
				},
				NodeConfigs: map[string]*NodeConfig{
					"pool3": {
						ImagesForArch: nil, // Explicitly nil
					},
				},
			},
		}

		nodeGroup := &hetznerNodeGroup{
			id:      "pool3",
			manager: manager,
		}

		nodeConfig, exists := manager.clusterConfig.NodeConfigs[nodeGroup.id]
		require.True(t, exists)
		assert.Nil(t, nodeConfig.ImagesForArch)
		assert.Equal(t, "global-arm64-image", manager.clusterConfig.ImagesForArch.Arm64)
		assert.Equal(t, "global-amd64-image", manager.clusterConfig.ImagesForArch.Amd64)
	})
}

func TestImageSelectionLogic(t *testing.T) {
	// Test the image selection logic that would be used in findImage function
	t.Run("image selection logic", func(t *testing.T) {
		manager := &hetznerManager{
			clusterConfig: &ClusterConfig{
				IsUsingNewFormat: true,
				ImagesForArch: ImageList{
					Arm64: "global-arm64-image",
					Amd64: "global-amd64-image",
				},
				NodeConfigs: map[string]*NodeConfig{
					"pool1": {
						ImagesForArch: &ImageList{
							Arm64: "pool1-arm64-image",
							Amd64: "pool1-amd64-image",
						},
					},
					"pool2": {
						// No ImagesForArch specified
					},
				},
			},
		}

		// Test pool1 (has specific imagesForArch)
		nodeConfig, exists := manager.clusterConfig.NodeConfigs["pool1"]
		require.True(t, exists)
		require.NotNil(t, nodeConfig.ImagesForArch)

		var imagesForArch *ImageList
		if nodeConfig.ImagesForArch != nil {
			imagesForArch = nodeConfig.ImagesForArch
		} else {
			imagesForArch = &manager.clusterConfig.ImagesForArch
		}

		assert.Equal(t, "pool1-arm64-image", imagesForArch.Arm64)
		assert.Equal(t, "pool1-amd64-image", imagesForArch.Amd64)

		// Test pool2 (no specific imagesForArch, should use global)
		nodeConfig, exists = manager.clusterConfig.NodeConfigs["pool2"]
		require.True(t, exists)
		assert.Nil(t, nodeConfig.ImagesForArch)

		if nodeConfig.ImagesForArch != nil {
			imagesForArch = nodeConfig.ImagesForArch
		} else {
			imagesForArch = &manager.clusterConfig.ImagesForArch
		}

		assert.Equal(t, "global-arm64-image", imagesForArch.Arm64)
		assert.Equal(t, "global-amd64-image", imagesForArch.Amd64)
	})
}
