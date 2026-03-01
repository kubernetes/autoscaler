/*
Copyright 2025 The Kubernetes Authors.

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

package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	controller2 "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	capacitybuffer "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
)

func TestControllers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping controller tests in short mode.")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var cfg *rest.Config
var testEnv *envtest.Environment
var k8sClient *kubernetes.Clientset
var buffersClient *capacitybuffer.Clientset
var ctx context.Context
var cancel context.CancelFunc

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "apis", "config", "crd")},
		ErrorIfCRDPathMissing: true,
	}

	if binaryDir := getFirstFoundEnvTestBinaryDir(); binaryDir != "" {
		testEnv.BinaryAssetsDirectory = binaryDir
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	buffersClient, err = capacitybuffer.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(buffersClient).NotTo(BeNil())

	client, err := cbclient.NewCapacityBufferClientFromConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	controller := controller2.NewDefaultBufferController(client)

	go controller.Run(ctx.Done())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	if err != nil {
		logf.Log.Error(err, "failed to stop test environment")
	}
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}
