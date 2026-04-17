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

package inmemory

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestProvReqFullLifecycle(t *testing.T) {
	config := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(10*time.Minute),
			integration.WithProvisioningRequestEnabled(),
		)

	options := config.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n := test.BuildTestNode("ng1-node-0", 1000, 1000, test.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithNode(n))
		fakes.K8s.AddPod(test.BuildScheduledTestPod("p1", 600, 100, n.Name))

		// Create a ProvisioningRequest
		wrapper := provreqwrapper.BuildValidTestProvisioningRequestFromOptions(provreqwrapper.TestProvReqOptions{
			Name:     "test-pr",
			CPU:      "600m",
			Memory:   "100",
			PodCount: 1,
			Class:    v1.ProvisioningClassBestEffortAtomicScaleUp,
		})

		fakes.K8s.AddPodTemplate(wrapper.PodTemplates[0])
		_, err = fakes.PRClient.AutoscalingV1().ProvisioningRequests(wrapper.ProvisioningRequest.Namespace).Create(ctx, wrapper.ProvisioningRequest, metav1.CreateOptions{})
		assert.NoError(t, err)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, tg1)

		synctestutils.MustRunOnceAfter(t, autoscaler, 10*time.Second)

		// The NodeGroup should have scaled up by 1 due to the PR
		tg1After, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1After)

		// Scale-Down
		n2 := test.BuildTestNode("ng1-node-1", 1000, 1000, test.IsReady(true))
		fakes.K8s.AddNode(n2)

		err = fakes.PRClient.AutoscalingV1().ProvisioningRequests(wrapper.ProvisioningRequest.Namespace).Delete(ctx, wrapper.ProvisioningRequest.Name, metav1.DeleteOptions{})
		assert.NoError(t, err)

		// Run CA once to trigger the unneeded evaluation
		synctestutils.MustRunOnceAfter(t, autoscaler, 10*time.Second)

		// Step time forward by the ScaleDownUnneededTime (10 mins) + buffer (5 mins)
		synctestutils.MustRunOnceAfter(t, autoscaler, 15*time.Minute)

		tg1Final, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, tg1Final)
	})
}
