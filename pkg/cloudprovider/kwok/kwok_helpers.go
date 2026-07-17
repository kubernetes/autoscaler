/*
Copyright 2023 The Kubernetes Authors.

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

package kwok

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

const (
	templatesKey               = "templates"
	defaultTemplatesConfigName = "kwok-provider-templates"
)

type listerFn func(lister v1lister.NodeLister, filter func(*apiv1.Node) bool) kube_util.NodeLister

func loadNodeTemplatesFromCluster(kc *KwokProviderConfig,
	kubeClient kubernetes.Interface,
	lister kube_util.NodeLister) ([]*apiv1.Node, error) {

	if lister != nil {
		return lister.List()
	}

	nodeList, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nos := []*apiv1.Node{}
	// note: not using _, node := range nodeList.Items here because it leads to unexpected behavior
	// more info: https://stackoverflow.com/a/38693163/6874596
	for i := range nodeList.Items {
		nos = append(nos, &(nodeList.Items[i]))
	}

	return nos, nil
}

// LoadNodeTemplatesFromConfigMap loads template nodes from a k8s configmap
// check https://github.com/vadafoss/node-templates for more info on the parsing logic
func LoadNodeTemplatesFromConfigMap(configMapName string,
	kubeClient kubernetes.Interface) ([]*apiv1.Node, error) {
	currentNamespace := getCurrentNamespace()
	nodeTemplates := []*apiv1.Node{}

	c, err := kubeClient.CoreV1().ConfigMaps(currentNamespace).Get(context.Background(), configMapName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap '%s': %v", configMapName, err)
	}

	if c.Data[templatesKey] == "" {
		return nil, fmt.Errorf("configmap '%s' doesn't have 'templates' key", configMapName)
	}

	scheme := runtime.NewScheme()
	clientscheme.AddToScheme(scheme)

	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	multiDocReader := yaml.NewYAMLReader(bufio.NewReader(strings.NewReader(c.Data[templatesKey])))

	objs := []runtime.Object{}

	for {
		buf, err := multiDocReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		obj, _, err := decoder.Decode(buf, nil, nil)
		if err != nil {
			return nil, err
		}

		objs = append(objs, obj)
	}

	if len(objs) > 1 {
		for _, obj := range objs {
			if node, ok := obj.(*apiv1.Node); ok {
				nodeTemplates = append(nodeTemplates, node)
			}
		}

	} else if nodelist, ok := objs[0].(*apiv1.List); ok {
		for _, item := range nodelist.Items {

			o, _, err := decoder.Decode(item.Raw, nil, nil)
			if err != nil {
				return nil, err
			}

			if node, ok := o.(*apiv1.Node); ok {
				nodeTemplates = append(nodeTemplates, node)
			}
		}
	} else {
		return nil, errors.New("invalid templates file (found something other than nodes in the file)")
	}

	return nodeTemplates, nil
}

func createNodegroups(nodes []*apiv1.Node, kubeClient kubernetes.Interface, kc *KwokProviderConfig, initCustomLister listerFn,
	allNodeLister v1lister.NodeLister) []*NodeGroup {
	ngs := map[string]*NodeGroup{}

	// note: not using _, node := range nodes here because it leads to unexpected behavior
	// more info: https://stackoverflow.com/a/38693163/6874596
	for i := range nodes {

		belongsToNg := ((kc.status.groupNodesBy == groupNodesByAnnotation &&
			nodes[i].GetAnnotations()[kc.status.key] != "") ||
			(kc.status.groupNodesBy == groupNodesByLabel &&
				nodes[i].GetLabels()[kc.status.key] != ""))
		if !belongsToNg {
			continue
		}

		ngName := getNGName(nodes[i], kc)
		if ngName == "" {
			klog.Fatalf("%s '%s' for node '%s' not present in the manifest",
				kc.status.groupNodesBy, kc.status.key,
				nodes[i].GetName())
		}

		if ngs[ngName] != nil {
			ngs[ngName].targetSize += 1
			continue
		}

		ng := parseAnnotations(nodes[i], kc)
		ng.name = ngName
		sanitizeNode(nodes[i])
		prepareNode(nodes[i], ng.name)
		ng.nodeTemplate = nodes[i]

		filterFn := func(no *apiv1.Node) bool {
			return no.GetAnnotations()[NGNameAnnotation] == ng.name
		}

		ng.kubeClient = kubeClient
		ng.lister = initCustomLister(allNodeLister, filterFn)

		ngs[ngName] = ng
	}

	result := []*NodeGroup{}
	for i := range ngs {
		result = append(result, ngs[i])
	}
	return result
}

// sanitizeNode cleans the node
func sanitizeNode(no *apiv1.Node) {
	no.ResourceVersion = ""
	no.Generation = 0
	no.UID = ""
	no.CreationTimestamp = v1.Time{}
	no.Status.NodeInfo.KubeletVersion = "fake"

}

// prepareNode prepares node as a kwok template node
func prepareNode(no *apiv1.Node, ngName string) {
	// add prefix in the name to make it clear that this node is different
	// from the ones already existing in the cluster (in case there is a name clash)
	no.Name = fmt.Sprintf("kwok-fake-%s", no.GetName())
	no.Annotations[KwokManagedAnnotation] = "fake"
	no.Annotations[NGNameAnnotation] = ngName
	no.Spec.ProviderID = getProviderID(no.GetName())
}

func getProviderID(nodeName string) string {
	return fmt.Sprintf("kwok:%s", nodeName)
}

func parseAnnotations(no *apiv1.Node, kc *KwokProviderConfig) *NodeGroup {
	min := 0
	max := 200
	target := min
	if no.GetAnnotations()[NGMinSizeAnnotation] != "" {
		if mi, err := strconv.Atoi(no.GetAnnotations()[NGMinSizeAnnotation]); err == nil {
			min = mi
		} else {
			klog.Fatalf("invalid value for annotation key '%s' for node '%s'", NGMinSizeAnnotation, no.GetName())
		}
	}

	if no.GetAnnotations()[NGMaxSizeAnnotation] != "" {
		if ma, err := strconv.Atoi(no.GetAnnotations()[NGMaxSizeAnnotation]); err == nil {
			max = ma
		} else {
			klog.Fatalf("invalid value for annotation key '%s' for node '%s'", NGMaxSizeAnnotation, no.GetName())
		}
	}

	if no.GetAnnotations()[NGDesiredSizeAnnotation] != "" {
		if ta, err := strconv.Atoi(no.GetAnnotations()[NGDesiredSizeAnnotation]); err == nil {
			target = ta
		} else {
			klog.Fatalf("invalid value for annotation key '%s' for node '%s'", NGDesiredSizeAnnotation, no.GetName())
		}
	}

	if max < min {
		log.Fatalf("min-count '%d' cannot be lesser than max-count '%d' for the node '%s'", min, max, no.GetName())
	}

	if target > max || target < min {
		log.Fatalf("desired-count '%d' cannot be lesser than min-count '%d' or greater than max-count '%d' for the node '%s'", target, min, max, no.GetName())
	}

	return &NodeGroup{
		minSize:    min,
		maxSize:    max,
		targetSize: target,
	}
}

// getNGName returns the node group name of the given k8s node object.
// Return empty string if no node group is found.
func getNGName(no *apiv1.Node, kc *KwokProviderConfig) string {

	if no.GetAnnotations()[NGNameAnnotation] != "" {
		return no.GetAnnotations()[NGNameAnnotation]
	}

	var ngName string
	switch kc.status.groupNodesBy {
	case "annotation":
		ngName = no.GetAnnotations()[kc.status.key]
	case "label":
		ngName = no.GetLabels()[kc.status.key]
	default:
		klog.Warning("grouping criteria for nodes is not set (expected: 'annotation' or 'label')")
	}

	return ngName
}
