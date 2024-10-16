package e2e_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AgentPool Properties", func() {
	var (
		namespace *corev1.Namespace
	)

	BeforeEach(func() {
		Eventually(allVMSSStable, "10m", "30s").Should(Succeed())

		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "azure-e2e-",
			},
		}
		Expect(k8s.Create(ctx, namespace)).To(Succeed())
	})

	Context("MinCount", func() {
		It("should enforce min count if nodepool count is less than min size", func() {
			ensureHelmValues(map[string]interface{}{
			"extraArgs": map[string]interface{}{
				"scale-down-delay-after-add":       "10s",
				"scale-down-unneeded-time":         "10s",
				"scale-down-candidates-pool-ratio": "1.0",
				"unremovable-node-recheck-timeout": "10s",
				"enforce-node-group-min-size": "true",
				"skip-nodes-with-system-pods":      "false",
				"skip-nodes-with-local-storage":    "false",
			},
		})

			agentPoolName := "enforcemin"
			agentPool := armcontainerservice.AgentPool{
				Name: lo.ToPtr(agentPoolName),
				Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
					MinCount:      lo.ToPtr[int32](5),
					Count:         lo.ToPtr[int32](1),
					MaxCount:      lo.ToPtr[int32](10),
					ScaleDownMode: lo.ToPtr(armcontainerservice.ScaleDownModeDelete),
				},
			}
			
			ap, err := CreateAgentpool(ctx, agentPoolClient, customerResourceGroup, clusterName, agentPoolName, agentPool)
			if err != nil {
				fmt.Fprintf(GinkgoWriter, "Error creating agent pool: %v\n", err)
			}
			Expect(err).To(BeNil())
			// Sanity Checks
			Expect(ap.Properties.MinCount).To(Equal(5))
			Expect(ap.Properties.MaxCount).To(Equal(10))
			Expect(ap.Properties.ScaleDownMode).To(Equal(armcontainerservice.ScaleDownModeDelete))
			Expect(ap.Name).To(Equal(agentPoolName))

			Eventually(func() (int, error) {
				nodes := &corev1.NodeList{}
				if err := k8s.List(ctx, nodes); err != nil {
					return 0, err
				}

				count := 0
				for _, node := range nodes.Items {
					if val, ok := node.Labels["agentpool"]; ok && val == "enforcemin" {
						count++
					}
				}
				return count, nil
			}).WithOffset(1).WithTimeout(10 * time.Minute).Should(Equal(5))

		})

	})
	AfterEach(func() {
		Expect(k8s.Delete(ctx, namespace)).To(Succeed())
		Eventually(func() bool {
			err := k8s.Get(ctx, client.ObjectKeyFromObject(namespace), &corev1.Namespace{})
			return apierrors.IsNotFound(err)
		}, "1m", "5s").Should(BeTrue(), "Namespace "+namespace.Name+" still exists")
	})

})

func CreateAgentpool(ctx context.Context, client *armcontainerservice.AgentPoolsClient, rg, clusterName, agentPoolName string, agentPool armcontainerservice.AgentPool) (*armcontainerservice.AgentPool, error) {
	poller, err := client.BeginCreateOrUpdate(context.TODO(), rg, clusterName, agentPoolName, agentPool, nil)
	if err != nil {
		return nil, err
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &res.AgentPool, nil
}
