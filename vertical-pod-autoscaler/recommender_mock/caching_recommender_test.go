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

package recommender

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/vertical-pod-autoscaler/test"
)

func TestGetWithCache(t *testing.T) {
	apiMock := &test.RecommenderAPIMock{}
	rec := test.Recommendation("test", "", "")
	pod := test.BuildTestPod("test", "", "", "", nil)
	apiMock.On("GetRecommendation", &pod.Spec).Return(rec, nil)
	recommender := NewCachingRecommender(10*time.Second, apiMock)

	result, err := recommender.Get(&pod.Spec)

	assert.Equal(t, rec, result)
	assert.Equal(t, nil, err)

	// test get from cache
	for i := 0; i < 5; i++ {
		result, err = recommender.Get(&pod.Spec)
	}
	apiMock.AssertNumberOfCalls(t, "GetRecommendation", 1)
}

func TestGetCacheExpired(t *testing.T) {
	apiMock := &test.RecommenderAPIMock{}
	rec := test.Recommendation("test", "", "")
	pod := test.BuildTestPod("test", "", "", "", nil)
	apiMock.On("GetRecommendation", &pod.Spec).Return(rec, nil)
	recommender := NewCachingRecommender(time.Second, apiMock)

	result, err := recommender.Get(&pod.Spec)
	assert.Equal(t, rec, result)
	assert.Equal(t, nil, err)

	<-time.After(2 * time.Second)

	result, err = recommender.Get(&pod.Spec)
	apiMock.AssertNumberOfCalls(t, "GetRecommendation", 2)

}

func TestNoRec(t *testing.T) {
	apiMock := &test.RecommenderAPIMock{}
	pod := test.BuildTestPod("test", "", "", "", nil)
	apiMock.On("GetRecommendation", &pod.Spec).Return(nil, nil)
	recommender := NewCachingRecommender(time.Second, apiMock)

	result, err := recommender.Get(&pod.Spec)
	assert.Nil(t, result)
	assert.Nil(t, err)

	// check nil response not chached
	result, err = recommender.Get(&pod.Spec)
	apiMock.AssertNumberOfCalls(t, "GetRecommendation", 2)
}

func TestError(t *testing.T) {
	apiMock := &test.RecommenderAPIMock{}
	pod := test.BuildTestPod("test", "", "", "", nil)
	err := fmt.Errorf("Expected Fail")
	apiMock.On("GetRecommendation", &pod.Spec).Return(nil, err)
	recommender := NewCachingRecommender(time.Second, apiMock)

	result, err := recommender.Get(&pod.Spec)
	assert.Nil(t, result)
	assert.Error(t, err)

	// check error response not chached
	result, err = recommender.Get(&pod.Spec)
	apiMock.AssertNumberOfCalls(t, "GetRecommendation", 2)
}
