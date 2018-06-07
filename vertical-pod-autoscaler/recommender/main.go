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
	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/initialization"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/input/history"
	"k8s.io/client-go/rest"
	kube_restclient "k8s.io/client-go/rest"
)

var (
	metricsFetcherInterval = flag.Duration("recommender-interval", 1*time.Minute, `How often metrics should be fetched`)
	checkpointsGCInterval  = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	prometheusAddress      = flag.String("prometheus-address", "", `Where to reach for Prometheus metrics`)
	storage                = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
)

func main() {
	kube_flag.InitFlags()
	glog.V(1).Infof("Vertical Pod Autoscaler %s Recommender", common.VerticalPodAutoscalerVersion)

	config := createKubeConfig()

	useCheckpoints := *storage != "prometheus"
	recommender := initialization.NewRecommender(config, *checkpointsGCInterval, useCheckpoints)
	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints()
	} else {
		recommender.GetClusterStateFeeder().InitFromHistoryProvider(history.NewPrometheusHistoryProvider(*prometheusAddress))
	}

	for {
		select {
		case <-time.After(*metricsFetcherInterval):
			{
				recommender.RunOnce()
			}
		}
	}

}

func createKubeConfig() *rest.Config {
	config, err := kube_restclient.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	return config
}
