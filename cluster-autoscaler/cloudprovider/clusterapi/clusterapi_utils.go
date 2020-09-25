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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	deprecatedNodeGroupMinSizeAnnotationKey = "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size"
	deprecatedNodeGroupMaxSizeAnnotationKey = "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size"
	nodeGroupMinSizeAnnotationKey           = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"
	nodeGroupMaxSizeAnnotationKey           = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"
	clusterNameLabel                        = "cluster.x-k8s.io/cluster-name"
	deprecatedClusterNameLabel              = "cluster.k8s.io/cluster-name"
)

var (
	// errMissingMinAnnotation is the error returned when a
	// machine set does not have an annotation keyed by
	// nodeGroupMinSizeAnnotationKey.
	errMissingMinAnnotation = errors.New("missing min annotation")

	// errMissingMaxAnnotation is the error returned when a
	// machine set does not have an annotation keyed by
	// nodeGroupMaxSizeAnnotationKey.
	errMissingMaxAnnotation = errors.New("missing max annotation")

	// errInvalidMinAnnotationValue is the error returned when a
	// machine set has a non-integral min annotation value.
	errInvalidMinAnnotation = errors.New("invalid min annotation")

	// errInvalidMaxAnnotationValue is the error returned when a
	// machine set has a non-integral max annotation value.
	errInvalidMaxAnnotation = errors.New("invalid max annotation")
)

type normalizedProviderID string

// minSize returns the minimum value encoded in the annotations keyed
// by nodeGroupMinSizeAnnotationKey. Returns errMissingMinAnnotation
// if the annotation doesn't exist or errInvalidMinAnnotation if the
// value is not of type int.
func minSize(annotations map[string]string) (int, error) {
	val, found := annotations[nodeGroupMinSizeAnnotationKey]
	if !found {
		val, found = annotations[deprecatedNodeGroupMinSizeAnnotationKey]
	}
	if !found {
		return 0, errMissingMinAnnotation
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.Wrapf(err, "%s", errInvalidMinAnnotation)
	}
	return i, nil
}

// maxSize returns the maximum value encoded in the annotations keyed
// by nodeGroupMaxSizeAnnotationKey. Returns errMissingMaxAnnotation
// if the annotation doesn't exist or errInvalidMaxAnnotation if the
// value is not of type int.
func maxSize(annotations map[string]string) (int, error) {
	val, found := annotations[nodeGroupMaxSizeAnnotationKey]
	if !found {
		val, found = annotations[deprecatedNodeGroupMaxSizeAnnotationKey]
	}
	if !found {
		return 0, errMissingMaxAnnotation
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.Wrapf(err, "%s", errInvalidMaxAnnotation)
	}
	return i, nil
}

func parseScalingBounds(annotations map[string]string) (int, int, error) {
	minSize, err := minSize(annotations)
	if err != nil && err != errMissingMinAnnotation {
		return 0, 0, err
	}

	if minSize < 0 {
		return 0, 0, errInvalidMinAnnotation
	}

	maxSize, err := maxSize(annotations)
	if err != nil && err != errMissingMaxAnnotation {
		return 0, 0, err
	}

	if maxSize < 0 {
		return 0, 0, errInvalidMaxAnnotation
	}

	if maxSize < minSize {
		return 0, 0, errInvalidMaxAnnotation
	}

	return minSize, maxSize, nil
}

func getOwnerForKind(u *unstructured.Unstructured, kind string) *metav1.OwnerReference {
	for _, ref := range u.GetOwnerReferences() {
		if ref.Kind == kind && ref.Name != "" {
			return ref.DeepCopy()
		}
	}

	return nil
}

func machineOwnerRef(machine *unstructured.Unstructured) *metav1.OwnerReference {
	return getOwnerForKind(machine, machineSetKind)
}

func machineSetOwnerRef(machineSet *unstructured.Unstructured) *metav1.OwnerReference {
	return getOwnerForKind(machineSet, machineDeploymentKind)
}

func machineSetHasMachineDeploymentOwnerRef(machineSet *unstructured.Unstructured) bool {
	return machineSetOwnerRef(machineSet) != nil
}

// normalizedProviderString splits s on '/' returning everything after
// the last '/'.
func normalizedProviderString(s string) normalizedProviderID {
	split := strings.Split(s, "/")
	return normalizedProviderID(split[len(split)-1])
}

func clusterNameFromResource(r *unstructured.Unstructured) string {
	// Use Spec.ClusterName if defined (only available on v1alpha3+ types)
	clusterName, found, err := unstructured.NestedString(r.Object, "spec", "clusterName")
	if err != nil {
		return ""
	}

	if found {
		return clusterName
	}

	// Fallback to value of clusterNameLabel
	if clusterName, ok := r.GetLabels()[clusterNameLabel]; ok {
		return clusterName
	}

	// fallback for backward compatibility for deprecatedClusterNameLabel
	if clusterName, ok := r.GetLabels()[deprecatedClusterNameLabel]; ok {
		return clusterName
	}

	// fallback for cluster-api v1alpha1 cluster linking
	templateLabels, found, err := unstructured.NestedStringMap(r.UnstructuredContent(), "spec", "template", "metadata", "labels")
	if found {
		if clusterName, ok := templateLabels[deprecatedClusterNameLabel]; ok {
			return clusterName
		}
	}

	return ""
}
