/*
Copyright 2026 Oracle and/or its affiliates.
*/

package nodepools

import (
	"strconv"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	klog "k8s.io/klog/v2"
)

const (
	nodeTemplateLabelTagPrefix              = "cluster-autoscaler/node-template/label/"
	nodeTemplateTaintTagPrefix              = "cluster-autoscaler/node-template/taint/"
	nodeTemplateResourcesTagPrefix          = "cluster-autoscaler/node-template/resources/"
	nodeTemplateAutoscalingOptionsTagPrefix = "cluster-autoscaler/node-template/autoscaling-options/"
)

func extractNodeTemplateLabelsFromTags(tags map[string]string) map[string]string {
	var labels map[string]string
	for key, value := range tags {
		labelName, found := strings.CutPrefix(key, nodeTemplateLabelTagPrefix)
		if !found || labelName == "" {
			continue
		}
		if labels == nil {
			labels = make(map[string]string)
		}
		if tagValueLabelName, labelValue, hasExplicitName := strings.Cut(value, "="); hasExplicitName && tagValueLabelName != "" {
			labels[tagValueLabelName] = labelValue
			continue
		}
		labels[labelName] = value
	}
	return labels
}

func extractNodeTemplateResourcesFromTags(tags map[string]string) apiv1.ResourceList {
	var resources apiv1.ResourceList
	for key, value := range tags {
		resourceName, found := strings.CutPrefix(key, nodeTemplateResourcesTagPrefix)
		if !found || resourceName == "" {
			continue
		}
		quantityValue := value
		if tagValueResourceName, tagValueQuantity, hasExplicitName := strings.Cut(value, "="); hasExplicitName && tagValueResourceName != "" {
			resourceName = tagValueResourceName
			quantityValue = tagValueQuantity
		}
		quantity, err := resource.ParseQuantity(quantityValue)
		if err != nil {
			klog.Warningf("failed to parse node template resource quantity %q for resource %q: %v", quantityValue, resourceName, err)
			continue
		}
		if resources == nil {
			resources = apiv1.ResourceList{}
		}
		resources[apiv1.ResourceName(resourceName)] = quantity
	}
	return resources
}

func extractNodeTemplateTaintsFromTags(tags map[string]string) []apiv1.Taint {
	var taints []apiv1.Taint
	for key, value := range tags {
		taintKey, found := strings.CutPrefix(key, nodeTemplateTaintTagPrefix)
		if !found || taintKey == "" {
			continue
		}
		taintValue := value
		if tagValueTaintKey, tagValueTaint, hasExplicitName := strings.Cut(value, "="); hasExplicitName && tagValueTaintKey != "" {
			taintKey = tagValueTaintKey
			taintValue = tagValueTaint
		}
		parts := strings.SplitN(taintValue, ":", 2)
		if len(parts) != 2 {
			klog.Warningf("failed to parse node template taint %q for key %q", taintValue, taintKey)
			continue
		}
		effect := apiv1.TaintEffect(parts[1])
		if !isSupportedTaintEffect(effect) {
			klog.Warningf("unsupported node template taint effect %q for key %q", parts[1], taintKey)
			continue
		}
		taints = append(taints, apiv1.Taint{
			Key:    taintKey,
			Value:  parts[0],
			Effect: effect,
		})
	}
	return taints
}

func isSupportedTaintEffect(effect apiv1.TaintEffect) bool {
	switch effect {
	case apiv1.TaintEffectNoSchedule, apiv1.TaintEffectNoExecute, apiv1.TaintEffectPreferNoSchedule:
		return true
	default:
		return false
	}
}

func extractNodeTemplateAutoscalingOptionsFromTags(tags map[string]string) map[string]string {
	var options map[string]string
	for key, value := range tags {
		optionName, found := strings.CutPrefix(key, nodeTemplateAutoscalingOptionsTagPrefix)
		if !found || optionName == "" {
			continue
		}
		if options == nil {
			options = make(map[string]string)
		}
		options[optionName] = value
	}
	return options
}

func buildNodeGroupAutoscalingOptionsFromTags(tags map[string]string, defaults config.NodeGroupAutoscalingOptions, nodePoolID string) *config.NodeGroupAutoscalingOptions {
	options := extractNodeTemplateAutoscalingOptionsFromTags(tags)
	if len(options) == 0 {
		return &defaults
	}

	defaults.ScaleDownUtilizationThreshold = parseFloatOption(options, config.DefaultScaleDownUtilizationThresholdKey, defaults.ScaleDownUtilizationThreshold, nodePoolID)
	defaults.ScaleDownGpuUtilizationThreshold = parseFloatOption(options, config.DefaultScaleDownGpuUtilizationThresholdKey, defaults.ScaleDownGpuUtilizationThreshold, nodePoolID)
	defaults.ScaleDownUnneededTime = parseDurationOption(options, config.DefaultScaleDownUnneededTimeKey, defaults.ScaleDownUnneededTime, nodePoolID)
	defaults.ScaleDownUnreadyTime = parseDurationOption(options, config.DefaultScaleDownUnreadyTimeKey, defaults.ScaleDownUnreadyTime, nodePoolID)
	defaults.IgnoreDaemonSetsUtilization = parseBoolOption(options, config.DefaultIgnoreDaemonSetsUtilizationKey, defaults.IgnoreDaemonSetsUtilization, nodePoolID)

	return &defaults
}

func parseFloatOption(options map[string]string, key string, fallback float64, nodePoolID string) float64 {
	value, found := options[key]
	if !found {
		return fallback
	}
	opt, err := strconv.ParseFloat(value, 64)
	if err != nil {
		klog.Warningf("failed to convert nodepool %s %s tag to float: %v", nodePoolID, key, err)
		return fallback
	}
	return opt
}

func parseDurationOption(options map[string]string, key string, fallback time.Duration, nodePoolID string) time.Duration {
	value, found := options[key]
	if !found {
		return fallback
	}
	opt, err := time.ParseDuration(value)
	if err != nil {
		klog.Warningf("failed to convert nodepool %s %s tag to duration: %v", nodePoolID, key, err)
		return fallback
	}
	return opt
}

func parseBoolOption(options map[string]string, key string, fallback bool, nodePoolID string) bool {
	value, found := options[key]
	if !found {
		return fallback
	}
	opt, err := strconv.ParseBool(value)
	if err != nil {
		klog.Warningf("failed to convert nodepool %s %s tag to bool: %v", nodePoolID, key, err)
		return fallback
	}
	return opt
}
