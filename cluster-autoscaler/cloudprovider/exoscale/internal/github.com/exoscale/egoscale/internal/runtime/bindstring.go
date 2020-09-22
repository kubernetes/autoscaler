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

// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package runtime

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/deepmap/oapi-codegen/pkg/types"
)

// This function takes a string, and attempts to assign it to the destination
// interface via whatever type conversion is necessary. We have to do this
// via reflection instead of a much simpler type switch so that we can handle
// type aliases. This function was the easy way out, the better way, since we
// know the destination type each place that we use this, is to generate code
// to read each specific type.
func BindStringToObject(src string, dst interface{}) error {
	var err error

	v := reflect.ValueOf(dst)
	t := reflect.TypeOf(dst)

	// We need to dereference pointers
	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}

	// The resulting type must be settable. reflect will catch issues like
	// passing the destination by value.
	if !v.CanSet() {
		return errors.New("destination is not settable")
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var val int64
		val, err = strconv.ParseInt(src, 10, 64)
		if err == nil {
			v.SetInt(val)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var val uint64
		val, err = strconv.ParseUint(src, 10, 64)
		if err == nil {
			v.SetUint(val)
		}
	case reflect.String:
		v.SetString(src)
		err = nil
	case reflect.Float64, reflect.Float32:
		var val float64
		val, err = strconv.ParseFloat(src, 64)
		if err == nil {
			v.SetFloat(val)
		}
	case reflect.Bool:
		var val bool
		val, err = strconv.ParseBool(src)
		if err == nil {
			v.SetBool(val)
		}
	case reflect.Struct:
		switch dstType := dst.(type) {
		case *time.Time:
			// Don't fail on empty string.
			if src == "" {
				return nil
			}
			// Time is a special case of a struct that we handle
			parsedTime, err := time.Parse(time.RFC3339Nano, src)
			if err != nil {
				parsedTime, err = time.Parse(types.DateFormat, src)
				if err != nil {
					return fmt.Errorf("error parsing '%s' as RFC3339 or 2006-01-02 time: %s", src, err)
				}
			}
			*dstType = parsedTime
			return nil
		case *types.Date:
			// Don't fail on empty string.
			if src == "" {
				return nil
			}
			parsedTime, err := time.Parse(types.DateFormat, src)
			if err != nil {
				return fmt.Errorf("error parsing '%s' as date: %s", src, err)
			}
			dstType.Time = parsedTime
			return nil
		}
		fallthrough
	default:
		// We've got a bunch of types unimplemented, don't fail silently.
		err = fmt.Errorf("can not bind to destination of type: %s", t.Kind())
	}
	if err != nil {
		return fmt.Errorf("error binding string parameter: %s", err)
	}
	return nil
}
