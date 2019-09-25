/*
Copyright 2018 The Kubernetes Authors.

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

package baiducloud

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterAsg(t *testing.T) {
	asg := &Asg{
		baiducloudManager: testBaiducloudManager,
		minSize:           1,
		maxSize:           10,
		BaiducloudRef: BaiducloudRef{
			Name: "test-name",
		},
	}
	testBaiducloudManager.RegisterAsg(asg)
}

func TestBuildNodeFromTemplate(t *testing.T) {
	BaiduManager := &BaiducloudManager{}
	asg := &Asg{}
	template := &asgTemplate{}

	_, err := BaiduManager.buildNodeFromTemplate(asg, template)
	assert.NoError(t, err)

	asg.Name = "test-asg"
	template = &asgTemplate{
		InstanceType: 10,
		Region:       "test",
		Zone:         "test",
		CPU:          2,
		Memory:       2048,
		GpuCount:     0,
	}
	node, err := BaiduManager.buildNodeFromTemplate(asg, template)
	assert.NoError(t, err)
	if !strings.Contains(node.ObjectMeta.Name, "test-asg") {
		t.Errorf("Generate node name err, get: %s", node.ObjectMeta.Name)
	}
}
