/*
Copyright 2018 The Kubernetes Authors.

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

package hcloud

import (
	"testing"
	"time"
)

const apiTimestampFormat = "2006-01-02T15:04:05-07:00"

func mustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()

	ts, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("parse time: layout %v: value %v: %v", layout, value, err)
	}
	return ts
}
