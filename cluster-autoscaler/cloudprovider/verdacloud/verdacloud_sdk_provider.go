/*
Copyright 2019 The Kubernetes Authors.

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

package verdacloud

import (
	"os"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda"
	klog "k8s.io/klog/v2"
)

const (
	// UserAgent identifies the cluster autoscaler in API requests
	autoscalerUserAgent = "cluster-autoscaler/verdacloud"
)

type verdacloudSDKProvider struct {
	client *verda.Client
}

func createVerdacloudSDKProvider(cfg *cloudConfig) (*verdacloudSDKProvider, error) {
	clientID := os.Getenv("VERDA_CLIENT_ID")
	clientSecret := os.Getenv("VERDA_CLIENT_SECRET")
	baseURL := os.Getenv("VERDA_BASE_URL")

	verdaDebug := os.Getenv("VERDA_DEBUG")
	detailedDebugEnabled := verdaDebug == "true" || verdaDebug == "1"

	var logger verda.Logger
	if cfg.Debug || detailedDebugEnabled {
		logger = verda.NewStdLogger(true)
	} else {
		logger = &verda.NoOpLogger{}
	}

	clientOpts := []verda.ClientOption{
		verda.WithClientID(clientID),
		verda.WithClientSecret(clientSecret),
		verda.WithDebugLogging(cfg.Debug),
		verda.WithLogger(logger),
		verda.WithUserAgent(autoscalerUserAgent),
	}

	if baseURL != "" {
		clientOpts = append(clientOpts, verda.WithBaseURL(baseURL))
		klog.V(4).Infof("Using VerdaCloud API base URL from VERDA_BASE_URL: %s", baseURL)
	} else {
		klog.V(4).Infof("Using default VerdaCloud API base URL: https://api.verda.com/v1")
	}

	client, err := verda.NewClient(clientOpts...)
	if err != nil {
		return nil, err
	}

	client.AddRequestMiddleware(
		verda.ExponentialBackoffRetryMiddleware(3, time.Second, logger),
	)

	if detailedDebugEnabled {
		verda.AddDetailedDebugLogging(client)
		klog.V(4).Info("VerdaCloud SDK detailed debug logging enabled (VERDA_DEBUG=true)")
	}

	klog.V(4).Info("VerdaCloud SDK client created with retry middleware enabled (max 3 retries, exponential backoff)")

	return &verdacloudSDKProvider{
		client: client,
	}, nil
}
