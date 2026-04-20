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

package predicate

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store/fort"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/interpodaffinity"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

// affinityTermKey uniquely identifies an AffinityTerm by its criteria and topology.
type affinityTermKey struct {
	antiAffinity bool
	termId       string
}

// termEntry stores a compiled AffinityTerm and its anti-affinity status.
type termEntry struct {
	term         *schedulerinterface.AffinityTerm
	antiAffinity bool
}

// PodAffinityContext captures all information required to evaluate a pod's affinity
// against nodes. It is designed to be extracted serially (PreFilter) and then used
// concurrently across multiple nodes (Filter) without further map lookups.
type PodAffinityContext struct {
	pod               *apiv1.Pod
	ownedAntiAffinity []termWithTopology
	ownedAffinity     []termWithTopology
	// matchIDs are the IDs of terms in the system that this pod matches.
	matchIDs []int
}

// termWithTopology pairs a term ID with its target topology key.
type termWithTopology struct {
	id          int
	topologyKey string
}

// matchCacheEntry stores the result of getMatchingTerms to avoid re-evaluating Matches().
type matchCacheEntry struct {
	matchIDs []int
	// termsCount tracks the number of terms in the system when this entry was created,
	// allowing for incremental cache updates as new terms are registered.
	termsCount int
}

// PredicateEvaluator is a high-performance, namespace-aware evaluator for Pod Affinity
// and Anti-Affinity predicates. It uses incremental indexing and Copy-on-Write (CoW)
// data structures to sustain Cluster Autoscaler simulations at the 5,000+ node scale.
//
// Conceptually, it implements a "Two-Way Label Index" (indexed by term and pod labels)
// to avoid O(N^2) matching, and relies on dense integer slices for fast per-node lookups.
type PredicateEvaluator struct {
	// lock is shared with the underlying ClusterSnapshotStore to ensure atomic mutations.
	lock fort.LockGroup

	podInformer  cache.SharedInformer
	nodeInformer cache.SharedInformer

	// Global registry of terms present in the system.
	termRegistry *common.PatchSet[affinityTermKey, int]
	// termsByID is an append-only slice of term definitions.
	termsByID []termEntry
	// termCounts tracks how many pods own each term.
	termCounts []int

	// Label-based indexing for terms to speed up matching new pods against the registry.
	// key -> value -> []termID
	termLabelIndex *common.PatchSet[string, []int]
	// complexTerms contains IDs of terms that cannot be indexed by a simple label.
	complexTerms []int

	// Presence and Forbidden maps track which terms are satisfied or violated in each domain.
	// (TopologyKey \x00 TopologyValue) -> []int32 (counts indexed by term ID)
	presence  fort.BTreeMap[[]int32]
	forbidden fort.BTreeMap[[]int32]

	// hashDomainCounts tracks how many pods of a specific type (hash) exist in each domain.
	// podHash\x00domain -> count
	hashDomainCounts *common.PatchSet[string, int32]
	// hashDomains provides a fast list of domains a pod type occupies.
	hashDomains *common.PatchSet[string, []string]
	// hashRepresentatives stores a sample pod for each unique pod type to run Matches() against.
	hashRepresentatives *common.PatchSet[string, *apiv1.Pod]
	// podLabelIndex indexes pod types by their labels to speed up backfilling new terms.
	// key\x00value -> set[podHash]
	podLabelIndex *common.PatchSet[string, sets.Set[string]]
	// matchCache stores which terms each pod hash matches, preserved across forks.
	matchCache *common.PatchSet[string, matchCacheEntry]

	// podsWithNodes is the reactive joiner that powers the incremental updates.
	podsWithNodes fort.CloneableSharedInformerQuery
	registration  cache.ResourceEventHandlerRegistration
}

// NewPredicateEvaluator constructs a new evaluator and attaches it to the provided informers.
func NewPredicateEvaluator(podInformer, nodeInformer cache.SharedInformer) *PredicateEvaluator {
	var lock fort.LockGroup
	if ci, ok := podInformer.(fort.CloneableSharedInformerQuery); ok {
		lock = ci.GetLockGroup()
	} else {
		lock = fort.NewLockGroup()
	}

	e := &PredicateEvaluator{
		lock:                lock,
		podInformer:         podInformer,
		nodeInformer:        nodeInformer,
		termRegistry:        common.NewPatchSet[affinityTermKey, int](),
		termLabelIndex:      common.NewPatchSet[string, []int](),
		presence:            fort.NewBTreeMap[[]int32](),
		forbidden:           fort.NewBTreeMap[[]int32](),
		hashDomainCounts:    common.NewPatchSet[string, int32](),
		hashDomains:         common.NewPatchSet[string, []string](),
		hashRepresentatives: common.NewPatchSet[string, *apiv1.Pod](),
		podLabelIndex:       common.NewPatchSet[string, sets.Set[string]](),
		matchCache:          common.NewPatchSet[string, matchCacheEntry](),
	}

	// The pipeline joins pods with their nodes to maintain topology-aware indexing.
	e.podsWithNodes = fort.QueryInformer(&fort.Join[framework.PodWithNode, *apiv1.Pod, *apiv1.Node]{
		Lock: lock,
		From: podInformer,
		Join: nodeInformer,
		On: func(p *apiv1.Pod, n *apiv1.Node) any {
			if p != nil {
				return p.Spec.NodeName
			}
			if n != nil {
				return n.Name
			}
			return ""
		},
		Select: func(p *apiv1.Pod, n *apiv1.Node) (framework.PodWithNode, error) {
			return framework.PodWithNode{Pod: p, Node: n}, nil
		},
	})

	e.registration, _ = e.podsWithNodes.AddEventHandler(e)

	return e
}

// PreparePod extracts the pod's affinity state in a thread-safe manner (PreFilter).
// It registers any novel terms the pod owns and pre-calculates matching terms.
func (e *PredicateEvaluator) PreparePod(pod *apiv1.Pod) *PodAffinityContext {
	if pod == nil {
		return nil
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	podInfo := framework.NewPodInfo(pod, nil)

	// Extract selectors from spec to power the Two-Way Label Index.
	var specAff, specAnti []apiv1.PodAffinityTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAffinity != nil {
			specAff = pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			specAnti = pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}

	// ensureTermRegistered handles lazy backfilling of new rules against existing pods.
	prepareTerms := func(terms []schedulerinterface.AffinityTerm, specTerms []apiv1.PodAffinityTerm, anti bool) []termWithTopology {
		var res []termWithTopology
		for i := range terms {
			t := &terms[i]
			tk := affinityTermKey{antiAffinity: anti, termId: getTermId(t)}
			var ls *metav1.LabelSelector
			if i < len(specTerms) {
				ls = specTerms[i].LabelSelector
			}
			id := e.ensureTermRegistered(t, tk, ls)
			res = append(res, termWithTopology{id: id, topologyKey: t.TopologyKey})
		}
		return res
	}

	ownedAff := prepareTerms(podInfo.GetRequiredAffinityTerms(), specAff, false)
	ownedAnti := prepareTerms(podInfo.GetRequiredAntiAffinityTerms(), specAnti, true)

	// matchIDs allows the node loop to use integer slice lookups instead of Matches() string evaluations.
	matchIDs := e.getMatchingTerms(pod)

	return &PodAffinityContext{
		pod:               pod,
		ownedAffinity:     ownedAff,
		ownedAntiAffinity: ownedAnti,
		matchIDs:          matchIDs,
	}
}

func (e *PredicateEvaluator) getPodMatchHash(pod *apiv1.Pod) string {
	if pod == nil {
		return ""
	}
	lblHash := framework.GetPodLabelSetHash(pod)
	// We include namespace to correctly support cross-namespace affinity rules.
	if lblHash == "" && len(pod.Labels) == 0 {
		return pod.Namespace + "\x00"
	}
	return pod.Namespace + "\x00" + lblHash
}

// OnAdd updates all reactive indices when a pod is scheduled onto a node.
func (e *PredicateEvaluator) OnAdd(obj any, isInInitialList bool) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.OnAddLocked(obj, isInInitialList)
}

func (e *PredicateEvaluator) OnAddLocked(obj any, _ bool) {
	pwn := obj.(framework.PodWithNode)
	if pwn.Pod == nil || pwn.Node == nil {
		return
	}

	podInfo := framework.NewPodInfo(pwn.Pod, nil)
	hash := e.getPodMatchHash(pwn.Pod)

	// Tracking unique pod types (hashes) ensures backfilling doesn't become O(Pods).
	if !e.hashRepresentatives.InCurrentPatch(hash) {
		if _, found := e.hashRepresentatives.FindValue(hash); !found {
			e.hashRepresentatives.SetCurrent(hash, pwn.Pod)

			// Index the pod's labels to power the Two-Way Label Index during future rule registration.
			for k, v := range pwn.Pod.Labels {
				kv := k + "\x00" + v
				hashes, found := e.podLabelIndex.FindValue(kv)
				if !found || hashes == nil {
					hashes = sets.New[string]()
				} else {
					hashes = hashes.Clone()
				}
				hashes.Insert(hash)
				e.podLabelIndex.SetCurrent(kv, hashes)
			}
		}
	}

	// 1. Extract and register affinity terms owned by this pod.
	var specAff, specAnti []apiv1.PodAffinityTerm
	if pwn.Pod.Spec.Affinity != nil {
		if pwn.Pod.Spec.Affinity.PodAffinity != nil {
			specAff = pwn.Pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
		if pwn.Pod.Spec.Affinity.PodAntiAffinity != nil {
			specAnti = pwn.Pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}

	register := func(terms []schedulerinterface.AffinityTerm, specTerms []apiv1.PodAffinityTerm, anti bool) []int {
		var ids []int
		for i := range terms {
			t := &terms[i]
			tk := affinityTermKey{antiAffinity: anti, termId: getTermId(t)}
			var ls *metav1.LabelSelector
			if i < len(specTerms) {
				ls = specTerms[i].LabelSelector
			}
			id := e.ensureTermRegistered(t, tk, ls)
			e.termCounts[id]++
			ids = append(ids, id)
		}
		return ids
	}

	ownedAffinity := podInfo.GetRequiredAffinityTerms()
	ownedAntiAffinity := podInfo.GetRequiredAntiAffinityTerms()
	allOwnedIDs := append(register(ownedAffinity, specAff, false), register(ownedAntiAffinity, specAnti, true)...)

	// 2. Identify which terms in the system this pod satisfies.
	matchingTerms := e.getMatchingTerms(pwn.Pod)

	// 3. Update reactive topology indices for all domains the node belongs to.
	for k, v := range pwn.Node.Labels {
		domain := k + "\x00" + v
		e.updateHashDomain(hash, domain, 1)
		e.updatePresence(domain, matchingTerms, 1)
	}

	// 4. Update the Forbidden B-Tree for symmetry checks (anti-affinity).
	e.updateForbidden(pwn.Node.Labels, allOwnedIDs, 1)
}

// ensureTermRegistered adds a term to the global registry and lazily backfills it.
// It uses the podLabelIndex to only scan pod types that could possibly match, avoiding O(N).
func (e *PredicateEvaluator) ensureTermRegistered(term *schedulerinterface.AffinityTerm, tk affinityTermKey, ls *metav1.LabelSelector) int {
	if id, ok := e.termRegistry.FindValue(tk); ok {
		return id
	}

	id := len(e.termsByID)
	e.termRegistry.SetCurrent(tk, id)
	e.termsByID = append(e.termsByID, termEntry{term: term, antiAffinity: tk.antiAffinity})
	e.termCounts = append(e.termCounts, 0)

	// Index the term definition to speed up future matching of incoming pods.
	indexed := false
	if ls != nil && len(ls.MatchExpressions) == 0 && len(ls.MatchLabels) > 0 {
		for k, v := range ls.MatchLabels {
			kv := k + "\x00" + v
			ids, _ := e.termLabelIndex.FindValue(kv)
			newIDs := append([]int(nil), ids...)
			newIDs = append(newIDs, id)
			e.termLabelIndex.SetCurrent(kv, newIDs)
			indexed = true
		}
	}
	if !indexed {
		e.complexTerms = append(e.complexTerms, id)
	}

	// BACKFILL: Run the new rule against all existing pods in the cluster.
	// We narrow down candidates using the podLabelIndex (The Two-Way Index).
	var candidateHashes sets.Set[string]
	if ls != nil && len(ls.MatchExpressions) == 0 && len(ls.MatchLabels) > 0 {
		candidateHashes = sets.New[string]()
		for k, v := range ls.MatchLabels {
			kv := k + "\x00" + v
			if hashes, ok := e.podLabelIndex.FindValue(kv); ok && hashes != nil {
				for h := range hashes {
					candidateHashes.Insert(h)
				}
			}
			break
		}
	} else {
		// Fallback for complex selectors: check all known unique pod types.
		candidateHashes = sets.New[string]()
		e.hashRepresentatives.ForEach(func(hash string, _ *apiv1.Pod) bool {
			candidateHashes.Insert(hash)
			return true
		})
	}

	for hash := range candidateHashes {
		rep, found := e.hashRepresentatives.FindValue(hash)
		if found && rep != nil && term.Matches(rep, nil) {
			domains, _ := e.hashDomains.FindValue(hash)
			for _, domain := range domains {
				hdKey := hash + "\x00" + domain
				if count, ok := e.hashDomainCounts.FindValue(hdKey); ok && count > 0 {
					e.updatePresence(domain, []int{id}, count)
				}
			}
		}
	}

	return id
}

// updatePresence increments/decrements term satisfy counts in the presence B-Tree.
func (e *PredicateEvaluator) updatePresence(domain string, matchingTerms []int, delta int32) {
	if len(matchingTerms) == 0 {
		return
	}
	m, _ := e.presence.Get(domain)
	m = cloneCounts(m, len(e.termsByID))
	for _, id := range matchingTerms {
		m[id] += delta
	}
	e.presence.Set(domain, m)
}

// updateForbidden updates the anti-affinity violation counts in the forbidden B-Tree.
func (e *PredicateEvaluator) updateForbidden(nodeLabels map[string]string, ownedIDs []int, delta int32) {
	for _, id := range ownedIDs {
		term := e.termsByID[id].term
		if topoVal, ok := nodeLabels[term.TopologyKey]; ok {
			key := term.TopologyKey + "\x00" + topoVal
			f, _ := e.forbidden.Get(key)
			f = cloneCounts(f, len(e.termsByID))
			f[id] += delta
			e.forbidden.Set(key, f)
		}
	}
}

// updateHashDomain manages the association between a pod type (hash) and a topology domain.
func (e *PredicateEvaluator) updateHashDomain(hash, domain string, delta int32) int32 {
	hdKey := hash + "\x00" + domain
	count, _ := e.hashDomainCounts.FindValue(hdKey)
	newCount := count + delta

	if count == 0 && newCount > 0 {
		domains, _ := e.hashDomains.FindValue(hash)
		newDomains := make([]string, len(domains), len(domains)+1)
		copy(newDomains, domains)
		newDomains = append(newDomains, domain)
		e.hashDomains.SetCurrent(hash, newDomains)
	} else if count > 0 && newCount <= 0 {
		// pop domain from hashDomains if it hits zero.
		domains, _ := e.hashDomains.FindValue(hash)
		newDomains := make([]string, 0, len(domains))
		for _, d := range domains {
			if d != domain {
				newDomains = append(newDomains, d)
			}
		}
		e.hashDomains.SetCurrent(hash, newDomains)
		e.hashDomainCounts.DeleteCurrent(hdKey)
		return 0
	}

	e.hashDomainCounts.SetCurrent(hdKey, newCount)
	return newCount
}

func (e *PredicateEvaluator) OnUpdate(oldObj, newObj any) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.OnUpdateLocked(oldObj, newObj)
}

func (e *PredicateEvaluator) OnUpdateLocked(oldObj, newObj any) {
	e.OnDeleteLocked(oldObj)
	e.OnAddLocked(newObj, false)
}

func (e *PredicateEvaluator) OnDelete(obj any) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.OnDeleteLocked(obj)
}

func (e *PredicateEvaluator) OnDeleteLocked(obj any) {
	pwn := obj.(framework.PodWithNode)
	if pwn.Pod == nil || pwn.Node == nil {
		return
	}

	podInfo := framework.NewPodInfo(pwn.Pod, nil)
	hash := e.getPodMatchHash(pwn.Pod)
	matchingTerms := e.getMatchingTerms(pwn.Pod)

	ownedAffKeys := e.getKeys(podInfo.GetRequiredAffinityTerms(), false)
	ownedAntiKeys := e.getKeys(podInfo.GetRequiredAntiAffinityTerms(), true)

	var allOwnedIDs []int
	for _, tk := range append(ownedAffKeys, ownedAntiKeys...) {
		if id, ok := e.termRegistry.FindValue(tk); ok {
			allOwnedIDs = append(allOwnedIDs, id)
		}
	}

	// 1. Remove anti-affinity constraints.
	e.updateForbidden(pwn.Node.Labels, allOwnedIDs, -1)

	// 2. Remove term satisfaction presence.
	for k, v := range pwn.Node.Labels {
		domain := k + "\x00" + v
		e.updatePresence(domain, matchingTerms, -1)
		e.updateHashDomain(hash, domain, -1)
	}

	// 3. Clean up unique pod type tracking if the last pod of this type is gone.
	totalHashCount := int32(0)
	domains, _ := e.hashDomains.FindValue(hash)
	for _, d := range domains {
		hdKey := hash + "\x00" + d
		if c, ok := e.hashDomainCounts.FindValue(hdKey); ok {
			totalHashCount += c
		}
	}

	if totalHashCount == 0 {
		for k, v := range pwn.Pod.Labels {
			kv := k + "\x00" + v
			if hashes, ok := e.podLabelIndex.FindValue(kv); ok && hashes != nil {
				hashes = hashes.Clone()
				hashes.Delete(hash)
				if len(hashes) == 0 {
					e.podLabelIndex.DeleteCurrent(kv)
				} else {
					e.podLabelIndex.SetCurrent(kv, hashes)
				}
			}
		}
		e.hashRepresentatives.DeleteCurrent(hash)
		e.hashDomains.DeleteCurrent(hash)
		e.matchCache.DeleteCurrent(hash)
	}

	// 4. Update ref counts for owned terms.
	for _, id := range allOwnedIDs {
		e.termCounts[id]--
	}
}

// Clear resets the evaluator to an empty base state.
func (e *PredicateEvaluator) Clear() {
	e.termRegistry = common.NewPatchSet[affinityTermKey, int]()
	e.termsByID = nil
	e.termCounts = nil
	e.termLabelIndex = common.NewPatchSet[string, []int]()
	e.complexTerms = nil
	e.presence = fort.NewBTreeMap[[]int32]()
	e.forbidden = fort.NewBTreeMap[[]int32]()
	e.hashDomainCounts = common.NewPatchSet[string, int32]()
	e.hashDomains = common.NewPatchSet[string, []string]()
	e.hashRepresentatives = common.NewPatchSet[string, *apiv1.Pod]()
	e.podLabelIndex = common.NewPatchSet[string, sets.Set[string]]()
	e.matchCache = common.NewPatchSet[string, matchCacheEntry]()
}

// Fork creates a lightweight child evaluator sharing the parent's PatchSets.
func (e *PredicateEvaluator) Fork(newPods, newNodes cache.SharedInformer) *PredicateEvaluator {
	e.termRegistry.Fork()
	e.termLabelIndex.Fork()
	e.hashDomainCounts.Fork()
	e.hashDomains.Fork()
	e.hashRepresentatives.Fork()
	e.podLabelIndex.Fork()
	e.matchCache.Fork()

	res := &PredicateEvaluator{
		lock:                e.lock,
		podInformer:         newPods,
		nodeInformer:        newNodes,
		termRegistry:        e.termRegistry,
		termsByID:           e.termsByID[:len(e.termsByID):len(e.termsByID)], // CoW slice
		termCounts:          make([]int, len(e.termCounts)),
		termLabelIndex:      e.termLabelIndex,
		complexTerms:        make([]int, len(e.complexTerms)),
		presence:            e.presence.Clone(),
		forbidden:           e.forbidden.Clone(),
		hashDomainCounts:    e.hashDomainCounts,
		hashDomains:         e.hashDomains,
		hashRepresentatives: e.hashRepresentatives,
		podLabelIndex:       e.podLabelIndex,
		matchCache:          e.matchCache,
	}
	copy(res.termCounts, e.termCounts)
	copy(res.complexTerms, e.complexTerms)

	memo := make(map[cache.SharedInformer]cache.SharedInformer)
	memo[e.podInformer] = newPods
	memo[e.nodeInformer] = newNodes

	res.podsWithNodes = fort.ClonePipeline(e.podsWithNodes, memo).(fort.CloneableSharedInformerQuery)
	if ms, ok := res.podsWithNodes.(fort.ManualSharedInformer); ok {
		res.registration, _ = ms.AddEventHandlerNoReplay(res)
	} else {
		res.registration, _ = res.podsWithNodes.AddEventHandler(res)
	}

	return res
}

func (e *PredicateEvaluator) Commit() {
	e.termRegistry.Commit()
	e.termLabelIndex.Commit()
	e.hashDomainCounts.Commit()
	e.hashDomains.Commit()
	e.hashRepresentatives.Commit()
	e.podLabelIndex.Commit()
	e.matchCache.Commit()
}

func (e *PredicateEvaluator) Revert() {
	e.termRegistry.Revert()
	e.termLabelIndex.Revert()
	e.hashDomainCounts.Revert()
	e.hashDomains.Revert()
	e.hashRepresentatives.Revert()
	e.podLabelIndex.Revert()
	e.matchCache.Revert()
}

// FastCheckAffinity executes the affinity checks for a pod on a node (Filter).
// It relies entirely on pre-calculated context and integer slice lookups.
func (e *PredicateEvaluator) FastCheckAffinity(podCtx *PodAffinityContext, node *apiv1.Node) clustersnapshot.SchedulingError {
	if podCtx == nil || node == nil {
		return nil
	}

	e.lock.RLock()
	defer e.lock.RUnlock()

	// 1. Incoming Anti-Affinity: check if any existing pods violate OUR anti-affinity rules.
	for _, twt := range podCtx.ownedAntiAffinity {
		if topoVal, ok := node.Labels[twt.topologyKey]; ok {
			key := twt.topologyKey + "\x00" + topoVal
			if m, ok := e.presence.Get(key); ok {
				if twt.id < len(m) && m[twt.id] > 0 {
					return clustersnapshot.NewFailingPredicateError(podCtx.pod, "InterPodAffinity", []string{interpodaffinity.ErrReasonAntiAffinityRulesNotMatch}, "", "")
				}
			}
		}
	}

	// 2. Incoming Affinity: check if any existing pods satisfy OUR affinity requirements.
	for _, twt := range podCtx.ownedAffinity {
		topoVal, ok := node.Labels[twt.topologyKey]
		if !ok {
			return clustersnapshot.NewFailingPredicateError(podCtx.pod, "InterPodAffinity", []string{interpodaffinity.ErrReasonAffinityRulesNotMatch}, "", "node lacks topology key")
		}

		key := twt.topologyKey + "\x00" + topoVal
		m, hasPresence := e.presence.Get(key)
		if !hasPresence || twt.id >= len(m) || m[twt.id] <= 0 {
			// Affinity Fallback: a pod can satisfy its own affinity rule if it matches its own selector.
			selfMatches := false
			for _, mid := range podCtx.matchIDs {
				if mid == twt.id {
					selfMatches = true
					break
				}
			}
			if !selfMatches {
				return clustersnapshot.NewFailingPredicateError(podCtx.pod, "InterPodAffinity", []string{interpodaffinity.ErrReasonAffinityRulesNotMatch}, "", "")
			}
		}
	}

	// 3. Existing Anti-Affinity (Symmetry): check if WE violate any existing pod's anti-affinity rules.
	for k, v := range node.Labels {
		key := k + "\x00" + v
		if f, ok := e.forbidden.Get(key); ok {
			for _, mid := range podCtx.matchIDs {
				if mid < len(f) && f[mid] > 0 && e.termsByID[mid].antiAffinity {
					return clustersnapshot.NewFailingPredicateError(podCtx.pod, "InterPodAffinity", []string{interpodaffinity.ErrReasonAntiAffinityRulesNotMatch}, "", "existing pod anti-affinity")
				}
			}
		}
	}

	return nil
}

// getMatchingTerms identifies which terms in the system a pod matches.
// It is fully incremental: it caches previous results and only evaluates newly registered terms.
func (e *PredicateEvaluator) getMatchingTerms(pod *apiv1.Pod) []int {
	hash := e.getPodMatchHash(pod)
	var res []int
	startIndex := 0

	// Check CoW cache first.
	if hash != "" {
		if val, ok := e.matchCache.FindValue(hash); ok {
			if val.termsCount == len(e.termsByID) {
				return val.matchIDs
			}
			res = append([]int(nil), val.matchIDs...)
			startIndex = val.termsCount
		}
	}

	// Incrementally evaluate newly added terms using the Two-Way Label Index.
	if startIndex < len(e.termsByID) {
		candidates := make(map[int]struct{})
		for k, v := range pod.Labels {
			kv := k + "\x00" + v
			if ids, ok := e.termLabelIndex.FindValue(kv); ok && ids != nil {
				for _, id := range ids {
					if id >= startIndex {
						candidates[id] = struct{}{}
					}
				}
			}
		}
		for _, id := range e.complexTerms {
			if id >= startIndex {
				candidates[id] = struct{}{}
			}
		}

		for id := startIndex; id < len(e.termsByID); id++ {
			if _, ok := candidates[id]; ok {
				if e.termsByID[id].term.Matches(pod, nil) {
					res = append(res, id)
				}
			}
		}

		// Update the CoW cache for subsequent forks.
		if hash != "" {
			e.matchCache.SetCurrent(hash, matchCacheEntry{
				matchIDs:   res,
				termsCount: len(e.termsByID),
			})
		}
	}

	return res
}

func (e *PredicateEvaluator) getKeys(terms []schedulerinterface.AffinityTerm, anti bool) []affinityTermKey {
	var keys []affinityTermKey
	for i := range terms {
		keys = append(keys, affinityTermKey{antiAffinity: anti, termId: getTermId(&terms[i])})
	}
	return keys
}

func getTermId(term *schedulerinterface.AffinityTerm) string {
	return fmt.Sprintf("%v/%v/%v/%v", term.Selector, sets.List(term.Namespaces), term.NamespaceSelector, term.TopologyKey)
}

// cloneCounts performs a fast, zero-allocation-aware copy of a count slice.
func cloneCounts(counts []int32, minLen int) []int32 {
	l := len(counts)
	if minLen > l {
		l = minLen
	}
	if l == 0 {
		return make([]int32, minLen)
	}
	res := make([]int32, l)
	copy(res, counts)
	return res
}
