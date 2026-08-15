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

// Package environment provides a shared test environment for Azure CAS e2e tests.
// The environment provides K8s and Azure clients for test assertions,
// and optionally deploys CAS via Helm (for CI) or assumes it's already running (for local dev).
package environment

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const casReleaseName = "cluster-autoscaler"

// HelmConfig holds optional Helm deployment configuration.
// When populated, BeforeSuite deploys CAS via Helm (CI path).
// When empty, tests assume CAS is already running (local dev path).
type HelmConfig struct {
	ChartPath             string
	ClusterName           string
	ClientID              string
	CASNamespace          string
	CASServiceAccountName string
	CASImageRepository    string
	CASImageTag           string
}

// IsEnabled returns true if Helm deployment is configured.
func (h *HelmConfig) IsEnabled() bool {
	return h != nil && h.CASImageRepository != "" && h.CASImageTag != ""
}

// Environment holds all clients and configuration for an e2e test suite.
type Environment struct {
	Ctx            context.Context
	K8s            client.Client
	VMSS           *armcompute.VirtualMachineScaleSetsClient
	ResourceGroup  string
	SubscriptionID string
	TenantID       string
	Helm           *HelmConfig
}

// NewEnvironment creates a fully initialized Environment.
func NewEnvironment(resourceGroup string, helm *HelmConfig) *Environment {
	env := &Environment{
		Ctx:            context.Background(),
		ResourceGroup:  resourceGroup,
		SubscriptionID: os.Getenv("AZURE_SUBSCRIPTION_ID"),
		TenantID:       os.Getenv("AZURE_TENANT_ID"),
		Helm:           helm,
	}

	azCred, err := azidentity.NewDefaultAzureCredential(nil)
	Expect(err).NotTo(HaveOccurred())

	env.VMSS, err = armcompute.NewVirtualMachineScaleSetsClient(env.SubscriptionID, azCred, nil)
	Expect(err).NotTo(HaveOccurred())

	restConfig, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())
	env.K8s, err = client.New(restConfig, client.Options{})
	Expect(err).NotTo(HaveOccurred())

	return env
}

// --- Helm helpers ---

// EnsureHelmRelease deploys or updates CAS via Helm if HelmConfig is enabled.
// If Helm is not configured, this is a no-op (CAS is managed externally).
func (env *Environment) EnsureHelmRelease(extraValues map[string]interface{}) {
	if !env.Helm.IsEnabled() {
		GinkgoLogr.Info("Helm not configured — assuming CAS is already deployed (e.g., via skaffold)")
		return
	}

	values := map[string]interface{}{
		"cloudProvider":                     "azure",
		"azureTenantID":                     env.TenantID,
		"azureSubscriptionID":               env.SubscriptionID,
		"azureUseWorkloadIdentityExtension": true,
		"azureResourceGroup":                env.ResourceGroup,
		"podLabels": map[string]interface{}{
			"azure.workload.identity/use": "true",
		},
		"rbac": map[string]interface{}{
			"serviceAccount": map[string]interface{}{
				"name": env.Helm.CASServiceAccountName,
				"annotations": map[string]interface{}{
					"azure.workload.identity/tenant-id": env.TenantID,
					"azure.workload.identity/client-id": env.Helm.ClientID,
				},
			},
		},
		"autoDiscovery": map[string]interface{}{
			"clusterName": env.Helm.ClusterName,
		},
		"nodeSelector": map[string]interface{}{
			"kubernetes.io/os": "linux",
		},
		"image": map[string]interface{}{
			"repository": env.Helm.CASImageRepository,
			"tag":        env.Helm.CASImageTag,
			"pullPolicy": "Always",
		},
	}

	// Merge extra values (e.g., extraArgs for scale-down tuning)
	for k, v := range extraValues {
		values[k] = v
	}

	helmEnv := cli.New()
	helmCfg := new(action.Configuration)
	Expect(helmCfg.Init(helmEnv.RESTClientGetter(), env.Helm.CASNamespace, "secret", func(format string, v ...interface{}) {
		GinkgoLogr.Info(fmt.Sprintf(format, v...))
	})).To(Succeed())

	chart, err := loader.Load(env.Helm.ChartPath)
	Expect(err).NotTo(HaveOccurred())

	get := action.NewGet(helmCfg)
	_, err = get.Run(casReleaseName)
	if errors.Is(err, driver.ErrReleaseNotFound) {
		install := action.NewInstall(helmCfg)
		install.Timeout = 5 * time.Minute
		install.Wait = true
		install.CreateNamespace = true
		install.ReleaseName = casReleaseName
		install.Namespace = env.Helm.CASNamespace
		_, err = install.Run(chart, values)
		Expect(err).NotTo(HaveOccurred())
		return
	}
	Expect(err).NotTo(HaveOccurred())

	upgrade := action.NewUpgrade(helmCfg)
	upgrade.Timeout = 5 * time.Minute
	upgrade.Wait = true
	upgrade.ResetValues = true
	_, err = upgrade.Run(casReleaseName, chart, values)
	Expect(err).NotTo(HaveOccurred())
}

// --- VMSS helpers ---

// AllVMSSStable checks that all VMSS in the resource group have Succeeded
// provisioning state and the number of Ready K8s nodes matches total VMSS capacity.
func (env *Environment) AllVMSSStable(g Gomega) {
	pager := env.VMSS.NewListPager(env.ResourceGroup, nil)
	expectedNodes := 0
	for pager.More() {
		page, err := pager.NextPage(env.Ctx)
		g.Expect(err).NotTo(HaveOccurred())
		for _, scaleset := range page.Value {
			g.Expect(*scaleset.Properties.ProvisioningState).To(Equal("Succeeded"))
			expectedNodes += int(*scaleset.SKU.Capacity)
		}
	}

	nodes := &corev1.NodeList{}
	g.Expect(env.K8s.List(env.Ctx, nodes)).To(Succeed())
	g.Expect(nodes.Items).To(SatisfyAll(
		HaveLen(expectedNodes),
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

// --- K8s helpers ---

// ReadyNodeCount returns the number of Ready nodes in the cluster.
func (env *Environment) ReadyNodeCount() (int, error) {
	readyCount := 0
	nodes := &corev1.NodeList{}
	if err := env.K8s.List(env.Ctx, nodes); err != nil {
		return 0, err
	}
	for _, node := range nodes.Items {
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				readyCount++
				break
			}
		}
	}
	return readyCount, nil
}
