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

package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clockutil "k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	capacitybuffer "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned"
	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
	cbapi "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	cbctrl "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/controller"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/fakepods"
	cbmetrics "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	cqctrl "k8s.io/autoscaler/cluster-autoscaler/resourcequotas/capacityquota"
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
var crClient client.Client
var ctx context.Context
var cancel context.CancelFunc
var reconciliationCache *cbmetrics.ReconciliationCache
var clock = clockutil.RealClock{}

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

	scheme := runtime.NewScheme()
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = cqv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	buffersClient, err = capacitybuffer.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(buffersClient).NotTo(BeNil())

	crClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(crClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = cqctrl.NewCapacityQuotaReconciler(mgr.GetClient(), cqctrl.ReconcilerOptions{NodeFilter: utils.VirtualKubeletNodeFilter{}}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

	client, err := cbclient.NewCapacityBufferClientFromConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	resolver := fakepods.NewDryRunResolver(k8sClient)
	reconciliationCache = cbmetrics.NewReconciliationCache()
	controller := cbctrl.NewDefaultBufferController(
		client,
		resolver,
		[]string{cbapi.ActiveProvisioningStrategy, ""},
		reconciliationCache,
		clock,
	)

	go controller.Run(ctx.Done())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	Eventually(func() error {
		return testEnv.Stop()
	}, time.Minute, time.Second).Should(Succeed())
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
