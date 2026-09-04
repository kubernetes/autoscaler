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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
)

func TestPickReason(t *testing.T) {
	tests := []struct {
		name   string
		errors []googleapi.ErrorItem
		want   string
	}{
		{
			"simple test, all reasons from map",
			[]googleapi.ErrorItem{
				{Reason: "badRequest"},
				{Reason: "quotaExceeded"},
				{Reason: "requestExceedsCapacity"},
				{Reason: "insufficientCapacity"},
				{Reason: "backendError"},
			},
			"requestExceedsCapacity",
		},
		{
			"simple test, some reasons from map",
			[]googleapi.ErrorItem{
				{Reason: "insufficientCapacity"},
				{Reason: "badRequest"},
			},
			"insufficientCapacity",
		},
		{
			"simple test, some reasons outside from map",
			[]googleapi.ErrorItem{
				{Reason: "badRequest"},
				{Reason: "test2"},
				{Reason: "quotaExceeded"},
				{Reason: "test3"},
			},
			"quotaExceeded",
		},
		{
			"simple test, all reasons outside from map",
			[]googleapi.ErrorItem{
				{Reason: "ctest"},
				{Reason: "test2"},
				{Reason: "btest"},
				{Reason: "test2"},
			},
			"btest",
		},
		{
			"empty list",
			[]googleapi.ErrorItem{},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PickReason(tt.errors); got != tt.want {
				t.Errorf("PickReason() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineResponseStatusForLatencyMetric(t *testing.T) {
	type args struct {
		rsp        any
		err        error
		withReason bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple test, nil err, nil response",
			args: args{
				rsp:        nil,
				err:        nil,
				withReason: false,
			},
			want: "200",
		},
		{
			name: "simple test, nil err, not nil response",
			args: args{
				rsp: &googleapi.ServerResponse{
					HTTPStatusCode: http.StatusOK,
				},
				err:        nil,
				withReason: false,
			},
			want: "200",
		},
		{
			name: "simple test, err is context.DeadlineExceeded",
			args: args{
				rsp:        nil,
				err:        context.DeadlineExceeded,
				withReason: false,
			},
			want: "timeout",
		},
		{
			name: "simple test, err is context.Canceled",
			args: args{
				rsp:        nil,
				err:        context.Canceled,
				withReason: false,
			},
			want: "timeout",
		},
		{
			name: "simple test, err is googleapi.Error",
			args: args{
				rsp: nil,
				err: &googleapi.Error{
					Code: http.StatusNotFound,
				},
				withReason: false,
			},
			want: "404",
		},
		{
			name: "simple test, err is googleapi.Error, withReason is false",
			args: args{
				rsp: nil,
				err: &googleapi.Error{
					Code: http.StatusBadRequest,
					Errors: []googleapi.ErrorItem{
						{Reason: "quotaExceeded"},
					},
				},
				withReason: false,
			},
			want: "400",
		},
		{
			name: "simple test, err is googleapi.Error, withReason is true, with errors",
			args: args{
				rsp: nil,
				err: &googleapi.Error{
					Code: http.StatusBadRequest,
					Errors: []googleapi.ErrorItem{
						{Reason: "badRequest"},
						{Reason: "SomeRandomString"},
						{Reason: "quotaExceeded"},
					},
				},
				withReason: true,
			},
			want: "400_quotaExceeded",
		},
		{
			name: "simple test, err is unknown error",
			args: args{
				rsp:        nil,
				err:        errors.New("some other error"),
				withReason: false,
			},
			want: "internal_error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineResponseStatusForLatencyMetric(tt.args.rsp, tt.args.err, tt.args.withReason)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStructFieldByName(t *testing.T) {
	type testStruct struct {
		ExportedField   string
		unexportedField string
	}

	testCases := []struct {
		name        string
		obj         any
		fieldname   string
		expectedVal any
		expectedErr bool
	}{
		{
			name: "get exported field from struct",
			obj: testStruct{
				ExportedField:   "exported value",
				unexportedField: "unexported value",
			},
			fieldname:   "ExportedField",
			expectedVal: "exported value",
			expectedErr: false,
		},
		{
			name: "get exported field from pointer to struct",
			obj: &testStruct{
				ExportedField:   "exported value",
				unexportedField: "unexported value",
			},
			fieldname:   "ExportedField",
			expectedVal: "exported value",
			expectedErr: false,
		},
		{
			name:        "nil object",
			obj:         nil,
			fieldname:   "ExportedField",
			expectedVal: nil,
			expectedErr: false,
		},
		{
			name:        "nil pointer to struct",
			obj:         (*testStruct)(nil),
			fieldname:   "ExportedField",
			expectedVal: nil,
			expectedErr: false,
		},
		{
			name:        "not a struct",
			obj:         "this is a string",
			fieldname:   "ExportedField",
			expectedVal: nil,
			expectedErr: true,
		},
		{
			name: "field does not exist",
			obj: testStruct{
				ExportedField:   "exported value",
				unexportedField: "unexported value",
			},
			fieldname:   "NonExistentField",
			expectedVal: nil,
			expectedErr: true,
		},
		{
			name: "unexported field",
			obj: testStruct{
				ExportedField:   "exported value",
				unexportedField: "unexported value",
			},
			fieldname:   "unexportedField",
			expectedVal: nil,
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := structFieldByName(tc.obj, tc.fieldname)
			assert.Equal(t, tc.expectedVal, val)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerResponseFromAny(t *testing.T) {
	type structWithServerResponse struct {
		ServerResponse googleapi.ServerResponse
	}

	type structWithWrongType struct {
		ServerResponse string
	}

	type structWithoutServerResponse struct {
		SomeOtherField string
	}

	testCases := []struct {
		name     string
		obj      any
		expected *googleapi.ServerResponse
	}{
		{
			name: "struct with ServerResponse",
			obj: &structWithServerResponse{
				ServerResponse: googleapi.ServerResponse{
					HTTPStatusCode: 200,
				},
			},
			expected: &googleapi.ServerResponse{
				HTTPStatusCode: 200,
			},
		},
		{
			name: "direct ServerResponse struct",
			obj: googleapi.ServerResponse{
				HTTPStatusCode: 200,
			},
			expected: &googleapi.ServerResponse{
				HTTPStatusCode: 200,
			},
		},
		{
			name: "direct ServerResponse pointer",
			obj: &googleapi.ServerResponse{
				HTTPStatusCode: 200,
			},
			expected: &googleapi.ServerResponse{
				HTTPStatusCode: 200,
			},
		},
		{
			name: "struct with wrong type for ServerResponse",
			obj: &structWithWrongType{
				ServerResponse: "not a server response",
			},
			expected: nil,
		},
		{
			name: "struct without ServerResponse field",
			obj: &structWithoutServerResponse{
				SomeOtherField: "some value",
			},
			expected: nil,
		},
		{
			name:     "nil object",
			obj:      nil,
			expected: nil,
		},
		{
			name:     "nil pointer to struct",
			obj:      (*structWithServerResponse)(nil),
			expected: nil,
		},
		{
			name:     "not a struct",
			obj:      "this is a string",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := serverResponseFromAny(tc.obj)
			assert.Equal(t, tc.expected, result)
		})
	}
}
