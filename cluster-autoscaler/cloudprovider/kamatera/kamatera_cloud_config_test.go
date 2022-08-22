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

package kamatera

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudConfig_getSizeLimits(t *testing.T) {
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

func TestCloudConfig_buildCloudConfig(t *testing.T) {
	cfg := strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
default-min-size=1
default-max-size=10
default-name-prefix=test
default-password=Aa123456!
default-ssh-key=ssh-rsa AAAA...
default-datacenter=IL
default-image=ubuntu-1604-x64-server-2016-03-01
default-cpu=1a
default-ram=1024
default-disk=size=10
default-disk=size=20
default-network=name=wan,ip=auto
default-billingcycle=hourly
default-monthlypackage=t5000
default-script-base64=ZGVmYXVsdAo=

[nodegroup "default"]

[nodegroup "highcpu"]
min-size=3
name-prefix=highcpu
password=Bb654321!
ssh-key=ssh-rsa BBBB...
datacenter=US
image=ubuntu-2204
cpu=2a
ram=2048
disk=size=50
dailybackup=true
managed=true
network=name=wan,ip=auto
network=name=lan-12345-lan,ip=auto
billingcycle=monthly
monthlypackage=t10000
script-base64=aGlnaGJwdQo=

[nodegroup "highram"]
max-size=2
`)
	config, err := buildCloudConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "1a222bbb3ccc44d5555e6ff77g88hh9i", config.apiClientId)
	assert.Equal(t, "9ii88h7g6f55555ee4444444dd33eee2", config.apiSecret)
	assert.Equal(t, "aaabbb", config.clusterName)
	assert.Equal(t, 1, config.defaultMinSize)
	assert.Equal(t, 10, config.defaultMaxSize)
	assert.Equal(t, 3, len(config.nodeGroupCfg))
	assert.Equal(t, 1, config.nodeGroupCfg["default"].minSize)
	assert.Equal(t, 10, config.nodeGroupCfg["default"].maxSize)
	assert.Equal(t, 3, config.nodeGroupCfg["highcpu"].minSize)
	assert.Equal(t, 10, config.nodeGroupCfg["highcpu"].maxSize)
	assert.Equal(t, 1, config.nodeGroupCfg["highram"].minSize)
	assert.Equal(t, 2, config.nodeGroupCfg["highram"].maxSize)

	// default server configurations
	assert.Equal(t, "test", config.nodeGroupCfg["default"].NamePrefix)
	assert.Equal(t, "Aa123456!", config.nodeGroupCfg["default"].Password)
	assert.Equal(t, "ssh-rsa AAAA...", config.nodeGroupCfg["default"].SshKey)
	assert.Equal(t, "IL", config.nodeGroupCfg["default"].Datacenter)
	assert.Equal(t, "ubuntu-1604-x64-server-2016-03-01", config.nodeGroupCfg["default"].Image)
	assert.Equal(t, "1a", config.nodeGroupCfg["default"].Cpu)
	assert.Equal(t, "1024", config.nodeGroupCfg["default"].Ram)
	assert.Equal(t, []string{"size=10", "size=20"}, config.nodeGroupCfg["default"].Disks)
	assert.False(t, config.nodeGroupCfg["default"].Dailybackup)
	assert.False(t, config.nodeGroupCfg["default"].Managed)
	assert.Equal(t, []string{"name=wan,ip=auto"}, config.nodeGroupCfg["default"].Networks)
	assert.Equal(t, "hourly", config.nodeGroupCfg["default"].BillingCycle)
	assert.Equal(t, "t5000", config.nodeGroupCfg["default"].MonthlyPackage)
	assert.Equal(t, "ZGVmYXVsdAo=", config.nodeGroupCfg["default"].ScriptBase64)

	// highcpu server configurations
	assert.Equal(t, "highcpu", config.nodeGroupCfg["highcpu"].NamePrefix)
	assert.Equal(t, "Bb654321!", config.nodeGroupCfg["highcpu"].Password)
	assert.Equal(t, "ssh-rsa BBBB...", config.nodeGroupCfg["highcpu"].SshKey)
	assert.Equal(t, "US", config.nodeGroupCfg["highcpu"].Datacenter)
	assert.Equal(t, "ubuntu-2204", config.nodeGroupCfg["highcpu"].Image)
	assert.Equal(t, "2a", config.nodeGroupCfg["highcpu"].Cpu)
	assert.Equal(t, "2048", config.nodeGroupCfg["highcpu"].Ram)
	assert.Equal(t, []string{"size=50"}, config.nodeGroupCfg["highcpu"].Disks)
	assert.True(t, config.nodeGroupCfg["highcpu"].Dailybackup)
	assert.True(t, config.nodeGroupCfg["highcpu"].Managed)
	assert.Equal(t, []string{"name=wan,ip=auto", "name=lan-12345-lan,ip=auto"}, config.nodeGroupCfg["highcpu"].Networks)
	assert.Equal(t, "monthly", config.nodeGroupCfg["highcpu"].BillingCycle)
	assert.Equal(t, "t10000", config.nodeGroupCfg["highcpu"].MonthlyPackage)
	assert.Equal(t, "aGlnaGJwdQo=", config.nodeGroupCfg["highcpu"].ScriptBase64)

	cfg = strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
default-min-size=1
default-max-size=10

[nodegroup "default"]

[nodegroup "highcpu"]
min-size=3

[nodegroup "highram"]
max-size=2a
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error on size of a specific node group is not an integer string")

	cfg = strings.NewReader(`
[global]
cluster-name=aaabbb
default-min-size=1
default-max-size=10
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i

[nodegroup "default"]

[nodegroup "highcpu"]
min-size=3

[nodegroup "highram"]
max-size=2
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error on missing kamatera api secret")
	assert.Contains(t, err.Error(), "kamatera api secret is not set")

	cfg = strings.NewReader(`
[global]
cluster-name=aaabbb
default-min-size=1
default-max-size=10

[nodegroup "default"]

[nodegroup "highcpu"]
min-size=3

[nodegroup "highram"]
max-size=2
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error on missing kamatera api client id")
	assert.Contains(t, err.Error(), "kamatera api client id is not set")

	cfg = strings.NewReader(`
[gglobal]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
default-min-size=1
default-max-size=10

[nodegroup "default"]

[nodegroup "highcpu"]
min-size=3

[nodegroup "highram"]
max-size=2
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error when config has no global section")
	assert.Contains(t, err.Error(), "can't store data at section \"gglobal\"")

	cfg = strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=aaabbb
default-min-size=1
default-max-size=10

[nodegroup "1234567890123456"]
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error when nodegroup name is more then 15 characters")
	assert.Contains(t, err.Error(), "node group name must be at most 15 characters long")

	cfg = strings.NewReader(`
[global]
kamatera-api-client-id=1a222bbb3ccc44d5555e6ff77g88hh9i
kamatera-api-secret=9ii88h7g6f55555ee4444444dd33eee2
cluster-name=1234567890123456
default-min-size=1
default-max-size=10

[nodegroup "default"]
`)
	config, err = buildCloudConfig(cfg)
	assert.Error(t, err, "no error when cluster name is more then 15 characters")
	assert.Contains(t, err.Error(), "cluster name must be at most 15 characters long")
}
