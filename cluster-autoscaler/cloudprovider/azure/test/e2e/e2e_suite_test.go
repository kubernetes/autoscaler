/*
Copyright 2024 The Kubernetes Authors.

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

package e2e_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx           = context.Background()
	vmss          *armcompute.VirtualMachineScaleSetsClient
	k8s           client.Client
	resourceGroup string
)

func init() {
	flag.StringVar(&resourceGroup, "resource-group", "", "resource group containing cluster-autoscaler-managed resources")
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e Suite")
}

var _ = BeforeSuite(func() {
	azCred, err := azidentity.NewDefaultAzureCredential(nil)
	Expect(err).NotTo(HaveOccurred())
	vmss, err = armcompute.NewVirtualMachineScaleSetsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), azCred, nil)
	Expect(err).NotTo(HaveOccurred())

	k8sConfig := genericclioptions.NewConfigFlags(false)
	restConfig, err := k8sConfig.ToRESTConfig()
	Expect(err).NotTo(HaveOccurred())
	k8s, err = client.New(restConfig, client.Options{})
	Expect(err).NotTo(HaveOccurred())
})

func allVMSSStable(g Gomega) {
	pager := vmss.NewListPager(resourceGroup, nil)
	expectedNodes := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		for _, scaleset := range page.Value {
			g.Expect(*scaleset.Properties.ProvisioningState).To(Equal("Succeeded"))
			expectedNodes += int(*scaleset.SKU.Capacity)
		}
	}

	nodes := &corev1.NodeList{}
	g.Expect(k8s.List(ctx, nodes)).To(Succeed())
	g.Expect(nodes.Items).To(SatisfyAll(
		HaveLen(int(expectedNodes)),
		ContainElements(Satisfy(func(node corev1.Node) bool {
			for _, cond := range node.Status.Conditions {
				if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		})),
	))
}
