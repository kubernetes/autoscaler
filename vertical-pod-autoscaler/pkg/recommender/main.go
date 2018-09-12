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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/routines"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	"k8s.io/client-go/rest"
)

var (
	metricsFetcherInterval = flag.Duration("recommender-interval", 1*time.Minute, `How often metrics should be fetched`)
	checkpointsGCInterval  = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	prometheusAddress      = flag.String("prometheus-address", "", `Where to reach for Prometheus metrics`)
	storage                = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	address                = flag.String("address", ":8942", "The address to expose Prometheus metrics.")
	kubeApiQps             = flag.Float64("kube-api-qps", 5.0, `QPS limit when making requests to Kubernetes apiserver`)
	kubeApiBurst           = flag.Float64("kube-api-burst", 10.0, `QPS burst limit when making requests to Kubernetes apiserver`)
)

func main() {
	kube_flag.InitFlags()
	glog.V(1).Infof("Vertical Pod Autoscaler %s Recommender", common.VerticalPodAutoscalerVersion)

	config := createKubeConfig(float32(*kubeApiQps), int(*kubeApiBurst))

	metrics.Initialize(*address)
	metrics_recommender.Register()

	useCheckpoints := *storage != "prometheus"
	recommender := routines.NewRecommender(config, *checkpointsGCInterval, useCheckpoints)
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

func createKubeConfig(kubeApiQps float32, kubeApiBurst int) *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	config.QPS = kubeApiQps
	config.Burst = kubeApiBurst
	return config
}
