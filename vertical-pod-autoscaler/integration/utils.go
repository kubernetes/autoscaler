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
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	kubeapiservertesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/test/integration/framework"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/app"
)

// installVPACRDs installs the VPA CustomResourceDefinitions from the YAML file
func installVPACRDs(ctx context.Context, client apiextensionsclientset.Interface) error {
	// Get the path to the CRD YAML file relative to this test file
	_, thisFile, _, _ := runtime.Caller(0)
	crdPath := filepath.Join(filepath.Dir(thisFile), "..", "deploy", "vpa-v1-crd-gen.yaml")

	file, err := os.Open(crdPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var crdNames []string
	reader := yaml.NewYAMLReader(bufio.NewReader(file))
	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal(data, crd); err != nil {
			return err
		}

		// Skip empty documents
		if crd.Name == "" {
			continue
		}

		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		crdNames = append(crdNames, crd.Name)
	}

	// Wait for all CRDs to be established
	for _, name := range crdNames {
		if err := waitForCRDEstablished(ctx, client, name); err != nil {
			return err
		}
	}

	return nil
}

// waitForCRDEstablished waits for a CRD to be established and ready to accept resources
func waitForCRDEstablished(ctx context.Context, client apiextensionsclientset.Interface, name string) error {
	return wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}

// sigs.k8s.io/controller-runtime/pkg/envtest
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

func makeVPACP(ns string) *vpa_types.VerticalPodAutoscalerCheckpoint {
	return &vpa_types.VerticalPodAutoscalerCheckpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vpa",
			Namespace: ns,
		},
		Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
			VPAObjectName: "test-vpa",
			ContainerName: "A",
		},
	}
}

func createTestServerAndInstallCRDsWithClients(t *testing.T) (*kubeapiservertesting.TestServer, *clientset.Clientset, *vpa_clientset.Clientset) {
	etcdOptions := framework.DefaultEtcdOptions()
	server := kubeapiservertesting.StartTestServerOrDie(t, nil, framework.DefaultTestServerFlags(), &etcdOptions.StorageConfig)

	apiextensionsClient := apiextensionsclientset.NewForConfigOrDie(server.ClientConfig)

	if err := installVPACRDs(context.Background(), apiextensionsClient); err != nil {
		t.Fatalf("Failed to install VPA CRD: %v", err)
	}

	kubeClient := clientset.NewForConfigOrDie(server.ClientConfig)

	vpaClientConfig := rest.CopyConfig(server.ClientConfig)
	vpaClientConfig.ContentType = "application/json"
	vpaClient := vpa_clientset.NewForConfigOrDie(vpaClientConfig)

	return server, kubeClient, vpaClient
}

func setupKubeconfig(t *testing.T, restConfig *rest.Config) (kubeconfigPath string, cleanup func()) {
	t.Helper()
	kubeconfig := createKubeconfigFileForRestConfig(restConfig)
	t.Logf("Using kubeconfig file: %s", kubeconfig)
	return kubeconfig, func() {
		os.Remove(kubeconfig)
	}
}

// startRecommender creates and starts a recommender app with the given config.
// It returns a context for the recommender and a cancel function that should be deferred.
// The recommender runs in a background goroutine.
func startRecommender(t *testing.T, config *app.RecommenderConfig) (recommenderCtx context.Context, cancel func()) {
	t.Helper()

	recommenderApp, err := app.NewRecommenderApp(config)
	if err != nil {
		t.Fatalf("Failed to create recommender app: %v", err)
	}

	recommenderCtx, recommenderCancel := context.WithTimeout(context.Background(), 15*time.Second)

	// Start the recommender in a goroutine
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})
	leaderElection := app.DefaultLeaderElectionConfiguration()
	leaderElection.LeaderElect = false // Disable leader election for testing

	go func() {
		defer close(doneChan)
		t.Log("Starting recommender app...")
		err := recommenderApp.Run(recommenderCtx, leaderElection)
		if err != nil && recommenderCtx.Err() == nil {
			errChan <- err
		}
	}()

	return recommenderCtx, recommenderCancel
}

// newHamsterDeployment creates a simple hamster deployment for testing.
func newHamsterDeployment(ns string, replicas int32, labels map[string]string) *appsv1.Deployment {
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
