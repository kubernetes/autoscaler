/*
Copyright 2026 The Kubernetes Authors.

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

package gce

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"google.golang.org/api/googleapi"
	"k8s.io/klog/v2"
)

// reasonWeights contains an integer for each of the important error reasons,
// the higher the number the more important the reason. Those outside of the
// map will get 0, values with the same weight will be sorted lexicographically.
var reasonWeights = map[string]int{
	"badRequest":             1,
	"backendError":           2,
	"quotaExceeded":          3,
	"insufficientCapacity":   4,
	"requestExceedsCapacity": 5,
}

// PickReason returns the most significant reason from a list of ErrorItems.
func PickReason(errors []googleapi.ErrorItem) string {
	if len(errors) == 0 {
		return ""
	}
	reasons := make([]string, 0, len(errors))
	for _, err := range errors {
		reasons = append(reasons, err.Reason)
	}
	sort.Slice(reasons, func(i, j int) bool {
		iWeight := reasonWeights[reasons[i]]
		jWeight := reasonWeights[reasons[j]]
		if iWeight == jWeight {
			return reasons[i] < reasons[j]
		}
		return iWeight > jWeight
	})
	return reasons[0]
}

func determineResponseStatusForLatencyMetric(rsp any, err error, withReason bool) string {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "timeout"
		} else if gErr, ok := err.(*googleapi.Error); ok {
			if withReason && len(gErr.Errors) > 0 {
				return fmt.Sprintf("%d_%s", gErr.Code, PickReason(gErr.Errors))
			}
			return fmt.Sprintf("%d", gErr.Code)
		}
		return "internal_error"
	}

	serverResponse := serverResponseFromAny(rsp)
	if serverResponse != nil {
		return fmt.Sprintf("%d", serverResponse.HTTPStatusCode)
	}

	// If no ServerResponse was found and there was no error, assume 200 OK
	return "200"
}

func serverResponseFromAny(obj any) *googleapi.ServerResponse {
	if obj == nil {
		return nil
	}
	if sr, ok := obj.(*googleapi.ServerResponse); ok {
		return sr
	}
	if sr, ok := obj.(googleapi.ServerResponse); ok {
		return &sr
	}

	fieldname := "ServerResponse"
	fieldInterface, err := structFieldByName(obj, fieldname)
	if err != nil {
		return nil
	}
	if fieldInterface == nil {
		return nil
	}
	serverResponse, ok := fieldInterface.(googleapi.ServerResponse)
	if !ok {
		klog.V(5).Infof("Field value for fieldname %s is not of expected type googleapi.ServerResponse, but %T", fieldname, fieldInterface)
		return nil
	}
	return &serverResponse
}

func structFieldByName(obj any, fieldname string) (any, error) {
	if obj == nil {
		return nil, nil
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %T is not a struct", obj)
	}
	fieldVal := val.FieldByName(fieldname)
	if !fieldVal.IsValid() {
		return nil, fmt.Errorf("for struct of type %T, field value for fieldname %s is not valid", obj, fieldname)
	}
	if !fieldVal.CanInterface() {
		return nil, fmt.Errorf("for struct of type %T, field value for fieldname %s is not addressable", obj, fieldname)
	}
	return fieldVal.Interface(), nil
}
