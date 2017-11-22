/*
Copyright 2017 The Kubernetes Authors.

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

package api

import "strconv"

func stringRefToFloat64(p *string) (float64, error) {
	if p == nil {
		return 0, nil
	}
	return strconv.ParseFloat(*p, 64)
}

func stringRefToStringSlice(in ...*string) []string {
	vs := make([]string, len(in))

	for i, v := range in {
		vs[i] = *v
	}

	return vs
}

func stringToStringSliceRef(in ...string) []*string {
	vs := make([]*string, len(in))

	for i, v := range in {
		vs[i] = &v
	}

	return vs
}
