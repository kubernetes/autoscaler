/*
Copyright 2016 The Kubernetes Authors.

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

package linode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCludConfig_getSizeLimits(t *testing.T) {
	_, _, err := getSizeLimits("3", "2", 1, 2)
	assert.Error(t, err, "no errors on minSize > maxSize")

	_, _, err = getSizeLimits("4", "", 2, 3)
	assert.Error(t, err, "no errors on minSize > maxSize using defaults")

	_, _, err = getSizeLimits("", "4", 5, 10)
	assert.Error(t, err, "no errors on minSize > maxSize using defaults")

	_, _, err = getSizeLimits("-1", "4", 5, 10)
	assert.Error(t, err, "no errors on minSize <= 0")

	_, _, err = getSizeLimits("1", "4a", 5, 10)
	assert.Error(t, err, "no error on malformed integer string")

	_, _, err = getSizeLimits("1.0", "4", 5, 10)
	assert.Error(t, err, "no error on malformed integer string")

	min, max, err := getSizeLimits("", "", 1, 2)
	assert.Equal(t, 1, min)
	assert.Equal(t, 2, max)

	min, max, err = getSizeLimits("", "3", 1, 2)
	assert.Equal(t, 1, min)
	assert.Equal(t, 3, max)

	min, max, err = getSizeLimits("6", "8", 1, 2)
	assert.Equal(t, 6, min)
	assert.Equal(t, 8, max)
}

func TestCludConfig_buildCloudConfig(t *testing.T) {
	cfg := strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
defaut-min-size-per-linode-type=2
defaut-max-size-per-linode-type=10
do-not-import-pool-id=888
do-not-import-pool-id=999

[nodegroup "g6-standard-1"]
min-size=1
max-size=2

[nodegroup "g6-standard-2"]
min-size=4
max-size=5
`)
	config, err := buildCloudConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "123123123", config.token)
	assert.Equal(t, 456456, config.clusterID)
	assert.Equal(t, 2, config.defaultMinSize)
	assert.Equal(t, 10, config.defaultMaxSize)
	assert.Equal(t, true, config.excludedPoolIDs[999])
	assert.Equal(t, true, config.excludedPoolIDs[888])
	assert.Equal(t, 1, config.nodeGroupCfg["g6-standard-1"].minSize)
	assert.Equal(t, 2, config.nodeGroupCfg["g6-standard-1"].maxSize)
	assert.Equal(t, 4, config.nodeGroupCfg["g6-standard-2"].minSize)
	assert.Equal(t, 5, config.nodeGroupCfg["g6-standard-2"].maxSize)

	cfg = strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
defaut-max-size-per-linode-type=20

[nodegroup "g6-standard-1"]
max-size=10

[nodegroup "g6-standard-2"]
min-size=3
`)
	config, err = buildCloudConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, config.nodeGroupCfg["g6-standard-1"].minSize)
	assert.Equal(t, 10, config.nodeGroupCfg["g6-standard-1"].maxSize)
	assert.Equal(t, 3, config.nodeGroupCfg["g6-standard-2"].minSize)
	assert.Equal(t, 20, config.nodeGroupCfg["g6-standard-2"].maxSize)

	cfg = strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
defaut-max-size-per-linode-type=20

[nodegroup "g6-standard-1"]
max-size=10a
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error on size of a specific node group is not an integer string")

	cfg = strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
do-not-import-pool-id=99f
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error on excluded pool id is not an integer string")

	cfg = strings.NewReader(`
[global]
lke-cluster-id=456456
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error on missing linode token")

	cfg = strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456aaa
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error when lke cluster id is not an integer string")

	cfg = strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
defaut-max-size-per-linode-type=20.0
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error when default max size in global section is not an integer string")

	cfg = strings.NewReader(`
[gglobal]
linode-token=123123123
lke-cluster-id=456456
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error when config has no global section")
}
