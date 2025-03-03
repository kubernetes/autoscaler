package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"k8s.io/client-go/util/retry"

	gin "github.com/onsi/ginkgo/v2"
	gom "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1storage "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const (
	// dwdAnnotation annotation which makes dwd skip the scaling of component
	dwdAnnotation string = "dependency-watchdog.gardener.cloud/ignore-scaling"
	// ignoreScaleDownAnnotation is the annotation to skip the scaling down of extra/unused node.
	ignoreScaleDownAnnotation string = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
	mcmPriorityAnnotation            = "machinepriority.machine.sapcloud.io"
	mcdNameLabel                     = "name"
	workerLabelKey                   = "worker.garden.sapcloud.io/group"
	workerWithOneZone                = "one-zone"
	workerWithThreeZones             = "three-zones"
	workerForSystemComponents        = "sys-comp"
	pollingTimeout                   = 300 * time.Second
	pollingInterval                  = 2 * time.Second
	initialNumberOfNodes             = 2
)

var (
	criticalAddonsOnlyTaint             = "CriticalAddonsOnly"
	disabledTaint                       = "DisabledForAutoscalingTest"
	blockInitialNodesForSchedulingTaint = "testing.node.gardener.cloud/initial-node-blocked"
	smallMemory                         = *resource.NewQuantity(50*1024*1024, resource.BinarySI)
	mediumMemory                        = *resource.NewQuantity(150*1024*1024, resource.BinarySI)
	largeMemory                         = *resource.NewQuantity(500*1024*1024, resource.BinarySI)
	smallCPU                            = *resource.NewMilliQuantity(500, resource.DecimalSI)
	cpuResource                         *resource.Quantity
	tolerationsToInitialNodeTaint       = []v1.Toleration{
		{
			Key:      blockInitialNodesForSchedulingTaint,
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}
)

// rotateLogFile takes file name as input and returns a file object obtained by os.Create
// If the file exists already then it renames it so that a new file can be created
func rotateLogFile(fileName string) (*os.File, error) {

	if _, err := os.Stat(fileName); err == nil { // !strings.Contains(err.Error(), "no such file or directory") {
		for i := 9; i > 0; i-- {
			_ = os.Rename(fmt.Sprintf("%s.%d", fileName, i), fmt.Sprintf("%s.%d", fileName, i+1))
		}
		_ = os.Rename(fileName, fmt.Sprintf("%s.%d", fileName, 1))
	}

	return os.Create(fileName) //#nosec G304 (CWE-22) -- this is used only for tests. Cannot be exploited
}

func (driver *Driver) addTaintsToInitialNodes() error {
	gin.By("Marking nodes present before the tests as unschedulable")
	nodes, _ := driver.targetCluster.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	for _, n := range nodes.Items {
		if err := driver.addTaintsToNode(&n, map[string]bool{blockInitialNodesForSchedulingTaint: true}); err != nil {
			return fmt.Errorf("some initial nodes might be tainted, tainting node %s failed with err: %q , aborting operation", n.Name, err)
		}
	}
	return nil
}

func (driver *Driver) removeTaintsFromInitialNodes() error {
	gin.By("Turning nodes present before the tests, back to schedulable")
	nodes, _ := driver.targetCluster.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	// marking every node schedulable as this method is called either in BeforeSuite() or AfterSuite()
	for _, n := range nodes.Items {
		if err := driver.removeTaintsFromNode(&n, false, map[string]bool{blockInitialNodesForSchedulingTaint: true}); err != nil {
			return fmt.Errorf("some initial nodes might be left tainted, removing taint from node %s failed with err: %q , aborting operation", n.Name, err)
		}
	}
	return nil
}

func (driver *Driver) adjustNodeGroups() error {

	if driver.targetCluster.getNumberOfReadyNodes() == 1 {
		return nil
	}

	machineDeployments, err := driver.controlCluster.MCMClient.MachineV1alpha1().MachineDeployments(controlClusterNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	isFirstMDOfWorkerWithThreeZones := true
	for _, machineDeployment := range machineDeployments.Items {
		scaleDownMachineDeployment := machineDeployment.DeepCopy()
		if strings.Contains(machineDeployment.Name, workerForSystemComponents) {
			continue
		}
		if strings.Contains(machineDeployment.Name, workerWithThreeZones) && isFirstMDOfWorkerWithThreeZones {
			isFirstMDOfWorkerWithThreeZones = false
			scaleDownMachineDeployment.Spec.Replicas = 1
		} else {
			scaleDownMachineDeployment.Spec.Replicas = 0
		}
		_, err := driver.controlCluster.MCMClient.MachineV1alpha1().MachineDeployments(controlClusterNamespace).Update(context.Background(), scaleDownMachineDeployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	gin.By("Adjusting node groups to initial required size")
	gom.Eventually(
		driver.targetCluster.getNumberOfReadyNodes,
		pollingTimeout,
		pollingInterval).
		Should(gom.BeNumerically("==", initialNumberOfNodes))

	return nil
}

// getNumberOfReadyNodes tries to retrieve the list of node objects in the cluster.
func (c *Cluster) getNumberOfReadyNodes() int {
	nodes, _ := c.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	count := 0
	for _, n := range nodes.Items {
		cpuResource = n.Status.Capacity.Cpu()
		for _, nodeCondition := range n.Status.Conditions {
			if nodeCondition.Type == "Ready" && nodeCondition.Status == "True" {
				count++
			}
		}
	}
	return count
}

func (driver *Driver) scaleAutoscaler(replicas int32) error {
	autoscalerDeployment, err := driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Get(context.Background(), "cluster-autoscaler", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if replicas > 1 {
		replicas = 1
	}
	if replicas == 0 {
		autoscalerDeployment.ObjectMeta.Annotations[dwdAnnotation] = "true"
	} else if replicas == 1 {
		delete(autoscalerDeployment.ObjectMeta.Annotations, dwdAnnotation)
	}
	if autoscalerDeployment.Spec.Replicas != pointer.Int32Ptr(replicas) {
		autoscalerDeployment.Spec.Replicas = pointer.Int32Ptr(replicas)
		fmt.Printf("Scaling Cluster Autoscaler to %d replicas\n", replicas)
		_, err = driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Update(context.Background(), autoscalerDeployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// time.Sleep(30 * time.Second)

	return nil
}

// runAutoscaler run the machine controller and machine controller manager binary locally
func (driver *Driver) runAutoscaler() {

	machineDeployments, err := driver.controlCluster.MCMClient.MachineV1alpha1().MachineDeployments(controlClusterNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}

	if len(machineDeployments.Items) != 5 {
		fmt.Printf("Cluster node group configuration is improper. Setup Before Suite might not have successfully run. Please check!")
		return
	}

	gin.By("Starting Cluster Autoscaler....")
	args := strings.Fields(
		fmt.Sprintf(
			"make --directory=%s start TARGET_KUBECONFIG=%s MACHINE_DEPLOYMENT_1_ZONE_1=%s MACHINE_DEPLOYMENT_2_ZONE_1=%s MACHINE_DEPLOYMENT_2_ZONE_2=%s MACHINE_DEPLOYMENT_2_ZONE_3=%s LEADER_ELECT=%s",
			"../",
			driver.targetCluster.KubeConfigFilePath,
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[0].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[2].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[3].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[4].Name),
			"false",
		),
	)

	outputFile, err := rotateLogFile(CALogFile)
	gom.Expect(err).ShouldNot(gom.HaveOccurred())
	autoscalerSession, err = gexec.Start(exec.Command(args[0], args[1:]...), outputFile, outputFile) //#nosec G204 (CWE-78) -- this is used only for tests. Cannot be exploited
	gom.Expect(err).ShouldNot(gom.HaveOccurred())
	gom.Expect(autoscalerSession.ExitCode()).Should(gom.Equal(-1))
}

func getStorageClassObject(class string) (*v1storage.StorageClass, error) {
	provider := os.Getenv("CLUSTER_PROVIDER")
	var provisioner string
	switch provider {
	case "aws":
		provisioner = "ebs.csi.aws.com"
	case "azure":
		provisioner = "disk.csi.azure.com"
	case "gcp":
		provisioner = "pd.csi.storage.gke.io"
	default:
		return nil, fmt.Errorf("invalid cluster provider")
	}
	allowExpansion := true
	volumeReclaimPolicy := v1.PersistentVolumeReclaimDelete
	volumeBindingMode := v1storage.VolumeBindingImmediate
	var zone = []string{os.Getenv("VOLUME_ZONE")}
	var expressions []v1.TopologySelectorLabelRequirement
	if provider == "aws" {
		expressions = []v1.TopologySelectorLabelRequirement{{Key: "topology.ebs.csi.aws.com/zone", Values: zone}}
	} else if provider == "gcp" {
		expressions = []v1.TopologySelectorLabelRequirement{{Key: "topology.gke.io/zone", Values: zone}}
	}
	var topologies []v1.TopologySelectorTerm = []v1.TopologySelectorTerm{{MatchLabelExpressions: expressions}}

	storageClass := &v1storage.StorageClass{
		AllowVolumeExpansion: &allowExpansion,
		ObjectMeta: metav1.ObjectMeta{
			Name: class,
			Labels: map[string]string{
				"shoot.gardener.cloud/no-cleanup": "true",
			},
			Annotations: map[string]string{
				"resources.gardener.cloud/delete-on-invalid-update": "true",
			},
		},
		Provisioner:       provisioner,
		ReclaimPolicy:     &volumeReclaimPolicy,
		VolumeBindingMode: &volumeBindingMode,
		AllowedTopologies: topologies,
	}
	return storageClass, nil
}

func getPvcObject(claimName string, class string) *v1.PersistentVolumeClaim {
	namespace := "default"
	cap := "2Gi"
	var mode []v1.PersistentVolumeAccessMode
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources:        v1.VolumeResourceRequirements{Requests: v1.ResourceList{v1.ResourceName(v1.ResourceStorage): resource.MustParse(cap)}},
			AccessModes:      append(mode, v1.ReadWriteOnce),
			StorageClassName: &class,
		},
	}
	return pvc
}

func getDeploymentObjectWithVolumeReq(deploymentName string, claimName string) *appv1.Deployment {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: "default",
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Replicas: pointer.Int32Ptr(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "ngnix-container",
							Image: "nginx:latest",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									MountPath: "/var/www/html",
									Name:      "mypd",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "mypd",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: claimName,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func getDeploymentObject(replicas int32, resourceCPU resource.Quantity, resourceMemory resource.Quantity, workloadName, workerName string, tolerations []v1.Toleration) *appv1.Deployment {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: "default",
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Replicas: pointer.Int32Ptr(replicas),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "ngnix-container",
							Image: "nginx:latest",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resourceCPU,
									v1.ResourceMemory: resourceMemory,
								},
							},
						},
					},
					Tolerations:  tolerations,
					NodeSelector: map[string]string{workerLabelKey: workerName},
				},
			},
		},
	}
	return deployment
}

func (driver *Driver) deployWorkload(replicas int32, workloadName, workerName string, canTolerateTaintPlacedOnInitialNodes bool) error {
	// TODO(himanshu-kun): Remove such a dependency on approximating system components space . This changes over time.
	assumedSystemComponentsUsedSpace := *resource.NewMilliQuantity(1000, resource.DecimalSI)
	approxCPURequested := cpuResource.DeepCopy()
	approxCPURequested.Sub(assumedSystemComponentsUsedSpace)
	var deployment *appv1.Deployment
	if canTolerateTaintPlacedOnInitialNodes {
		deployment = getDeploymentObject(replicas, approxCPURequested, mediumMemory, workloadName, workerName, tolerationsToInitialNodeTaint)
	} else {
		deployment = getDeploymentObject(replicas, approxCPURequested, mediumMemory, workloadName, workerName, nil)
	}
	_, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) deployLargeWorkload(replicas int32, workloadName, workerName string, canTolerateTaintPlacedOnInitialNodes bool) error {
	// TODO(himanshu-kun): Remove such an approximation
	extraCPURequest := *resource.NewMilliQuantity(1000, resource.DecimalSI)
	cpuRequested := cpuResource.DeepCopy()
	cpuRequested.Add(extraCPURequest)
	var deployment *appv1.Deployment
	if canTolerateTaintPlacedOnInitialNodes {
		deployment = getDeploymentObject(replicas, cpuRequested, largeMemory, "large-"+workloadName, workerName, tolerationsToInitialNodeTaint)
	} else {
		deployment = getDeploymentObject(replicas, cpuRequested, largeMemory, "large-"+workloadName, workerName, nil)
	}
	_, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) deploySmallWorkload(replicas int32, workloadName, workerName string, canTolerateTaintPlacedOnInitialNodes bool) error {
	var deployment *appv1.Deployment
	if canTolerateTaintPlacedOnInitialNodes {
		deployment = getDeploymentObject(replicas, smallCPU, smallMemory, "small-"+workloadName, workerName, tolerationsToInitialNodeTaint)
	} else {
		deployment = getDeploymentObject(replicas, smallCPU, smallMemory, "small-"+workloadName, workerName, nil)
	}
	_, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) scaleWorkload(workloadName string, replicas int32) (err error) {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Get(context.Background(), workloadName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		deployment.Spec.Replicas = pointer.Int32Ptr(replicas)

		_, err = driver.targetCluster.Clientset.AppsV1().Deployments("default").Update(context.Background(), deployment, metav1.UpdateOptions{})
		return err
	})
}

func (driver *Driver) getOldestAndLatestNode() (*v1.Node, *v1.Node, error) {
	nodeList, err := driver.targetCluster.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{workerLabelKey: workerWithThreeZones},
	})})
	if err != nil {
		return nil, nil, err
	}
	if len(nodeList.Items) < 1 {
		err = fmt.Errorf("no nodes found")
	}
	if err != nil {
		return nil, nil, err
	}
	// sorting in ascending order of creation timeStamp
	sort.Slice(nodeList.Items, func(i, j int) bool {
		return nodeList.Items[i].ObjectMeta.CreationTimestamp.Before(&nodeList.Items[j].ObjectMeta.CreationTimestamp)
	})
	oldestNode := nodeList.Items[0]
	latestNode := nodeList.Items[len(nodeList.Items)-1]
	return &oldestNode, &latestNode, nil
}

func (driver *Driver) addAnnotationToNode(node *v1.Node) error {
	node.ObjectMeta.Annotations[ignoreScaleDownAnnotation] = "true"
	_, err := driver.targetCluster.Clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) removeAnnotationFromNode(node *v1.Node) error {
	delete(node.ObjectMeta.Annotations, ignoreScaleDownAnnotation)
	_, err := driver.targetCluster.Clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// CriticalAddonsOnlyError implements the `error` interface, and signifies the
// presence of the `CriticalAddonsOnly` taint on the node.
type CriticalAddonsOnlyError struct{}

func (CriticalAddonsOnlyError) Error() string {
	return "criticalAddonsOnly taint found on node"
}

func (driver *Driver) addTaintsToNode(node *v1.Node, taintKeysToAdd map[string]bool) error {
	gin.By(fmt.Sprintf("Taint node %s", node.Name))
	for j := 0; j < 3; j++ {
		freshNode, err := driver.targetCluster.Clientset.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for taintKey := range taintKeysToAdd {
			freshNode.Spec.Taints = append(freshNode.Spec.Taints, v1.Taint{
				Key:    taintKey,
				Value:  "true",
				Effect: v1.TaintEffectNoSchedule,
			})
		}

		_, err = driver.targetCluster.Clientset.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}
		if !apierrors.IsConflict(err) {
			return err
		}
		klog.Warningf("Got 409 conflict when trying to taint node, retries left: %v", 3-j)
	}
	return fmt.Errorf("failed to taint node in allowed number of retries")
}

func (driver *Driver) removeTaintsFromNode(node *v1.Node, failIfCriticalAddonsOnlyTaintPresent bool, taintKeysToRemove map[string]bool) error {
	if len(taintKeysToRemove) == 0 {
		return nil
	}

	gin.By(fmt.Sprintf("Removing taint(s) from node %s", node.Name))
	for j := 0; j < 3; j++ {
		freshNode, err := driver.targetCluster.Clientset.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		var newTaints []v1.Taint
		for _, taint := range freshNode.Spec.Taints {
			if failIfCriticalAddonsOnlyTaintPresent && taint.Key == criticalAddonsOnlyTaint {
				return CriticalAddonsOnlyError{}
			}
			if _, present := taintKeysToRemove[taint.Key]; !present {
				newTaints = append(newTaints, taint)
			}
		}

		freshNode.Spec.Taints = newTaints
		_, err = driver.targetCluster.Clientset.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}
		if !apierrors.IsConflict(err) {
			return err
		}
		klog.Warningf("Got 409 conflict when trying to taint node, retries left: %v", 3-j)
	}
	return fmt.Errorf("failed to remove taint from node in allowed number of retries")
}
