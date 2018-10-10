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

package endpoints

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMappingResolver_TryResolve(t *testing.T) {
	AddEndpointMapping("cn-hangzhou", "Ecs", "unreachable.aliyuncs.com")
	resolveParam := &ResolveParam{
		RegionId: "cn-hangzhou",
		Product:  "ecs",
	}
	endpoint, err := Resolve(resolveParam)
	assert.Nil(t, err)
	assert.Equal(t, endpoint, "unreachable.aliyuncs.com")
	fmt.Println("finished")
}
