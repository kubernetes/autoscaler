/*
Copyright 2020 The Kubernetes Authors.

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

package kubernetes

import (
	clientv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
)

type eventSinkLoggingWrapper struct {
	actualSink record.EventSink
}

// Create wraps EventSink's Create().
func (s eventSinkLoggingWrapper) Create(event *clientv1.Event) (*clientv1.Event, error) {
	logEvent(event)
	return s.actualSink.Create(event)
}

// Update wraps EventSink's Update().
func (s eventSinkLoggingWrapper) Update(event *clientv1.Event) (*clientv1.Event, error) {
	logEvent(event)
	return s.actualSink.Update(event)
}

// Patch wraps EventSink's Patch().
func (s eventSinkLoggingWrapper) Patch(oldEvent *clientv1.Event, data []byte) (*clientv1.Event, error) {
	logEvent(oldEvent)
	return s.actualSink.Patch(oldEvent, data)
}

func logEvent(e *clientv1.Event) {
	klog.V(4).Infof("Event(%#v): type: '%v' reason: '%v' %v", e.InvolvedObject, e.Type, e.Reason, e.Message)
}

// WrapEventSinkWithLogging adds logging each event via klog to an existing event sink.
func WrapEventSinkWithLogging(sink record.EventSink) record.EventSink {
	return eventSinkLoggingWrapper{actualSink: sink}
}
