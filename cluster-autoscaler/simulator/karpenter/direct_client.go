/*
Copyright 2026 The Kubernetes Authors.

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

package karpenter

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	csitranslation "k8s.io/csi-translation-lib"
	"sigs.k8s.io/controller-runtime/pkg/client"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

var (
	translator         = csitranslation.New()
	directClientScheme = func() *runtime.Scheme {
		scheme := runtime.NewScheme()
		_ = apiv1.AddToScheme(scheme)
		_ = storagev1.AddToScheme(scheme)
		gv := schema.GroupVersion{Group: "karpenter.sh", Version: "v1"}
		scheme.AddKnownTypes(gv, &karpenterv1.NodePool{}, &karpenterv1.NodePoolList{}, &karpenterv1.NodeClaim{}, &karpenterv1.NodeClaimList{})
		return scheme
	}()
)

// DirectClient is a read-only facade over a ClusterSnapshot implementing controller-runtime's client.Client.
// WHY: Karpenter controllers and scheduling primitives rely on controller-runtime's client.Client interface
// to query nodes, pods, storage classes, and CSINodes. DirectClient wraps CA's ClusterSnapshot to expose a lightweight,
// in-memory client without requiring a real API server or etcd connection.
type DirectClient struct {
	snapshot clustersnapshot.ClusterSnapshot
	scheme   *runtime.Scheme

	// WHY (In-Memory Multi-Indexing): Karpenter's topology and scheduling algorithms query pods heavily by node name
	// (spec.nodeName) and label selectors during every simulation iteration. Standard linear scans over all pods in the snapshot
	// would incur O(P * N) overhead per scheduling pass. Pre-building podsByName, podsByNode, and podsByLabel indexes upon construction
	// reduces field/label selector lookups to O(1) time complexity, significantly speeding up Karpenter solver iterations.
	podsByName  map[types.NamespacedName]*apiv1.Pod
	podsByNode  map[string][]*apiv1.Pod
	podsByLabel map[string]map[string][]*apiv1.Pod // labelKey -> labelValue -> pods
	nodesByName map[string]*apiv1.Node
	nodePools   []*karpenterv1.NodePool
	nodeClaims  []*karpenterv1.NodeClaim
}

// NewDirectClient returns a new DirectClient populated from the given pods and nodes.
func NewDirectClient(snapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, nodes []*apiv1.Node, nodePools []*karpenterv1.NodePool, nodeClaims []*karpenterv1.NodeClaim) *DirectClient {
	c := &DirectClient{
		snapshot:    snapshot,
		scheme:      directClientScheme,
		podsByName:  make(map[types.NamespacedName]*apiv1.Pod),
		podsByNode:  make(map[string][]*apiv1.Pod),
		podsByLabel: make(map[string]map[string][]*apiv1.Pod),
		nodesByName: make(map[string]*apiv1.Node),
		nodePools:   nodePools,
		nodeClaims:  nodeClaims,
	}

	for _, n := range nodes {
		if n != nil {
			c.nodesByName[n.Name] = n
		}
	}

	for _, p := range pods {
		if p == nil {
			continue
		}
		c.podsByName[types.NamespacedName{Namespace: p.Namespace, Name: p.Name}] = p
		if p.Spec.NodeName != "" {
			c.podsByNode[p.Spec.NodeName] = append(c.podsByNode[p.Spec.NodeName], p)
		}
		for k, v := range p.Labels {
			if _, ok := c.podsByLabel[k]; !ok {
				c.podsByLabel[k] = make(map[string][]*apiv1.Pod)
			}
			c.podsByLabel[k][v] = append(c.podsByLabel[k][v], p)
		}
	}
	return c
}

// Get retrieves an obj for the given object key.
func (c *DirectClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	switch o := obj.(type) {
	case *apiv1.Node:
		if n, ok := c.nodesByName[key.Name]; ok {
			*o = *n
			return nil
		}
		nodeInfo, err := c.snapshot.GetNodeInfo(key.Name)
		if err != nil || nodeInfo == nil || nodeInfo.Node() == nil {
			return apierrors.NewNotFound(apiv1.Resource("nodes"), key.Name)
		}
		*o = *nodeInfo.Node()
		return nil
	case *apiv1.Pod:
		if p, ok := c.podsByName[key]; ok {
			*o = *p
			return nil
		}
		return apierrors.NewNotFound(apiv1.Resource("pods"), key.Name)
	case *karpenterv1.NodePool:
		for _, np := range c.nodePools {
			if np != nil && np.Name == key.Name {
				*o = *np
				return nil
			}
		}
		return apierrors.NewNotFound(schema.GroupResource{Group: "karpenter.sh", Resource: "nodepools"}, key.Name)
	case *apiv1.PersistentVolume:
		pv, err := c.snapshot.GetPV(key.Name)
		if err != nil || pv == nil {
			return apierrors.NewNotFound(apiv1.Resource("persistentvolumes"), key.Name)
		}
		*o = *pv
		return nil
	case *apiv1.PersistentVolumeClaim:
		pvc, err := c.snapshot.GetPVC(key.Namespace, key.Name)
		if err != nil || pvc == nil {
			return apierrors.NewNotFound(apiv1.Resource("persistentvolumeclaims"), key.Name)
		}
		*o = *pvc
		return nil
	case *storagev1.StorageClass:
		sc, err := c.snapshot.GetStorageClass(key.Name)
		if err != nil || sc == nil {
			return apierrors.NewNotFound(storagev1.Resource("storageclasses"), key.Name)
		}
		*o = *sc
		return nil
	case *storagev1.CSINode:
		csi, err := c.snapshot.GetCSINode(key.Name)
		if err != nil || csi == nil {
			csi = c.mockCSINode(key.Name)
		}
		if csi != nil {
			*o = *csi
			return nil
		}
		return apierrors.NewNotFound(storagev1.Resource("csinodes"), key.Name)
	default:
		return fmt.Errorf("unsupported Get for type %T", obj)
	}
}

// List retrieves list of objects for a given namespace and list options.
func (c *DirectClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listOpts)
	}

	switch l := list.(type) {
	case *apiv1.NodeList:
		nodeInfos, err := c.snapshot.NodeInfos().List()
		if err != nil {
			return err
		}
		l.Items = make([]apiv1.Node, 0, len(nodeInfos))
		for _, ni := range nodeInfos {
			if ni == nil || ni.Node() == nil {
				continue
			}
			if n, ok := c.nodesByName[ni.Node().Name]; ok {
				l.Items = append(l.Items, *n)
			} else {
				l.Items = append(l.Items, *ni.Node())
			}
		}
		return nil
	case *apiv1.PodList:
		var candidatePods []*apiv1.Pod
		indexed := false

		// 1. Try Node Index
		if listOpts.FieldSelector != nil {
			if val, ok := listOpts.FieldSelector.RequiresExactMatch("spec.nodeName"); ok {
				candidatePods = c.podsByNode[val]
				indexed = true
			}
		}

		// 2. Try Label Index (Equality matches only)
		if !indexed && listOpts.LabelSelector != nil {
			if requirements, ok := listOpts.LabelSelector.Requirements(); ok {
				for _, req := range requirements {
					if req.Operator() == selection.Equals || req.Operator() == selection.DoubleEquals || req.Operator() == selection.In {
						values := req.Values().UnsortedList()
						if len(values) == 1 {
							if pods, ok := c.podsByLabel[req.Key()]; ok {
								candidatePods = pods[values[0]]
								indexed = true
								break
							}
						}
					}
				}
			}
		}

		// 3. Fallback to full scan
		if !indexed {
			candidatePods = make([]*apiv1.Pod, 0, len(c.podsByName))
			for _, p := range c.podsByName {
				candidatePods = append(candidatePods, p)
			}
		}

		// Filter and copy
		var matched []apiv1.Pod
		for _, p := range candidatePods {
			if p == nil {
				continue
			}
			if listOpts.Namespace != "" && p.Namespace != listOpts.Namespace {
				continue
			}
			if listOpts.LabelSelector != nil && !listOpts.LabelSelector.Matches(labels.Set(p.Labels)) {
				continue
			}
			matched = append(matched, *p)
		}
		l.Items = matched
		return nil
	case *karpenterv1.NodePoolList:
		l.Items = make([]karpenterv1.NodePool, 0, len(c.nodePools))
		for _, np := range c.nodePools {
			l.Items = append(l.Items, *np)
		}
		return nil
	case *karpenterv1.NodeClaimList:
		l.Items = make([]karpenterv1.NodeClaim, 0, len(c.nodeClaims))
		for _, nc := range c.nodeClaims {
			l.Items = append(l.Items, *nc)
		}
		return nil
	case *apiv1.PersistentVolumeList:
		pvs, err := c.snapshot.ListPVs()
		if err != nil {
			return err
		}
		l.Items = make([]apiv1.PersistentVolume, 0, len(pvs))
		for _, pv := range pvs {
			l.Items = append(l.Items, *pv)
		}
		return nil
	case *apiv1.PersistentVolumeClaimList:
		pvcs, err := c.snapshot.ListPVCs()
		if err != nil {
			return err
		}
		l.Items = make([]apiv1.PersistentVolumeClaim, 0, len(pvcs))
		for _, pvc := range pvcs {
			l.Items = append(l.Items, *pvc)
		}
		return nil
	case *storagev1.StorageClassList:
		scs, err := c.snapshot.ListStorageClasses()
		if err != nil {
			return err
		}
		l.Items = make([]storagev1.StorageClass, 0, len(scs))
		for _, sc := range scs {
			l.Items = append(l.Items, *sc)
		}
		return nil
	case *storagev1.CSINodeList:
		csis, err := c.snapshot.ListCSINodes()
		if err != nil {
			return err
		}
		l.Items = make([]storagev1.CSINode, 0, len(csis))
		for _, csi := range csis {
			l.Items = append(l.Items, *csi)
		}
		return nil
	default:
		return fmt.Errorf("unsupported List for type %T", list)
	}
}

// Create is not supported.
func (c *DirectClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "Create")
}

// Delete is not supported.
func (c *DirectClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "Delete")
}

// Update is not supported.
func (c *DirectClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "Update")
}

// Patch is not supported.
func (c *DirectClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "Patch")
}

// DeleteAllOf is not supported.
func (c *DirectClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "DeleteAllOf")
}

// Apply is not supported.
func (c *DirectClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "Apply")
}

// Status returns a SubResourceWriter which is not supported.
func (c *DirectClient) Status() client.SubResourceWriter {
	return &unsupportedSubResourceWriter{}
}

// SubResource returns a SubResourceClient which is not supported.
func (c *DirectClient) SubResource(subResource string) client.SubResourceClient {
	return &unsupportedSubResourceClient{}
}

// Scheme returns the Scheme this client was configured with.
func (c *DirectClient) Scheme() *runtime.Scheme {
	return c.scheme
}

// RESTMapper returns the RESTMapper this client was configured with.
func (c *DirectClient) RESTMapper() meta.RESTMapper {
	return nil
}

// GroupVersionKindFor returns the GVK for the given object.
func (c *DirectClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	kinds, _, err := c.scheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if len(kinds) == 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("no kinds registered for type %T", obj)
	}
	return kinds[0], nil
}

// IsObjectNamespaced returns true if the object is namespaced.
func (c *DirectClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	gvk, err := c.GroupVersionKindFor(obj)
	if err != nil {
		return false, err
	}
	return gvk.Kind != "Node" && gvk.Kind != "PersistentVolume" && gvk.Kind != "StorageClass", nil
}

type unsupportedSubResourceWriter struct{}

func (w *unsupportedSubResourceWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "CreateSubResource")
}

func (w *unsupportedSubResourceWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "UpdateSubResource")
}

func (w *unsupportedSubResourceWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "PatchSubResource")
}

func (w *unsupportedSubResourceWriter) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.SubResourceApplyOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "ApplySubResource")
}

type unsupportedSubResourceClient struct{}

func (c *unsupportedSubResourceClient) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "GetSubResource")
}

func (c *unsupportedSubResourceClient) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "CreateSubResource")
}

func (c *unsupportedSubResourceClient) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "UpdateSubResource")
}

func (c *unsupportedSubResourceClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "PatchSubResource")
}

func (c *unsupportedSubResourceClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.SubResourceApplyOption) error {
	return apierrors.NewMethodNotSupported(schema.GroupResource{}, "ApplySubResource")
}

// mockCSINode constructs a fallback CSINode object populated from snapshot StorageClasses.
// WHY: Karpenter's VolumeTopology constraint evaluator inspects CSINode objects to discover driver topology keys
// (e.g. topology.kubernetes.io/zone). In unit tests or minimal simulation environments where real CSINode objects
// are not populated in the snapshot, mockCSINode dynamically derives registered CSI drivers from active StorageClasses
// (using CSI translation for in-tree volume plugins) and equips them with standard zone and hostname topology keys,
// preventing volume topology scheduling validation failures.
func (c *DirectClient) mockCSINode(nodeName string) *storagev1.CSINode {
	csiNode := &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
	}
	scs, err := c.snapshot.ListStorageClasses()
	if err != nil {
		return csiNode
	}
	for _, sc := range scs {
		driverName := sc.Provisioner
		if csiName, err := translator.GetCSINameFromInTreeName(sc.Provisioner); err == nil {
			driverName = csiName
		}
		found := false
		for _, d := range csiNode.Spec.Drivers {
			if d.Name == driverName {
				found = true
				break
			}
		}
		if !found {
			csiNode.Spec.Drivers = append(csiNode.Spec.Drivers, storagev1.CSINodeDriver{
				Name:         driverName,
				NodeID:       nodeName,
				TopologyKeys: []string{apiv1.LabelZoneFailureDomainStable, apiv1.LabelHostname},
			})
		}
	}
	return csiNode
}
