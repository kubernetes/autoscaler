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
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/app"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
)

// TestEnvironment holds the resources for a single test's environment.
// Each test gets its own isolated API server, clients, and kubeconfig.
type testEnvironment struct {
	TestEnv    *envtest.Environment
	RestConfig *rest.Config
	KubeClient *clientset.Clientset
	VPAClient  *vpa_clientset.Clientset
	Kubeconfig string
}

// Cleanup stops the test environment and removes temporary files.
func (te *testEnvironment) Cleanup() {
	if te.Kubeconfig != "" {
		_ = os.Remove(te.Kubeconfig)
	}
	if te.TestEnv != nil {
		_ = te.TestEnv.Stop()
	}
}

// SetupTestEnvironment creates a new isolated envtest environment for a single test.
// This allows tests to run in parallel without interfering with each other.
// The caller should defer cleanup() to ensure proper resource cleanup.
func SetupTestEnvironment(t *testing.T) *testEnvironment {
	t.Helper()

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
	env := &envtest.Environment{
		CRDDirectoryPaths:     []string{crdPath},
		ErrorIfCRDPathMissing: true,
	}

	restConfig, err := env.Start()
	if err != nil {
		t.Fatalf("Failed to start envtest: %v. Make sure KUBEBUILDER_ASSETS is set or binaries are installed. "+
			"Run: go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest && "+
			"eval $(setup-envtest use -p env)", err)
	}

	kubeClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		_ = env.Stop()
		t.Fatalf("Failed to create kube client: %v", err)
	}

	vpaClientConfig := rest.CopyConfig(restConfig)
	vpaClientConfig.ContentType = "application/json"
	vpaClient, err := vpa_clientset.NewForConfig(vpaClientConfig)
	if err != nil {
		_ = env.Stop()
		t.Fatalf("Failed to create VPA client: %v", err)
	}

	// Create a kubeconfig file for the recommender to use
	kubeconfig := createKubeconfigFileForRestConfig(restConfig)

	return &testEnvironment{
		TestEnv:    env,
		RestConfig: restConfig,
		KubeClient: kubeClient,
		VPAClient:  vpaClient,
		Kubeconfig: kubeconfig,
	}
}

// getFreePort returns a free port number that can be used for the metrics server.
// This is necessary for parallel test execution to avoid port conflicts.
func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close() //nolint:errcheck
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func createKubeconfigFileForRestConfig(restConfig *rest.Config) string {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		TLSServerName:            restConfig.ServerName,
		CertificateAuthorityData: restConfig.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
		Token:                 restConfig.BearerToken,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	kubeConfigFile, _ := os.CreateTemp("", "kubeconfig")
	_ = clientcmd.WriteToFile(clientConfig, kubeConfigFile.Name())
	return kubeConfigFile.Name()
}

// StartRecommender creates and starts a recommender app with the given config.
// It returns a context for the recommender and a cancel function that should be deferred.
// The recommender runs in a background goroutine.
func StartRecommender(t *testing.T, config *recommender_config.RecommenderConfig) (recommenderCtx context.Context, cancel func()) {
	t.Helper()

	// Get a free port for the metrics server to avoid conflicts in parallel tests
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port for metrics: %v", err)
	}
	config.Address = fmt.Sprintf(":%d", port)

	recommenderApp, err := app.NewRecommenderApp(config)
	if err != nil {
		t.Fatalf("Failed to create recommender app: %v", err)
	}

	recommenderCtx, recommenderCancel := context.WithCancel(context.Background())

	// Start the recommender in a goroutine
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})
	leaderElection := app.DefaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = false // Disable leader election for testing

	go func() {
		defer close(doneChan)
		t.Logf("Starting recommender app on port %d...", port)
		err := recommenderApp.Run(recommenderCtx, leaderElection)
		if err != nil && recommenderCtx.Err() == nil {
			errChan <- err
		}
	}()

	return recommenderCtx, recommenderCancel
}

// NewHamsterDeployment creates a simple hamster deployment for testing.
func NewHamsterDeployment(ns string, replicas int32, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hamster-deployment",
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{{
						Name:  "hamster",
						Image: "busybox",
					}},
				},
			},
		},
	}
}
