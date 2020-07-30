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

package internal

import (
	"reflect"
	"strings"
)

// RemainingKeys will inspect a struct and compare it to a map. Any struct
// field that does not have a JSON tag that matches a key in the map or
// a matching lower-case field in the map will be returned as an extra.
//
// This is useful for determining the extra fields returned in response bodies
// for resources that can contain an arbitrary or dynamic number of fields.
func RemainingKeys(s interface{}, m map[string]interface{}) (extras map[string]interface{}) {
	extras = make(map[string]interface{})
	for k, v := range m {
		extras[k] = v
	}

	valueOf := reflect.ValueOf(s)
	typeOf := reflect.TypeOf(s)
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)

		lowerField := strings.ToLower(field.Name)
		delete(extras, lowerField)

		if tagValue := field.Tag.Get("json"); tagValue != "" && tagValue != "-" {
			delete(extras, tagValue)
		}
	}

	return
}
