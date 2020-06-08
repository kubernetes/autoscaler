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

package priority

import (
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v2"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	caserrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	// PriorityConfigMapName defines a name of the ConfigMap used to store priority expander configuration
	PriorityConfigMapName = "cluster-autoscaler-priority-expander"
	// ConfigMapKey defines the key used in the ConfigMap to configure priorities
	ConfigMapKey = "priorities"
)

type priorities map[int][]*regexp.Regexp

type priority struct {
	logRecorder      record.EventRecorder
	fallbackStrategy expander.Strategy
	okConfigUpdates  int
	badConfigUpdates int
	configMapLister  v1lister.ConfigMapNamespaceLister
}

// NewStrategy returns an expansion strategy that picks node groups based on user-defined priorities
func NewStrategy(configMapLister v1lister.ConfigMapNamespaceLister,
	logRecorder record.EventRecorder) (expander.Strategy, caserrors.AutoscalerError) {
	res := &priority{
		logRecorder:      logRecorder,
		fallbackStrategy: random.NewStrategy(),
		configMapLister:  configMapLister,
	}
	return res, nil
}

func (p *priority) reloadConfigMap() (priorities, *apiv1.ConfigMap, error) {
	cm, err := p.configMapLister.Get(PriorityConfigMapName)
	if err != nil {
		return nil, nil, fmt.Errorf("Priority expander config map %s not found: %v", PriorityConfigMapName, err)
	}

	prioString, found := cm.Data[ConfigMapKey]
	if !found {
		msg := fmt.Sprintf("Wrong configmap for priority expander, doesn't contain %s key. Ignoring update.",
			ConfigMapKey)
		p.logConfigWarning(cm, "PriorityConfigMapInvalid", msg)
		return nil, cm, errors.New(msg)
	}

	newPriorities, err := p.parsePrioritiesYAMLString(prioString)
	if err != nil {
		msg := fmt.Sprintf("Wrong configuration for priority expander: %v. Ignoring update.", err)
		p.logConfigWarning(cm, "PriorityConfigMapInvalid", msg)
		return nil, cm, err
	}

	return newPriorities, cm, nil
}

func (p *priority) logConfigWarning(cm *apiv1.ConfigMap, reason, msg string) {
	p.logRecorder.Event(cm, apiv1.EventTypeWarning, reason, msg)
	klog.Warning(msg)
	p.badConfigUpdates++
}

func (p *priority) parsePrioritiesYAMLString(prioritiesYAML string) (priorities, error) {
	if prioritiesYAML == "" {
		return nil, fmt.Errorf("priority configuration in %s configmap is empty; please provide valid configuration",
			PriorityConfigMapName)
	}
	var config map[int][]string
	if err := yaml.Unmarshal([]byte(prioritiesYAML), &config); err != nil {
		return nil, fmt.Errorf("Can't parse YAML with priorities in the configmap: %v", err)
	}

	newPriorities := make(map[int][]*regexp.Regexp)
	for prio, reList := range config {
		for _, re := range reList {
			regexp, err := regexp.Compile(re)
			if err != nil {
				return nil, fmt.Errorf("Can't compile regexp rule for priority %d and rule %s: %v", prio, re, err)
			}
			newPriorities[prio] = append(newPriorities[prio], regexp)
		}
	}

	p.okConfigUpdates++
	msg := "Successfully loaded priority configuration from configmap."
	klog.V(4).Info(msg)

	return newPriorities, nil
}

func (p *priority) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
	if len(expansionOptions) <= 0 {
		return nil
	}

	priorities, cm, err := p.reloadConfigMap()
	if err != nil {
		return nil
	}

	maxPrio := -1
	best := []expander.Option{}
	for _, option := range expansionOptions {
		id := option.NodeGroup.Id()
		found := false
		for prio, nameRegexpList := range priorities {
			if prio < maxPrio {
				continue
			}
			if !p.groupIDMatchesList(id, nameRegexpList) {
				continue
			}
			if prio > maxPrio {
				maxPrio = prio
				best = nil
			}
			best = append(best, option)
			found = true
			break
		}
		if !found {
			msg := fmt.Sprintf("Priority expander: node group %s not found in priority expander configuration. "+
				"The group won't be used.", id)
			p.logConfigWarning(cm, "PriorityConfigMapNotMatchedGroup", msg)
		}
	}

	if len(best) == 0 {
		msg := "Priority expander: no priorities info found for any of the expansion options. Falling back to random choice."
		p.logConfigWarning(cm, "PriorityConfigMapNoGroupMatched", msg)
		return p.fallbackStrategy.BestOption(expansionOptions, nodeInfo)
	}

	for _, opt := range best {
		klog.V(2).Infof("priority expander: %s chosen as the highest available", opt.NodeGroup.Id())
	}
	return p.fallbackStrategy.BestOption(best, nodeInfo)
}

func (p *priority) groupIDMatchesList(id string, nameRegexpList []*regexp.Regexp) bool {
	for _, re := range nameRegexpList {
		if re.FindStringIndex(id) != nil {
			return true
		}
	}
	return false
}
