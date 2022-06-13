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

package rancher

import "testing"

func TestNewConfig(t *testing.T) {
	cfg, err := newConfig("./examples/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.URL) == 0 {
		t.Fatal("expected url to be set")
	}

	if len(cfg.Token) == 0 {
		t.Fatal("expected token to be set")
	}

	if len(cfg.ClusterName) == 0 {
		t.Fatal("expected cluster name to be set")
	}

	if len(cfg.ClusterNamespace) == 0 {
		t.Fatal("expected cluster namespace to be set")
	}
}
