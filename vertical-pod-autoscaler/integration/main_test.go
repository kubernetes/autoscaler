//go:build integration
// +build integration

/*
Copyright The Kubernetes Authors.

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

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
)

// Global test environment shared across all tests
var (
	testEnv    *envtest.Environment
	restConfig *rest.Config
	kubeClient *clientset.Clientset
	vpaClient  *vpa_clientset.Clientset
	kubeconfig string
)

// TestMain sets up the integration test environment once for all tests.
// This is more efficient than setting up envtest for each individual test.
func TestMain(m *testing.M) {
	var err error
	var code int

	defer func() {
		// Cleanup
		if kubeconfig != "" {
			_ = os.Remove(kubeconfig)
		}
		if testEnv != nil {
			_ = testEnv.Stop()
		}
		os.Exit(code)
	}()

	// Setup envtest
	if err = setupTestEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test environment: %v\n", err)
		code = 1
		return
	}

	// Run tests
	code = m.Run()
}

func setupTestEnv() error {
	// Get the path to the CRD YAML file relative to this test file
	_, thisFile, _, _ := runtime.Caller(0)
	crdPath := filepath.Join(filepath.Dir(thisFile), "..", "deploy", "vpa-v1-crd-gen.yaml")

	// envtest looks for binaries in the following order:
	// 1. KUBEBUILDER_ASSETS environment variable
	// 2. Default path: /usr/local/kubebuilder/bin
	// To install the binaries, run:
	//   go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	//   setup-envtest use --bin-dir /usr/local/kubebuilder/bin
	// Or set KUBEBUILDER_ASSETS to point to the directory containing etcd and kube-apiserver
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{crdPath},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	restConfig, err = testEnv.Start()
	if err != nil {
		return fmt.Errorf("failed to start envtest: %w. Make sure KUBEBUILDER_ASSETS is set or binaries are installed. "+
			"Run: go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest && "+
			"eval $(setup-envtest use -p env)", err)
	}

	kubeClient, err = clientset.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	vpaClientConfig := rest.CopyConfig(restConfig)
	vpaClientConfig.ContentType = "application/json"
	vpaClient, err = vpa_clientset.NewForConfig(vpaClientConfig)
	if err != nil {
		return fmt.Errorf("failed to create VPA client: %w", err)
	}

	// Create a kubeconfig file for the recommender to use
	kubeconfig = createKubeconfigFileForRestConfig(restConfig)

	return nil
}
