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

package k8sclient

// (Heavily) based on the existing CPVPA code:
// https://github.com/kubernetes-incubator/cluster-proportional-vertical-autoscaler/tree/master/pkg/autoscaler/k8sclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type ResourcePatcher interface {
	UpdateResources(kind, namespace, name string, update *corev1.PodSpec, dryRun bool) error
}

type kubernetesPatcher struct {
	client        kubernetes.Interface
	groupVersions map[string]bool
}

var _ ResourcePatcher = &kubernetesPatcher{}

func NewKubernetesPatcher(client kubernetes.Interface) (ResourcePatcher, error) {
	resourceLists, err := client.Discovery().ServerResources()
	if err != nil {
		return nil, fmt.Errorf("failed to query server preferred namespaced resources: %v", err)
	}

	// TODO: Periodically refresh API versions?

	groupVersions := map[string]bool{}
	for _, resourceList := range resourceLists {
		for _, res := range resourceList.APIResources {
			gvk := resourceList.GroupVersion + "/" + res.Kind
			groupVersions[gvk] = true
		}
	}

	return &kubernetesPatcher{
		client:        client,
		groupVersions: groupVersions,
	}, nil
}

// Captures the namespace and name to patch, and calls the best
// resource-specific patch method.
type patchFunc func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error

// findPatcher returns a groupVersion string and a patch function for the
// specified kind.  This is needed because, at least in theory, the schema of a
// resource could change dramatically, and we should use statically versioned
// types everywhere.  In practice, it's unlikely that the bits we care about
// would change (since we PATCH).  Alas, there's not a great way to dynamically
// use whatever is "latest".  The fallout of this is that we will need to update
// this program when new API group-versions are introduced.
func (k *kubernetesPatcher) findPatcher(kind string) (schema.GroupVersion, patchFunc, error) {
	switch strings.ToLower(kind) {
	case "deployment":
		return findDeploymentPatcher(k.groupVersions)
	case "daemonset":
		return findDaemonSetPatcher(k.groupVersions)
	case "replicaset":
		return findReplicaSetPatcher(k.groupVersions)
	}
	return schema.GroupVersion{}, nil, fmt.Errorf("unknown target kind: %s", kind)
}

func findDeploymentPatcher(groupVersions map[string]bool) (schema.GroupVersion, patchFunc, error) {
	// Find the best API to use - newest API first.
	if groupVersions["apps/v1beta2/Deployment"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.AppsV1beta2().Deployments(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "apps", Version: "v1beta2"}, patchFunc(fn), nil
	}
	if groupVersions["apps/v1beta1/Deployment"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.AppsV1beta1().Deployments(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "apps", Version: "v1beta1"}, patchFunc(fn), nil
	}
	if groupVersions["extensions/v1beta1/Deployment"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.ExtensionsV1beta1().Deployments(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "extensions", Version: "v1beta1"}, patchFunc(fn), nil
	}
	return schema.GroupVersion{}, nil, fmt.Errorf("no supported API group for Deployment: %v", groupVersions)
}

func findDaemonSetPatcher(groupVersions map[string]bool) (schema.GroupVersion, patchFunc, error) {
	// Find the best API to use - newest API first.
	if groupVersions["apps/v1beta2/DaemonSet"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.AppsV1beta2().DaemonSets(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "apps", Version: "v1beta2"}, patchFunc(fn), nil
	}
	if groupVersions["extensions/v1beta1/DaemonSet"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.ExtensionsV1beta1().DaemonSets(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "extensions", Version: "v1beta1"}, patchFunc(fn), nil
	}
	return schema.GroupVersion{}, nil, fmt.Errorf("no supported API group for DaemonSet: %v", groupVersions)
}

func findReplicaSetPatcher(groupVersions map[string]bool) (schema.GroupVersion, patchFunc, error) {
	// Find the best API to use - newest API first.
	if groupVersions["apps/v1beta2/ReplicaSet"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.AppsV1beta2().ReplicaSets(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "extensions", Version: "v1beta2"}, patchFunc(fn), nil
	}
	if groupVersions["extensions/v1beta1/ReplicaSet"] {
		fn := func(client kubernetes.Interface, namespace, name string, pt types.PatchType, data []byte) error {
			_, err := client.ExtensionsV1beta1().ReplicaSets(namespace).Patch(name, pt, data)
			return err
		}
		return schema.GroupVersion{Group: "extensions", Version: "v1beta1"}, patchFunc(fn), nil
	}
	return schema.GroupVersion{}, nil, fmt.Errorf("no supported API group for ReplicaSet: %v", groupVersions)
}

func (k *kubernetesPatcher) UpdateResources(kind, namespace, name string, update *corev1.PodSpec, dryRun bool) error {
	gv, patcher, err := k.findPatcher(kind)
	if err != nil {
		return err
	}

	ctrs := []interface{}{}
	for i := range update.Containers {
		container := &update.Containers[i]
		ctrs = append(ctrs, map[string]interface{}{
			"name":      container.Name,
			"resources": container.Resources,
		})
	}
	podSpec := map[string]interface{}{
		"containers": ctrs,
	}

	spec := make(map[string]interface{})

	switch strings.ToLower(kind) {
	case "replicaset", "deployment", "daemonset":
		spec["template"] = map[string]interface{}{
			"spec": podSpec,
		}

	default:
		return fmt.Errorf("unhandled type: %s", kind)
	}

	patch := map[string]interface{}{
		"apiVersion": gv.String(),
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": spec,
	}

	jb, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("can't marshal patch to JSON: %v", err)
	}

	if dryRun {
		glog.Infof("Performing dry-run, only printing updates:")
		glog.Infof("patch: %s", string(jb))
		return nil
	}
	glog.Infof("patching %s %s/%s: %s", kind, namespace, name, string(jb))
	if err := patcher(k.client, namespace, name, types.StrategicMergePatchType, jb); err != nil {
		return fmt.Errorf("patch failed: %v", err)
	}

	return nil
}
