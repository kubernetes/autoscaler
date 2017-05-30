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

package price

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// TODO: add preferred node
type priceBased struct {
	pricingModel cloudprovider.PricingModel
}

// NewStrategy returns an expansion strategy that picks nodes based on price and preferred node type.
func NewStrategy(pricingModel cloudprovider.PricingModel) expander.Strategy {
	return &priceBased{
		pricingModel: pricingModel,
	}
}

// BestOption selects option based on cost and preferred node type.
func (p *priceBased) BestOption(expansionOptions []expander.Option, nodeInfos map[string]*schedulercache.NodeInfo) *expander.Option {
	var bestOption *expander.Option
	bestOptionScore := 0.0
	now := time.Now()
	then := now.Add(time.Hour)

nextoption:
	for i, option := range expansionOptions {
		nodeInfo, found := nodeInfos[option.NodeGroup.Id()]
		if !found {
			glog.Warningf("No node info for %s", option.NodeGroup.Id())
			continue
		}
		nodePrice, err := p.pricingModel.NodePrice(nodeInfo.Node(), now, then)
		if err != nil {
			glog.Warningf("Failed to calculate node price for %s: %v", option.NodeGroup.Id(), err)
			continue
		}
		totalNodePrice := nodePrice * float64(option.NodeCount)
		totalPodPrice := 0.0
		for _, pod := range option.Pods {
			podPrice, err := p.pricingModel.PodPrice(pod, now, then)
			if err != nil {
				glog.Warningf("Failed to calculate pod price for %s/%s: %v", pod.Namespace, pod.Name, err)
				continue nextoption
			}
			totalPodPrice += podPrice
		}
		if totalPodPrice == 0 {
			glog.Warningf("Total pod price is 0, skipping %s", option.NodeGroup.Id())
			continue
		}

		optionScore := totalNodePrice / totalPodPrice
		glog.V(5).Infof("Price of %s expansion is %f - ratio %f", option.NodeGroup.Id(), totalNodePrice)

		if bestOption == nil || bestOptionScore > optionScore {
			bestOption = &expansionOptions[i]
			bestOptionScore = optionScore
		}
	}
	return bestOption
}
