/*
Copyright 2016 The Kubernetes Authors.

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
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	kube_rest "k8s.io/client-go/rest"
	kube_client_cmd "k8s.io/client-go/tools/clientcmd"
)

// This code was borrowed from Heapster to push the work forward and contains some functionality
// that may not be needed in Kubernetes.
// TODO(mwielgus): revisit this once we have the basic structure ready.

const (
	defaultUseServiceAccount  = false
	defaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultInClusterConfig    = true
)

func getConfigOverrides(uri *url.URL) (*kube_client_cmd.ConfigOverrides, error) {
	kubeConfigOverride := kube_client_cmd.ConfigOverrides{}
	if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
		kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
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
				Host: configOverrides.ClusterInfo.Server,
			}
		}
	}
	if len(kubeConfig.Host) == 0 {
		return nil, fmt.Errorf("invalid kubernetes master url specified")
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

// NewDynamicKubeConfigRoundTripper wraps the http.RoundTripper to provide the latest bearer token from the kube config on disk
func NewDynamicKubeConfigRoundTripper(kubeconfigPath string, rt http.RoundTripper) http.RoundTripper {
	return &bearerAuthRoundTripper{
		kubeconfigPath: kubeconfigPath,
		rt:             rt,
	}
}

type bearerAuthRoundTripper struct {
	kubeconfigPath string
	rt             http.RoundTripper

	lastModified time.Time
	bearerToken  string
}

func (rt *bearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	stat, err := os.Stat(rt.kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat kubeconfig path: %w", err)
	}
	if stat.ModTime().After(rt.lastModified) {
		kubeConfig, err := kube_client_cmd.BuildConfigFromFlags("", rt.kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("cannot build kube cluster config: %w", err)
		}
		rt.bearerToken = kubeConfig.BearerToken
		rt.lastModified = stat.ModTime()
	}
	req = utilnet.CloneRequest(req)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rt.bearerToken))
	return rt.rt.RoundTrip(req)
}

var _ http.RoundTripper = &bearerAuthRoundTripper{}
