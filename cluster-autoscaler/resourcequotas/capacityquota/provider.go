/*
Copyright 2025 The Kubernetes Authors.

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

package capacityquota

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider provides quotas from CapacityQuota custom resource.
type Provider struct {
	kubeClient client.Client
}

// NewCapacityQuotasProvider returns a new Provider.
func NewCapacityQuotasProvider(kubeClient client.Client) *Provider {
	return &Provider{kubeClient: kubeClient}
}

// Quotas returns quotas built from CapacityQuota resources in the cluster.
func (p *Provider) Quotas() ([]resourcequotas.Quota, error) {
	capacityQuotas := &cqv1alpha1.CapacityQuotaList{}
	err := p.kubeClient.List(context.TODO(), capacityQuotas)
	if err != nil {
		return nil, err
	}
	var quotas []resourcequotas.Quota
	for _, cq := range capacityQuotas.Items {
		quota, err := newFromCapacityQuota(cq)
		if err != nil {
			klog.Errorf("Skipping CapacityQuota %q, err: %v", cq.Name, err)
			continue
		}
		quotas = append(quotas, quota)
	}
	return quotas, nil
}

type labelSelectorQuota struct {
	id       string
	selector labels.Selector
	limits   map[string]int64
}

func (lsq *labelSelectorQuota) ID() string {
	return lsq.id
}

func (lsq *labelSelectorQuota) AppliesTo(node *apiv1.Node) bool {
	return lsq.selector.Matches(labels.Set(node.Labels))
}

func (lsq *labelSelectorQuota) Limits() map[string]int64 {
	return lsq.limits
}

func newFromCapacityQuota(cq cqv1alpha1.CapacityQuota) (*labelSelectorQuota, error) {
	selector, err := labelSelectorAsSelector(cq.Spec.Selector)
	if err != nil {
		return nil, err
	}
	limits := make(map[string]int64, len(cq.Spec.Limits.Resources))
	for resource, limit := range cq.Spec.Limits.Resources {
		limits[string(resource)] = limit.Value()
	}
	return &labelSelectorQuota{
		id:       fmt.Sprintf("CapacityQuota/%s", cq.Name),
		selector: selector,
		limits:   limits,
	}, nil
}

func labelSelectorAsSelector(ls *v1.LabelSelector) (labels.Selector, error) {
	if ls == nil {
		return labels.Everything(), nil
	}
	selector, err := v1.LabelSelectorAsSelector(ls)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}
	return selector, nil
}
