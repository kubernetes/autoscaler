package integration

import (
	"context"
	"fmt"
	"os"
	"time"

	ginkgo "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	controlKubeconfig = os.Getenv("CONTROL_KUBECONFIG")
	targetKubeconfig  = os.Getenv("TARGET_KUBECONFIG")

	pollingTimeout  = 300 * time.Second
	pollingInterval = 2 * time.Second

	scaleUpWorkload      = "scale-up-pod"
	initialNumberOfNodes = 1
)

var driver = NewDriver(controlKubeconfig, targetKubeconfig)

var _ = ginkgo.BeforeSuite(driver.setupBeforeSuite)
var _ = ginkgo.AfterSuite(driver.cleanup)

var _ = ginkgo.Describe("Machine controllers test", func() {
	driver.beforeEachCheck(checkIfClusterAutoscalerUp)
	driver.controllerTests()
})

func (driver *Driver) beforeEachCheck(fn func()) {
	ginkgo.BeforeEach(fn)
}

func checkIfClusterAutoscalerUp() {
	ginkgo.By("Checking autoscaler process is running")
	gomega.Expect(autoscalerSession.ExitCode()).Should(gomega.Equal(-1))
}

func (driver *Driver) setupBeforeSuite() {
	driver.scaleAutoscaler(0)
	driver.adjustNodeGroups()
	driver.runAutoscaler()
}

func (driver *Driver) cleanup() {
	ginkgo.By("Running CleanUp")

	ginkgo.By("Deleting workload")
	err := driver.targetCluster.Clientset.AppsV1().Deployments("default").DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error cleaning up deployments in target cluster: %s", err.Error())
	}

	driver.adjustNodeGroups()

	ginkgo.By("Scaling CA back up to 1 in the Shoot namespace")
	err = driver.scaleAutoscaler(1)
	if err != nil {
		fmt.Printf("error scaling up cluster autoscaler deployment in control cluster Shoot namespace: %s", err.Error())
	}
}

func (driver *Driver) controllerTests() {
	ginkgo.Describe("Scale up and down nodes", func() {
		ginkgo.Context("by deploying new workload requesting more resources", func() {
			ginkgo.It("should not lead to any errors and add 1 more node in target cluster", func() {

				ginkgo.By("Deploying workload...")
				gomega.Expect(driver.deployWorkload(1)).To(gomega.BeNil())

				ginkgo.By("Validating Scale up")
				gomega.Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(gomega.BeNumerically("==", initialNumberOfNodes+1))
			})
		})

		ginkgo.Context("by scaling deployed workload to 3 replicas", func() {
			ginkgo.It("should not lead to any errors and add 3 more node in target cluster", func() {

				ginkgo.By("Scaling up workload to 3 replicas...")
				gomega.Expect(driver.scaleWorkload(scaleUpWorkload, 3)).To(gomega.BeNil())

				ginkgo.By("Validating Scale up")
				gomega.Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(gomega.BeNumerically("==", initialNumberOfNodes+3))
			})
		})

		ginkgo.Context("by scaling down the deployed workload to 0", func() {
			ginkgo.It("should not lead to any errors and 3 nodes to be removed from the target cluster", func() {

				ginkgo.By("Scaling down workload to zero...")
				gomega.Expect(driver.scaleWorkload(scaleUpWorkload, 0)).To(gomega.BeNil())

				ginkgo.By("Validating Scale down")
				gomega.Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(gomega.BeNumerically("==", initialNumberOfNodes))
			})
		})
	})
}
