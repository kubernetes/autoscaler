/*
Copyright 2020 The Kubernetes Authors.

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

package clusterapi

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	authv1 "k8s.io/api/authorization/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	klog "k8s.io/klog/v2"
)

type unstructuredScalableResource struct {
	controller         *machineController
	unstructured       *unstructured.Unstructured
	maxSize            int
	minSize            int
	autoscalingOptions map[string]string
}

func (r unstructuredScalableResource) ID() string {
	return path.Join(r.Kind(), r.Namespace(), r.Name())
}

func (r unstructuredScalableResource) MaxSize() int {
	return r.maxSize
}

func (r unstructuredScalableResource) MinSize() int {
	return r.minSize
}

func (r unstructuredScalableResource) Kind() string {
	return r.unstructured.GetKind()
}

func (r unstructuredScalableResource) GroupVersionResource() (schema.GroupVersionResource, error) {
	switch r.Kind() {
	case machineDeploymentKind:
		return r.controller.machineDeploymentResource, nil
	case machineSetKind:
		return r.controller.machineSetResource, nil
	case machinePoolKind:
		return r.controller.machinePoolResource, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("unknown scalable resource kind %s", r.Kind())
	}
}

func (r unstructuredScalableResource) Name() string {
	return r.unstructured.GetName()
}

func (r unstructuredScalableResource) Namespace() string {
	return r.unstructured.GetNamespace()
}

func (r unstructuredScalableResource) ProviderIDs() ([]string, error) {
	providerIds, err := r.controller.scalableResourceProviderIDs(r.unstructured)
	if err != nil {
		return nil, err
	}

	return providerIds, nil
}

func (r unstructuredScalableResource) Replicas() (int, error) {
	gvr, err := r.GroupVersionResource()
	if err != nil {
		return 0, err
	}

	s, err := r.controller.managementScaleClient.Scales(r.Namespace()).Get(context.TODO(), gvr.GroupResource(), r.Name(), metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	if s == nil {
		return 0, fmt.Errorf("failed to fetch resource scale: unknown %s %s/%s", r.Kind(), r.Namespace(), r.Name())
	}
	return int(s.Spec.Replicas), nil
}

func (r unstructuredScalableResource) SetSize(nreplicas int) error {
	switch {
	case nreplicas > r.maxSize:
		return fmt.Errorf("size increase too large - desired:%d max:%d", nreplicas, r.maxSize)
	case nreplicas < r.minSize:
		return fmt.Errorf("size decrease too large - desired:%d min:%d", nreplicas, r.minSize)
	}

	gvr, err := r.GroupVersionResource()
	if err != nil {
		return err
	}

	s, err := r.controller.managementScaleClient.Scales(r.Namespace()).Get(context.TODO(), gvr.GroupResource(), r.Name(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	if s == nil {
		return fmt.Errorf("unknown %s %s/%s", r.Kind(), r.Namespace(), r.Name())
	}

	s.Spec.Replicas = int32(nreplicas)
	_, updateErr := r.controller.managementScaleClient.Scales(r.Namespace()).Update(context.TODO(), gvr.GroupResource(), s, metav1.UpdateOptions{})

	if updateErr == nil {
		updateErr = unstructured.SetNestedField(r.unstructured.UnstructuredContent(), int64(nreplicas), "spec", "replicas")
	}

	return updateErr
}

func (r unstructuredScalableResource) UnmarkMachineForDeletion(machine *unstructured.Unstructured) error {
	u, err := r.controller.managementClient.Resource(r.controller.machineResource).Namespace(machine.GetNamespace()).Get(context.TODO(), machine.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	annotations := u.GetAnnotations()
	delete(annotations, machineDeleteAnnotationKey)
	u.SetAnnotations(annotations)
	_, updateErr := r.controller.managementClient.Resource(r.controller.machineResource).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})

	return updateErr
}

func (r unstructuredScalableResource) MarkMachineForDeletion(machine *unstructured.Unstructured) error {
	u, err := r.controller.managementClient.Resource(r.controller.machineResource).Namespace(machine.GetNamespace()).Get(context.TODO(), machine.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	u = u.DeepCopy()

	annotations := u.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[machineDeleteAnnotationKey] = time.Now().String()
	u.SetAnnotations(annotations)

	_, updateErr := r.controller.managementClient.Resource(r.controller.machineResource).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})

	return updateErr
}

func (r unstructuredScalableResource) Labels() map[string]string {
	annotations := r.unstructured.GetAnnotations()
	// annotation value of the form "key1=value1,key2=value2"
	if val, found := annotations[labelsKey]; found {
		labels := strings.Split(val, ",")
		kv := make(map[string]string, len(labels))
		for _, label := range labels {
			split := strings.SplitN(label, "=", 2)
			if len(split) == 2 {
				kv[split[0]] = split[1]
			}
		}
		return kv
	}
	return nil
}

func (r unstructuredScalableResource) Taints() []apiv1.Taint {
	annotations := r.unstructured.GetAnnotations()
	// annotation value the form of "key1=value1:condition,key2=value2:condition"
	if val, found := annotations[taintsKey]; found {
		taints := strings.Split(val, ",")
		ret := make([]apiv1.Taint, 0, len(taints))
		for _, taintStr := range taints {
			taint, err := parseTaint(taintStr)
			if err == nil {
				ret = append(ret, taint)
			}
		}
		return ret
	}
	return nil
}

// A node group can scale from zero if it can inform about the CPU and memory
// capacity of the nodes within the group.
func (r unstructuredScalableResource) CanScaleFromZero() bool {
	capacity, err := r.InstanceCapacity()
	if err != nil {
		return false
	}
	// CPU and memory are the minimum necessary for scaling from zero
	_, cpuOk := capacity[corev1.ResourceCPU]
	_, memOk := capacity[corev1.ResourceMemory]

	return cpuOk && memOk
}

// Inspect the annotations on the scalable resource, and the status.capacity
// field of the machine template infrastructure resource to build the projected
// capacity for this node group. The returned map will be empty if the
// provider does not support scaling from zero, or the annotations have not
// been added.
func (r unstructuredScalableResource) InstanceCapacity() (map[corev1.ResourceName]resource.Quantity, error) {
	capacityAnnotations := map[corev1.ResourceName]resource.Quantity{}

	cpu, err := r.InstanceCPUCapacityAnnotation()
	if err != nil {
		return nil, err
	}
	if !cpu.IsZero() {
		capacityAnnotations[corev1.ResourceCPU] = cpu
	}

	mem, err := r.InstanceMemoryCapacityAnnotation()
	if err != nil {
		return nil, err
	}
	if !mem.IsZero() {
		capacityAnnotations[corev1.ResourceMemory] = mem
	}

	disk, err := r.InstanceEphemeralDiskCapacityAnnotation()
	if err != nil {
		return nil, err
	}
	if !disk.IsZero() {
		capacityAnnotations[corev1.ResourceEphemeralStorage] = disk
	}

	gpuCount, err := r.InstanceGPUCapacityAnnotation()
	if err != nil {
		return nil, err
	}
	gpuType := r.InstanceGPUTypeAnnotation()
	if !gpuCount.IsZero() && gpuType != "" {
		capacityAnnotations[corev1.ResourceName(gpuType)] = gpuCount
	}

	maxPods, err := r.InstanceMaxPodsCapacityAnnotation()
	if err != nil {
		return nil, err
	}
	if maxPods.IsZero() {
		maxPods = *resource.NewQuantity(defaultMaxPods, resource.DecimalSI)
	}
	capacityAnnotations[corev1.ResourcePods] = maxPods

	infraObj, err := r.readInfrastructureReferenceResource()
	if err != nil || infraObj == nil {
		// err="machinetemplates.infrastructure.cluster.x-k8s.io \"TestMachineTemplate\" not found"
		// infraObj=nil
		// because it is possible that the infrastructure provider does not implement
		// the capacity in the infrastructure reference, if there are annotations we
		// should return them here.
		// Check against 1 here because the max pods is always set.
		if len(capacityAnnotations) > 1 {
			return capacityAnnotations, nil
		}
		return nil, err
	}
	capacityInfraStatus := resourceCapacityFromInfrastructureObject(infraObj)

	// The annotations should override any values from the status block of the machine template.
	// We loop through the status block capacity first, then overwrite any values with the
	// annotation capacities.
	capacity := map[corev1.ResourceName]resource.Quantity{}
	for k, v := range capacityInfraStatus {
		capacity[k] = v
	}
	for k, v := range capacityAnnotations {
		capacity[k] = v
	}

	return capacity, nil
}

func (r unstructuredScalableResource) InstanceEphemeralDiskCapacityAnnotation() (resource.Quantity, error) {
	return parseEphemeralDiskCapacity(r.unstructured.GetAnnotations())
}

func (r unstructuredScalableResource) InstanceCPUCapacityAnnotation() (resource.Quantity, error) {
	return parseCPUCapacity(r.unstructured.GetAnnotations())
}

func (r unstructuredScalableResource) InstanceMemoryCapacityAnnotation() (resource.Quantity, error) {
	return parseMemoryCapacity(r.unstructured.GetAnnotations())
}

func (r unstructuredScalableResource) InstanceGPUCapacityAnnotation() (resource.Quantity, error) {
	return parseGPUCount(r.unstructured.GetAnnotations())
}

func (r unstructuredScalableResource) InstanceGPUTypeAnnotation() string {
	return parseGPUType(r.unstructured.GetAnnotations())
}

func (r unstructuredScalableResource) InstanceMaxPodsCapacityAnnotation() (resource.Quantity, error) {
	return parseMaxPodsCapacity(r.unstructured.GetAnnotations())
}

func (r unstructuredScalableResource) readInfrastructureReferenceResource() (*unstructured.Unstructured, error) {
	infraref, found, err := unstructured.NestedStringMap(r.unstructured.Object, "spec", "template", "spec", "infrastructureRef")
	if !found || err != nil {
		return nil, nil
	}

	apiversion, ok := infraref["apiVersion"]
	if !ok {
		return nil, nil
	}
	kind, ok := infraref["kind"]
	if !ok {
		return nil, nil
	}
	name, ok := infraref["name"]
	if !ok {
		return nil, nil
	}
	// kind needs to be lower case and plural
	kind = fmt.Sprintf("%ss", strings.ToLower(kind))
	gvk := schema.FromAPIVersionAndKind(apiversion, kind)
	res := schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: gvk.Kind}
	for _, v := range []string{"list", "watch"} {
		action := authv1.ResourceAttributes{
			Namespace: r.Namespace(),
			Verb:      v,
			Resource:  gvk.Kind,
			Version:   gvk.Version,
			Group:     gvk.Group,
		}
		selfCheck := authv1.SelfSubjectAccessReview{
			Spec: authv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &action,
			},
		}
		resp, err := r.controller.managementClientset.AuthorizationV1().
			SelfSubjectAccessReviews().
			Create(context.TODO(), &selfCheck, metav1.CreateOptions{})
		if err != nil || !resp.Status.Allowed {
			klog.V(4).Infof("Unable to validate infrastructure reference authorization, error: %v", err)
			return nil, err
		}
	}

	infra, err := r.controller.getInfrastructureResource(res, name, r.Namespace())
	if err != nil {
		klog.V(4).Infof("Unable to read infrastructure reference, error: %v", err)
		return nil, err
	}

	return infra, nil
}

func newUnstructuredScalableResource(controller *machineController, u *unstructured.Unstructured) (*unstructuredScalableResource, error) {
	annotations := u.GetAnnotations()
	minSize, maxSize, err := parseScalingBounds(annotations)
	if err != nil {
		return nil, errors.Wrap(err, "error validating min/max annotations")
	}

	return &unstructuredScalableResource{
		controller:         controller,
		unstructured:       u,
		maxSize:            maxSize,
		minSize:            minSize,
		autoscalingOptions: autoscalingOptions(annotations),
	}, nil
}

func resourceCapacityFromInfrastructureObject(infraobj *unstructured.Unstructured) map[corev1.ResourceName]resource.Quantity {
	capacity := map[corev1.ResourceName]resource.Quantity{}

	infracap, found, err := unstructured.NestedStringMap(infraobj.Object, "status", "capacity")
	if !found || err != nil {
		return capacity
	}

	for k, v := range infracap {
		// if we cannot parse the quantity, don't add it to the capacity
		if value, err := resource.ParseQuantity(v); err == nil {
			capacity[corev1.ResourceName(k)] = value
		}
	}

	return capacity
}

// adapted from https://github.com/kubernetes/kubernetes/blob/release-1.25/pkg/util/taints/taints.go#L39
func parseTaint(st string) (apiv1.Taint, error) {
	var taint apiv1.Taint

	var key string
	var value string
	var effect apiv1.TaintEffect

	parts := strings.Split(st, ":")
	switch len(parts) {
	case 1:
		key = parts[0]
	case 2:
		effect = apiv1.TaintEffect(parts[1])

		partsKV := strings.Split(parts[0], "=")
		if len(partsKV) > 2 {
			return taint, fmt.Errorf("invalid taint spec: %v", st)
		}
		key = partsKV[0]
		if len(partsKV) == 2 {
			value = partsKV[1]
			if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
				return taint, fmt.Errorf("invalid taint spec: %v, %s", st, strings.Join(errs, "; "))
			}
		}
	default:
		return taint, fmt.Errorf("invalid taint spec: %v", st)
	}

	if errs := validation.IsQualifiedName(key); len(errs) > 0 {
		return taint, fmt.Errorf("invalid taint spec: %v, %s", st, strings.Join(errs, "; "))
	}

	taint.Key = key
	taint.Value = value
	taint.Effect = effect

	return taint, nil
}
