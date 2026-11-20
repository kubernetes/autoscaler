/*
Copyright The Kubernetes Authors.

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

package hack

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// Builds the memory histogram checkpoint VPA would hold after 8 daily peaks per replica at a
// ~216MB working set (the anon=200 regime), 2 replicas. PoC helper, not for upstream.
func TestMakeAgedHistory(t *testing.T) {
	model.InitializeAggregationsConfig(model.NewAggregationsConfig(24*time.Hour, 8, 24*time.Hour, 24*time.Hour, 1.2, 100*1024*1024))
	s := model.NewAggregateContainerState()
	now := time.Now().UTC()
	for d := 8; d >= 1; d-- {
		end := now.Add(-time.Duration(d-1) * 24 * time.Hour)
		for range 2 {
			s.AggregateMemoryPeaks.AddSample(216*1024*1024, 1.0, end)
		}
	}
	h, err := s.AggregateMemoryPeaks.SaveToChekpoint()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(h)
	fmt.Println("HIST", string(b))
}
