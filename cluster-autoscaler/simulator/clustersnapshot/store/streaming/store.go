/*
Copyright 2024 The Kubernetes Authors.

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

package streaming

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store/fort"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

type StreamingSnapshotStore struct {
	lock fort.LockGroup

	nodesInformer fort.ManualSharedInformer
	podsInformer  fort.ManualSharedInformer

	// Cache for NodeInfo objects
	nodeInfoCache fort.BTreeMap[*framework.NodeInfo]

	// Index: PVC Key -> Count of pods using it
	pvcUsage fort.CloneableSharedInformerQuery

	// Stack for Fork/Revert
	forkStack []snapshotState
}

type snapshotState struct {
	nodes         fort.ManualSharedInformer
	pods          fort.ManualSharedInformer
	nodeInfoCache fort.BTreeMap[*framework.NodeInfo]
	pvcUsage      fort.CloneableSharedInformerQuery
}

func NewStreamingSnapshotStore() *StreamingSnapshotStore {
	lock := fort.NewLockGroup()
	nodes := fort.NewManualSharedInformerWithOptions(lock, fort.DefaultKeyFunc)
	pods := fort.NewManualSharedInformerWithOptions(lock, fort.DefaultKeyFunc)
	nodes.SetHasSynced()
	pods.SetHasSynced()

	// Track PVC usage
	pvcUsage := fort.QueryInformer(&fort.FlatMap[string, *apiv1.Pod]{
		Lock: lock,
		Over: pods,
		Map: func(p *apiv1.Pod) ([]string, error) {
			res := []string{}
			for _, v := range p.Spec.Volumes {
				if v.PersistentVolumeClaim != nil {
					res = append(res, fmt.Sprintf("%s/%s", p.Namespace, v.PersistentVolumeClaim.ClaimName))
				}
			}
			return res, nil
		},
	})

	// Count occurrences of each PVC key
	pvcCounts := fort.QueryInformer(&fort.GroupBy[int, string]{
		Lock: lock,
		From: pvcUsage,
		GroupBy: func(pvcKey string) (any, []fort.GroupField) {
			return pvcKey, []fort.GroupField{fort.Count()}
		},
		Select: func(fields []fort.GroupField) (int, error) {
			return int(fields[0].(int64)), nil
		},
	})

	s := &StreamingSnapshotStore{
		lock:          lock,
		nodesInformer: nodes,
		podsInformer:  pods,
		nodeInfoCache: fort.NewBTreeMap[*framework.NodeInfo](),
		pvcUsage:      pvcCounts,
	}
	return s
}

func (s *StreamingSnapshotStore) GetPodInformer() cache.SharedInformer {
	return s.podsInformer
}

func (s *StreamingSnapshotStore) GetNodeInformer() cache.SharedInformer {
	return s.nodesInformer
}

// Implement ClusterSnapshotStore interface

func (s *StreamingSnapshotStore) NodeInfos() schedulerinterface.NodeInfoLister {
	return s
}

func (s *StreamingSnapshotStore) StorageInfos() schedulerinterface.StorageInfoLister {
	return s
}

func (s *StreamingSnapshotStore) IsPVCUsedByPods(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	count, ok, _ := s.pvcUsage.GetStore().GetByKey(key)
	return ok && count.(int) > 0
}

func (s *StreamingSnapshotStore) List() ([]schedulerinterface.NodeInfo, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	items := s.nodeInfoCache.List()
	res := make([]schedulerinterface.NodeInfo, 0, len(items))
	for _, ni := range items {
		res = append(res, ni)
	}
	return res, nil
}

func (s *StreamingSnapshotStore) HavePodsWithAffinityList() ([]schedulerinterface.NodeInfo, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	items := s.nodeInfoCache.List()
	res := make([]schedulerinterface.NodeInfo, 0)
	for _, ni := range items {
		if len(ni.PodsWithAffinity) > 0 {
			res = append(res, ni)
		}
	}
	return res, nil
}

func (s *StreamingSnapshotStore) HavePodsWithRequiredAntiAffinityList() ([]schedulerinterface.NodeInfo, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	items := s.nodeInfoCache.List()
	res := make([]schedulerinterface.NodeInfo, 0)
	for _, ni := range items {
		if len(ni.PodsWithRequiredAntiAffinity) > 0 {
			res = append(res, ni)
		}
	}
	return res, nil
}

func (s *StreamingSnapshotStore) Get(nodeName string) (schedulerinterface.NodeInfo, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	ni, ok := s.nodeInfoCache.Get(nodeName)
	if !ok {
		return nil, clustersnapshot.ErrNodeNotFound
	}
	return ni, nil
}

func (s *StreamingSnapshotStore) Fork() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.forkStack = append(s.forkStack, snapshotState{
		nodes:         s.nodesInformer,
		pods:          s.podsInformer,
		nodeInfoCache: s.nodeInfoCache,
		pvcUsage:      s.pvcUsage,
	})

	s.nodesInformer = s.nodesInformer.Clone(nil).(fort.ManualSharedInformer)
	s.podsInformer = s.podsInformer.Clone(nil).(fort.ManualSharedInformer)
	s.nodeInfoCache = s.nodeInfoCache.Clone()

	memo := make(map[cache.SharedInformer]cache.SharedInformer)
	memo[s.forkStack[len(s.forkStack)-1].pods] = s.podsInformer
	s.pvcUsage = fort.ClonePipeline(s.pvcUsage, memo).(fort.CloneableSharedInformerQuery)
}
func (s *StreamingSnapshotStore) Revert() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.forkStack) == 0 {
		return
	}

	last := s.forkStack[len(s.forkStack)-1]
	s.forkStack = s.forkStack[:len(s.forkStack)-1]

	s.nodesInformer = last.nodes
	s.podsInformer = last.pods
	s.nodeInfoCache = last.nodeInfoCache
	s.pvcUsage = last.pvcUsage
}

func (s *StreamingSnapshotStore) Commit() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.forkStack) == 0 {
		return nil
	}

	s.forkStack = s.forkStack[:len(s.forkStack)-1]
	return nil
}

func (s *StreamingSnapshotStore) StorePodInfo(podInfo *framework.PodInfo, nodeName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pod := podInfo.Pod
	s.podsInformer.OnAddLocked(pod, false)

	// Update cache
	if ni, ok := s.nodeInfoCache.Get(nodeName); ok {
		newNi := ni.DeepCopy()
		newNi.AddPod(framework.NewPodInfo(pod, podInfo.NeededResourceClaims))
		s.nodeInfoCache.Set(nodeName, newNi)
	} else {
		return clustersnapshot.ErrNodeNotFound
	}
	return nil
}

func (s *StreamingSnapshotStore) RemovePodInfo(namespace string, podName string, nodeName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, podName)
	obj, exists, _ := s.podsInformer.GetStore().GetByKey(key)
	if exists {
		s.podsInformer.OnDeleteLocked(obj)

		// Update cache
		if ni, ok := s.nodeInfoCache.Get(nodeName); ok {
			newNi := ni.DeepCopy()
			newNi.RemovePod(klog.Background(), obj.(*apiv1.Pod))
			s.nodeInfoCache.Set(nodeName, newNi)
		} else {
			return clustersnapshot.ErrNodeNotFound
		}
	} else {
		return fmt.Errorf("pod %s/%s not found", namespace, podName)
	}
	return nil
}

func (s *StreamingSnapshotStore) StoreNodeInfo(nodeInfo *framework.NodeInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.nodeInfoCache.Get(nodeInfo.Node().Name); ok {
		return fmt.Errorf("node %s already exists", nodeInfo.Node().Name)
	}

	s.nodesInformer.OnAddLocked(nodeInfo.Node(), false)

	// Create a new NodeInfo and add all pods from nodeInfo to it.
	newNi := framework.NewNodeInfo(nodeInfo.Node(), nodeInfo.LocalResourceSlices)
	if nodeInfo.CSINode != nil {
		newNi.SetCSINode(nodeInfo.CSINode)
	}

	for _, podInfo := range nodeInfo.Pods() {
		pod := podInfo.Pod
		s.podsInformer.OnAddLocked(pod, false)
		newNi.AddPod(framework.NewPodInfo(pod, podInfo.NeededResourceClaims))
	}

	s.nodeInfoCache.Set(nodeInfo.Node().Name, newNi)
	return nil
}

func (s *StreamingSnapshotStore) RemoveNodeInfo(nodeName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	obj, exists, _ := s.nodesInformer.GetStore().GetByKey(nodeName)
	if exists {
		// Remove all pods of this node from podsInformer
		if ni, ok := s.nodeInfoCache.Get(nodeName); ok {
			for _, podInfo := range ni.Pods() {
				s.podsInformer.OnDeleteLocked(podInfo.Pod)
			}
		}

		s.nodesInformer.OnDeleteLocked(obj)
		s.nodeInfoCache.Delete(nodeName)
	} else {
		return clustersnapshot.ErrNodeNotFound
	}
	return nil
}

func (s *StreamingSnapshotStore) Clear() {
	s.lock.Lock()
	s.forkStack = nil
	s.nodeInfoCache = fort.NewBTreeMap[*framework.NodeInfo]()
	s.lock.Unlock()

	s.nodesInformer.Clear()
	s.podsInformer.Clear()

	// pvcUsage is already linked to podsInformer, so it will be cleared automatically via events.
}
