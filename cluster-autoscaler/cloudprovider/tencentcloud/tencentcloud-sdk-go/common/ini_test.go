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
	"testing"
)

func Test_openFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
		wantErr  bool
	}{
		{
			"invalid path", args{path: "./testdata"}, []byte{}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := openFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("openFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_parse(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *sections
		wantErr bool
	}{
		{"valid ini", args{path: "./testdata_valid.ini"}, nil, false},
		{"invalid ini", args{"./testdata_invalid.ini"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.section("default").key("key1").string() == "" {
					t.Errorf("parse() error:not get default.key1 value")
				}
				if _, e := got.section("custom").key("customKey2").bool(); e != nil {
					t.Errorf("parse() error:not get custom.customKey2 value")
				}
			}
		})
	}
}
