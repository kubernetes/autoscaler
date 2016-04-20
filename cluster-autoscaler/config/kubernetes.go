/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"k8s.io/kubernetes/pkg/api/unversioned"
	kube_rest "k8s.io/kubernetes/pkg/client/restclient"
	kube_client_cmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kube_client_cmd_api "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

// This code was borrowed from Heapster to push the work forward and contains some functionality
// that may not be needed in Kubernetes.
// TODO(mwielgus): revisit this once we have the basic structur ready.

const (
	// APIVersion to be used.
	APIVersion = "v1"

	defaultUseServiceAccount  = false
	defaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultInClusterConfig    = true
)

func getConfigOverrides(uri *url.URL) (*kube_client_cmd.ConfigOverrides, error) {
	kubeConfigOverride := kube_client_cmd.ConfigOverrides{
		ClusterInfo: kube_client_cmd_api.Cluster{
			APIVersion: APIVersion,
		},
	}
	if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
		kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
	}

	opts := uri.Query()

	if len(opts["apiVersion"]) >= 1 {
		kubeConfigOverride.ClusterInfo.APIVersion = opts["apiVersion"][0]
	}

	if len(opts["insecure"]) > 0 {
		insecure, err := strconv.ParseBool(opts["insecure"][0])
		if err != nil {
			return nil, err
		}
		kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
	}

	return &kubeConfigOverride, nil
}

// GetKubeClientConfig returns rest client configuration based on the passed url.
func GetKubeClientConfig(uri *url.URL) (*kube_rest.Config, error) {
	var (
		kubeConfig *kube_rest.Config
		err        error
	)

	opts := uri.Query()
	configOverrides, err := getConfigOverrides(uri)
	if err != nil {
		return nil, err
	}

	inClusterConfig := defaultInClusterConfig
	if len(opts["inClusterConfig"]) > 0 {
		inClusterConfig, err = strconv.ParseBool(opts["inClusterConfig"][0])
		if err != nil {
			return nil, err
		}
	}

	if inClusterConfig {
		kubeConfig, err = kube_rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		if configOverrides.ClusterInfo.Server != "" {
			kubeConfig.Host = configOverrides.ClusterInfo.Server
		}
		kubeConfig.GroupVersion = &unversioned.GroupVersion{Version: configOverrides.ClusterInfo.APIVersion}
		kubeConfig.Insecure = configOverrides.ClusterInfo.InsecureSkipTLSVerify
		if configOverrides.ClusterInfo.InsecureSkipTLSVerify {
			kubeConfig.TLSClientConfig.CAFile = ""
		}
	} else {
		authFile := ""
		if len(opts["auth"]) > 0 {
			authFile = opts["auth"][0]
		}

		if authFile != "" {
			if kubeConfig, err = kube_client_cmd.NewNonInteractiveDeferredLoadingClientConfig(
				&kube_client_cmd.ClientConfigLoadingRules{ExplicitPath: authFile},
				configOverrides).ClientConfig(); err != nil {
				return nil, err
			}
		} else {
			kubeConfig = &kube_rest.Config{
				Host:     configOverrides.ClusterInfo.Server,
				Insecure: configOverrides.ClusterInfo.InsecureSkipTLSVerify,
			}
			kubeConfig.GroupVersion = &unversioned.GroupVersion{Version: configOverrides.ClusterInfo.APIVersion}
		}
	}
	if len(kubeConfig.Host) == 0 {
		return nil, fmt.Errorf("invalid kubernetes master url specified")
	}
	if len(kubeConfig.GroupVersion.Version) == 0 {
		return nil, fmt.Errorf("invalid kubernetes API version specified")
	}

	useServiceAccount := defaultUseServiceAccount
	if len(opts["useServiceAccount"]) >= 1 {
		useServiceAccount, err = strconv.ParseBool(opts["useServiceAccount"][0])
		if err != nil {
			return nil, err
		}
	}

	if useServiceAccount {
		// If a readable service account token exists, then use it
		if contents, err := ioutil.ReadFile(defaultServiceAccountFile); err == nil {
			kubeConfig.BearerToken = string(contents)
		}
	}

	return kubeConfig, nil
}
