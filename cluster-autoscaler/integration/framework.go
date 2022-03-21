package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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

	ginkgo.By("Adjusting node groups to initial required size")
	gomega.Eventually(
		driver.targetCluster.getNumberOfReadyNodes,
		pollingTimeout,
		pollingInterval).
		Should(gomega.BeNumerically("==", initialNumberOfNodes))

	return nil
}

//getNumberOfReadyNodes tries to retrieve the list of node objects in the cluster.
func (c *Cluster) getNumberOfReadyNodes() int16 {
	nodes, _ := c.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	count := 0
	for _, n := range nodes.Items {
		for _, nodeCondition := range n.Status.Conditions {
			if nodeCondition.Type == "Ready" && nodeCondition.Status == "True" {
				count++
			}
		}
	}
	return int16(count)
}

func (driver *Driver) scaleAutoscaler(replicas int32) error {
	autoScalerDeployment, err := driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Get(context.Background(), "cluster-autoscaler", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if replicas > 1 {
		replicas = 1
	}

	if autoScalerDeployment.Spec.Replicas != pointer.Int32Ptr(replicas) {
		autoScalerDeployment.Spec.Replicas = pointer.Int32Ptr(replicas)
		fmt.Printf("Scaling Cluster Autoscaler to %d replicas\n", replicas)
		_, err = driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Update(context.Background(), autoScalerDeployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// time.Sleep(30 * time.Second)

	return nil
}

// runAutoscaler run the machine controller and machine controller manager binary locally
func (c *Driver) runAutoscaler() {

	machineDeployments, err := c.controlCluster.MCMClient.MachineV1alpha1().MachineDeployments(controlClusterNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}

	if len(machineDeployments.Items) > 3 {
		fmt.Printf("Cluster node group configuration is improper. Setup Before Suite might not have successfully run. Please check!")
		return
	}

	ginkgo.By("Starting Cluster Autoscaler....")
	args := strings.Fields(
		fmt.Sprintf(
			"make --directory=%s start TARGET_KUBECONFIG=%s MACHINE_DEPLOYMENT_ZONE_1=%s MACHINE_DEPLOYMENT_ZONE_2=%s MACHINE_DEPLOYMENT_ZONE_3=%s LEADER_ELECT=%s",
			"../",
			c.targetCluster.KubeConfigFilePath,
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[0].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[1].Name),
			fmt.Sprintf("%s.%s", controlClusterNamespace, machineDeployments.Items[2].Name),
			"false",
		),
	)

	outputFile, err := rotateLogFile(CALogFile)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	autoscalerSession, err = gexec.Start(exec.Command(args[0], args[1:]...), outputFile, outputFile)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(autoscalerSession.ExitCode()).Should(gomega.Equal(-1))
}

func getDeploymentObject(replicas int32) *appv1.Deployment {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scaleUpWorkload,
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
							// TODO: this is the object to be dyamically changed based on the machine type
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("3"),
									v1.ResourceMemory: resource.MustParse("150Mi"),
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

func (driver *Driver) deployWorkload(replicas int32) error {
	deployment := getDeploymentObject(replicas)
	_, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) scaleWorkload(workloadName string, replicas int32) error {
	deployment, err := driver.targetCluster.Clientset.AppsV1().Deployments("default").Get(context.Background(), workloadName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Replicas = pointer.Int32Ptr(replicas)

	_, err = driver.targetCluster.Clientset.AppsV1().Deployments("default").Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
