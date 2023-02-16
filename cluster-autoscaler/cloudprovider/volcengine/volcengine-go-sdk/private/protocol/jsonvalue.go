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

package protocol

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
)

// EscapeMode is the mode that should be use for escaping a value
type EscapeMode uint

// The modes for escaping a value before it is marshaled, and unmarshaled.
const (
	NoEscape EscapeMode = iota
	Base64Escape
	QuotedEscape
)

// EncodeJSONValue marshals the value into a JSON string, and optionally base64
// encodes the string before returning it.
//
// Will panic if the escape mode is unknown.
func EncodeJSONValue(v volcengine.JSONValue, escape EscapeMode) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	switch escape {
	case NoEscape:
		return string(b), nil
	case Base64Escape:
		return base64.StdEncoding.EncodeToString(b), nil
	case QuotedEscape:
		return strconv.Quote(string(b)), nil
	}

	panic(fmt.Sprintf("EncodeJSONValue called with unknown EscapeMode, %v", escape))
}

// DecodeJSONValue will attempt to decode the string input as a JSONValue.
// Optionally decoding base64 the value first before JSON unmarshaling.
//
// Will panic if the escape mode is unknown.
func DecodeJSONValue(v string, escape EscapeMode) (volcengine.JSONValue, error) {
	var b []byte
	var err error

	switch escape {
	case NoEscape:
		b = []byte(v)
	case Base64Escape:
		b, err = base64.StdEncoding.DecodeString(v)
	case QuotedEscape:
		var u string
		u, err = strconv.Unquote(v)
		b = []byte(u)
	default:
		panic(fmt.Sprintf("DecodeJSONValue called with unknown EscapeMode, %v", escape))
	}

	if err != nil {
		return nil, err
	}

	m := volcengine.JSONValue{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
