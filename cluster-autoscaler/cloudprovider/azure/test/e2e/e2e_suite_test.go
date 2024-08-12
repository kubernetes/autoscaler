//go:build e2e

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
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	casReleaseName = "cluster-autoscaler"
)

var (
	ctx = context.Background()

	vmss          *armcompute.VirtualMachineScaleSetsClient
	vmssVMsClient *armcompute.VirtualMachineScaleSetVMsClient

	k8s     client.Client
	helmEnv = cli.New()

	nodeResourceGroup     string
	clusterResourceGroup  string
	clusterName           string
	clientID              string
	casNamespace          string
	casServiceAccountName string
	casImageRepository    string
	casImageTag           string
)

func init() {
	flag.StringVar(&nodeResourceGroup, "node-resource-group", "", "resource group containing cluster-autoscaler-managed resources")
	flag.StringVar(&clusterResourceGroup, "cluster-resource-group", "", "cluster resource group contains the managed cluster we end up creating")
	flag.StringVar(&clusterName, "cluster-name", "", "Cluster API Cluster name for the cluster to be managed by cluster-autoscaler")
	flag.StringVar(&clientID, "client-id", "", "Azure client ID to be used by cluster-autoscaler")
	flag.StringVar(&casNamespace, "cas-namespace", "", "Namespace in which to install cluster-autoscaler")
	flag.StringVar(&casServiceAccountName, "cas-serviceaccount-name", "", "Name of the ServiceAccount to be used by cluster-autoscaler")
	flag.StringVar(&casImageRepository, "cas-image-repository", "", "Repository of the container image for cluster-autoscaler")
	flag.StringVar(&casImageTag, "cas-image-tag", "", "Tag of the container image for cluster-autoscaler")
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e Suite")
}

var _ = BeforeSuite(func() {
	azCred, err := azidentity.NewDefaultAzureCredential(nil)
	Expect(err).NotTo(HaveOccurred())
	sub := os.Getenv("AZURE_SUBSCRIPTION_ID")

	vmss, err = armcompute.NewVirtualMachineScaleSetsClient(sub, azCred, nil)
	Expect(err).NotTo(HaveOccurred())

	vmssVMsClient, err = armcompute.NewVirtualMachineScaleSetVMsClient(sub, azCred, nil)
	Expect(err).NotTo(HaveOccurred())

	restConfig, err := helmEnv.RESTClientGetter().ToRESTConfig()
	Expect(err).NotTo(HaveOccurred())
	k8s, err = client.New(restConfig, client.Options{})
	Expect(err).NotTo(HaveOccurred())

	ensureHelmValues(map[string]interface{}{
		"cloudProvider":                     "azure",
		"azureTenantID":                     os.Getenv("AZURE_TENANT_ID"),
		"azureSubscriptionID":               os.Getenv("AZURE_SUBSCRIPTION_ID"),
		"azureUseWorkloadIdentityExtension": true,
		"azureResourceGroup":                nodeResourceGroup,
		"podLabels": map[string]interface{}{
			"azure.workload.identity/use": "true",
		},
		"rbac": map[string]interface{}{
			"serviceAccount": map[string]interface{}{
				"name": casServiceAccountName,
				"annotations": map[string]interface{}{
					"azure.workload.identity/tenant-id": os.Getenv("AZURE_TENANT_ID"),
					"azure.workload.identity/client-id": clientID,
				},
			},
		},
		"autoDiscovery": map[string]interface{}{
			"clusterName": clusterName,
		},
		"nodeSelector": map[string]interface{}{
			"kubernetes.io/os":          "linux",
			"kubernetes.azure.com/mode": "system",
		},
		"image": map[string]interface{}{
			"repository": casImageRepository,
			"tag":        casImageTag,
			"pullPolicy": "Always",
		},
	})
})

func allVMSSStable(g Gomega) {
	pager := vmss.NewListPager(nodeResourceGroup, nil)
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

func ensureHelmValues(values map[string]interface{}) {
	helmCfg := new(action.Configuration)
	Expect(helmCfg.Init(helmEnv.RESTClientGetter(), casNamespace, "secret", func(format string, v ...interface{}) {
		GinkgoLogr.Info(fmt.Sprintf(format, v...))
	})).To(Succeed())

	chart, err := loader.Load("../../../../../charts/cluster-autoscaler")
	Expect(err).NotTo(HaveOccurred())

	get := action.NewGet(helmCfg)
	_, err = get.Run(casReleaseName)
	if errors.Is(err, driver.ErrReleaseNotFound) {
		install := action.NewInstall(helmCfg)
		install.Timeout = 5 * time.Minute
		install.Wait = true
		install.CreateNamespace = true
		install.ReleaseName = casReleaseName
		install.Namespace = casNamespace
		_, err := install.Run(chart, values)
		Expect(err).NotTo(HaveOccurred())
		return
	} else {
		Expect(err).NotTo(HaveOccurred())
	}

	upgrade := action.NewUpgrade(helmCfg)
	upgrade.Timeout = 5 * time.Minute
	upgrade.Wait = true
	upgrade.ReuseValues = true
	_, err = upgrade.Run(casReleaseName, chart, values)
	Expect(err).NotTo(HaveOccurred())
}
