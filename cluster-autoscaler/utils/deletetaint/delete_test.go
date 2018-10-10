/*
Copyright 2016 The Kubernetes Authors.

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

package deletetaint

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkNodes(t *testing.T) {
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClient(t, node)
	err := MarkToBeDeleted(node, fakeClient)
	assert.NoError(t, err)

	updatedNode, err := fakeClient.Core().Nodes().Get("node", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.True(t, HasToBeDeletedTaint(updatedNode))
}

func TestCheckNodes(t *testing.T) {
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClient(t, node)
	err := MarkToBeDeleted(node, fakeClient)
	assert.NoError(t, err)

	updatedNode, err := fakeClient.Core().Nodes().Get("node", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.True(t, HasToBeDeletedTaint(updatedNode))

	val, err := GetToBeDeletedTime(updatedNode)
	assert.NoError(t, err)
	assert.True(t, time.Now().Sub(*val) < 10*time.Second)
}

func TestCleanNodes(t *testing.T) {
	node := BuildTestNode("node", 1000, 1000)
	addToBeDeletedTaint(node)
	fakeClient := buildFakeClient(t, node)

	cleaned, err := CleanToBeDeleted(node, fakeClient)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	updatedNode, err := fakeClient.Core().Nodes().Get("node", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(updatedNode))
}

func buildFakeClient(t *testing.T, node *apiv1.Node) *fake.Clientset {
	fakeClient := fake.NewSimpleClientset()

	_, err := fakeClient.CoreV1().Nodes().Create(node)
	assert.NoError(t, err)

	// return a 'Conflict' error on the first upadte, then pass it through, then return a Conflict again
	var returnedConflict int32
	fakeClient.Fake.PrependReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)

		if atomic.LoadInt32(&returnedConflict) == 0 {
			// allow the next update
			atomic.StoreInt32(&returnedConflict, 1)
			return true, nil, errors.NewConflict(apiv1.Resource("node"), obj.GetName(), fmt.Errorf("concurrent update on %s", obj.GetName()))
		}

		// return a conflict on next update
		atomic.StoreInt32(&returnedConflict, 0)
		return false, nil, nil
	})

	return fakeClient
}
