/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	kube_restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/rest"
)

var (
	metricsFetcherInterval = flag.Duration("updater-interval", 1*time.Minute, `How often metrics should be fetched`)
)

func main() {
	glog.Infof("Running VPA Recommender")
	kube_flag.InitFlags()

	// TODO monitoring

	config := createKubeConfig()
	recommender := NewRecommender(config, metricsFetcherInterval)
	recommender.Run()
}

func createKubeConfig() *rest.Config {
	config, err := kube_restclient.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	return config
}
