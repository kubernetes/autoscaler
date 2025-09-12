/*
Copyright 2022 The Kubernetes Authors.

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

package deepcopy

import (
	"reflect"
	"sync"
)

type deepCopyInterface interface {
	DeepCopy() interface{}
}

// Copy deepcopies from v.
func Copy(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	if fromSyncMap, ok := src.(*sync.Map); ok {
		to := copySyncMap(fromSyncMap)
		return to
	}

	return copyNormal(src)
}

// copySyncMap copies with sync.Map but not nested
// Targets are vmssVMCache, vmssFlexVMCache, etc.
func copySyncMap(from *sync.Map) *sync.Map {
	to := &sync.Map{}

	from.Range(func(k, v interface{}) bool {
		vm, ok := v.(*sync.Map)
		if ok {
			to.Store(k, copySyncMap(vm))
		} else {
			to.Store(k, copyNormal(v))
		}
		return true
	})

	return to
}

func copyNormal(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	from := reflect.ValueOf(src)

	to := reflect.New(from.Type()).Elem()

	copyCustomimpl(from, to)

	return to.Interface()
}

func copyCustomimpl(from, to reflect.Value) {
	// Check if DeepCopy() is already implemented for the interface
	if from.CanInterface() {
		if deepcopy, ok := from.Interface().(deepCopyInterface); ok {
			to.Set(reflect.ValueOf(deepcopy.DeepCopy()))
			return
		}
	}

	switch from.Kind() {
	case reflect.Pointer:
		fromValue := from.Elem()
		if !fromValue.IsValid() {
			return
		}

		to.Set(reflect.New(fromValue.Type()))
		copyCustomimpl(fromValue, to.Elem())

	case reflect.Interface:
		if from.IsNil() {
			return
		}

		fromValue := from.Elem()
		toValue := reflect.New(fromValue.Type()).Elem()
		copyCustomimpl(fromValue, toValue)
		to.Set(toValue)

	case reflect.Struct:
		for i := 0; i < from.NumField(); i++ {
			if from.Type().Field(i).PkgPath != "" {
				// It is an unexported field.
				continue
			}
			copyCustomimpl(from.Field(i), to.Field(i))
		}

	case reflect.Slice:
		if from.IsNil() {
			return
		}

		to.Set(reflect.MakeSlice(from.Type(), from.Len(), from.Cap()))
		for i := 0; i < from.Len(); i++ {
			copyCustomimpl(from.Index(i), to.Index(i))
		}

	case reflect.Map:
		if from.IsNil() {
			return
		}

		to.Set(reflect.MakeMap(from.Type()))
		for _, key := range from.MapKeys() {
			fromValue := from.MapIndex(key)
			toValue := reflect.New(fromValue.Type()).Elem()
			copyCustomimpl(fromValue, toValue)
			copiedKey := Copy(key.Interface())
			to.SetMapIndex(reflect.ValueOf(copiedKey), toValue)
		}

	default:
		to.Set(from)
	}
}
