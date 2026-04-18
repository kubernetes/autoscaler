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

package recommender

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	aruntime "runtime"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/test/integration/fixtures"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kubeapiservertesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/test/integration/framework"
	"k8s.io/kubernetes/test/utils/ktesting"

	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/routines"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

func loadCRDs(path string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	yamlDecoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024)

	var crds []*apiextensionsv1.CustomResourceDefinition

	for {
		var raw runtime.RawExtension
		if err := yamlDecoder.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// skip empty docs
		if len(raw.Raw) == 0 {
			continue
		}

		obj, gvk, err := decoder.Decode(raw.Raw, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}

		crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
		if !ok {
			return nil, fmt.Errorf("unexpected type: %s", gvk)
		}

		crds = append(crds, crd)
	}

	return crds, nil
}

func installVPACRDs(t *testing.T, config *rest.Config) {
	t.Helper()

	crdClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating apiextensions client: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating dynamic client: %v", err)
	}

	_, thisFile, _, _ := aruntime.Caller(0)
	crdPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "deploy", "vpa-v1-crd-gen.yaml")

	crds, err := loadCRDs(crdPath)
	if err != nil {
		t.Fatalf("Error loading CRDs: %v", err)
	}

	for _, crd := range crds {
		_, _ = fixtures.CreateNewV1CustomResourceDefinition(crd, crdClient, dynamicClient)
	}
}

func recommenderSetup(t *testing.T, recommenderConfig *recommender_config.RecommenderConfig) (context.Context, kubeapiservertesting.TearDownFunc, *routines.RecommenderController, informers.SharedInformerFactory, clientset.Interface, *vpa_clientset.Clientset) {
	tCtx := ktesting.Init(t)
	// Disable ServiceAccount admission plugin as we don't have serviceaccount controller running.
	server := kubeapiservertesting.StartTestServerOrDie(t, nil, framework.DefaultTestServerFlags(), framework.SharedEtcd())

	config := rest.CopyConfig(server.ClientConfig)

	// Install VPA CRDs so the VPA informers can sync.
	installVPACRDs(t, config)

	clientSet, err := clientset.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error in create clientset: %v", err)
	}
	resyncPeriod := 12 * time.Hour
	informers := informers.NewSharedInformerFactory(clientset.NewForConfigOrDie(rest.AddUserAgent(config, "vpa-updater-informers")), resyncPeriod)

	vpaClientConfig := rest.CopyConfig(config)
	vpaClientConfig.ContentType = "application/json"

	vpaClient := vpa_clientset.NewForConfigOrDie(rest.AddUserAgent(vpaClientConfig, "vpa-updater-vpa-client"))
	healthCheck := metrics.NewHealthCheck(1 * 5)

	recommenderConfig.MetricsFetcherInterval = 1 * time.Second // Short interval for testing

	stopCh := make(chan struct{})

	rm, err := routines.NewRecommenderController(
		tCtx,
		config,
		clientSet,
		vpaClient,
		informers,
		recommenderConfig,
		healthCheck,
		stopCh,
	)
	if err != nil {
		close(stopCh) // Clean up any started routines
		t.Fatalf("Error creating recommender controller: %v", err)
	}

	newTeardown := func() {
		close(stopCh)
		tCtx.Cancel("tearing down controller")
		server.TearDownFn()
	}

	return tCtx, newTeardown, rm, informers, clientSet, vpaClient
}

func runControllerAndInformers(ctx context.Context, rm *routines.RecommenderController, informers informers.SharedInformerFactory) func() {
	ctx, cancel := context.WithCancel(ctx)
	informers.Start(ctx.Done())
	go func() {
		_ = rm.Run(ctx)
	}()
	return cancel
}
