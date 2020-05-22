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

package magnum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAutoDiscoverySpec(t *testing.T) {
	specs := []struct {
		Spec  string
		Roles []string
		Err   bool
	}{
		{Spec: "magnum:role=autoscaling", Roles: []string{"autoscaling"}, Err: false},
		{Spec: "magnum:role=autoscaling,worker", Roles: []string{"autoscaling", "worker"}, Err: false},
		{Spec: "magnum:role=autoscaling,", Roles: nil, Err: true},
		{Spec: "magnum:role=,,", Roles: nil, Err: true},
		{Spec: "magnum:role=,", Roles: nil, Err: true},
		{Spec: "magnum:role=", Roles: nil, Err: true},
		{Spec: "magnum:role", Roles: nil, Err: true},
		{Spec: "magnum:", Roles: nil, Err: true},
		{Spec: "magnum", Roles: nil, Err: true},
		{Spec: "", Roles: nil, Err: true},

		{Spec: "abc:role=autoscaling", Roles: nil, Err: true},
		{Spec: "magnum:abc=autoscaling", Roles: nil, Err: true},
	}

	for _, s := range specs {
		cfg, err := parseMagnumAutoDiscoverySpec(s.Spec)
		if s.Err {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, s.Roles, cfg.Roles)
	}
}
