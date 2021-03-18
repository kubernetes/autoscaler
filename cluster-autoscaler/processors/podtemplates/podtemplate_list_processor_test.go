/*
Copyright 2018 The Kubernetes Authors.

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

package podtemplates

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/fake"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
)

func TestNewActivePodTemplateListProcessor(t *testing.T) {
	client := fake.NewSimpleClientset()
	processor := NewActivePodTemplateListProcessor(client)
	defer processor.CleanUp()
}

func Test_activePodTemplateListProcessor_ExtraDaemonsets(t *testing.T) {
	podTplMatch := &apiv1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels:    map[string]string{PodTemplateDaemonSetLabelKey: PodTemplateDaemonSetLabelValueTrue},
		},
	}

	daemonSetWant := newDaemonSet(podTplMatch)

	type args struct {
		context *ca_context.AutoscalingContext
	}
	tests := []struct {
		name           string
		initListerFunc func() (v1lister.PodTemplateLister, error)
		args           args

		want    []*appsv1.DaemonSet
		wantErr bool
	}{
		{
			name: "no PodTemplate, should return 0 Daemonset",
			args: args{
				context: &ca_context.AutoscalingContext{},
			},
			initListerFunc: func() (v1lister.PodTemplateLister, error) {
				return newTestDaemonSetLister(nil)
			},
			want:    []*appsv1.DaemonSet{},
			wantErr: false,
		},
		{
			name: "1 matching PodTemplate, should return 1 Daemonset",
			args: args{
				context: &ca_context.AutoscalingContext{},
			},
			initListerFunc: func() (v1lister.PodTemplateLister, error) {
				return newTestDaemonSetLister([]*apiv1.PodTemplate{podTplMatch})
			},
			want:    []*appsv1.DaemonSet{daemonSetWant},
			wantErr: false,
		},
		{
			name: "Lister returns an error",
			args: args{
				context: &ca_context.AutoscalingContext{},
			},
			initListerFunc: func() (v1lister.PodTemplateLister, error) {
				podLister := &podTemplateListerMock{}
				podLister.On("List").Return([]*apiv1.PodTemplate{}, fmt.Errorf("unable to list"))
				return podLister, nil
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelFunc := context.WithCancel(context.Background())
			lister, err := tt.initListerFunc()
			assert.Nil(t, err, "lister creation should not return an error")
			p := activePodTemplateListProcessor{
				podTemplateLister: lister,
				ctx:               ctx,
				cancelFunc:        cancelFunc,
			}
			defer p.CleanUp()
			got, err := p.ExtraDaemonsets(tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("activePodTemplateListProcessor.ExtraDaemonsets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "they should be equal")
			t.Logf("got: %v", got)
		})
	}
}

func Test_noOpPodTemplateListProcessor_ExtraDaemonsets(t *testing.T) {
	type args struct {
		context *ca_context.AutoscalingContext
	}
	tests := []struct {
		name    string
		args    args
		want    []*appsv1.DaemonSet
		wantErr bool
	}{
		{
			name: "always return empty list with no error",
			args: args{
				context: &ca_context.AutoscalingContext{},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewDefaultPodTemplateListProcessor()
			got, err := p.ExtraDaemonsets(tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("noOpPodTemplateListProcessor.ExtraDaemonsets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("noOpPodTemplateListProcessor.ExtraDaemonsets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newTestDaemonSetLister(pts []*apiv1.PodTemplate) (v1lister.PodTemplateLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, pt := range pts {
		err := store.Add(pt)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1lister.NewPodTemplateLister(store), nil
}

type podTemplateListerMock struct {
	mock.Mock
}

// List lists all PodTemplates in the indexer.
func (p *podTemplateListerMock) List(selector labels.Selector) (ret []*apiv1.PodTemplate, err error) {
	args := p.Called()
	return args.Get(0).([]*apiv1.PodTemplate), args.Error(1)
}

// PodTemplates returns an object that can list and get PodTemplates.
func (p *podTemplateListerMock) PodTemplates(namespace string) v1lister.PodTemplateNamespaceLister {
	args := p.Called(namespace)
	return args.Get(0).(v1lister.PodTemplateNamespaceLister)
}
