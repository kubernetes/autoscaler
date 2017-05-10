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

package cloudprovider

import (
	"fmt"
	"testing"
)

func TestNodeGroupDiscoveryOptionsValidate(t *testing.T) {
	o := NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpec: "asg:tag=foobar",
		NodeGroupSpecs:             []string{"myasg:0:10"},
	}

	err := o.Validate()
	if err == nil {
		t.Errorf("Expected validation error didn't occur with NodeGroupDiscoveryOptions: %+v", o)
		t.FailNow()
	}
	if msg := fmt.Sprintf("%v", err); msg != `Either node group specs([myasg:0:10]) or node group auto discovery spec(asg:tag=foobar) can be specified but not both` {
		t.Errorf("Unexpected validation error message: %s", msg)
	}
}
