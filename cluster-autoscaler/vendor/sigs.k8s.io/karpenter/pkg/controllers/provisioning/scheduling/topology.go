/*
Copyright The Kubernetes Authors.

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

package scheduling

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/awslabs/operatorpkg/option"
	"github.com/awslabs/operatorpkg/serrors"
	"github.com/samber/lo"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/utils/pod"
	"sigs.k8s.io/karpenter/pkg/utils/pretty"
)

type Topology struct {
	kubeClient       client.Client
	preferencePolicy PreferencePolicy
	// Both the topologyGroups and inverseTopologies are maps of the hash from TopologyGroup.Hash() to the topology group
	// itself. This is used to allow us to store one topology group that tracks the topology of many pods instead of
	// having a 1<->1 mapping between topology groups and pods owned/selected by that group.
	topologyGroups map[uint64]*TopologyGroup
	// Anti-affinity works both ways (if a zone has a pod foo with anti-affinity to a pod bar, we can't schedule bar to
	// that zone, even though bar has no anti affinity terms on it. For this to work, we need to separately track the
	// topologies of pods with anti-affinity terms, so we can prevent scheduling the pods they have anti-affinity to
	// in some cases.
	inverseTopologyGroups map[uint64]*TopologyGroup
	// The universe of domains by topology key
	domainGroups map[string]TopologyDomainGroup
	// excludedPods are the pod UIDs of pods that are excluded from counting.  This is used so we can simulate
	// moving pods to prevent them from being double counted.
	excludedPods sets.Set[string]
	cluster      *state.Cluster
	stateNodes   []*state.StateNode
}

func NewTopology(
	ctx context.Context,
	kubeClient client.Client,
	cluster *state.Cluster,
	stateNodes []*state.StateNode,
	nodePools []*v1.NodePool,
	instanceTypes map[string][]*cloudprovider.InstanceType,
	pods []*corev1.Pod,
	opts ...Options,
) (*Topology, error) {
	t := &Topology{
		kubeClient:            kubeClient,
		preferencePolicy:      option.Resolve(opts...).preferencePolicy,
		cluster:               cluster,
		stateNodes:            stateNodes,
		domainGroups:          buildDomainGroups(nodePools, instanceTypes),
		topologyGroups:        map[uint64]*TopologyGroup{},
		inverseTopologyGroups: map[uint64]*TopologyGroup{},
		excludedPods:          sets.New[string](),
	}

	// these are the pods that we intend to schedule, so if they are currently in the cluster we shouldn't count them for
	// topology purposes
	for _, p := range pods {
		t.excludedPods.Insert(string(p.UID))
	}

	errs := t.updateInverseAffinities(ctx)
	for i := range pods {
		errs = multierr.Append(errs, t.Update(ctx, pods[i]))
	}
	if errs != nil {
		return nil, errs
	}
	return t, nil
}

func buildDomainGroups(nodePools []*v1.NodePool, instanceTypes map[string][]*cloudprovider.InstanceType) map[string]TopologyDomainGroup {
	nodePoolIndex := lo.SliceToMap(nodePools, func(np *v1.NodePool) (string, *v1.NodePool) {
		return np.Name, np
	})
	domainGroups := map[string]TopologyDomainGroup{}
	for npName, its := range instanceTypes {
		np := nodePoolIndex[npName]
		for _, it := range its {
			// We need to intersect the instance type requirements with the current nodePool requirements.  This
			// ensures that something like zones from an instance type don't expand the universe of valid domains.
			requirements := scheduling.NewNodeSelectorRequirementsWithMinValues(np.Spec.Template.Spec.Requirements...)
			requirements.Add(scheduling.NewLabelRequirements(np.Spec.Template.Labels).Values()...)
			requirements.Add(it.Requirements.Values()...)

			for topologyKey, requirement := range requirements {
				if _, ok := domainGroups[topologyKey]; !ok {
					domainGroups[topologyKey] = NewTopologyDomainGroup()
				}
				for _, domain := range requirement.Values() {
					domainGroups[topologyKey].Insert(domain, np.Spec.Template.Spec.Taints...)
				}
			}
		}

		requirements := scheduling.NewNodeSelectorRequirementsWithMinValues(np.Spec.Template.Spec.Requirements...)
		requirements.Add(scheduling.NewLabelRequirements(np.Spec.Template.Labels).Values()...)
		for key, requirement := range requirements {
			if requirement.Operator() == corev1.NodeSelectorOpIn {
				if _, ok := domainGroups[key]; !ok {
					domainGroups[key] = NewTopologyDomainGroup()
				}
				for _, value := range requirement.Values() {
					domainGroups[key].Insert(value, np.Spec.Template.Spec.Taints...)
				}
			}
		}
	}
	return domainGroups
}

// topologyError allows lazily generating the error string in the topology error.  If a pod fails to schedule, most often
// we are only interested in the fact that it failed to schedule and not why.
type topologyError struct {
	topology    *TopologyGroup
	podDomains  *scheduling.Requirement
	nodeDomains *scheduling.Requirement
}

func (t topologyError) Error() string {
	return fmt.Sprintf("unsatisfiable topology constraint for %s, key=%s (counts = %s, podDomains = %v, nodeDomains = %v", t.topology.Type, t.topology.Key,
		pretty.Map(t.topology.domains, 25), t.podDomains, t.nodeDomains)
}

// Update unregisters the pod as the owner of all affinities and then creates any new topologies based on the pod spec
// registered the pod as the owner of all associated affinities, new or old.  This allows Update() to be called after
// relaxation of a preference to properly break the topology <-> owner relationship so that the preferred topology will
// no longer influence scheduling.
func (t *Topology) Update(ctx context.Context, p *corev1.Pod) error {
	for _, topology := range t.topologyGroups {
		topology.RemoveOwner(p.UID)
	}

	if (t.preferencePolicy == PreferencePolicyIgnore && pod.HasRequiredPodAntiAffinity(p)) ||
		(t.preferencePolicy == PreferencePolicyRespect && pod.HasPodAntiAffinity(p)) {
		if err := t.updateInverseAntiAffinity(ctx, p, nil); err != nil {
			return fmt.Errorf("updating inverse anti-affinities, %w", err)
		}
	}

	topologies := t.newForTopologies(p)
	affinities, err := t.newForAffinities(ctx, p)
	if err != nil {
		return fmt.Errorf("updating affinities, %w", err)
	}

	for _, tg := range append(topologies, affinities...) {
		hash := tg.Hash()
		// Avoid recomputing topology counts if we've already seen this group
		if existing, ok := t.topologyGroups[hash]; !ok {
			if err := t.countDomains(ctx, tg); err != nil {
				return err
			}
			t.topologyGroups[hash] = tg
		} else {
			tg = existing
		}
		tg.AddOwner(p.UID)
	}
	return nil
}

// Record records the topology changes given that pod p schedule on a node with the given requirements
func (t *Topology) Record(p *corev1.Pod, taints []corev1.Taint, requirements scheduling.Requirements, compatibilityOptions ...option.Function[scheduling.CompatibilityOptions]) {
	// once we've committed to a domain, we record the usage in every topology that cares about it
	for _, tg := range t.topologyGroups {
		if tg.Counts(p, taints, requirements, compatibilityOptions...) {
			domains := requirements.Get(tg.Key)
			if tg.Type == TopologyTypePodAntiAffinity {
				// for anti-affinity topologies we need to block out all possible domains that the pod could land in
				tg.Record(domains.Values()...)
			} else {
				// but for affinity & topology spread, we can only record the domain if we know the specific domain we land in
				if domains.Len() == 1 {
					tg.Record(domains.Values()[0])
				}
			}
		}
	}
	// for anti-affinities, we record where the pods could be, even if
	// requirements haven't collapsed to a single value.
	for _, tg := range t.inverseTopologyGroups {
		if tg.IsOwnedBy(p.UID) {
			tg.Record(requirements.Get(tg.Key).Values()...)
		}
	}
}

// AddRequirements tightens the input requirements by adding additional requirements that are being enforced by topology spreads
// affinities, anti-affinities or inverse anti-affinities.  The nodeHostname is the hostname that we are currently considering
// placing the pod on.  It returns these newly tightened requirements, or an error in the case of a set of requirements that
// cannot be satisfied.
func (t *Topology) AddRequirements(p *corev1.Pod, taints []corev1.Taint, podRequirements, nodeRequirements scheduling.Requirements, compatibilityOptions ...option.Function[scheduling.CompatibilityOptions]) (scheduling.Requirements, error) {
	requirements := scheduling.NewRequirements(nodeRequirements.Values()...)
	for _, topology := range t.getMatchingTopologies(p, taints, nodeRequirements, compatibilityOptions...) {
		podDomains := scheduling.NewRequirement(topology.Key, corev1.NodeSelectorOpExists)
		if podRequirements.Has(topology.Key) {
			podDomains = podRequirements.Get(topology.Key)
		}
		nodeDomains := scheduling.NewRequirement(topology.Key, corev1.NodeSelectorOpExists)
		if nodeRequirements.Has(topology.Key) {
			nodeDomains = nodeRequirements.Get(topology.Key)
		}
		domains, _ := topology.Get(p, podDomains, nodeDomains)
		if domains.Len() == 0 {
			return nil, topologyError{
				topology:    topology,
				podDomains:  podDomains,
				nodeDomains: nodeDomains,
			}
		}
		requirements.Add(domains)
	}
	return requirements, nil
}

// GetTopologyZoneConstraints returns the set of valid zones from all topology constraints
// that use the zone topology key for the given pod, along with whether the constraints are satisfiable.
func (t *Topology) GetTopologyZoneConstraints(p *corev1.Pod, podRequirements scheduling.Requirements) (sets.Set[string], bool) {
	var result sets.Set[string]

	for _, topology := range t.topologyGroups {
		if !topology.IsOwnedBy(p.UID) || topology.Key != corev1.LabelTopologyZone {
			continue
		}

		podDomains := scheduling.NewRequirement(topology.Key, corev1.NodeSelectorOpExists)
		if podRequirements.Has(topology.Key) {
			podDomains = podRequirements.Get(topology.Key)
		}
		nodeDomains := scheduling.NewRequirement(topology.Key, corev1.NodeSelectorOpExists)
		_, validDomains := topology.Get(p, podDomains, nodeDomains)

		if validDomains.Len() == 0 {
			return nil, false
		}
		if result == nil {
			result = validDomains
		} else {
			for zone := range result {
				if !validDomains.Has(zone) {
					delete(result, zone)
				}
			}
		}
	}
	return result, true
}

// Register is used to register a domain as available across topologies for the given topology key.
func (t *Topology) Register(topologyKey string, domain string) {
	for _, tg := range t.topologyGroups {
		if tg.Key == topologyKey {
			tg.Register(domain)
		}
	}
	for _, tg := range t.inverseTopologyGroups {
		if tg.Key == topologyKey {
			tg.Register(domain)
		}
	}
}

// Unregister is used to unregister a domain as available across topologies for the given topology key.
func (t *Topology) Unregister(topologyKey string, domain string) {
	for _, topology := range t.topologyGroups {
		if topology.Key == topologyKey {
			topology.Unregister(domain)
		}
	}
	for _, topology := range t.inverseTopologyGroups {
		if topology.Key == topologyKey {
			topology.Unregister(domain)
		}
	}
}

// updateInverseAffinities is used to identify pods with anti-affinity terms so we can track those topologies.  We
// have to look at every pod in the cluster as there is no way to query for a pod with anti-affinity terms.
func (t *Topology) updateInverseAffinities(ctx context.Context) error {
	var errs error
	t.cluster.ForPodsWithAntiAffinity(func(pod *corev1.Pod, node *corev1.Node) bool {
		// don't count the pod we are excluding
		if t.excludedPods.Has(string(pod.UID)) {
			return true
		}
		if err := t.updateInverseAntiAffinity(ctx, pod, node.Labels); err != nil {
			errs = multierr.Append(errs, fmt.Errorf("tracking existing pod anti-affinity, %w", err))
		}
		return true
	})
	return errs
}

// updateInverseAntiAffinity is used to track topologies of inverse anti-affinities. Here the domains & counts track the
// pods with the anti-affinity.
func (t *Topology) updateInverseAntiAffinity(ctx context.Context, pod *corev1.Pod, domains map[string]string) error {
	// We intentionally don't track inverse anti-affinity preferences. We're not
	// required to enforce them so it just adds complexity for very little
	// value.  The problem with them comes from the relaxation process, the pod
	// we are relaxing is not the pod with the anti-affinity term.
	for _, term := range pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
		namespaces, err := t.buildNamespaceList(ctx, pod.Namespace, term.Namespaces, term.NamespaceSelector)
		if err != nil {
			return err
		}

		tg := NewTopologyGroup(TopologyTypePodAntiAffinity, term.TopologyKey, pod, namespaces, term.LabelSelector, math.MaxInt32, nil, nil, nil, t.domainGroups[term.TopologyKey])

		hash := tg.Hash()
		if existing, ok := t.inverseTopologyGroups[hash]; !ok {
			t.inverseTopologyGroups[hash] = tg
		} else {
			tg = existing
		}
		if domain, ok := domains[tg.Key]; ok {
			tg.Record(domain)
		}
		tg.AddOwner(pod.UID)
	}
	return nil
}

// countDomains initializes the topology group by registereding any well known domains and performing pod counts
// against the cluster for any existing pods.
//
//nolint:gocyclo
func (t *Topology) countDomains(ctx context.Context, tg *TopologyGroup) error {
	podList := &corev1.PodList{}

	// collect the pods from all the specified namespaces (don't see a way to query multiple namespaces
	// simultaneously)
	var pods []corev1.Pod
	for _, ns := range tg.namespaces.UnsortedList() {
		if err := t.kubeClient.List(ctx, podList, TopologyListOptions(ns, tg.rawSelector)); err != nil {
			return fmt.Errorf("listing pods, %w", err)
		}
		pods = append(pods, podList.Items...)
	}

	// capture new domain values from existing nodes that may not have any pods selected by the topology group
	// scheduled to them already
	// Note: long term we should handle this when constructing the domain groups, but that would require domain groups
	// to handle affinity in addition to taints / tolerations.
	for _, n := range t.stateNodes {
		// ignore state nodes which are tracking in-flight NodeClaims
		if n.Node == nil {
			continue
		}
		// ignore the node if it doesn't match the topology group
		if !tg.nodeFilter.Matches(n.Node.Spec.Taints, scheduling.NewLabelRequirements(n.Node.Labels)) {
			continue
		}
		domain, exists := n.Labels()[tg.Key]
		if !exists {
			continue
		}
		if _, ok := tg.domains[domain]; !ok {
			tg.domains[domain] = 0
			tg.emptyDomains.Insert(domain)
		}
	}

	// sort our pods by the node they are scheduled to
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Spec.NodeName < pods[j].Spec.NodeName
	})
	var previousNode *corev1.Node
	var previousNodeRequirements scheduling.Requirements

	for i, p := range pods {
		if IgnoredForTopology(&pods[i]) {
			continue
		}
		// pod is excluded for counting purposes
		if t.excludedPods.Has(string(p.UID)) {
			continue
		}
		var node *corev1.Node
		var nodeRequirements scheduling.Requirements
		if previousNode != nil && previousNode.Name == p.Spec.NodeName {
			// no need to look up the node since we already have it
			node = previousNode
			nodeRequirements = previousNodeRequirements
		} else {
			node = &corev1.Node{}
			if err := t.kubeClient.Get(ctx, types.NamespacedName{Name: p.Spec.NodeName}, node); err != nil {
				// Pods that cannot be evicted can be leaked in the API Server after
				// a Node is removed. Since pod bindings are immutable, these pods
				// cannot be recovered, and will be deleted by the pod lifecycle
				// garbage collector. These pods are not running, and should not
				// impact future topology calculations.
				if errors.IsNotFound(err) {
					continue
				}
				return serrors.Wrap(fmt.Errorf("getting node, %w", err), "Node", klog.KRef("", p.Spec.NodeName))
			}
			nodeRequirements = scheduling.NewLabelRequirements(node.Labels)

			// assign back to previous node so we can hopefully re-use these in the next iteration
			previousNode = node
			previousNodeRequirements = nodeRequirements
		}

		domain, ok := node.Labels[tg.Key]
		// Kubelet sets the hostname label, but the node may not be ready yet so there is no label.  We fall back and just
		// treat the node name as the label.  It probably is in most cases, but even if not we at least count the existence
		// of the pods in some domain, even if not in the correct one.  This is needed to handle the case of pods with
		// self-affinity only fulfilling that affinity if all domains are empty.
		if !ok && tg.Key == corev1.LabelHostname {
			domain = node.Name
			ok = true
		}
		if !ok {
			continue // Don't include pods if node doesn't contain domain https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/#conventions
		}

		// nodes may or may not be considered for counting purposes for topology spread constraints depending on if they
		// are selected by the pod's node selectors and required node affinities.  If these are unset, the node always counts.
		if !tg.nodeFilter.Matches(node.Spec.Taints, nodeRequirements) {
			continue
		}
		tg.Record(domain)
	}
	return nil
}

func (t *Topology) newForTopologies(p *corev1.Pod) []*TopologyGroup {
	var topologyGroups []*TopologyGroup
	for _, tsc := range p.Spec.TopologySpreadConstraints {
		if t.preferencePolicy == PreferencePolicyIgnore && tsc.WhenUnsatisfiable != corev1.DoNotSchedule {
			continue
		}
		for _, key := range tsc.MatchLabelKeys {
			if value, ok := p.Labels[key]; ok {
				tsc.LabelSelector.MatchExpressions = append(tsc.LabelSelector.MatchExpressions, metav1.LabelSelectorRequirement{
					Key:      key,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{value},
				})
			}
		}
		topologyGroups = append(topologyGroups, NewTopologyGroup(
			TopologyTypeSpread,
			tsc.TopologyKey,
			p,
			sets.New(p.Namespace),
			tsc.LabelSelector,
			tsc.MaxSkew,
			tsc.MinDomains,
			tsc.NodeTaintsPolicy,
			tsc.NodeAffinityPolicy,
			t.domainGroups[tsc.TopologyKey],
		))
	}
	return topologyGroups
}

// newForAffinities returns a list of topology groups that have been constructed based on the input pod and required/preferred affinity terms
func (t *Topology) newForAffinities(ctx context.Context, p *corev1.Pod) ([]*TopologyGroup, error) {
	var topologyGroups []*TopologyGroup
	// No affinity defined
	if p.Spec.Affinity == nil {
		return topologyGroups, nil
	}
	affinityTerms := map[TopologyType][]corev1.PodAffinityTerm{}

	// include both soft and hard affinity terms
	if p.Spec.Affinity.PodAffinity != nil {
		affinityTerms[TopologyTypePodAffinity] = append(affinityTerms[TopologyTypePodAffinity], p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
		if t.preferencePolicy == PreferencePolicyRespect {
			for _, term := range p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				affinityTerms[TopologyTypePodAffinity] = append(affinityTerms[TopologyTypePodAffinity], term.PodAffinityTerm)
			}
		}
	}

	// include both soft and hard antiaffinity terms
	if p.Spec.Affinity.PodAntiAffinity != nil {
		affinityTerms[TopologyTypePodAntiAffinity] = append(affinityTerms[TopologyTypePodAntiAffinity], p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
		if t.preferencePolicy == PreferencePolicyRespect {
			for _, term := range p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				affinityTerms[TopologyTypePodAntiAffinity] = append(affinityTerms[TopologyTypePodAntiAffinity], term.PodAffinityTerm)
			}
		}
	}

	// build topologies
	for topologyType, terms := range affinityTerms {
		for _, term := range terms {
			namespaces, err := t.buildNamespaceList(ctx, p.Namespace, term.Namespaces, term.NamespaceSelector)
			if err != nil {
				return nil, err
			}
			topologyGroups = append(topologyGroups, NewTopologyGroup(topologyType, term.TopologyKey, p, namespaces, term.LabelSelector, math.MaxInt32, nil, nil, nil, t.domainGroups[term.TopologyKey]))
		}
	}
	return topologyGroups, nil
}

// buildNamespaceList constructs a unique list of namespaces consisting of the pod's namespace and the optional list of
// namespaces and those selected by the namespace selector
func (t *Topology) buildNamespaceList(ctx context.Context, namespace string, namespaces []string, selector *metav1.LabelSelector) (sets.Set[string], error) {
	if len(namespaces) == 0 && selector == nil {
		return sets.New(namespace), nil
	}
	if selector == nil {
		return sets.New(namespaces...), nil
	}
	var namespaceList corev1.NamespaceList
	labelSelector, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("parsing selector, %w", err)
	}
	if err := t.kubeClient.List(ctx, &namespaceList, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
		return nil, fmt.Errorf("listing namespaces, %w", err)
	}
	selected := sets.New[string]()
	for _, namespace := range namespaceList.Items {
		selected.Insert(namespace.Name)
	}
	selected.Insert(namespaces...)
	return selected, nil
}

// getMatchingTopologies returns a sorted list of topologies that either control the scheduling of pod p, or for which
// the topology selects pod p and the scheduling of p affects the count per topology domain
func (t *Topology) getMatchingTopologies(p *corev1.Pod, taints []corev1.Taint, requirements scheduling.Requirements, compatibilityOptions ...option.Function[scheduling.CompatibilityOptions]) []*TopologyGroup {
	var matchingTopologies []*TopologyGroup
	for _, tg := range t.topologyGroups {
		if tg.IsOwnedBy(p.UID) {
			matchingTopologies = append(matchingTopologies, tg)
		}
	}
	for _, tg := range t.inverseTopologyGroups {
		if tg.Counts(p, taints, requirements, compatibilityOptions...) {
			matchingTopologies = append(matchingTopologies, tg)
		}
	}
	return matchingTopologies
}

func TopologyListOptions(namespace string, labelSelector *metav1.LabelSelector) *client.ListOptions {
	selector := labels.Everything()
	if labelSelector == nil {
		return &client.ListOptions{Namespace: namespace, LabelSelector: selector}
	}
	for key, value := range labelSelector.MatchLabels {
		requirement, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return &client.ListOptions{Namespace: namespace, LabelSelector: labels.Nothing()}
		}
		selector = selector.Add(*requirement)
	}
	for _, expression := range labelSelector.MatchExpressions {
		requirement, err := labels.NewRequirement(expression.Key, mapOperator(expression.Operator), expression.Values)
		if err != nil {
			return &client.ListOptions{Namespace: namespace, LabelSelector: labels.Nothing()}
		}
		selector = selector.Add(*requirement)
	}
	return &client.ListOptions{Namespace: namespace, LabelSelector: selector}
}

func mapOperator(operator metav1.LabelSelectorOperator) selection.Operator {
	switch operator {
	case metav1.LabelSelectorOpIn:
		return selection.In
	case metav1.LabelSelectorOpNotIn:
		return selection.NotIn
	case metav1.LabelSelectorOpExists:
		return selection.Exists
	case metav1.LabelSelectorOpDoesNotExist:
		return selection.DoesNotExist
	}
	// this shouldn't occur as we cover all valid cases of LabelSelectorOperator that the API allows.  If it still
	// does occur somehow we'll panic just later when the requirement throws an error.,
	return ""
}

func IgnoredForTopology(p *corev1.Pod) bool {
	return !pod.IsScheduled(p) || pod.IsTerminal(p) || pod.IsTerminating(p)
}
