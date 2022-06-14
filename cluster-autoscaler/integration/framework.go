package integration

import (
	"context"
	"fmt"
	"k8s.io/client-go/util/retry"
	"os"
	"os/exec"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	//annotaion which makes dwd skip the scaling of component
	dwdAnnotation string = "dependency-watchdog.gardener.cloud/ignore-scaling"
	//annotaion to skip the scaling down of exrta/unused node.
	ignoreScaledownAnnotation string = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
)

var (
	criticalAddonsOnlyTaint = "CriticalAddonsOnly"
	disabledTaint           = "DisabledForAutoscalingTest"
	smallMemory             = "50Mi"
	mediumMemory            = "150Mi"
	largeMemory             = "500Mi"
	smallCPU                = "500m"
	cpuResource             int64
)

// rotateLogFile takes file name as input and returns a file object obtained by os.Create
// If the file exists already then it renames it so that a new file can be created
func rotateLogFile(fileName string) (*os.File, error) {

	if _, err := os.Stat(fileName); err == nil { // !strings.Contains(err.Error(), "no such file or directory") {
		for i := 9; i > 0; i-- {
			os.Rename(fmt.Sprintf("%s.%d", fileName, i), fmt.Sprintf("%s.%d", fileName, i+1))
		}
		os.Rename(fileName, fmt.Sprintf("%s.%d", fileName, 1))
	}

	return os.Create(fileName)
}

func (driver *Driver) adjustNodeGroups() error {

	if driver.targetCluster.getNumberOfReadyNodes() == 1 {
		return nil
	}

	machineDeployments, err := driver.controlCluster.MCMClient.MachineV1alpha1().MachineDeployments(controlClusterNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for index, machineDeployment := range machineDeployments.Items {
		scaleDownMachineDeployment := machineDeployment.DeepCopy()
		if index > 0 && machineDeployment.Spec.Replicas != 0 {
			scaleDownMachineDeployment.Spec.Replicas = 0
		} else if index == 0 && machineDeployment.Spec.Replicas > 1 {
			scaleDownMachineDeployment.Spec.Replicas = 1
		}

		_, err := driver.controlCluster.MCMClient.MachineV1alpha1().MachineDeployments(controlClusterNamespace).Update(context.Background(), scaleDownMachineDeployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	By("Adjusting node groups to initial required size")
	Eventually(
		driver.targetCluster.getNumberOfReadyNodes,
		pollingTimeout,
		pollingInterval).
		Should(BeNumerically("==", initialNumberOfNodes))

	return nil
}

//getNumberOfReadyNodes tries to retrieve the list of node objects in the cluster.
func (c *Cluster) getNumberOfReadyNodes() int16 {
	nodes, _ := c.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	count := 0
	for _, n := range nodes.Items {
		cpuResource, _ = (n.Status.Capacity.Cpu()).AsInt64()
		for _, nodeCondition := range n.Status.Conditions {
			if nodeCondition.Type == "Ready" && nodeCondition.Status == "True" {
				count++
			}
		}
	}
	return int16(count)
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

	if len(machineDeployments.Items) > 3 {
		fmt.Printf("Cluster node group configuration is improper. Setup Before Suite might not have successfully run. Please check!")
		return
	}

	By("Starting Cluster Autoscaler....")
	args := strings.Fields(
		fmt.Sprintf(
			"make --directory=%s start TARGET_KUBECONFIG=%s MACHINE_DEPLOYMENT_ZONE_1=%s MACHINE_DEPLOYMENT_ZONE_2=%s MACHINE_DEPLOYMENT_ZONE_3=%s LEADER_ELECT=%s",
			"../",
			driver.targetCluster.KubeConfigFilePath,
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[0].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[1].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[2].Name),
			"false",
		),
	)

	outputFile, err := rotateLogFile(CALogFile)
	Expect(err).ShouldNot(HaveOccurred())
	autoscalerSession, err = gexec.Start(exec.Command(args[0], args[1:]...), outputFile, outputFile)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(autoscalerSession.ExitCode()).Should(Equal(-1))
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
	var zone []string = []string{os.Getenv("VOLUME_ZONE")}
	var expressions []v1.TopologySelectorLabelRequirement = []v1.TopologySelectorLabelRequirement{{Key: "topology.kubernetes.io/zone", Values: zone}}
	// for GKE
	// var expressions []v1.TopologySelectorLabelRequirement = []v1.TopologySelectorLabelRequirement{{Key: "topology.gke.io/zone", Values: zone}}
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
			Resources:        v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceName(v1.ResourceStorage): resource.MustParse(cap)}},
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

func getDeploymentObject(replicas int32, resourceCpu string, resourceMemory string, workloadName string) *appv1.Deployment {
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
									v1.ResourceCPU:    resource.MustParse(resourceCpu),
									v1.ResourceMemory: resource.MustParse(resourceMemory),
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

func (driver *Driver) deployWorkload(replicas int32, workloadName string) error {
	deployment := getDeploymentObject(replicas, fmt.Sprint(cpuResource-1), mediumMemory, workloadName)
	_, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) deployLargeWorkload(replicas int32, workloadName string) error {
	deployment := getDeploymentObject(replicas, fmt.Sprint(cpuResource+1), largeMemory, "large-"+workloadName)
	_, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) deploySmallWorkload(replicas int32, workloadName string) error {
	deployment := getDeploymentObject(replicas, smallCPU, smallMemory, "small-"+workloadName)
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
	nodeList, err := driver.targetCluster.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	if len(nodeList.Items) < 1 {
		err = fmt.Errorf("no nodes found")
	}
	if err != nil {
		return nil, nil, err
	}
	//sorting in ascending order of creation timeStamp
	sort.Slice(nodeList.Items, func(i, j int) bool {
		return nodeList.Items[i].ObjectMeta.CreationTimestamp.Before(&nodeList.Items[j].ObjectMeta.CreationTimestamp)
	})
	oldestNode := nodeList.Items[0]
	latestNode := nodeList.Items[len(nodeList.Items)-1]
	return &oldestNode, &latestNode, nil
}

func (driver *Driver) addAnnotationToNode(node *v1.Node) error {
	node.ObjectMeta.Annotations[ignoreScaledownAnnotation] = "true"
	_, err := driver.targetCluster.Clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) removeAnnotationFromNode(node *v1.Node) error {
	delete(node.ObjectMeta.Annotations, ignoreScaledownAnnotation)
	_, err := driver.targetCluster.Clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) makeNodeUnschedulable(node *v1.Node) error {
	By(fmt.Sprintf("Taint node %s", node.Name))
	for j := 0; j < 3; j++ {
		freshNode, err := driver.targetCluster.Clientset.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
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

// CriticalAddonsOnlyError implements the `error` interface, and signifies the
// presence of the `CriticalAddonsOnly` taint on the node.
type CriticalAddonsOnlyError struct{}

func (CriticalAddonsOnlyError) Error() string {
	return "criticalAddonsOnly taint found on node"
}

func (driver *Driver) makeNodeSchedulable(node *v1.Node, failOnCriticalAddonsOnly bool) error {
	By(fmt.Sprintf("Remove taint from node %s", node.Name))
	for j := 0; j < 3; j++ {
		freshNode, err := driver.targetCluster.Clientset.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
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
