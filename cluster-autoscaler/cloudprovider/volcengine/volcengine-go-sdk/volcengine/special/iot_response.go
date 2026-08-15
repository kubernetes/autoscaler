/*
Copyright 2023 The Kubernetes Authors.

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

package special

import (
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/response"
)

func iotResponse(response response.VolcengineResponse, i interface{}) interface{} {
	_, ok1 := reflect.TypeOf(i).Elem().FieldByName("ResponseMetadata")
	_, ok2 := reflect.TypeOf(i).Elem().FieldByName("Result")
	if ok1 && ok2 {
		return response
	}
	return response.Result
}
