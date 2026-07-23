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

package autoscaling

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	scopeutil "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/scope"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/autoscaler/vertical-pod-autoscaler/test/e2e/utils"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"
)

// scopeLabelKey is the node label key used to partition DaemonSet recommendations
// in the scoped-VPA e2e tests. It mirrors the "one recommendation per node class"
// use case described in the AEP (for example: nodes with vs. without a GPU).
const scopeLabelKey = "e2e.vpa.scope/group"

// daemonSetName is the name of the hamster DaemonSet targeted by the scoped VPA.
const daemonSetName = "hamster-daemonset"

var _ = utils.RecommenderE2eDescribe("Scoped DaemonSet", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	var vpaClientSet vpa_clientset.Interface

	ginkgo.BeforeEach(func() {
		vpaClientSet = utils.GetVpaClientSet(f)
	})

	f.It("serves per-scope-value recommendations and a global fallback", framework.WithFeatureGate(features.DaemonSetScope), framework.WithSlow(), func(ctx context.Context) {
		ginkgo.By("Labelling nodes with distinct scope values")
		expectedScopeValues := labelNodesForScope(ctx, f)

		ginkgo.By("Setting up a hamster DaemonSet")
		ds := setupHamsterDaemonSet(ctx, f, "100m", "100Mi")
		expectedScopeValues = restrictToScheduledNodes(ctx, f, ds, expectedScopeValues)
		gomega.Expect(expectedScopeValues).NotTo(gomega.BeEmpty(), "DaemonSet should run on at least one node")

		ginkgo.By("Setting up a scoped VPA CRD targeting the DaemonSet")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
			Name:       daemonSetName,
		}
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-scoped-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			WithScope(vpa_types.VerticalPodAutoscalerScopeType(scopeLabelKey)).
			WithUpdateMode(vpa_types.UpdateModeOff).
			Get()
		utils.InstallVPA(f, vpaCRD)

		ginkgo.By(fmt.Sprintf("Waiting for a recommendation group per scope value %v", sortedKeys(expectedScopeValues)))
		vpa, err := waitForRecommendationGroups(vpaClientSet, vpaCRD, expectedScopeValues)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Verifying the global recommendation is kept as a fallback")
		gomega.Expect(vpa.Status.Recommendation).NotTo(gomega.BeNil(),
			"scoped DaemonSet VPAs must keep a global recommendation so clients that ignore recommendationGroups (or run with the feature gate disabled) still get a value")
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).NotTo(gomega.BeEmpty())

		ginkgo.By("Verifying every group carries a CPU recommendation for the hamster container")
		for _, group := range vpa.Status.RecommendationGroups {
			gomega.Expect(group.ContainerRecommendations).To(gomega.HaveLen(1),
				fmt.Sprintf("group %q should recommend the single hamster container", group.ScopeValue))
			cpu := group.ContainerRecommendations[0].Target[apiv1.ResourceCPU]
			gomega.Expect(cpu.MilliValue()).To(gomega.BeNumerically(">", 0),
				fmt.Sprintf("group %q should have a positive CPU target", group.ScopeValue))
		}
	})
})

// labelNodesForScope labels the cluster nodes with distinct values of scopeLabelKey
// and returns the set of scope values that are expected to appear as recommendation
// groups. When more than one node is available the first node is intentionally left
// unlabelled so the AbsentLabelValue ("__absent__") group is exercised as well.
// A DeferCleanup is registered to remove the labels afterwards.
func labelNodesForScope(ctx context.Context, f *framework.Framework) map[string]bool {
	nodeList, err := f.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error listing nodes")

	nodeNames := make([]string, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		if isNodeReady(&node) {
			nodeNames = append(nodeNames, node.Name)
		}
	}
	gomega.Expect(nodeNames).NotTo(gomega.BeEmpty(), "expected at least one ready node")
	sort.Strings(nodeNames)

	// Alternate between two values to build several groups while remaining
	// deterministic regardless of the cluster size.
	values := []string{"group-a", "group-b"}
	expected := map[string]bool{}
	labelled := map[string]string{}
	for i, name := range nodeNames {
		if i == 0 && len(nodeNames) > 1 {
			// Leave the first node unlabelled to produce the __absent__ group.
			expected[scopeutil.AbsentLabelValue] = true
			continue
		}
		value := values[(i)%len(values)]
		setNodeLabel(ctx, f, name, scopeLabelKey, value)
		labelled[name] = value
		expected[value] = true
	}

	ginkgo.DeferCleanup(func(ctx context.Context) {
		for name := range labelled {
			removeNodeLabel(ctx, f, name, scopeLabelKey)
		}
	})
	return expected
}

// restrictToScheduledNodes narrows the expected scope-value set to the values that
// are actually present on the nodes where DaemonSet pods got scheduled. This keeps
// the assertion accurate even if some node did not receive a DaemonSet pod.
func restrictToScheduledNodes(ctx context.Context, f *framework.Framework, ds *appsv1.DaemonSet, expected map[string]bool) map[string]bool {
	nodeIndex := map[string]*apiv1.Node{}
	nodeList, err := f.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	for i := range nodeList.Items {
		nodeIndex[nodeList.Items[i].Name] = &nodeList.Items[i]
	}

	podList, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).List(ctx, metav1.ListOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	scheduled := map[string]bool{}
	for i := range podList.Items {
		pod := &podList.Items[i]
		if !metav1.IsControlledBy(pod, ds) || pod.Spec.NodeName == "" {
			continue
		}
		node, ok := nodeIndex[pod.Spec.NodeName]
		if !ok {
			continue
		}
		if value, found := node.Labels[scopeLabelKey]; found {
			scheduled[value] = true
		} else {
			scheduled[scopeutil.AbsentLabelValue] = true
		}
	}

	// Intersect the pre-computed expectation with what was actually scheduled.
	result := map[string]bool{}
	for value := range scheduled {
		if expected[value] {
			result[value] = true
		}
	}
	return result
}

// setupHamsterDaemonSet creates a hamster DaemonSet that burns CPU on every node
// (including control-plane nodes) and waits for it to become ready.
func setupHamsterDaemonSet(ctx context.Context, f *framework.Framework, cpu, memory string) *appsv1.DaemonSet {
	container := SetupHamsterContainer(cpu, memory)
	container.Args = []string{"-c", "yes >/dev/null"}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonSetName,
			Namespace: f.Namespace.Name,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: utils.HamsterLabels},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: utils.HamsterLabels},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{container},
					// Tolerate every taint so the DaemonSet also runs on
					// control-plane nodes, maximising the number of scope groups.
					Tolerations: []apiv1.Toleration{{Operator: apiv1.TolerationOpExists}},
				},
			},
		},
	}

	created, err := f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Create(ctx, ds, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error creating hamster DaemonSet")

	gomega.Expect(waitForDaemonSetReady(ctx, f, created.Name)).To(gomega.Succeed(),
		"unexpected error waiting for hamster DaemonSet to become ready")

	created, err = f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Get(ctx, created.Name, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return created
}

// waitForDaemonSetReady polls until all desired DaemonSet pods are ready.
func waitForDaemonSetReady(ctx context.Context, f *framework.Framework, name string) error {
	return wait.PollUntilContextTimeout(ctx, utils.PollInterval, utils.PollTimeout, true, func(ctx context.Context) (bool, error) {
		ds, err := f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if ds.Status.DesiredNumberScheduled == 0 {
			return false, nil
		}
		return ds.Status.NumberReady == ds.Status.DesiredNumberScheduled, nil
	})
}

// waitForRecommendationGroups waits until the VPA status exposes a recommendation
// group (with a container recommendation) for exactly the expected scope values and
// a non-empty global recommendation.
func waitForRecommendationGroups(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler, expected map[string]bool) (*vpa_types.VerticalPodAutoscaler, error) {
	return utils.WaitForVPAMatch(c, vpa, func(vpa *vpa_types.VerticalPodAutoscaler) bool {
		if vpa.Status.Recommendation == nil || len(vpa.Status.Recommendation.ContainerRecommendations) == 0 {
			return false
		}
		got := map[string]bool{}
		for _, group := range vpa.Status.RecommendationGroups {
			if len(group.ContainerRecommendations) == 0 {
				continue
			}
			got[group.ScopeValue] = true
		}
		return stringSetsEqual(got, expected)
	})
}

func setNodeLabel(ctx context.Context, f *framework.Framework, nodeName, key, value string) {
	patch := map[string]any{"metadata": map[string]any{"labels": map[string]string{key: value}}}
	patchNode(ctx, f, nodeName, patch)
}

func removeNodeLabel(ctx context.Context, f *framework.Framework, nodeName, key string) {
	patch := map[string]any{"metadata": map[string]any{"labels": map[string]any{key: nil}}}
	patchNode(ctx, f, nodeName, patch)
}

func patchNode(ctx context.Context, f *framework.Framework, nodeName string, patch map[string]any) {
	bytes, err := json.Marshal(patch)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = f.ClientSet.CoreV1().Nodes().Patch(ctx, nodeName, types.MergePatchType, bytes, metav1.PatchOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf("unexpected error patching node %q", nodeName))
}

func isNodeReady(node *apiv1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == apiv1.NodeReady {
			return cond.Status == apiv1.ConditionTrue
		}
	}
	return false
}

func stringSetsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
