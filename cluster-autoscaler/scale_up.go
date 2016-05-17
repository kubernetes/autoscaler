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

package main

import (
	"fmt"

	"k8s.io/contrib/cluster-autoscaler/config"
	"k8s.io/contrib/cluster-autoscaler/estimator"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/contrib/cluster-autoscaler/utils/gce"
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/golang/glog"
)

// ExpansionOption describes an option to expand the cluster.
type ExpansionOption struct {
	migConfig *config.MigConfig
	estimator *estimator.BasicNodeEstimator
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occured.
func ScaleUp(unschedulablePods []*kube_api.Pod, nodes []*kube_api.Node, migConfigs []*config.MigConfig,
	gceManager *gce.GceManager, kubeClient *kube_client.Client,
	predicateChecker *simulator.PredicateChecker) (bool, error) {

	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		glog.V(1).Info("No unschedulable pods")
		return false, nil
	}

	for _, pod := range unschedulablePods {
		glog.V(1).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}

	expansionOptions := make([]ExpansionOption, 0)
	nodeInfos, err := GetNodeInfosForMigs(nodes, gceManager, kubeClient)
	if err != nil {
		return false, fmt.Errorf("failed to build node infors for migs: %v", err)
	}

	for _, migConfig := range migConfigs {

		currentSize, err := gceManager.GetMigSize(migConfig)
		if err != nil {
			glog.Errorf("Failed to get MIG size: %v", err)
			continue
		}
		if currentSize >= int64(migConfig.MaxSize) {
			// skip this mig.
			glog.V(4).Infof("Skipping MIG %s - max size reached", migConfig.Url())
			continue
		}

		option := ExpansionOption{
			migConfig: migConfig,
			estimator: estimator.NewBasicNodeEstimator(),
		}
		migHelpsSomePods := false

		nodeInfo, found := nodeInfos[migConfig.Url()]
		if !found {
			glog.Errorf("No node info for: %s", migConfig.Url())
			continue
		}

		for _, pod := range unschedulablePods {
			err = predicateChecker.CheckPredicates(pod, nodeInfo)
			if err == nil {
				migHelpsSomePods = true
				option.estimator.Add(pod)
			} else {
				glog.V(2).Infof("Scale-up predicate failed: %v", err)
			}
		}
		if migHelpsSomePods {
			expansionOptions = append(expansionOptions, option)
		}
	}

	// Pick some expansion option.
	bestOption := BestExpansionOption(expansionOptions)
	if bestOption != nil && bestOption.estimator.GetCount() > 0 {
		glog.V(1).Infof("Best option to resize: %s", bestOption.migConfig.Url())
		nodeInfo, found := nodeInfos[bestOption.migConfig.Url()]
		if !found {
			return false, fmt.Errorf("no sample node for: %s", bestOption.migConfig.Url())

		}
		node := nodeInfo.Node()
		estimate, report := bestOption.estimator.Estimate(node)
		glog.V(1).Info(bestOption.estimator.GetDebug())
		glog.V(1).Info(report)
		glog.V(1).Infof("Estimated %d nodes needed in %s", estimate, bestOption.migConfig.Url())

		currentSize, err := gceManager.GetMigSize(bestOption.migConfig)
		if err != nil {
			return false, fmt.Errorf("failed to get MIG size: %v", err)
		}
		newSize := currentSize + int64(estimate)
		if newSize >= int64(bestOption.migConfig.MaxSize) {
			newSize = int64(bestOption.migConfig.MaxSize)
		}
		glog.V(1).Infof("Setting %s size to %d", bestOption.migConfig.Url(), newSize)

		if err := gceManager.SetMigSize(bestOption.migConfig, newSize); err != nil {
			return false, fmt.Errorf("failed to set MIG size: %v", err)
		}
		return true, nil
	}
	return false, nil
}
