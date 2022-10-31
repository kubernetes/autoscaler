/*
Copyright 2017 The Kubernetes Authors.

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

package aws

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/eks"
	"k8s.io/client-go/tools/cache"
	test_clock "k8s.io/utils/clock/testing"
)

func TestManagedNodegroupCache(t *testing.T) {
	nodegroupName := "nodegroupName"
	clusterName := "clusterName"
	labelKey := "label key 1"
	labelValue := "label value 1"
	taintEffect := "effect 1"
	taintKey := "key 1"
	taintValue := "value 1"
	taint := apiv1.Taint{
		Effect: apiv1.TaintEffect(taintEffect),
		Key:    taintKey,
		Value:  taintValue,
	}

	c := newManagedNodeGroupCache(nil)
	err := c.Add(managedNodegroupCachedObject{
		name:        nodegroupName,
		clusterName: clusterName,
		taints:      []apiv1.Taint{taint},
		labels:      map[string]string{labelKey: labelValue},
	})
	require.NoError(t, err)
	obj, ok, err := c.GetByKey(nodegroupName)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, nodegroupName, obj.(managedNodegroupCachedObject).name)
	assert.Equal(t, clusterName, obj.(managedNodegroupCachedObject).clusterName)
	assert.Equal(t, len(obj.(managedNodegroupCachedObject).labels), 1)
	assert.Equal(t, labelValue, obj.(managedNodegroupCachedObject).labels[labelKey])
	assert.Equal(t, len(obj.(managedNodegroupCachedObject).taints), 1)
	assert.Equal(t, apiv1.TaintEffect(taintEffect), obj.(managedNodegroupCachedObject).taints[0].Effect)
	assert.Equal(t, taintKey, obj.(managedNodegroupCachedObject).taints[0].Key)
	assert.Equal(t, taintValue, obj.(managedNodegroupCachedObject).taints[0].Value)
}

func TestGetManagedNodegroupWithError(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(nil, errors.New("AccessDenied"))

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	// Make sure there's an error but no cache object returned
	_, err := c.getManagedNodegroup(nodegroupName, clusterName)
	require.Error(t, err)

	// Make sure an empty cache object was saved
	obj, ok, err := c.GetByKey(nodegroupName)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, nodegroupName, obj.(managedNodegroupCachedObject).name)
	assert.Equal(t, clusterName, obj.(managedNodegroupCachedObject).clusterName)
	assert.Nil(t, obj.(managedNodegroupCachedObject).labels)
	assert.Nil(t, obj.(managedNodegroupCachedObject).taints)
}

func TestGetManagedNodegroupNoTaintsOrLabels(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"
	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      nil,
		Labels:        nil,
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        nil,
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	cacheObj, err := c.getManagedNodegroup(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, cacheObj.name, nodegroupName)
	assert.Equal(t, cacheObj.clusterName, clusterName)
	assert.Equal(t, len(cacheObj.taints), 0)
	assert.Equal(t, len(cacheObj.labels), 4)
	assert.Equal(t, cacheObj.labels["amiType"], amiType)
	assert.Equal(t, cacheObj.labels["capacityType"], capacityType)
	assert.Equal(t, cacheObj.labels["k8sVersion"], k8sVersion)
	assert.Equal(t, cacheObj.labels["eks.amazonaws.com/nodegroup"], nodegroupName)
}

func TestGetManagedNodegroupWithTaintsAndLabels(t *testing.T) {
	k := &eksMock{}

	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"
	diskSize := int64(100)
	nodegroupName := "testNodegroup"
	clusterName := "testCluster"

	labelKey1 := "labelKey 1"
	labelKey2 := "labelKey 2"
	labelValue1 := "testValue 1"
	labelValue2 := "testValue 2"

	taintEffect1 := "effect 1"
	taintKey1 := "key 1"
	taintValue1 := "value 1"
	taint1 := eks.Taint{
		Effect: &taintEffect1,
		Key:    &taintKey1,
		Value:  &taintValue1,
	}

	taintEffect2 := "effect 2"
	taintKey2 := "key 2"
	taintValue2 := "value 2"
	taint2 := eks.Taint{
		Effect: &taintEffect2,
		Key:    &taintKey2,
		Value:  &taintValue2,
	}

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        map[string]*string{labelKey1: &labelValue1, labelKey2: &labelValue2},
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        []*eks.Taint{&taint1, &taint2},
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	cacheObj, err := c.getManagedNodegroup(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, cacheObj.name, nodegroupName)
	assert.Equal(t, cacheObj.clusterName, clusterName)
	assert.Equal(t, len(cacheObj.taints), 2)
	assert.Equal(t, cacheObj.taints[0].Effect, apiv1.TaintEffect(taintEffect1))
	assert.Equal(t, cacheObj.taints[0].Key, taintKey1)
	assert.Equal(t, cacheObj.taints[0].Value, taintValue1)
	assert.Equal(t, cacheObj.taints[1].Effect, apiv1.TaintEffect(taintEffect2))
	assert.Equal(t, cacheObj.taints[1].Key, taintKey2)
	assert.Equal(t, cacheObj.taints[1].Value, taintValue2)
	assert.Equal(t, len(cacheObj.labels), 7)
	assert.Equal(t, cacheObj.labels[labelKey1], labelValue1)
	assert.Equal(t, cacheObj.labels[labelKey2], labelValue2)
	assert.Equal(t, cacheObj.labels["diskSize"], strconv.FormatInt(diskSize, 10))
	assert.Equal(t, cacheObj.labels["amiType"], amiType)
	assert.Equal(t, cacheObj.labels["capacityType"], capacityType)
	assert.Equal(t, cacheObj.labels["k8sVersion"], k8sVersion)
	assert.Equal(t, cacheObj.labels["eks.amazonaws.com/nodegroup"], nodegroupName)
}

func TestGetManagedNodegroupInfoObjectWithError(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(nil, errors.New("AccessDenied"))

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	// Make sure there's an error
	mngInfoObject, err := c.getManagedNodegroupInfoObject(nodegroupName, clusterName)
	require.Error(t, err)

	// Make sure an object isn't returned
	assert.Nil(t, mngInfoObject)
}

func TestGetManagedNodegroupInfoObjectWithCachedNodegroup(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "nodegroupName"
	clusterName := "clusterName"
	labelKey := "label key 1"
	labelValue := "label value 1"
	taintEffect := "effect 1"
	taintKey := "key 1"
	taintValue := "value 1"
	taint := apiv1.Taint{
		Effect: apiv1.TaintEffect(taintEffect),
		Key:    taintKey,
		Value:  taintValue,
	}

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})
	err := c.Add(managedNodegroupCachedObject{
		name:        nodegroupName,
		clusterName: clusterName,
		taints:      []apiv1.Taint{taint},
		labels:      map[string]string{labelKey: labelValue},
	})

	mngInfoObject, err := c.getManagedNodegroupInfoObject(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(mngInfoObject.labels), 1)
	assert.Equal(t, mngInfoObject.labels[labelKey], labelValue)
	k.AssertNotCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}

func TestGetManagedNodegroupInfoObjectNoCachedNodegroup(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"
	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"
	diskSize := int64(100)

	labelKey1 := "labelKey 1"
	labelKey2 := "labelKey 2"
	labelValue1 := "testValue 1"
	labelValue2 := "testValue 2"

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        map[string]*string{labelKey1: &labelValue1, labelKey2: &labelValue2},
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        nil,
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	mngInfoObject, err := c.getManagedNodegroupInfoObject(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(mngInfoObject.labels), 7)
	assert.Equal(t, mngInfoObject.labels[labelKey1], labelValue1)
	assert.Equal(t, mngInfoObject.labels[labelKey2], labelValue2)
	assert.Equal(t, mngInfoObject.labels["diskSize"], strconv.FormatInt(diskSize, 10))
	assert.Equal(t, mngInfoObject.labels["amiType"], amiType)
	assert.Equal(t, mngInfoObject.labels["capacityType"], capacityType)
	assert.Equal(t, mngInfoObject.labels["k8sVersion"], k8sVersion)
	assert.Equal(t, mngInfoObject.labels["eks.amazonaws.com/nodegroup"], nodegroupName)
	k.AssertCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}

func TestGetManagedNodegroupLabelsWithCachedNodegroup(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "nodegroupName"
	clusterName := "clusterName"
	labelKey := "label key 1"
	labelValue := "label value 1"
	taintEffect := "effect 1"
	taintKey := "key 1"
	taintValue := "value 1"
	taint := apiv1.Taint{
		Effect: apiv1.TaintEffect(taintEffect),
		Key:    taintKey,
		Value:  taintValue,
	}

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})
	err := c.Add(managedNodegroupCachedObject{
		name:        nodegroupName,
		clusterName: clusterName,
		taints:      []apiv1.Taint{taint},
		labels:      map[string]string{labelKey: labelValue},
	})

	labelsMap, err := c.getManagedNodegroupLabels(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(labelsMap), 1)
	assert.Equal(t, labelsMap[labelKey], labelValue)
	k.AssertNotCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}

func TestGetManagedNodegroupLabelsNoCachedNodegroup(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"
	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"
	diskSize := int64(100)

	labelKey1 := "labelKey 1"
	labelKey2 := "labelKey 2"
	labelValue1 := "testValue 1"
	labelValue2 := "testValue 2"

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        map[string]*string{labelKey1: &labelValue1, labelKey2: &labelValue2},
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        nil,
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	labelsMap, err := c.getManagedNodegroupLabels(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(labelsMap), 7)
	assert.Equal(t, labelsMap[labelKey1], labelValue1)
	assert.Equal(t, labelsMap[labelKey2], labelValue2)
	assert.Equal(t, labelsMap["diskSize"], strconv.FormatInt(diskSize, 10))
	assert.Equal(t, labelsMap["amiType"], amiType)
	assert.Equal(t, labelsMap["capacityType"], capacityType)
	assert.Equal(t, labelsMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelsMap["eks.amazonaws.com/nodegroup"], nodegroupName)
	k.AssertCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}

func TestGetManagedNodegroupLabelsWithCachedNodegroupThatExpires(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"
	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"
	diskSize := int64(100)

	labelKey1 := "labelKey 1"
	labelKey2 := "labelKey 2"
	labelValue1 := "testValue 1"
	labelValue2 := "testValue 2"

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        map[string]*string{labelKey1: &labelValue1, labelKey2: &labelValue2},
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        nil,
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	fakeClock := test_clock.NewFakeClock(time.Unix(0, 0))
	fakeStore := cache.NewFakeExpirationStore(
		func(obj interface{}) (s string, e error) {
			return obj.(managedNodegroupCachedObject).name, nil
		},
		nil,
		&cache.TTLPolicy{
			TTL:   managedNodegroupCachedTTL,
			Clock: fakeClock,
		},
		fakeClock,
	)

	// Create cache with fake clock
	c := newManagedNodeGroupCacheWithClock(&awsWrapper{nil, nil, k}, fakeClock, fakeStore)

	// Add nodegroup entry that will expire
	err := c.Add(managedNodegroupCachedObject{
		name:        nodegroupName,
		clusterName: clusterName,
		taints:      make([]apiv1.Taint, 0),
		labels:      map[string]string{labelKey1: labelValue1},
	})
	require.NoError(t, err)
	obj, ok, err := c.GetByKey(nodegroupName)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, nodegroupName, obj.(managedNodegroupCachedObject).name)
	assert.Equal(t, clusterName, obj.(managedNodegroupCachedObject).clusterName)
	assert.Equal(t, len(obj.(managedNodegroupCachedObject).labels), 1)
	assert.Equal(t, labelValue1, obj.(managedNodegroupCachedObject).labels[labelKey1])
	assert.Equal(t, len(obj.(managedNodegroupCachedObject).taints), 0)

	// Query for nodegroup entry before it expires
	labelsMap, err := c.getManagedNodegroupLabels(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(labelsMap), 1)
	assert.Equal(t, labelsMap[labelKey1], labelValue1)
	k.AssertNotCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})

	// Expire nodegroup
	fakeClock.SetTime(time.Unix(0, 0).Add(managedNodegroupCachedTTL + 1*time.Minute))

	// Query for nodegroup entry after it expires - should have the new labels added
	newLabelsMap, err := c.getManagedNodegroupLabels(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(newLabelsMap), 7)
	assert.Equal(t, newLabelsMap[labelKey1], labelValue1)
	assert.Equal(t, newLabelsMap[labelKey2], labelValue2)
	assert.Equal(t, newLabelsMap["diskSize"], strconv.FormatInt(diskSize, 10))
	assert.Equal(t, newLabelsMap["amiType"], amiType)
	assert.Equal(t, newLabelsMap["capacityType"], capacityType)
	assert.Equal(t, newLabelsMap["k8sVersion"], k8sVersion)
	assert.Equal(t, newLabelsMap["eks.amazonaws.com/nodegroup"], nodegroupName)
	k.AssertCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}

func TestGetManagedNodegroupTaintsWithCachedNodegroup(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "nodegroupName"
	clusterName := "clusterName"
	labelKey := "label key 1"
	labelValue := "label value 1"
	taintEffect := "effect 1"
	taintKey := "key 1"
	taintValue := "value 1"
	taint := apiv1.Taint{
		Effect: apiv1.TaintEffect(taintEffect),
		Key:    taintKey,
		Value:  taintValue,
	}

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})
	err := c.Add(managedNodegroupCachedObject{
		name:        nodegroupName,
		clusterName: clusterName,
		taints:      []apiv1.Taint{taint},
		labels:      map[string]string{labelKey: labelValue},
	})

	taintsList, err := c.getManagedNodegroupTaints(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(taintsList), 1)
	assert.Equal(t, taintsList[0].Effect, apiv1.TaintEffect(taintEffect))
	assert.Equal(t, taintsList[0].Key, taintKey)
	assert.Equal(t, taintsList[0].Value, taintValue)
	k.AssertNotCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}

func TestGetManagedNodegroupTaintsNoCachedNodegroup(t *testing.T) {
	k := &eksMock{}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"
	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"
	diskSize := int64(100)

	taintEffect1 := "effect 1"
	taintKey1 := "key 1"
	taintValue1 := "value 1"
	taint1 := eks.Taint{
		Effect: &taintEffect1,
		Key:    &taintKey1,
		Value:  &taintValue1,
	}

	taintEffect2 := "effect 2"
	taintKey2 := "key 2"
	taintValue2 := "value 2"
	taint2 := eks.Taint{
		Effect: &taintEffect2,
		Key:    &taintKey2,
		Value:  &taintValue2,
	}

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        nil,
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        []*eks.Taint{&taint1, &taint2},
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	c := newManagedNodeGroupCache(&awsWrapper{nil, nil, k})

	taintsList, err := c.getManagedNodegroupTaints(nodegroupName, clusterName)
	require.NoError(t, err)
	assert.Equal(t, len(taintsList), 2)
	assert.Equal(t, taintsList[0].Effect, apiv1.TaintEffect(taintEffect1))
	assert.Equal(t, taintsList[0].Key, taintKey1)
	assert.Equal(t, taintsList[0].Value, taintValue1)
	assert.Equal(t, taintsList[1].Effect, apiv1.TaintEffect(taintEffect2))
	assert.Equal(t, taintsList[1].Key, taintKey2)
	assert.Equal(t, taintsList[1].Value, taintValue2)
	k.AssertCalled(t, "DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	})
}
