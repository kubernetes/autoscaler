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
	"context"
	"os"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/app"
)

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
