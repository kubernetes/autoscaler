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
	"fmt"
	"regexp"
	"sync"

	"gopkg.in/yaml.v2"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

type priority struct {
	fallbackStrategy expander.Strategy
	changesChan      <-chan watch.Event
	priorities       map[int][]*regexp.Regexp
	padlock          sync.RWMutex
	okConfigUpdates  int
	badConfigUpdates int
	logRecorder      EventRecorder
}

// NewStrategy returns an expansion strategy that picks node groups based on user-defined priorities
func NewStrategy(initialPriorities string, priorityChangesChan <-chan watch.Event,
	logRecorder EventRecorder) (expander.Strategy, errors.AutoscalerError) {
	res := &priority{
		fallbackStrategy: random.NewStrategy(),
		changesChan:      priorityChangesChan,
		logRecorder:      logRecorder,
	}
	if err := res.parsePrioritiesYAMLString(initialPriorities); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	go func() {
		// TODO: how to terminate on process shutdown?
		for event := range priorityChangesChan {
			cm, ok := event.Object.(*apiv1.ConfigMap)
			if !ok {
				klog.Exit("Unexpected object type received on the configmap update channel in priority expander")
			}

			if event.Type == watch.Deleted {
				msg := "Configmap for priority expander was deleted, no updates will be processed until recreated."
				res.logConfigWarning("PriorityConfigMapDeleted", msg)
				continue
			}

			prioString, found := cm.Data[ConfigMapKey]
			if !found {
				msg := fmt.Sprintf("Wrong configmap for priority expander, doesn't contain %s key. Ignoring update.",
					ConfigMapKey)
				res.logConfigWarning("PriorityConfigMapInvalid", msg)
				continue
			}
			if err := res.parsePrioritiesYAMLString(prioString); err != nil {
				msg := fmt.Sprintf("Wrong configuration for priority expander: %v. Ignoring update.", err)
				res.logConfigWarning("PriorityConfigMapInvalid", msg)
				continue
			}
		}
	}()
	return res, nil
}

func (p *priority) logConfigWarning(reason, msg string) {
	p.logRecorder.Event(apiv1.EventTypeWarning, reason, msg)
	klog.Warning(msg)
	p.badConfigUpdates++
}

func (p *priority) parsePrioritiesYAMLString(prioritiesYAML string) error {
	if prioritiesYAML == "" {
		p.badConfigUpdates++
		return fmt.Errorf("priority configuration in %s configmap is empty; please provide valid configuration", PriorityConfigMapName)
	}
	var config map[int][]string
	if err := yaml.Unmarshal([]byte(prioritiesYAML), &config); err != nil {
		p.badConfigUpdates++
		return fmt.Errorf("Can't parse YAML with priorities in the configmap: %v", err)
	}

	newPriorities := make(map[int][]*regexp.Regexp)
	for prio, reList := range config {
		for _, re := range reList {
			regexp, err := regexp.Compile(re)
			if err != nil {
				p.badConfigUpdates++
				return fmt.Errorf("Can't compile regexp rule for priority %d and rule %s: %v", prio, re, err)
			}
			newPriorities[prio] = append(newPriorities[prio], regexp)
		}
	}

	p.padlock.Lock()
	p.priorities = newPriorities
	p.okConfigUpdates++
	p.padlock.Unlock()

	msg := "Successfully reloaded priority configuration from configmap."
	klog.V(4).Info(msg)
	p.logRecorder.Event(apiv1.EventTypeNormal, "PriorityConfigMapReloaded", msg)

	return nil
}

func (p *priority) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulernodeinfo.NodeInfo) *expander.Option {
	if len(expansionOptions) <= 0 {
		return nil
	}

	maxPrio := -1
	best := []expander.Option{}
	p.padlock.RLock()
	for _, option := range expansionOptions {
		id := option.NodeGroup.Id()
		found := false
		for prio, nameRegexpList := range p.priorities {
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
			p.logConfigWarning("PriorityConfigMapNotMatchedGroup", msg)
		}
	}
	p.padlock.RUnlock()

	if len(best) == 0 {
		msg := "Priority expander: no priorities info found for any of the expansion options. Falling back to random choice."
		p.logConfigWarning("PriorityConfigMapNoGroupMatched", msg)
		return p.fallbackStrategy.BestOption(expansionOptions, nodeInfo)
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

// EventRecorder is an interface to abstract kubernetes event recording.
type EventRecorder interface {
	// Event records a new event of given type, reason and description given with message.
	Event(eventtype, reason, message string)

	// Events records a new event of given type, reason and description given with message,
	// which can be formatted using args.
	Eventf(eventtype, reason, message string, args ...interface{})
}
