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

package pagination

import (
	"fmt"
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
)

// LinkedPageBase may be embedded to implement a page that provides navigational "Next" and "Previous" links within its result.
type LinkedPageBase struct {
	PageResult

	// LinkPath lists the keys that should be traversed within a response to arrive at the "next" pointer.
	// If any link along the path is missing, an empty URL will be returned.
	// If any link results in an unexpected value type, an error will be returned.
	// When left as "nil", []string{"links", "next"} will be used as a default.
	LinkPath []string
}

// NextPageURL extracts the pagination structure from a JSON response and returns the "next" link, if one is present.
// It assumes that the links are available in a "links" element of the top-level response object.
// If this is not the case, override NextPageURL on your result type.
func (current LinkedPageBase) NextPageURL() (string, error) {
	var path []string
	var key string

	if current.LinkPath == nil {
		path = []string{"links", "next"}
	} else {
		path = current.LinkPath
	}

	submap, ok := current.Body.(map[string]interface{})
	if !ok {
		//		err := gophercloud.ErrUnexpectedType{}
		//		err.Expected = "map[string]interface{}"
		//		err.Actual = fmt.Sprintf("%v", reflect.TypeOf(current.Body))
		//		return "", err

		expected := "map[string]interface{}"
		actual := fmt.Sprintf("%v", reflect.TypeOf(current.Body))
		message := fmt.Sprintf(huaweicloudsdk.CEErrUnexpectedTypeMessage, expected, actual)
		err := huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CEErrUnexpectedTypeCode, message)
		return "", err
	}

	for {
		key, path = path[0], path[1:]

		value, ok := submap[key]
		if !ok {
			return "", nil
		}

		if len(path) > 0 {
			submap, ok = value.(map[string]interface{})
			if !ok {
				//				err := gophercloud.ErrUnexpectedType{}
				//				err.Expected = "map[string]interface{}"
				//				err.Actual = fmt.Sprintf("%v", reflect.TypeOf(value))
				//				return "", err

				expected := "map[string]interface{}"
				actual := fmt.Sprintf("%v", reflect.TypeOf(value))
				message := fmt.Sprintf(huaweicloudsdk.CEErrUnexpectedTypeMessage, expected, actual)
				err := huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CEErrUnexpectedTypeCode, message)
				return "", err
			}
		} else {
			if value == nil {
				// Actual null element.
				return "", nil
			}

			url, ok := value.(string)
			if !ok {
				//				err := gophercloud.ErrUnexpectedType{}
				//				err.Expected = "string"
				//				err.Actual = fmt.Sprintf("%v", reflect.TypeOf(value))
				//				return "", err

				expected := "string"
				actual := fmt.Sprintf("%v", reflect.TypeOf(value))
				message := fmt.Sprintf(huaweicloudsdk.CEErrUnexpectedTypeMessage, expected, actual)
				err := huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CEErrUnexpectedTypeCode, message)
				return "", err
			}

			return url, nil
		}
	}
}

// IsEmpty satisifies the IsEmpty method of the Page interface
func (current LinkedPageBase) IsEmpty() (bool, error) {
	if b, ok := current.Body.([]interface{}); ok {
		return len(b) == 0, nil
	}
	//	err := gophercloud.ErrUnexpectedType{}
	//	err.Expected = "[]interface{}"
	//	err.Actual = fmt.Sprintf("%v", reflect.TypeOf(current.Body))
	//	return true, err

	expected := "[]interface{}"
	actual := fmt.Sprintf("%v", reflect.TypeOf(current.Body))
	message := fmt.Sprintf(huaweicloudsdk.CEErrUnexpectedTypeMessage, expected, actual)
	err := huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CEErrUnexpectedTypeCode, message)
	return true, err
}

// GetBody returns the linked page's body. This method is needed to satisfy the
// Page interface.
func (current LinkedPageBase) GetBody() interface{} {
	return current.Body
}

// WrapNextPageURL function use makerID to warp next page url,it returns the full url for request.
func (current LinkedPageBase) WrapNextPageURL(markerID string) (string, error) {
	limit := current.URL.Query().Get("limit")

	if limit == "" {
		return "", nil
	}

	q := current.URL.Query()

	q.Set("marker", markerID)
	current.URL.RawQuery = q.Encode()
	return current.URL.String(), nil
}
