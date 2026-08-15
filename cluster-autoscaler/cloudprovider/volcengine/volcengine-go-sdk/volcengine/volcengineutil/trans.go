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

package volcengineutil

import "strings"

func ParameterToMap(body string, sensitive []string, enable bool) map[string]interface{} {
	if !enable {
		return nil
	}
	result := make(map[string]interface{})
	params := strings.Split(body, "&")
	for _, param := range params {
		values := strings.Split(param, "=")
		if values[0] == "Action" || values[0] == "Version" {
			continue
		}
		v := values[1]
		if sensitive != nil && len(sensitive) > 0 {
			for _, s := range sensitive {
				if strings.Contains(values[0], s) {
					v = "****"
					break
				}
			}
		}
		result[values[0]] = v
	}
	return result
}

func BodyToMap(input map[string]interface{}, sensitive []string, enable bool) map[string]interface{} {
	if !enable {
		return nil
	}
	result := make(map[string]interface{})
loop:
	for k, v := range input {
		if len(sensitive) > 0 {
			for _, s := range sensitive {
				if strings.Contains(k, s) {
					v = "****"
					result[k] = v
					continue loop
				}
			}
		}
		var (
			next    map[string]interface{}
			nextPtr *map[string]interface{}
			ok      bool
		)

		if next, ok = v.(map[string]interface{}); ok {
			result[k] = BodyToMap(next, sensitive, enable)
		} else if nextPtr, ok = v.(*map[string]interface{}); ok {
			result[k] = BodyToMap(*nextPtr, sensitive, enable)
		} else {
			result[k] = v
		}
	}
	return result
}
