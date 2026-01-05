/*
Copyright 2021 The Kubernetes Authors.

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

package debuggingsnapshot

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

func TestBasicSetterWorkflow(t *testing.T) {
	snapshot := &DebuggingSnapshotImpl{}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "Pod1",
		},
		Spec: v1.PodSpec{
			NodeName: "testNode",
		},
	}
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testNode",
		},
	}
	nodeInfo := framework.NewTestNodeInfo(node, pod)

	var nodeGroups []*framework.NodeInfo
	nodeGroups = append(nodeGroups, nodeInfo)
	timestamp := time.Now().In(time.UTC)
	snapshot.SetClusterNodes(nodeGroups)
	snapshot.SetEndTimestamp(timestamp)
	op, err := snapshot.GetOutputBytes()
	assert.False(t, err)

	type JSONList = []any
	type JSONMap = map[string]any
	var String = "test"

	var parsed map[string]any
	er := json.Unmarshal(op, &parsed)
	assert.NoError(t, er)
	assert.IsType(t, JSONMap{}, parsed)
	assert.IsType(t, JSONList{}, parsed["NodeList"])
	assert.Greater(t, len(parsed["NodeList"].([]any)), 0)
	assert.IsType(t, JSONMap{}, parsed["NodeList"].([]any)[0])
	pNodeInfo := parsed["NodeList"].([]any)[0].(map[string]any)
	assert.IsType(t, JSONMap{}, pNodeInfo["Node"].(map[string]any))
	pNode := pNodeInfo["Node"].(map[string]any)
	assert.IsType(t, JSONMap{}, pNode["metadata"].(map[string]any))
	pNodeObjectMeta := pNode["metadata"].(map[string]any)
	assert.IsType(t, String, pNodeObjectMeta["name"])
	pNodeName := pNodeObjectMeta["name"].(string)
	assert.Equal(t, pNodeName, "testNode")

	assert.IsType(t, JSONList{}, pNodeInfo["Pods"])
	assert.Greater(t, len(pNodeInfo["Pods"].([]any)), 0)
	assert.IsType(t, JSONMap{}, pNodeInfo["Pods"].([]any)[0])
	pPod := pNodeInfo["Pods"].([]any)[0].(map[string]any)
	assert.IsType(t, JSONMap{}, pPod["metadata"])
	pPodMeta := pPod["metadata"].(map[string]any)
	assert.IsType(t, String, pPodMeta["name"])
	pPodName := pPodMeta["name"].(string)
	assert.Equal(t, pPodName, "Pod1")

}

func TestEmptyDataNoError(t *testing.T) {
	snapshot := &DebuggingSnapshotImpl{}
	op, err := snapshot.GetOutputBytes()
	assert.False(t, err)
	assert.NotNil(t, op)
}
