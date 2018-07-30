/*
Copyright 2017 The Kubernetes Authors.

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

package aws

import (
	"testing"

	"time"

	"fmt"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestPriceModel_NodePrice(t *testing.T) {
	defaultFrom, err := time.Parse(time.RFC3339, "2017-01-02T15:00:00Z")
	assert.NoError(t, err)
	defaultTill, err := time.Parse(time.RFC3339, "2017-01-02T16:00:00Z")
	assert.NoError(t, err)

	pm := &priceModel{
		asgs: &fakeInstanceFinder{
			c: map[string]*asg{
				"node-a": {
					AwsRef: AwsRef{
						Name: "k8s-AutoscalingGroupWorker-AAAAAA",
					},
				},
				"node-b": {
					AwsRef: AwsRef{
						Name: "k8s-AutoscalingGroupWorker-BBBBBB",
					},
				},
				"node-c": {
					AwsRef: AwsRef{
						Name: "k8s-AutoscalingGroupWorker-CCCCCC",
					},
				},
			},
		},
		priceDescriptor: newFakePriceDescriptor(
			newASGPrice("k8s-AutoscalingGroupWorker-AAAAAA", 0.111),
			newASGPrice("k8s-AutoscalingGroupWorker-BBBBBB", 0.222),
		),
	}

	type testCase struct {
		node        *apiv1.Node
		from, till  time.Time
		expectError bool
		expectPrice float64
	}

	testCases := []testCase{
		{ // common case
			node:        buildNode(providerID("us-east-1", "node-a")),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: false,
			expectPrice: 0.111,
		},
		{ // common case
			node:        buildNode(providerID("us-east-1", "node-b")),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: false,
			expectPrice: 0.222,
		},
		{ // common case, zero duration
			node:        buildNode(providerID("us-east-1", "node-b")),
			from:        defaultFrom,
			till:        defaultFrom,
			expectError: false,
			expectPrice: 0.0,
		},
		{ // error case: no price could be found
			node:        buildNode(providerID("us-east-1", "node-c")),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: true,
			expectPrice: 0.0,
		},
		{ // error case: no asg found
			node:        buildNode(providerID("us-east-1", "node-d")),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: true,
			expectPrice: 0.0,
		},
		{ // error case: invalid provider id
			node:        buildNode("invalid-id"),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: true,
			expectPrice: 0.0,
		},
	}

	for n, tc := range testCases {
		price, err := pm.NodePrice(tc.node, tc.from, tc.till)
		if tc.expectError {
			assert.Error(t, err, fmt.Sprintf("case %d", n))
		} else {
			assert.NoError(t, err, fmt.Sprintf("case %d", n))
		}
		assert.Equal(t, tc.expectPrice, price, fmt.Sprintf("case %d", n))
	}
}

func TestPriceModel_PodPrice(t *testing.T) {
	pm := &priceModel{}

	defaultFrom, err := time.Parse(time.RFC3339, "2017-01-02T15:00:00Z")
	assert.NoError(t, err)
	defaultTill, err := time.Parse(time.RFC3339, "2017-01-02T16:00:00Z")
	assert.NoError(t, err)

	type testCase struct {
		pod         *apiv1.Pod
		from, till  time.Time
		expectError bool
		expectPrice float64
	}

	testCases := []testCase{
		{ // common case
			pod:         buildPod(podContainers{cpu: "0.5", memory: "1500m"}),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: false,
			expectPrice: 0.013875000012922101,
		},
		{ // with GPU
			pod:         buildPod(podContainers{cpu: "0.5", memory: "1500m", gpu: "1"}),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: false,
			expectPrice: 0.9658000000129221,
		},
		{ // multiple containers
			pod:         buildPod(podContainers{cpu: "0.5", memory: "1500m"}, podContainers{cpu: "0.8", memory: "300m"}),
			from:        defaultFrom,
			till:        defaultTill,
			expectError: false,
			expectPrice: 0.03607500001938315,
		},
		{ // no containers
			pod:         &apiv1.Pod{},
			from:        defaultFrom,
			till:        defaultTill,
			expectError: false,
			expectPrice: 0.0,
		},
		{ // zero gap between start and end time
			pod:         buildPod(podContainers{cpu: "0.5", memory: "1500m", gpu: "1"}),
			from:        defaultFrom,
			till:        defaultFrom,
			expectError: false,
			expectPrice: 0.0,
		},
	}

	for n, tc := range testCases {
		price, err := pm.PodPrice(tc.pod, tc.from, tc.till)
		if tc.expectError {
			assert.Error(t, err, fmt.Sprintf("case %d", n))
		} else {
			assert.NoError(t, err, fmt.Sprintf("case %d", n))
		}
		assert.Equal(t, tc.expectPrice, price, fmt.Sprintf("case %d", n))
	}
}

func newASGPrice(asgName string, price float64) asgPrice {
	return asgPrice{asgName, price}
}

type asgPrice struct {
	asgName string
	price   float64
}

func newFakePriceDescriptor(asgPrices ...asgPrice) *fakePriceDescriptor {
	c := make(map[string]float64)

	for _, a := range asgPrices {
		c[a.asgName] = a.price
	}

	return &fakePriceDescriptor{c}
}

type fakePriceDescriptor struct {
	c map[string]float64
}

func (pd *fakePriceDescriptor) Price(asgName string) (float64, error) {
	if p, found := pd.c[asgName]; found {
		return p, nil
	}
	return 0, fmt.Errorf("unable to determine price for asg %s", asgName)
}

func providerID(zone, name string) string {
	return fmt.Sprintf("aws:///%s/%s", zone, name)
}

func buildNode(providerID string) *apiv1.Node {
	return &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: providerID,
		},
	}
}

var _ instanceByASGFinder = &fakeInstanceFinder{}

type fakeInstanceFinder struct {
	c map[string]*asg
}

func (i *fakeInstanceFinder) GetAsgForInstance(instance AwsInstanceRef) *asg {
	if asg, found := i.c[instance.Name]; found {
		return asg
	}

	return nil
}

type podContainers struct {
	cpu    string
	gpu    string
	memory string
}

func buildPod(prs ...podContainers) *apiv1.Pod {
	return &apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: convertContainers(prs...),
		},
	}
}

func convertContainers(prs ...podContainers) []apiv1.Container {
	containers := make([]apiv1.Container, len(prs))

	for n, pr := range prs {
		containers[n].Resources.Requests = make(apiv1.ResourceList)

		if len(pr.cpu) != 0 {
			containers[n].Resources.Requests[apiv1.ResourceCPU] = resource.MustParse(pr.cpu)
		}
		if len(pr.gpu) != 0 {
			containers[n].Resources.Requests[resourceNvidiaGPU] = resource.MustParse(pr.gpu)
		}
		if len(pr.memory) != 0 {
			containers[n].Resources.Requests[apiv1.ResourceMemory] = resource.MustParse(pr.memory)
		}
	}

	return containers
}
