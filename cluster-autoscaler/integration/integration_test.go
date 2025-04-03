package integration

import (
	"context"
	"fmt"
	"os"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	controlKubeconfig = os.Getenv("CONTROL_KUBECONFIG")
	targetKubeconfig  = os.Getenv("TARGET_KUBECONFIG")

	scaleUpWorkload = "scale-up-pod"
	maxNodes        = 4
)

var driver = NewDriver(controlKubeconfig, targetKubeconfig)

var flag = false

var _ = BeforeSuite(driver.setupBeforeSuite)
var _ = AfterSuite(driver.cleanup)

var _ = Describe("Cluster Autoscaler test", func() {
	driver.beforeEachCheck(checkIfClusterAutoscalerUp)
	driver.afterEachCheck(removeWorkload)
	driver.controllerTests()
})

func (driver *Driver) beforeEachCheck(fn func()) {
	BeforeEach(fn)
}

func (driver *Driver) afterEachCheck(fn func()) {
	AfterEach(fn)
}

func removeWorkload() {
	if flag {
		Expect(driver.deleteWorkload()).To(BeNil())
		By("Checking that number of Ready nodes is equal to initial")
		Eventually(
			driver.targetCluster.getNumberOfReadyNodes,
			pollingTimeout,
			pollingInterval).
			Should(BeNumerically("==", initialNumberOfNodes))
	}
	flag = false
}

func checkIfClusterAutoscalerUp() {
	By("Checking autoscaler process is running")
	Expect(autoscalerSession.ExitCode()).Should(Equal(-1))
}

func (driver *Driver) setupBeforeSuite() {
	// TODO: add gomega assertions, available from MCM release v0.49.0. Example https://github.com/gardener/machine-controller-manager/blob/a03d0fbcd28ef2265adb57d095c5d0d3333d6043/pkg/test/integration/common/framework.go#L1230-L1233
	driver.scaleAutoscaler(0)
	driver.adjustNodeGroups()
	driver.addTaintsToInitialNodes()
	driver.runAutoscaler()
}
func (driver *Driver) deleteWorkload() error {
	By("Deleting workload")
	err := driver.targetCluster.Clientset.AppsV1().Deployments("default").DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error cleaning up deployments in target cluster: %s", err.Error())
	}
	return err
}

func (driver *Driver) cleanup() {
	By("Running CleanUp")

	// TODO: add gomega assertions, available from MCM release v0.49.0. Example https://github.com/gardener/machine-controller-manager/blob/a03d0fbcd28ef2265adb57d095c5d0d3333d6043/pkg/test/integration/common/framework.go#L1230-L1233
	driver.deleteWorkload()
	driver.adjustNodeGroups()
	driver.deleteWorkload()
	driver.removeTaintsFromInitialNodes()

	By("Scaling CA back up to 1 in the Shoot namespace")
	err := driver.scaleAutoscaler(int32(1))
	if err != nil {
		fmt.Printf("error scaling up cluster autoscaler deployment in control cluster Shoot namespace: %s", err.Error())
	}
}

func (driver *Driver) controllerTests() {
	Describe("Scale up and down nodes", func() {
		Context("by deploying new workload requesting more resources", func() {
			It("should not lead to any errors and add 1 more node in target cluster", func() {

				By("Deploying workload...")
				Expect(driver.deployWorkload(1, scaleUpWorkload, workerWithThreeZones, false)).To(BeNil())

				By("Validating Scale up")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes+1))
			})
		})

		Context("by scaling deployed workload to 3 replicas", func() {
			It("should not lead to any errors and add 3 more node in target cluster", func() {

				By("Scaling up workload to 3 replicas...")
				Expect(driver.scaleWorkload(scaleUpWorkload, 3)).To(BeNil())

				By("Validating Scale up")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes+3))
			})
		})

		Context("by scaling down the deployed workload to 0", func() {
			It("should not lead to any errors and 3 nodes to be removed from the target cluster", func() {

				By("Scaling down workload to zero...")
				Expect(driver.scaleWorkload(scaleUpWorkload, 0)).To(BeNil())

				By("Validating Scale down")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes))
				flag = true
			})
		})
	})
	Describe("testing annotation to skip scaledown", func() {
		Context("by adding annotation and then scaling the workload to zero", func() {
			It("should not scale down the extra node and should log correspondingly", func() {
				By("adding the annotation after deploy workload to 1")
				Expect(driver.deployWorkload(1, scaleUpWorkload, workerWithThreeZones, false)).To(BeNil())
				By("Validating Scale up")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes+1))
				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					By("getting the latest added node and adding annotation to it.")
					_, latestNode, err := driver.getOldestAndLatestNode()
					Expect(err).To(BeNil())
					return driver.addAnnotationToNode(latestNode)
				})
				Expect(err).To(BeNil())
				By("Scaling down workload to zero...")
				Expect(driver.scaleWorkload(scaleUpWorkload, 0)).To(BeNil())
				skippedRegexp, _ := regexp.Compile(` the node is marked as no scale down`)
				Eventually(func() bool {
					data, _ := os.ReadFile(CALogFile)
					return skippedRegexp.Match(data)
				}, pollingTimeout, pollingInterval).Should(BeTrue())
			})

			It("Should remove the unwanted node once scale down disable annotation is removed", func() {
				Expect(driver.targetCluster.getNumberOfReadyNodes()).Should(BeNumerically("==", initialNumberOfNodes+1))
				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					By("getting the latest added node and removing annotation from it.")
					_, latestNode, err := driver.getOldestAndLatestNode()
					Expect(err).To(BeNil())
					return driver.removeAnnotationFromNode(latestNode)
				})
				Expect(err).To(BeNil())
				By("Validating Scale down")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes))
				flag = true
			})
		})
	})
	Describe("testing min and max limit for Cluster autoscaler", func() {
		Context("by increasing the workload to above max", func() {
			It("shouldn't scale beyond max number of workers", func() {
				By("Deploying workload with replicas = max+4")
				Expect(driver.deployWorkload(int32(maxNodes+4), scaleUpWorkload, workerWithThreeZones, false)).To(BeNil())
				By("Validating Scale up")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", maxNodes))
			})
		})
		Context("by decreasing the workload to below min", func() {
			It("shouldn't scale down beyond min number of workers", func() {
				By("Scaling down workload to zero...")
				Expect(driver.scaleWorkload(scaleUpWorkload, 0)).To(BeNil())

				By("Validating Scale down")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes))
				flag = true
			})
		})
	})
	Describe("testing scaling due to taints", func() {
		Context("make current nodes unschedulable", func() {
			It("should spawn more nodes for accommodating new workload", func() {
				By("making the only node available, to be unschedulable")
				_, latestNode, err := driver.getOldestAndLatestNode()
				Expect(err).To(BeNil())

				driver.addTaintsToNode(latestNode, map[string]bool{disabledTaint: true})

				By("Increasing the workload")
				Expect(driver.deploySmallWorkload(1, scaleUpWorkload, workerWithThreeZones, true)).To(BeNil())

				By("Validating Scale up")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes+1))
			})
			It("should remove the node as the taint has been removed and node has low utilization", func() {
				By("making the node available to be schedulable")

				oldestNode, _, err := driver.getOldestAndLatestNode()
				Expect(err).To(BeNil())
				driver.removeTaintsFromNode(oldestNode, false, map[string]bool{disabledTaint: true})

				By("Validating Scale down")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes))
				flag = true
			})
		})
	})
	Describe("testing scaling due to volume pending", func() {
		Context("create a volume in a zone with no node and a pod requesting it", func() {
			It("should create a node in the zone with volume and hence scale up", func() {

				By("Creating StorageClass with topology restrictions")
				class := "myclass"
				provider := os.Getenv("CLUSTER_PROVIDER")
				if provider != "aws" && provider != "gcp" {
					return
				}
				storageClass, err1 := getStorageClassObject(class)

				if err1 != nil {
					fmt.Printf("StorageClassObject creation error: %v", err1)
					return
				}

				_, err := driver.targetCluster.Clientset.StorageV1().StorageClasses().Create(context.TODO(), storageClass, metav1.CreateOptions{})
				if err != nil {
					fmt.Printf("StorageClass Create API error: %v", err)
				}

				defer func() {
					By("Removing storage Class created earlier")
					err := driver.targetCluster.Clientset.StorageV1().StorageClasses().Delete(context.TODO(), class, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("StorageClass Delete API error: %v", err)
					}
				}()

				By("deploying PVC in zone with no nodes")

				claimName := "myclaim"
				pvc := getPvcObject(claimName, class)

				_, err = driver.targetCluster.Clientset.CoreV1().PersistentVolumeClaims("default").Create(context.TODO(), pvc, metav1.CreateOptions{})
				if err != nil {
					fmt.Printf("PVC Create API error: %v", err)
				}

				defer func() {
					By("Removing pvc created earlier")
					err := driver.targetCluster.Clientset.CoreV1().PersistentVolumeClaims("default").Delete(context.TODO(), claimName, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("StorageClass Delete API error: %v", err)
					}
				}()

				By("deploying the workload which requires the volume")

				deploymentName := "volume-pod"
				deployment := getDeploymentObjectWithVolumeReq(deploymentName, claimName)

				_, err = driver.targetCluster.Clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
				if err != nil {
					fmt.Printf("Deployment Create API error: %v", err)
				}

				By("Validation scale up to +1 in a new Zone")
				Eventually(
					driver.targetCluster.getNumberOfReadyNodes,
					pollingTimeout,
					pollingInterval).
					Should(BeNumerically("==", initialNumberOfNodes+1))
				flag = true
			})
		})
	})
	Describe("testing not able to scale due to excess demand", func() {
		Context("create a pod requiring more resources than a single machine can provide", func() {
			It("shouldn't scale up and log the error", func() {
				By("Deploying the workload")
				Expect(driver.deployLargeWorkload(1, scaleUpWorkload, workerWithThreeZones, false)).To(BeNil())
				By("checking that scale up didn't trigger because of no machine satisfying the requirement")
				skippedRegexp, _ := regexp.Compile("Pod default/large-scale-up-pod-.* can't be scheduled on .*, predicate checking error: .*predicate \"NodeResourcesFit\" didn't pass .*Insufficient cpu")
				Eventually(func() bool {
					data, _ := os.ReadFile(CALogFile)
					return skippedRegexp.Match(data)
				}, pollingTimeout, pollingInterval).Should(BeTrue())
				flag = true
			})
		})
	})

	Describe("testing CA behaviour when MCM is offline", func() {
		Context("When the available replicas of MCM are zero.", func() {
			It("The CA should suspend it's operations as long as MCM is offline", func() {
				By("Scaling down the MCM")
				deployment, err := driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Get(context.Background(), "machine-controller-manager", metav1.GetOptions{})
				Expect(err).Should(BeNil())
				deployment.Spec.Replicas = ptr.To(int32(0))
				deployment.ObjectMeta.Annotations[dwdAnnotation] = "true"
				_, err = driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
				Expect(err).Should(BeNil())
				skippedRegexp, _ := regexp.Compile("machine-controller-manager is offline. Cluster autoscaler operations would be suspended.")
				Eventually(func() bool {
					data, _ := os.ReadFile(CALogFile)
					return skippedRegexp.Match(data)
				}, pollingTimeout, pollingInterval).Should(BeTrue())
				deployment, err = driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Get(context.Background(), "machine-controller-manager", metav1.GetOptions{})
				Expect(err).Should(BeNil())
				deployment.Spec.Replicas = ptr.To(int32(1))
				delete(deployment.ObjectMeta.Annotations, dwdAnnotation)
				_, err = driver.controlCluster.Clientset.AppsV1().Deployments(controlClusterNamespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
				Expect(err).Should(BeNil())
				flag = true
			})
		})
	})
}
