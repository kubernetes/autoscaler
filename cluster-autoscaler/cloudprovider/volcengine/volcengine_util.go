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

package volcengine

// StringSliceInGroupsOf split arr into num parts
func StringSliceInGroupsOf(arr []string, num int64) [][]string {
	if arr == nil {
		return nil
	}
	sliceLen := int64(len(arr))
	if sliceLen <= num {
		return [][]string{arr}
	}
	var quantity int64
	if sliceLen%num == 0 {
		quantity = sliceLen / num
	} else {
		quantity = (sliceLen / num) + 1
	}
	var segments = make([][]string, 0)
	var start, end, i int64
	for i = 1; i <= quantity; i++ {
		end = i * num
		if i != quantity {
			segments = append(segments, arr[start:end])
		} else {
			segments = append(segments, arr[start:])
		}
		start = i * num
	}
	return segments
}
