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
	"reflect"
	"testing"
)

func Test_section_key(t *testing.T) {
	type fields struct {
		content map[string]*value
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *value
	}{
		{
			"contain key",
			fields{content: map[string]*value{
				"key1": {raw: "value1"},
			},
			},
			args{name: "key1"}, &value{raw: "value1"},
		},
		{
			"not contain key",
			fields{content: map[string]*value{}},
			args{name: "notkey"},
			&value{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &section{
				content: tt.fields.content,
			}
			if got := s.key(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("key() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sections_section(t *testing.T) {
	type fields struct {
		contains map[string]*section
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *section
	}{
		{
			"contain key",
			fields{contains: map[string]*section{
				"default": {content: map[string]*value{"key1": {raw: "value1"}}}},
			},
			args{name: "default"}, &section{content: map[string]*value{"key1": {raw: "value1"}}},
		},
		{
			"not contain key",
			fields{contains: map[string]*section{}},
			args{name: "notkey"},
			&section{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := sections{
				contains: tt.fields.contains,
			}
			if got := ss.section(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("section() = %v, want %v", got, tt.want)
			}
		})
	}
}
