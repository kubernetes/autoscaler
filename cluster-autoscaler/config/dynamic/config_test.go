/*
Copyright 2020 The Kubernetes Authors.

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

package dynamic

import (
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"

	"github.com/stretchr/testify/assert"
)

var configNodeGroupsInvalid = `
    {
      "nodeGroups": "name"
    }
`

var configValid = `
{
	"nodeGroups": [
	  {
		"minSize": 1,
		"maxSize": 100,
		"name": "aks-nodepool1-24160808-vmss",
        "scaleDownPolicy": "Delete"
	  }
	]
  }
`

func TestBuildConfig(t *testing.T) {
	t.Run("test BuildConfig invalid", func(t *testing.T) {
		yamlReader := strings.NewReader(configNodeGroupsInvalid)
		config, err := BuildConfig(yamlReader)
		assert.Error(t, err)
		assert.Empty(t, config)
	})

	t.Run("test BuildConfig valid", func(t *testing.T) {
		yamlReader := strings.NewReader(configValid)
		config, err := BuildConfig(yamlReader)
		assert.NoError(t, err)
		assert.Equal(t, "aks-nodepool1-24160808-vmss", config.NodeGroups[0].Name)
		assert.Equal(t, 1, config.NodeGroups[0].MinSize)
		assert.Equal(t, 100, config.NodeGroups[0].MaxSize)
		assert.Equal(t, deallocate.Delete, config.NodeGroups[0].ScaleDownPolicy)
	})
}

func TestUnmarshalConfig(t *testing.T) {
	t.Run("test UnmarshalConfig nodeGroups invalid", func(t *testing.T) {
		yamlReader := strings.NewReader(configNodeGroupsInvalid)
		config, err := umarshalConfig(yamlReader)
		assert.Error(t, err)
		assert.Empty(t, config)
	})

	t.Run("test UnmarshalConfig valid", func(t *testing.T) {
		yamlReader := strings.NewReader(configValid)
		config, err := umarshalConfig(yamlReader)
		assert.NoError(t, err)
		assert.Equal(t, "aks-nodepool1-24160808-vmss", config.NodeGroups[0].Name)
		assert.Equal(t, 1, config.NodeGroups[0].MinSize)
		assert.Equal(t, 100, config.NodeGroups[0].MaxSize)
		assert.Equal(t, deallocate.Delete, config.NodeGroups[0].ScaleDownPolicy)

	})
}
func TestValidate(t *testing.T) {
	t.Run("test validate valid", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:            "test1",
					MinSize:         1,
					MaxSize:         50,
					ScaleDownPolicy: deallocate.Delete,
				},
				{
					Name:            "test2",
					MinSize:         10,
					MaxSize:         70,
					ScaleDownPolicy: deallocate.Deallocate,
				},
			},
		}
		err := c.validate()
		assert.NoError(t, err)
	})

	t.Run("test validate empty name", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:    "",
					MinSize: 1,
					MaxSize: 50,
				},
				{
					Name:    "test2",
					MinSize: 10,
					MaxSize: 70,
				},
			},
		}
		err := c.validate()
		assert.EqualError(t, err, "invalid nodeGroup: name must not be blank")
	})

	t.Run("test validate wrong min/max", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:    "test1",
					MinSize: 100,
					MaxSize: 50,
				},
				{
					Name:    "test2",
					MinSize: 10,
					MaxSize: 70,
				},
			},
		}
		err := c.validate()
		assert.EqualError(t, err, "invalid nodeGroup: test1, max size must be greater or equal to min size")
	})

	t.Run("test validate invalid scaleDownPolicy", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:            "test1",
					MinSize:         1,
					MaxSize:         50,
					ScaleDownPolicy: "InvalidPolicy",
				},
				{
					Name:    "test2",
					MinSize: 10,
					MaxSize: 70,
				},
			},
		}
		err := c.validate()
		assert.EqualError(t, err, "invalid scaledown policy: InvalidPolicy. Valid values are: Delete, Deallocate")
	})
}

func TestNodeGroupSpecStrings(t *testing.T) {
	t.Run("no labels or taints", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:            "test1",
					MinSize:         1,
					MaxSize:         50,
					ScaleDownPolicy: deallocate.Delete,
				},
				{
					Name:            "test2",
					MinSize:         10,
					MaxSize:         70,
					ScaleDownPolicy: deallocate.Deallocate,
				},
			},
		}
		specStrings := c.NodeGroupSpecStrings()
		assert.Equal(t, []string{"1:50:Delete:test1:{}|", "10:70:Deallocate:test2:{}|"}, specStrings)
	})

	t.Run("only labels no taints", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:            "test1",
					MinSize:         1,
					MaxSize:         50,
					ScaleDownPolicy: deallocate.Deallocate,
				},
				{
					Name:            "test2",
					MinSize:         10,
					MaxSize:         70,
					ScaleDownPolicy: deallocate.Deallocate,
					Labels: map[string]string{
						"environment": "prod",
					},
				},
			},
		}
		specStrings := c.NodeGroupSpecStrings()
		assert.Equal(t, []string{"1:50:Deallocate:test1:{}|", "10:70:Deallocate:test2:{\"environment\":\"prod\"}|"}, specStrings)
	})

	t.Run("only taints", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:            "test1",
					MinSize:         1,
					MaxSize:         50,
					Taints:          "key1=value1:NoSchedule,key2=value2:NoSchedule",
					ScaleDownPolicy: deallocate.Delete,
				},
				{
					Name:            "test2",
					MinSize:         10,
					MaxSize:         70,
					ScaleDownPolicy: deallocate.Delete,
				},
			},
		}
		specStrings := c.NodeGroupSpecStrings()
		assert.Equal(t, []string{"1:50:Delete:test1:{}|key1=value1:NoSchedule,key2=value2:NoSchedule", "10:70:Delete:test2:{}|"}, specStrings)
	})

	t.Run("mix of labels and taints", func(t *testing.T) {
		c := &Config{
			NodeGroups: []NodeGroupSpec{
				{
					Name:    "test1",
					MinSize: 1,
					MaxSize: 50,
					Taints:  "key1=value1:NoSchedule,key2=value2:NoSchedule",
					Labels: map[string]string{
						"environment": "prod",
					},
					ScaleDownPolicy: deallocate.Delete,
				},
				{
					Name:    "test2",
					MinSize: 10,
					MaxSize: 70,
					Labels: map[string]string{
						"environment": "staging",
					},
					ScaleDownPolicy: deallocate.Delete,
				},
				{
					Name:    "test3",
					MinSize: 1,
					MaxSize: 5,
					Labels: map[string]string{
						"environment": "dev",
						"owner":       "myself",
					},
					Taints:          "key1=value1:NoSchedule,key2=value2:NoSchedule",
					ScaleDownPolicy: deallocate.Delete,
				},
			},
		}
		specStrings := c.NodeGroupSpecStrings()
		assert.Equal(t, []string{"1:50:Delete:test1:{\"environment\":\"prod\"}|key1=value1:NoSchedule,key2=value2:NoSchedule",
			"10:70:Delete:test2:{\"environment\":\"staging\"}|",
			"1:5:Delete:test3:{\"environment\":\"dev\",\"owner\":\"myself\"}|key1=value1:NoSchedule,key2=value2:NoSchedule"}, specStrings)
	})
}
