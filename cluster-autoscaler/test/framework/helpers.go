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

package framework

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2erc "k8s.io/kubernetes/test/e2e/framework/rc"
	testutils "k8s.io/kubernetes/test/utils"
	imageutils "k8s.io/kubernetes/test/utils/image"
)

func CreateHostPortPods(f *Framework, id string, replicas int, expectRunning bool) {
	f.T.Log("Running RC which reserves host port")
	config := &testutils.RCConfig{
		Client:    f.ClientSet,
		Name:      id,
		Namespace: f.Namespace.Name,
		Timeout:   defaultTimeout,
		Image:     imageutils.GetPauseImageName(),
		Replicas:  replicas,
		HostPorts: map[string]int{"port1": 4321},
	}
	err := e2erc.RunRC(*config)
	if expectRunning {
		assert.NoError(f.T, err)
	}
}

func runDrainTest(f *Framework, migSizes map[string]int, namespace string, podsPerNode, pdbSize int, verifyFunction func(int)) {
	increasedSize := manuallyIncreaseClusterSize(f, migSizes)

	nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{FieldSelector: fields.Set{
		"spec.unschedulable": "false",
	}.AsSelector().String()})
	assert.NoError(f.T, err)
	numPods := len(nodes.Items) * podsPerNode
	testID := string(uuid.NewUUID()) // So that we can label and find pods
	labelMap := map[string]string{"test_id": testID}
	assert.NoError(f.T, runReplicatedPodOnEachNode(f, nodes.Items, namespace, podsPerNode, "reschedulable-pods", labelMap, 0))

	defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, namespace, "reschedulable-pods")

	f.T.Log("Create a PodDisruptionBudget")
	minAvailable := intstr.FromInt(numPods - pdbSize)
	pdb := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_pdb",
			Namespace: namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector:     &metav1.LabelSelector{MatchLabels: labelMap},
			MinAvailable: &minAvailable,
		},
	}
	_, err = f.ClientSet.PolicyV1beta1().PodDisruptionBudgets(namespace).Create(context.TODO(), pdb, metav1.CreateOptions{})

	defer func() {
		f.ClientSet.PolicyV1beta1().PodDisruptionBudgets(namespace).Delete(context.TODO(), pdb.Name, metav1.DeleteOptions{})
	}()

	assert.NoError(f.T, err)
	verifyFunction(increasedSize)
}

func reserveMemory(f *Framework, id string, replicas, megabytes int, expectRunning bool, timeout time.Duration, selector map[string]string, tolerations []v1.Toleration, priorityClassName string) func() error {
	f.T.Logf("Running RC which reserves %v MB of memory", megabytes)
	request := int64(1024 * 1024 * megabytes / replicas)
	config := &testutils.RCConfig{
		Client:            f.ClientSet,
		Name:              id,
		Namespace:         f.Namespace.Name,
		Timeout:           timeout,
		Image:             imageutils.GetPauseImageName(),
		Replicas:          replicas,
		MemRequest:        request,
		NodeSelector:      selector,
		Tolerations:       tolerations,
		PriorityClassName: priorityClassName,
	}
	for start := time.Now(); time.Since(start) < rcCreationRetryTimeout; time.Sleep(rcCreationRetryDelay) {
		err := e2erc.RunRC(*config)
		if err != nil && strings.Contains(err.Error(), "Error creating replication controller") {
			f.T.Logf("Failed to create memory reservation: %v", err)
			continue
		}
		if expectRunning {
			assert.NoError(f.T, err)
		}
		return func() error {
			return e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, id)
		}
	}
	f.T.Fatal("Failed to reserve memory within timeout")
	return nil
}

func ReserveMemoryAsync(f *Framework, id string, replicas, megabytes int, expectRunning bool, timeout time.Duration) <-chan int {
	done := make(chan int)
	go func() {
		reserveMemory(f, id, replicas, megabytes, expectRunning, timeout, nil, nil, "")
		close(done)
	}()
	return done
}

// ReserveMemoryWithPriority creates a replication controller with pods with priority that, in summation,
// request the specified amount of memory.
func ReserveMemoryWithPriority(f *Framework, id string, replicas, megabytes int, expectRunning bool, timeout time.Duration, priorityClassName string) func() error {
	return reserveMemory(f, id, replicas, megabytes, expectRunning, timeout, nil, nil, priorityClassName)
}

// ReserveMemoryWithSelectorAndTolerations creates a replication controller with pods with node selector that, in summation,
// request the specified amount of memory.
func ReserveMemoryWithSelectorAndTolerations(f *Framework, id string, replicas, megabytes int, expectRunning bool, timeout time.Duration, selector map[string]string, tolerations []v1.Toleration) func() error {
	return reserveMemory(f, id, replicas, megabytes, expectRunning, timeout, selector, tolerations, "")
}

// ReserveMemory creates a replication controller with pods that, in summation,
// request the specified amount of memory.
func ReserveMemory(f *Framework, id string, replicas, megabytes int, expectRunning bool, timeout time.Duration) func() error {
	return reserveMemory(f, id, replicas, megabytes, expectRunning, timeout, nil, nil, "")
}

// WaitForClusterSizeFunc waits until the cluster size matches the given function.
func WaitForClusterSizeFunc(f *Framework, sizeFunc func(int) bool, timeout time.Duration) error {
	return WaitForClusterSizeFuncWithUnready(f, sizeFunc, timeout, 0)
}

// WaitForClusterSizeFuncWithUnready waits until the cluster size matches the given function and assumes some unready nodes.
func WaitForClusterSizeFuncWithUnready(f *Framework, sizeFunc func(int) bool, timeout time.Duration, expectedUnready int) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(20 * time.Second) {
		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector().String()})
		if err != nil {
			f.T.Logf("Failed to list nodes: %v", err)
			continue
		}
		numNodes := len(nodes.Items)

		// Filter out not-ready nodes.
		e2enode.Filter(nodes, func(node v1.Node) bool {
			return e2enode.IsConditionSetAsExpected(&node, v1.NodeReady, true)
		})
		numReady := len(nodes.Items)

		if numNodes == numReady+expectedUnready && sizeFunc(numNodes) {
			f.T.Logf("Cluster has reached the desired size")
			return nil
		}
		f.T.Logf("Waiting for cluster with func, current size %d, not ready nodes %d", numNodes, numNodes-numReady)
	}
	return fmt.Errorf("timeout waiting %v for appropriate cluster size", timeout)
}

func waitForCaPodsReadyInNamespace(f *Framework, tolerateUnreadyCount int) error {
	var notready []string
	for start := time.Now(); time.Now().Before(start.Add(scaleUpTimeout)); time.Sleep(20 * time.Second) {
		pods, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pods: %v", err)
		}
		notready = make([]string, 0)
		for _, pod := range pods.Items {
			ready := false
			for _, c := range pod.Status.Conditions {
				if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
					ready = true
				}
			}
			// Failed pods in this context generally mean that they have been
			// double scheduled onto a node, but then failed a constraint check.
			if pod.Status.Phase == v1.PodFailed {
				f.T.Logf("Pod has failed: %v", pod)
			}
			if !ready && pod.Status.Phase != v1.PodFailed {
				notready = append(notready, pod.Name)
			}
		}
		if len(notready) <= tolerateUnreadyCount {
			f.T.Logf("sufficient number of pods ready. Tolerating %d unready", tolerateUnreadyCount)
			return nil
		}
		f.T.Logf("Too many pods are not ready yet: %v", notready)
	}
	f.T.Logf("Timeout on waiting for pods being ready")

	// Some pods are still not running.
	return fmt.Errorf("Too many pods are still not running: %v", notready)
}

func waitForAllCaPodsReadyInNamespace(f *Framework) error {
	return waitForCaPodsReadyInNamespace(f, 0)
}

func setMigSizes(f *Framework, sizes map[string]int) bool {
	madeChanges := false
	f.T.Logf("got sizes %v", sizes)
	for mig, desiredSize := range sizes {
		currentSize, err := f.Provider.GroupSize(mig)
		assert.NoError(f.T, err)
		if desiredSize != currentSize {
			f.T.Logf("Setting size of %s to %d", mig, desiredSize)
			err = f.Provider.ResizeGroup(mig, int32(desiredSize))
			assert.NoError(f.T, err)
			madeChanges = true
		}
	}
	return madeChanges
}

func makeNodeUnschedulable(f *Framework, node *v1.Node) error {
	f.T.Logf("Taint node %s", node.Name)
	for j := 0; j < 3; j++ {
		freshNode, err := f.ClientSet.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, taint := range freshNode.Spec.Taints {
			if taint.Key == disabledTaint {
				return nil
			}
		}
		freshNode.Spec.Taints = append(freshNode.Spec.Taints, v1.Taint{
			Key:    disabledTaint,
			Value:  "DisabledForTest",
			Effect: v1.TaintEffectNoSchedule,
		})
		_, err = f.ClientSet.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}
		if !apierrors.IsConflict(err) {
			return err
		}
		f.T.Logf("Got 409 conflict when trying to taint node, retries left: %v", 3-j)
	}
	return fmt.Errorf("Failed to taint node in allowed number of retries")
}

// CriticalAddonsOnlyError implements the `error` interface, and signifies the
// presence of the `CriticalAddonsOnly` taint on the node.
type CriticalAddonsOnlyError struct{}

func (CriticalAddonsOnlyError) Error() string {
	return fmt.Sprintf("CriticalAddonsOnly taint found on node")
}

func makeNodeSchedulable(f *Framework, node *v1.Node, failOnCriticalAddonsOnly bool) error {
	f.T.Logf("Remove taint from node %s", node.Name)
	for j := 0; j < 3; j++ {
		freshNode, err := f.ClientSet.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		var newTaints []v1.Taint
		for _, taint := range freshNode.Spec.Taints {
			if failOnCriticalAddonsOnly && taint.Key == criticalAddonsOnlyTaint {
				return CriticalAddonsOnlyError{}
			}
			if taint.Key != disabledTaint {
				newTaints = append(newTaints, taint)
			}
		}

		if len(newTaints) == len(freshNode.Spec.Taints) {
			return nil
		}
		freshNode.Spec.Taints = newTaints
		_, err = f.ClientSet.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}
		if !apierrors.IsConflict(err) {
			return err
		}
		f.T.Logf("Got 409 conflict when trying to taint node, retries left: %v", 3-j)
	}
	return fmt.Errorf("Failed to remove taint from node in allowed number of retries")
}

// Create an RC running a given number of pods with anti-affinity
func runAntiAffinityPods(f *Framework, namespace string, pods int, id string, podLabels, antiAffinityLabels map[string]string) error {
	config := &testutils.RCConfig{
		Affinity:  buildAntiAffinity(antiAffinityLabels),
		Client:    f.ClientSet,
		Name:      id,
		Namespace: namespace,
		Timeout:   scaleUpTimeout,
		Image:     imageutils.GetPauseImageName(),
		Replicas:  pods,
		Labels:    podLabels,
	}
	err := e2erc.RunRC(*config)
	if err != nil {
		return err
	}
	_, err = f.ClientSet.CoreV1().ReplicationControllers(namespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

func runVolumeAntiAffinityPods(f *Framework, namespace string, pods int, id string, podLabels, antiAffinityLabels map[string]string, volumes []v1.Volume) error {
	config := &testutils.RCConfig{
		Affinity:  buildAntiAffinity(antiAffinityLabels),
		Volumes:   volumes,
		Client:    f.ClientSet,
		Name:      id,
		Namespace: namespace,
		Timeout:   scaleUpTimeout,
		Image:     imageutils.GetPauseImageName(),
		Replicas:  pods,
		Labels:    podLabels,
	}
	err := e2erc.RunRC(*config)
	if err != nil {
		return err
	}
	_, err = f.ClientSet.CoreV1().ReplicationControllers(namespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

var emptyDirVolumes = []v1.Volume{
	{
		Name: "empty-volume",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	},
}

func buildAntiAffinity(labels map[string]string) *v1.Affinity {
	return &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}
}

// Create an RC running a given number of pods on each node without adding any constraint forcing
// such pod distribution. This is meant to create a bunch of underutilized (but not unused) nodes
// with pods that can be rescheduled on different nodes.
// This is achieved using the following method:
// 1. disable scheduling on each node
// 2. create an empty RC
// 3. for each node:
// 3a. enable scheduling on that node
// 3b. increase number of replicas in RC by podsPerNode
func runReplicatedPodOnEachNode(f *Framework, nodes []v1.Node, namespace string, podsPerNode int, id string, labels map[string]string, memRequest int64) error {
	f.T.Log("Run a pod on each node")
	for _, node := range nodes {
		err := makeNodeUnschedulable(f, &node)

		defer func(n v1.Node) {
			makeNodeSchedulable(f, &n, false)
		}(node)

		if err != nil {
			return err
		}
	}
	config := &testutils.RCConfig{
		Client:     f.ClientSet,
		Name:       id,
		Namespace:  namespace,
		Timeout:    defaultTimeout,
		Image:      imageutils.GetPauseImageName(),
		Replicas:   0,
		Labels:     labels,
		MemRequest: memRequest,
	}
	err := e2erc.RunRC(*config)
	if err != nil {
		return err
	}
	rc, err := f.ClientSet.CoreV1().ReplicationControllers(namespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for i, node := range nodes {
		err = makeNodeSchedulable(f, &node, false)
		if err != nil {
			return err
		}

		// Update replicas count, to create new pods that will be allocated on node
		// (we retry 409 errors in case rc reference got out of sync)
		for j := 0; j < 3; j++ {
			*rc.Spec.Replicas = int32((i + 1) * podsPerNode)
			rc, err = f.ClientSet.CoreV1().ReplicationControllers(namespace).Update(context.TODO(), rc, metav1.UpdateOptions{})
			if err == nil {
				break
			}
			if !apierrors.IsConflict(err) {
				return err
			}
			f.T.Logf("Got 409 conflict when trying to scale RC, retries left: %v", 3-j)
			rc, err = f.ClientSet.CoreV1().ReplicationControllers(namespace).Get(context.TODO(), id, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		err = wait.PollImmediate(5*time.Second, podTimeout, func() (bool, error) {
			rc, err = f.ClientSet.CoreV1().ReplicationControllers(namespace).Get(context.TODO(), id, metav1.GetOptions{})
			if err != nil || rc.Status.ReadyReplicas < int32((i+1)*podsPerNode) {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return fmt.Errorf("failed to coerce RC into spawning a pod on node %s within timeout", node.Name)
		}
		err = makeNodeUnschedulable(f, &node)
		if err != nil {
			return err
		}
	}
	return nil
}

// Increase cluster size by newNodesForScaledownTests to create some unused nodes
// that can be later removed by cluster autoscaler.
func manuallyIncreaseClusterSize(f *Framework, originalSizes map[string]int) int {
	f.T.Log("Manually increase cluster size")
	increasedSize := 0
	newSizes := make(map[string]int)
	for key, val := range originalSizes {
		newSizes[key] = val + newNodesForScaledownTests
		increasedSize += val + newNodesForScaledownTests
	}
	setMigSizes(f, newSizes)

	checkClusterSize := func(size int) bool {
		if size >= increasedSize {
			return true
		}
		resized := setMigSizes(f, newSizes)
		if resized {
			f.T.Logf("Unexpected node group size while waiting for cluster resize. Setting size to target again.")
		}
		return false
	}

	assert.NoError(f.T, WaitForClusterSizeFunc(f, checkClusterSize, manualResizeTimeout))
	return increasedSize
}

type scaleUpStatus struct {
	status    string
	ready     int
	target    int
	timestamp time.Time
}

// Try to get timestamp from status.
// Status configmap is not parsing-friendly, so evil regexpery follows.
func getStatusTimestamp(status string) (time.Time, error) {
	timestampMatcher, err := regexp.Compile("Cluster-autoscaler status at \\s*([0-9\\-]+ [0-9]+:[0-9]+:[0-9]+\\.[0-9]+ \\+[0-9]+ [A-Za-z]+)")
	if err != nil {
		return time.Time{}, err
	}

	timestampMatch := timestampMatcher.FindStringSubmatch(status)
	if len(timestampMatch) < 2 {
		return time.Time{}, fmt.Errorf("Failed to parse CA status timestamp, raw status: %v", status)
	}

	timestamp, err := time.Parse(timestampFormat, timestampMatch[1])
	if err != nil {
		return time.Time{}, err
	}
	return timestamp, nil
}

// Try to get scaleup statuses of all node groups.
// Status configmap is not parsing-friendly, so evil regexpery follows.
func getScaleUpStatus(f *Framework) (*scaleUpStatus, error) {
	configMap, err := f.ClientSet.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "cluster-autoscaler-status", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	status, ok := configMap.Data["status"]
	if !ok {
		return nil, fmt.Errorf("Status information not found in configmap")
	}

	timestamp, err := getStatusTimestamp(status)
	if err != nil {
		return nil, err
	}

	matcher, err := regexp.Compile("s*ScaleUp:\\s*([A-Za-z]+)\\s*\\(ready=([0-9]+)\\s*cloudProviderTarget=([0-9]+)\\s*\\)")
	if err != nil {
		return nil, err
	}
	matches := matcher.FindAllStringSubmatch(status, -1)
	if len(matches) < 1 {
		return nil, fmt.Errorf("Failed to parse CA status configmap, raw status: %v", status)
	}

	result := scaleUpStatus{
		status:    caNoScaleUpStatus,
		ready:     0,
		target:    0,
		timestamp: timestamp,
	}
	for _, match := range matches {
		if match[1] == caOngoingScaleUpStatus {
			result.status = caOngoingScaleUpStatus
		}
		newReady, err := strconv.Atoi(match[2])
		if err != nil {
			return nil, err
		}
		result.ready += newReady
		newTarget, err := strconv.Atoi(match[3])
		if err != nil {
			return nil, err
		}
		result.target += newTarget
	}
	f.T.Logf("Cluster-Autoscaler scale-up status: %v (%v, %v)", result.status, result.ready, result.target)
	return &result, nil
}

func waitForScaleUpStatus(f *Framework, cond func(s *scaleUpStatus) bool, timeout time.Duration) (*scaleUpStatus, error) {
	var finalErr error
	var status *scaleUpStatus
	err := wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		status, finalErr = getScaleUpStatus(f)
		if finalErr != nil {
			return false, nil
		}
		if status.timestamp.Add(freshStatusLimit).Before(time.Now()) {
			// stale status
			finalErr = fmt.Errorf("Status too old")
			return false, nil
		}
		return cond(status), nil
	})
	if err != nil {
		err = fmt.Errorf("Failed to find expected scale up status: %v, last status: %v, final err: %v", err, status, finalErr)
	}
	return status, err
}

// This is a temporary fix to allow CA to migrate some kube-system pods
// TODO: Remove this when the PDB is added for some of those components
func addKubeSystemPdbs(f *Framework) (func(), error) {
	f.T.Log("Create PodDisruptionBudgets for kube-system components, so they can be migrated if required")

	var newPdbs []string
	cleanup := func() {
		var finalErr error
		for _, newPdbName := range newPdbs {
			f.T.Logf("Delete PodDisruptionBudget %v", newPdbName)
			err := f.ClientSet.PolicyV1beta1().PodDisruptionBudgets("kube-system").Delete(context.TODO(), newPdbName, metav1.DeleteOptions{})
			if err != nil {
				// log error, but attempt to remove other pdbs
				f.T.Logf("Failed to delete PodDisruptionBudget %v, err: %v", newPdbName, err)
				finalErr = err
			}
		}
		assert.NoErrorf(f.T, finalErr, "Error during PodDisruptionBudget cleanup: %v", finalErr)
	}

	type pdbInfo struct {
		label        string
		minAvailable int
	}
	pdbsToAdd := []pdbInfo{
		{label: "kube-dns", minAvailable: 1},
		{label: "kube-dns-autoscaler", minAvailable: 0},
		{label: "metrics-server", minAvailable: 0},
		{label: "kubernetes-dashboard", minAvailable: 0},
		{label: "glbc", minAvailable: 0},
	}
	for _, pdbData := range pdbsToAdd {
		f.T.Logf("Create PodDisruptionBudget for %v", pdbData.label)
		labelMap := map[string]string{"k8s-app": pdbData.label}
		pdbName := fmt.Sprintf("test-pdb-for-%v", pdbData.label)
		minAvailable := intstr.FromInt(pdbData.minAvailable)
		pdb := &policyv1beta1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pdbName,
				Namespace: "kube-system",
			},
			Spec: policyv1beta1.PodDisruptionBudgetSpec{
				Selector:     &metav1.LabelSelector{MatchLabels: labelMap},
				MinAvailable: &minAvailable,
			},
		}
		_, err := f.ClientSet.PolicyV1beta1().PodDisruptionBudgets("kube-system").Create(context.TODO(), pdb, metav1.CreateOptions{})
		newPdbs = append(newPdbs, pdbName)

		if err != nil {
			return cleanup, err
		}
	}
	return cleanup, nil
}

func createPriorityClasses(f *Framework) func() {
	priorityClasses := map[string]int32{
		expendablePriorityClassName: -15,
		highPriorityClassName:       1000,
	}
	for className, priority := range priorityClasses {
		_, err := f.ClientSet.SchedulingV1().PriorityClasses().Create(context.TODO(), &schedulingv1.PriorityClass{ObjectMeta: metav1.ObjectMeta{Name: className}, Value: priority}, metav1.CreateOptions{})
		if err != nil {
			f.T.Logf("Error creating priority class: %v", err)
		}
		assert.Equal(f.T, err == nil || apierrors.IsAlreadyExists(err), true)
	}

	return func() {
		for className := range priorityClasses {
			err := f.ClientSet.SchedulingV1().PriorityClasses().Delete(context.TODO(), className, metav1.DeleteOptions{})
			if err != nil {
				f.T.Logf("Error deleting priority class: %v", err)
			}
		}
	}
}
