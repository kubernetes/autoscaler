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

package xmlutil

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"encoding/xml"
	"strings"
)

type xmlAttrSlice []xml.Attr

func (x xmlAttrSlice) Len() int {
	return len(x)
}

func (x xmlAttrSlice) Less(i, j int) bool {
	spaceI, spaceJ := x[i].Name.Space, x[j].Name.Space
	localI, localJ := x[i].Name.Local, x[j].Name.Local
	valueI, valueJ := x[i].Value, x[j].Value

	spaceCmp := strings.Compare(spaceI, spaceJ)
	localCmp := strings.Compare(localI, localJ)
	valueCmp := strings.Compare(valueI, valueJ)

	if spaceCmp == -1 || (spaceCmp == 0 && (localCmp == -1 || (localCmp == 0 && valueCmp == -1))) {
		return true
	}

	return false
}

func (x xmlAttrSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}
