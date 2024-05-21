/*
Copyright 2021 The Kubernetes Authors.

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

package common

import (
	"crypto/rand"
	"fmt"
	"reflect"
)

const (
	fieldClientToken = "ClientToken"
)

func safeInjectClientToken(obj interface{}) {
	// obj Must be struct ptr
	getType := reflect.TypeOf(obj)
	if getType.Kind() != reflect.Ptr || getType.Elem().Kind() != reflect.Struct {
		return
	}

	// obj Must exist named field
	_, ok := getType.Elem().FieldByName(fieldClientToken)
	if !ok {
		return
	}

	field := reflect.ValueOf(obj).Elem().FieldByName(fieldClientToken)

	// field Must be string ptr
	if field.Kind() != reflect.Ptr {
		return
	}

	// Set if ClientToken is nil or empty
	if field.IsNil() || (field.Elem().Kind() == reflect.String && field.Elem().Len() == 0) {
		uuidVal := randomClientToken()
		field.Set(reflect.ValueOf(&uuidVal))
	}
}

// randomClientToken generate random string as ClientToken
// ref: https://stackoverflow.com/questions/15130321/is-there-a-method-to-generate-a-uuid-with-go-language
func randomClientToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
