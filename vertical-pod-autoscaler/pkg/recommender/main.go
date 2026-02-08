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
	"context"

	"github.com/spf13/pflag"
	kube_flag "k8s.io/component-base/cli/flag"
	componentbaseoptions "k8s.io/component-base/config/options"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/app"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
)

func main() {
	config := recommender_config.InitRecommenderFlags()
	klog.InitFlags(nil)
	common.InitLoggingFlags()
	features.MutableFeatureGate.AddFlag(pflag.CommandLine)

	recommender_config.ValidateRecommenderConfig(config)
	common.ValidateCommonConfig(config.CommonFlags)

	leaderElection := app.DefaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).InfoS("Vertical Pod Autoscaler Recommender", "version", common.VerticalPodAutoscalerVersion(), "recommenderName", config.RecommenderName)

	recommenderApp, err := app.NewRecommenderApp(config)
	if err != nil {
		klog.ErrorS(err, "Failed to create recommender app")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	ctx := context.Background()
	if err := recommenderApp.Run(ctx, leaderElection); err != nil {
		klog.ErrorS(err, "Error running recommender")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
