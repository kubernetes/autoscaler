/*
Copyright 2023 The Kubernetes Authors.

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

package configloader

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
)

type k8sSecretLoader[Type any] struct {
	KubeClient clientset.Interface
	configLoader[Type]
	*K8sSecretConfig
	decoderFactory[Type]
}

// cloud provider config secret
const (
	DefaultCloudProviderConfigSecName      = "azure-cloud-provider"
	DefaultCloudProviderConfigSecNamespace = "kube-system"
	DefaultCloudProviderConfigSecKey       = "cloud-config"
)

var (
	ErrNoKubeClient = errors.New("no kube client provided")
	ErrNoData       = errors.New("no data in secret")
)

type K8sSecretConfig struct {
	SecretName      string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
	SecretNamespace string `json:"secretNamespace,omitempty" yaml:"secretNamespace,omitempty"`
	CloudConfigKey  string `json:"cloudConfigKey,omitempty" yaml:"cloudConfigKey,omitempty"`
}

// newK8sSecretLoader returns a config loader which load config from k8s secret.
// If KubeClient is nil, it will return nil.
// decoderFactory is a function that creates a new loader from the content of the secret. it should never be nil.
func newK8sSecretLoader[Type any](config *K8sSecretConfig, KubeClient kubernetes.Interface, loader configLoader[Type], decoder decoderFactory[Type]) configLoader[Type] {
	if KubeClient == nil {
		return nil
	}
	if config == nil {
		return nil
	}
	if config.SecretName == "" {
		config.SecretName = DefaultCloudProviderConfigSecName
	}
	if config.SecretNamespace == "" {
		config.SecretNamespace = DefaultCloudProviderConfigSecNamespace
	}
	if config.CloudConfigKey == "" {
		config.CloudConfigKey = DefaultCloudProviderConfigSecKey
	}
	return &k8sSecretLoader[Type]{
		configLoader:    loader,
		K8sSecretConfig: config,
		decoderFactory:  decoder,
		KubeClient:      KubeClient,
	}
}

func (k *k8sSecretLoader[Type]) Load(ctx context.Context) (*Type, error) {
	if k.configLoader == nil {
		k.configLoader = newEmptyLoader[Type](nil)
	}

	if k.KubeClient == nil {
		return nil, ErrNoKubeClient
	}

	secret, err := k.KubeClient.CoreV1().Secrets(k.SecretNamespace).Get(ctx, k.SecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if len(secret.Data) <= 0 {
		return nil, ErrNoData
	}
	var ok bool
	var content []byte
	if content, ok = secret.Data[k.CloudConfigKey]; !ok {
		return nil, ErrNoData
	}
	loader := k.decoderFactory([]byte(content), k.configLoader)
	return loader.Load(ctx)
}
