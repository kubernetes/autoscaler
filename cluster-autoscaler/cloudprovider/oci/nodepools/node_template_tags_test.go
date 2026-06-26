/*
Copyright 2026 Oracle and/or its affiliates.
*/

package nodepools

import (
	"reflect"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/config"

	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	npconsts "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/nodepools/consts"
	ocisdk "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/core"
)

type fakeShapeGetter struct {
	shape *ocicommon.Shape
}

func (f fakeShapeGetter) GetNodePoolShape(*oke.NodePool, int64) (*ocicommon.Shape, error) {
	return f.shape, nil
}

func (f fakeShapeGetter) GetInstancePoolShape(*core.InstancePool) (*ocicommon.Shape, error) {
	return f.shape, nil
}

func (f fakeShapeGetter) Refresh() {}

func TestExtractNodeTemplateLabelsFromTags(t *testing.T) {
	tags := map[string]string{
		nodeTemplateLabelTagPrefix + "workload":     "batch",
		nodeTemplateLabelTagPrefix + "custom-label": "example.com/workload=custom",
		nodeTemplateLabelTagPrefix:                  "ignored",
		"unrelated":                                 "ignored",
	}

	got := extractNodeTemplateLabelsFromTags(tags)
	want := map[string]string{
		"workload":             "batch",
		"example.com/workload": "custom",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got labels %+v, want %+v", got, want)
	}
}

func TestExtractNodeTemplateResourcesFromTags(t *testing.T) {
	tags := map[string]string{
		nodeTemplateResourcesTagPrefix + "cpu":               "100m",
		nodeTemplateResourcesTagPrefix + "memory":            "100M",
		nodeTemplateResourcesTagPrefix + "ephemeral-storage": "20G",
		nodeTemplateResourcesTagPrefix + "custom-resource":   "example.com/custom-resource=5",
		nodeTemplateResourcesTagPrefix + "invalid-resource":  "not-a-quantity",
		nodeTemplateResourcesTagPrefix:                       "1",
		"unrelated":                                          "ignored",
	}

	got := extractNodeTemplateResourcesFromTags(tags)
	if len(got) != 4 {
		t.Fatalf("got %d resources, want 4: %+v", len(got), got)
	}

	want := apiv1.ResourceList{
		apiv1.ResourceCPU:                                 resource.MustParse("100m"),
		apiv1.ResourceMemory:                              resource.MustParse("100M"),
		apiv1.ResourceEphemeralStorage:                    resource.MustParse("20G"),
		apiv1.ResourceName("example.com/custom-resource"): resource.MustParse("5"),
	}
	for name, quantity := range want {
		gotQuantity := got[name]
		if gotQuantity.Cmp(quantity) != 0 {
			t.Fatalf("got resource %s=%s, want %s", name, gotQuantity.String(), quantity.String())
		}
	}
}

func TestExtractNodeTemplateTaintsFromTags(t *testing.T) {
	tags := map[string]string{
		nodeTemplateTaintTagPrefix + "dedicated": "batch:NoSchedule",
		nodeTemplateTaintTagPrefix + "group":     "system:NoExecute",
		nodeTemplateTaintTagPrefix + "app":       "fizz:PreferNoSchedule",
		nodeTemplateTaintTagPrefix + "custom":    "example.com/dedicated=custom:NoSchedule",
		nodeTemplateTaintTagPrefix + "bad":       "missing-effect",
		nodeTemplateTaintTagPrefix + "unknown":   "value:UnknownEffect",
		nodeTemplateTaintTagPrefix:               "value:NoSchedule",
		"unrelated":                              "ignored",
	}

	got := extractNodeTemplateTaintsFromTags(tags)
	want := []apiv1.Taint{
		{Key: "dedicated", Value: "batch", Effect: apiv1.TaintEffectNoSchedule},
		{Key: "group", Value: "system", Effect: apiv1.TaintEffectNoExecute},
		{Key: "app", Value: "fizz", Effect: apiv1.TaintEffectPreferNoSchedule},
		{Key: "example.com/dedicated", Value: "custom", Effect: apiv1.TaintEffectNoSchedule},
	}

	if !reflect.DeepEqual(taintSet(got), taintSet(want)) {
		t.Fatalf("got taints %+v, want %+v", got, want)
	}
}

func TestBuildNodeGroupAutoscalingOptionsFromTags(t *testing.T) {
	defaults := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.1,
		ScaleDownGpuUtilizationThreshold: 0.2,
		ScaleDownUnneededTime:            time.Second,
		ScaleDownUnreadyTime:             time.Minute,
		IgnoreDaemonSetsUtilization:      false,
	}

	tests := []struct {
		name string
		tags map[string]string
		want config.NodeGroupAutoscalingOptions
	}{
		{
			name: "use defaults on unspecified tags",
			tags: map[string]string{},
			want: defaults,
		},
		{
			name: "keep defaults on invalid tag values",
			tags: map[string]string{
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUtilizationThresholdKey: "not-a-float",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUnneededTimeKey:         "not-a-duration",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUnreadyTimeKey:          "",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultIgnoreDaemonSetsUtilizationKey:   "not-a-bool",
			},
			want: defaults,
		},
		{
			name: "use provided tags and fill missing with defaults",
			tags: map[string]string{
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUtilizationThresholdKey:    "0.42",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownGpuUtilizationThresholdKey: "0.7",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUnneededTimeKey:            "1h",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUnreadyTimeKey:             "25m",
				nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultIgnoreDaemonSetsUtilizationKey:      "true",
				nodeTemplateAutoscalingOptionsTagPrefix + "unknown":                                         "ignored",
			},
			want: config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold:    0.42,
				ScaleDownGpuUtilizationThreshold: 0.7,
				ScaleDownUnneededTime:            time.Hour,
				ScaleDownUnreadyTime:             25 * time.Minute,
				IgnoreDaemonSetsUtilization:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNodeGroupAutoscalingOptionsFromTags(tt.tags, defaults, "nodepool-id")
			if !reflect.DeepEqual(*got, tt.want) {
				t.Fatalf("got options %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestBuildNodeFromTemplateWithNodeTemplateTags(t *testing.T) {
	customResource := apiv1.ResourceName("example.com/custom-resource")
	manager := &ociManagerImpl{
		ociShapeGetter: fakeShapeGetter{
			shape: &ocicommon.Shape{
				Name:          "VM.Standard.E3.Flex",
				CPU:           4,
				MemoryInBytes: 16 * 1024 * 1024 * 1024,
				GPU:           0,
			},
		},
		ociTagsGetter:          ocicommon.CreateTagsGetter(),
		registeredTaintsGetter: CreateRegisteredTaintsGetter(),
	}
	nodePool := testNodePool(map[string]string{
		npconsts.EphemeralStorageSize:                          "1Gi",
		nodeTemplateResourcesTagPrefix + "custom-resource":     string(customResource) + "=1",
		nodeTemplateResourcesTagPrefix + "ephemeral-storage":   "2Gi",
		nodeTemplateLabelTagPrefix + "workload":                "example.com/workload=batch",
		nodeTemplateLabelTagPrefix + "instance-type":           apiv1.LabelInstanceTypeStable + "=tagged-shape",
		nodeTemplateTaintTagPrefix + "node-template-dedicated": "batch:NoSchedule",
	})
	nodePool.NodeMetadata = map[string]string{
		"kubelet-extra-args": "--register-with-taints=registered=yes:NoSchedule",
	}

	node, err := manager.buildNodeFromTemplate(nodePool)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if got := node.Status.Capacity[customResource]; got.Value() != 1 {
		t.Fatalf("got custom resource capacity %s, want 1", got.String())
	}
	if got := node.Status.Allocatable[customResource]; got.Value() != 1 {
		t.Fatalf("got custom resource allocatable %s, want 1", got.String())
	}
	ephemeralStorage := node.Status.Capacity[apiv1.ResourceEphemeralStorage]
	expectedEphemeralStorage := resource.MustParse("2Gi")
	if ephemeralStorage.Value() != expectedEphemeralStorage.Value() {
		t.Fatalf("got ephemeral-storage capacity %s, want 2Gi", ephemeralStorage.String())
	}
	if got := node.Labels["example.com/workload"]; got != "batch" {
		t.Fatalf("got template label %q, want batch", got)
	}
	if got := node.Labels[apiv1.LabelInstanceTypeStable]; got != "tagged-shape" {
		t.Fatalf("got instance type label %q, want tagged-shape", got)
	}
	if !hasTaint(node.Spec.Taints, apiv1.Taint{Key: "registered", Value: "yes", Effect: apiv1.TaintEffectNoSchedule}) {
		t.Fatalf("expected registered taint in %+v", node.Spec.Taints)
	}
	if !hasTaint(node.Spec.Taints, apiv1.Taint{Key: "node-template-dedicated", Value: "batch", Effect: apiv1.TaintEffectNoSchedule}) {
		t.Fatalf("expected node template taint in %+v", node.Spec.Taints)
	}
}

func TestBuildNodeFromTemplateWithoutNodeTemplateTags(t *testing.T) {
	manager := &ociManagerImpl{
		ociShapeGetter: fakeShapeGetter{
			shape: &ocicommon.Shape{
				Name:          "VM.Standard.E3.Flex",
				CPU:           4,
				MemoryInBytes: 16 * 1024 * 1024 * 1024,
				GPU:           0,
			},
		},
		ociTagsGetter:          ocicommon.CreateTagsGetter(),
		registeredTaintsGetter: CreateRegisteredTaintsGetter(),
	}

	node, err := manager.buildNodeFromTemplate(testNodePool(nil))
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if _, found := node.Status.Capacity[apiv1.ResourceName("example.com/custom-resource")]; found {
		t.Fatalf("unexpected custom resource capacity: %+v", node.Status.Capacity)
	}
	if _, found := node.Status.Capacity[apiv1.ResourceEphemeralStorage]; found {
		t.Fatalf("unexpected ephemeral-storage capacity: %+v", node.Status.Capacity)
	}
	if _, found := node.Labels["example.com/workload"]; found {
		t.Fatalf("unexpected template label: %+v", node.Labels)
	}
	if len(node.Spec.Taints) != 0 {
		t.Fatalf("got taints %+v, want none", node.Spec.Taints)
	}
}

func TestBuildNodeFromTemplateContinuesOnInvalidRegisteredTaints(t *testing.T) {
	manager := &ociManagerImpl{
		ociShapeGetter: fakeShapeGetter{
			shape: &ocicommon.Shape{
				Name:          "VM.Standard.E3.Flex",
				CPU:           4,
				MemoryInBytes: 16 * 1024 * 1024 * 1024,
				GPU:           0,
			},
		},
		ociTagsGetter:          ocicommon.CreateTagsGetter(),
		registeredTaintsGetter: CreateRegisteredTaintsGetter(),
	}
	nodePool := testNodePool(map[string]string{
		nodeTemplateTaintTagPrefix + "template": "batch:NoSchedule",
	})
	nodePool.NodeMetadata = map[string]string{
		"kubelet-extra-args": "--register-with-taints=invalid-taint",
	}

	node, err := manager.buildNodeFromTemplate(nodePool)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if !hasTaint(node.Spec.Taints, apiv1.Taint{Key: "template", Value: "batch", Effect: apiv1.TaintEffectNoSchedule}) {
		t.Fatalf("expected node template taint in %+v", node.Spec.Taints)
	}
}

func TestNodePoolGetOptions(t *testing.T) {
	cache := newNodePoolCache(nil)
	cache.cache["nodepool-id"] = testNodePool(map[string]string{
		nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUtilizationThresholdKey: "0.42",
		nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultScaleDownUnneededTimeKey:         "1h",
		nodeTemplateAutoscalingOptionsTagPrefix + config.DefaultIgnoreDaemonSetsUtilizationKey:   "true",
	})
	manager := &ociManagerImpl{
		nodePoolCache: cache,
		ociTagsGetter: ocicommon.CreateTagsGetter(),
	}
	np := &nodePool{
		id:      "nodepool-id",
		manager: manager,
	}
	defaults := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.1,
		ScaleDownGpuUtilizationThreshold: 0.2,
		ScaleDownUnneededTime:            time.Second,
		ScaleDownUnreadyTime:             time.Minute,
		IgnoreDaemonSetsUtilization:      false,
	}

	got, err := np.GetOptions(defaults)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	want := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.42,
		ScaleDownGpuUtilizationThreshold: 0.2,
		ScaleDownUnneededTime:            time.Hour,
		ScaleDownUnreadyTime:             time.Minute,
		IgnoreDaemonSetsUtilization:      true,
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("got options %+v, want %+v", *got, want)
	}
}

func testNodePool(freeformTags map[string]string) *oke.NodePool {
	return &oke.NodePool{
		Id:           ocisdk.String("nodepool-id"),
		NodeShape:    ocisdk.String("VM.Standard.E3.Flex"),
		FreeformTags: freeformTags,
		NodeConfigDetails: &oke.NodePoolNodeConfigDetails{
			PlacementConfigs: []oke.NodePoolPlacementConfigDetails{
				{AvailabilityDomain: ocisdk.String("Uocm:PHX-AD-1")},
			},
		},
		InitialNodeLabels: []oke.KeyValue{
			{Key: ocisdk.String("initial"), Value: ocisdk.String("label")},
		},
	}
}

func taintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := map[apiv1.Taint]bool{}
	for _, taint := range taints {
		set[taint] = true
	}
	return set
}

func hasTaint(taints []apiv1.Taint, want apiv1.Taint) bool {
	for _, taint := range taints {
		if taint == want {
			return true
		}
	}
	return false
}
