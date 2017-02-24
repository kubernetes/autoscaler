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

package core

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/contrib/cluster-autoscaler/config/dynamic"
	"testing"
	"time"
)

type AutoscalerMock struct {
	mock.Mock
}

func (m *AutoscalerMock) RunOnce(currentTime time.Time) {
	m.Called(currentTime)
}

func (m *AutoscalerMock) CleanUp() {
	m.Called()
}

type ConfigFetcherMock struct {
	mock.Mock
}

func (m *ConfigFetcherMock) FetchConfigIfUpdated() (*dynamic.Config, error) {
	args := m.Called()
	return args.Get(0).(*dynamic.Config), args.Error(1)
}

type AutoscalerBuilderMock struct {
	mock.Mock
}

func (m *AutoscalerBuilderMock) SetDynamicConfig(config dynamic.Config) AutoscalerBuilder {
	args := m.Called(config)
	return args.Get(0).(AutoscalerBuilder)
}

func (m *AutoscalerBuilderMock) Build() Autoscaler {
	args := m.Called()
	return args.Get(0).(Autoscaler)
}

func TestRunOnceWhenNoUpdate(t *testing.T) {
	currentTime := time.Now()

	autoscaler := &AutoscalerMock{}
	autoscaler.On("RunOnce", currentTime).Once()

	configFetcher := &ConfigFetcherMock{}
	configFetcher.On("FetchConfigIfUpdated").Return((*dynamic.Config)(nil), nil).Once()

	builder := &AutoscalerBuilderMock{}
	builder.On("Build").Return(autoscaler).Once()

	a := NewDynamicAutoscaler(builder, configFetcher)
	a.RunOnce(currentTime)

	autoscaler.AssertExpectations(t)
	configFetcher.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestRunOnceWhenUpdated(t *testing.T) {
	currentTime := time.Now()

	newConfig := dynamic.NewDefaultConfig()

	initialAutoscaler := &AutoscalerMock{}

	newAutoscaler := &AutoscalerMock{}
	newAutoscaler.On("RunOnce", currentTime).Once()

	configFetcher := &ConfigFetcherMock{}
	configFetcher.On("FetchConfigIfUpdated").Return(&newConfig, nil).Once()

	builder := &AutoscalerBuilderMock{}
	builder.On("Build").Return(initialAutoscaler).Once()
	builder.On("SetDynamicConfig", newConfig).Return(builder).Once()
	builder.On("Build").Return(newAutoscaler).Once()

	a := NewDynamicAutoscaler(builder, configFetcher)
	a.RunOnce(currentTime)

	initialAutoscaler.AssertNotCalled(t, "RunOnce", mock.AnythingOfType("time.Time"))
	newAutoscaler.AssertExpectations(t)
	configFetcher.AssertExpectations(t)
	builder.AssertExpectations(t)
}
