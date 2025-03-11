// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package fakeclient

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"

	fakeuntyped "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/fake"
	apipolicyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	policyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	fakepolicyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1/fake"
	k8stesting "k8s.io/client-go/testing"
)

// FakeObjectTracker implements both k8stesting.ObjectTracker as well as watch.Interface.
type FakeObjectTracker struct {
	*watch.FakeWatcher
	delegatee    k8stesting.ObjectTracker
	watchers     []*watcher
	trackerMutex sync.Mutex
	fakingOptions
}

// Add receives an add event with the object
func (t *FakeObjectTracker) Add(obj runtime.Object) error {
	if t.fakingOptions.failAll != nil {
		err := t.fakingOptions.failAll.RunFakeInvocations()
		if err != nil {
			return err
		}
	}

	return t.delegatee.Add(obj)
}

// Get receives a get event with the object
func (t *FakeObjectTracker) Get(gvr schema.GroupVersionResource, ns, name string, getOps ...metav1.GetOptions) (runtime.Object, error) {
	var err error
	if t.fakingOptions.failAll != nil {
		err = t.fakingOptions.failAll.RunFakeInvocations()
		if err != nil {
			return nil, err
		}
	}
	if t.fakingOptions.failAt != nil {
		if gvr.Resource == "nodes" {
			err = t.fakingOptions.failAt.Node.Get.RunFakeInvocations()
		} else if gvr.Resource == "machines" {
			err = t.fakingOptions.failAt.Machine.Get.RunFakeInvocations()
		} else if gvr.Resource == "machinesets" {
			err = t.fakingOptions.failAt.MachineSet.Get.RunFakeInvocations()
		} else if gvr.Resource == "machinedeployments" {
			err = t.fakingOptions.failAt.MachineDeployment.Get.RunFakeInvocations()
		}
		if err != nil {
			return nil, err
		}
	}

	return t.delegatee.Get(gvr, ns, name, getOps...)
}

// Create receives a create event with the object. Not needed for CA.
func (t *FakeObjectTracker) Create(gvr schema.GroupVersionResource, obj runtime.Object, ns string, createOps ...metav1.CreateOptions) error {
	return nil
}

// Apply receives an apply event with the object. Not needed for CA.
func (t *FakeObjectTracker) Apply(gvr schema.GroupVersionResource, obj runtime.Object, ns string, patchOps ...metav1.PatchOptions) error {
	return nil
}

// Patch receives a patch event with the object. Not needed for CA.
func (t *FakeObjectTracker) Patch(gvr schema.GroupVersionResource, obj runtime.Object, ns string, patchOps ...metav1.PatchOptions) error {
	var err error
	if t.fakingOptions.failAll != nil {
		err = t.fakingOptions.failAll.RunFakeInvocations()
		if err != nil {
			return err
		}
	}
	if t.fakingOptions.failAt != nil {
		if gvr.Resource == "nodes" {
			err = t.fakingOptions.failAt.Node.Update.RunFakeInvocations()
		} else if gvr.Resource == "machines" {
			err = t.fakingOptions.failAt.Machine.Update.RunFakeInvocations()
		} else if gvr.Resource == "machinedeployments" {
			err = t.fakingOptions.failAt.MachineDeployment.Update.RunFakeInvocations()
		}
		if err != nil {
			return err
		}
	}
	err = t.delegatee.Patch(gvr, obj, ns, patchOps...)
	if err != nil {
		return err
	}

	if t.FakeWatcher == nil {
		return errors.New("error sending event on a tracker with no watch support")
	}

	if t.IsStopped() {
		return errors.New("error sending event on a stopped tracker")
	}

	t.FakeWatcher.Modify(obj)
	return nil
}

// Update receives an update event with the object
func (t *FakeObjectTracker) Update(gvr schema.GroupVersionResource, obj runtime.Object, ns string, updateOps ...metav1.UpdateOptions) error {
	var err error
	if t.fakingOptions.failAll != nil {
		err = t.fakingOptions.failAll.RunFakeInvocations()
		if err != nil {
			return err
		}
	}
	if t.fakingOptions.failAt != nil {
		if gvr.Resource == "nodes" {
			err = t.fakingOptions.failAt.Node.Update.RunFakeInvocations()
		} else if gvr.Resource == "machines" {
			err = t.fakingOptions.failAt.Machine.Update.RunFakeInvocations()
		} else if gvr.Resource == "machinedeployments" {
			err = t.fakingOptions.failAt.MachineDeployment.Update.RunFakeInvocations()
		}
		if err != nil {
			return err
		}
	}
	err = t.delegatee.Update(gvr, obj, ns, updateOps...)
	if err != nil {
		return err
	}

	if t.FakeWatcher == nil {
		return errors.New("error sending event on a tracker with no watch support")
	}

	if t.IsStopped() {
		return errors.New("error sending event on a stopped tracker")
	}

	t.FakeWatcher.Modify(obj)
	return nil
}

// List receives a list event with the object
func (t *FakeObjectTracker) List(gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, ns string, listOps ...metav1.ListOptions) (runtime.Object, error) {
	if t.fakingOptions.failAll != nil {
		err := t.fakingOptions.failAll.RunFakeInvocations()
		if err != nil {
			return nil, err
		}
	}
	return t.delegatee.List(gvr, gvk, ns, listOps...)
}

// Delete receives an delete event with the object. Not needed for CA.
func (t *FakeObjectTracker) Delete(gvr schema.GroupVersionResource, ns, name string, deleteOps ...metav1.DeleteOptions) error {
	return nil
}

// Watch receives a watch event with the object
func (t *FakeObjectTracker) Watch(gvr schema.GroupVersionResource, name string, listOps ...metav1.ListOptions) (watch.Interface, error) {
	if t.fakingOptions.failAll != nil {
		err := t.fakingOptions.failAll.RunFakeInvocations()
		if err != nil {
			return nil, err
		}
	}
	return t.delegatee.Watch(gvr, name, listOps...)
}

func (t *FakeObjectTracker) watchReactionFunc(action k8stesting.Action) (bool, watch.Interface, error) {
	if t.FakeWatcher == nil {
		return false, nil, errors.New("cannot watch on a tracker with no watch support")
	}

	switch a := action.(type) {
	case k8stesting.WatchAction:
		w := &watcher{
			FakeWatcher: watch.NewFake(),
			action:      a,
		}
		go func() {
			err := w.dispatchInitialObjects(a, t)
			if err != nil {
				klog.Errorf("error dispatching initial objects, Err: %v", err)
			}
		}()
		t.trackerMutex.Lock()
		defer t.trackerMutex.Unlock()
		t.watchers = append(t.watchers, w)
		return true, w, nil
	default:
		return false, nil, fmt.Errorf("expected WatchAction but got %v", action)
	}
}

// Start begins tracking of an FakeObjectTracker
func (t *FakeObjectTracker) Start() error {
	if t.FakeWatcher == nil {
		return errors.New("tracker has no watch support")
	}

	for event := range t.ResultChan() {
		event := event.DeepCopy() // passing a deep copy to avoid race.
		t.dispatch(event)
	}

	return nil
}

func (t *FakeObjectTracker) dispatch(event *watch.Event) {
	for _, w := range t.watchers {
		go w.dispatch(event)
	}
}

// Stop terminates tracking of an FakeObjectTracker
func (t *FakeObjectTracker) Stop() {
	if t.FakeWatcher == nil {
		panic(errors.New("tracker has no watch support"))
	}

	t.trackerMutex.Lock()
	defer t.trackerMutex.Unlock()

	t.FakeWatcher.Stop()
	for _, w := range t.watchers {
		w.Stop()
	}
}

type watcher struct {
	*watch.FakeWatcher
	action      k8stesting.WatchAction
	updateMutex sync.Mutex
}

func (w *watcher) Stop() {
	w.updateMutex.Lock()
	defer w.updateMutex.Unlock()

	w.FakeWatcher.Stop()
}

func (w *watcher) handles(event *watch.Event) bool {
	if w.IsStopped() {
		return false
	}

	t, err := meta.TypeAccessor(event.Object)
	if err != nil {
		return false
	}

	gvr, _ := meta.UnsafeGuessKindToResource(schema.FromAPIVersionAndKind(t.GetAPIVersion(), t.GetKind()))
	if !(&k8stesting.SimpleWatchReactor{Resource: gvr.Resource}).Handles(w.action) {
		return false
	}

	o, err := meta.Accessor(event.Object)
	if err != nil {
		return false
	}

	info := w.action.GetWatchRestrictions()
	rv, fs, ls := info.ResourceVersion, info.Fields, info.Labels
	if rv != "" && o.GetResourceVersion() != rv {
		return false
	}

	if fs != nil && !fs.Matches(fields.Set{
		"metadata.name":      o.GetName(),
		"metadata.namespace": o.GetNamespace(),
	}) {
		return false
	}

	if ls != nil && !ls.Matches(labels.Set(o.GetLabels())) {
		return false
	}

	return true
}

func (w *watcher) dispatch(event *watch.Event) {
	w.updateMutex.Lock()
	defer w.updateMutex.Unlock()

	if !w.handles(event) {
		return
	}
	w.Action(event.Type, event.Object)
}

func (w *watcher) dispatchInitialObjects(action k8stesting.WatchAction, t k8stesting.ObjectTracker) error {
	listObj, err := t.List(action.GetResource(), action.GetResource().GroupVersion().WithKind(action.GetResource().Resource), action.GetNamespace())
	if err != nil {
		return err
	}

	itemsPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return err
	}

	items := itemsPtr.([]runtime.Object)
	for _, o := range items {
		w.dispatch(&watch.Event{
			Type:   watch.Added,
			Object: o,
		})
	}

	return nil
}

// ResourceActions contains of Kubernetes/Machine resources whose response can be faked
type ResourceActions struct {
	Node              Actions
	Machine           Actions
	MachineSet        Actions
	MachineDeployment Actions
}

// Actions contains the actions whose response can be faked
type Actions struct {
	Get    FakeResponse
	Update FakeResponse
}

// FakeResponse is the custom error response configuration that are used for responding to client calls
type FakeResponse struct {
	counter       int
	errorMsg      string
	responseDelay time.Duration
}

// fakingOptions are options that can be set while trying to fake object tracker returns
type fakingOptions struct {
	// Fail at different resource action
	failAt *ResourceActions
	// Fail every action
	failAll *FakeResponse
}

// CreateFakeResponse creates a fake response for an action
func CreateFakeResponse(counter int, errorMsg string, responseDelay time.Duration) FakeResponse {
	return FakeResponse{
		counter:       counter,
		errorMsg:      errorMsg,
		responseDelay: responseDelay,
	}
}

// DecrementCounter reduces the counter for the particular action response by 1
func (o *FakeResponse) DecrementCounter() {
	o.counter--
}

// RunFakeInvocations runs any custom fake configurations/methods before invoking standard ObjectTrackers
func (o *FakeResponse) RunFakeInvocations() error {
	if !o.IsFakingEnabled() {
		return nil
	}
	// decrement the counter
	o.DecrementCounter()

	// Delay while returning call
	if o.responseDelay != 0 {
		time.Sleep(o.responseDelay)
	}

	// If error message has been set
	if o.errorMsg != "" {
		return errors.New(o.errorMsg)
	}
	return nil
}

// IsFakingEnabled will return true if counter is positive for the fake response
func (o *FakeResponse) IsFakingEnabled() bool {
	return o.counter > 0
}

// SetFailAtFakeResourceActions sets up the errorMessage to be returned on specific calls
func (o *fakingOptions) SetFailAtFakeResourceActions(resourceActions *ResourceActions) {
	o.failAt = resourceActions
}

// SetFailAllFakeResponse sets the error message for all calls from the client
func (o *fakingOptions) SetFailAllFakeResponse(response *FakeResponse) {
	o.failAll = response
}

// NewMachineClientSet returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewMachineClientSet(objects ...runtime.Object) (*fakeuntyped.Clientset, *FakeObjectTracker) {
	var scheme = runtime.NewScheme()
	var codecs = serializer.NewCodecFactory(scheme)

	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	_ = fakeuntyped.AddToScheme(scheme)

	o := &FakeObjectTracker{
		FakeWatcher: watch.NewFake(),
		delegatee:   k8stesting.NewObjectTracker(scheme, codecs.UniversalDecoder()),
	}

	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &fakeuntyped.Clientset{}
	cs.Fake.AddReactor("*", "*", k8stesting.ObjectReaction(o))
	cs.Fake.AddWatchReactor("*", o.watchReactionFunc)

	return cs, o
}

// FakeObjectTrackers is a struct containing all the controller fake object trackers
type FakeObjectTrackers struct {
	ControlMachine, TargetCore, ControlApps *FakeObjectTracker
}

// NewFakeObjectTrackers initializes fakeObjectTrackers initializes the fake object trackers
func NewFakeObjectTrackers(controlMachine, targetCore, controlApps *FakeObjectTracker) *FakeObjectTrackers {

	fakeObjectTrackers := &FakeObjectTrackers{
		ControlMachine: controlMachine,
		TargetCore:     targetCore,
		ControlApps:    controlApps,
	}

	return fakeObjectTrackers
}

// Start starts all object trackers as go routines
func (o *FakeObjectTrackers) Start() {
	go func() {
		err := o.ControlMachine.Start()
		if err != nil {
			klog.Errorf("failed to start machine object tracker, Err: %v", err)
		}
	}()

	go func() {
		err := o.TargetCore.Start()
		if err != nil {
			klog.Errorf("failed to start target core object tracker, Err: %v", err)
		}
	}()
}

// Stop stops all object trackers
func (o *FakeObjectTrackers) Stop() {
	o.ControlMachine.Stop()
	o.TargetCore.Stop()
}

// NewCoreClientSet returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewCoreClientSet(objects ...runtime.Object) (*Clientset, *FakeObjectTracker) {

	var scheme = runtime.NewScheme()
	var codecs = serializer.NewCodecFactory(scheme)

	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	_ = k8sfake.AddToScheme(scheme)

	o := &FakeObjectTracker{
		FakeWatcher: watch.NewFake(),
		delegatee:   k8stesting.NewObjectTracker(scheme, codecs.UniversalDecoder()),
	}

	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{Clientset: &k8sfake.Clientset{}}
	cs.FakeDiscovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.Fake.AddReactor("*", "*", k8stesting.ObjectReaction(o))
	cs.Fake.AddWatchReactor("*", o.watchReactionFunc)

	return cs, o
}

// NewAppsClientSet returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewAppsClientSet(objects ...runtime.Object) (*Clientset, *FakeObjectTracker) {

	var scheme = runtime.NewScheme()
	var codecs = serializer.NewCodecFactory(scheme)

	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1", Group: "apps"})
	_ = k8sfake.AddToScheme(scheme)

	o := &FakeObjectTracker{
		FakeWatcher: watch.NewFake(),
		delegatee:   k8stesting.NewObjectTracker(scheme, codecs.UniversalDecoder()),
	}

	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{Clientset: &k8sfake.Clientset{}}
	cs.FakeDiscovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.Fake.AddReactor("*", "*", k8stesting.ObjectReaction(o))
	cs.Fake.AddWatchReactor("*", o.watchReactionFunc)

	return cs, o
}

// Clientset extends k8sfake.Clientset to override the Policy implementation.
// This is because the default Policy fake implementation does not propagate the
// eviction name.
type Clientset struct {
	*k8sfake.Clientset
	FakeDiscovery *fakediscovery.FakeDiscovery
}

// Discovery returns the fake discovery implementation.
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.FakeDiscovery
}

// PolicyV1beta1 retrieves the PolicyV1beta1Client
func (c *Clientset) PolicyV1beta1() policyv1beta1.PolicyV1beta1Interface {
	return &FakePolicyV1beta1{
		FakePolicyV1beta1: &fakepolicyv1beta1.FakePolicyV1beta1{
			Fake: &c.Fake,
		},
	}
}

// Policy retrieves the PolicyV1beta1Client
func (c *Clientset) Policy() policyv1beta1.PolicyV1beta1Interface {
	return c.PolicyV1beta1()
}

// FakePolicyV1beta1 extends fakepolicyv1beta1.FakePolicyV1beta1 to override the
// Policy implementation. This is because the default Policy fake implementation
// does not propagate the eviction name.
type FakePolicyV1beta1 struct {
	*fakepolicyv1beta1.FakePolicyV1beta1
}

// Evictions extends fakepolicyv1beta1.FakeEvictions to override the
// Policy implementation. This is because the default Policy fake implementation
// does not propagate the eviction name.
func (c *FakePolicyV1beta1) Evictions(namespace string) policyv1beta1.EvictionInterface {
	return &FakeEvictions{
		FakePolicyV1beta1: c.FakePolicyV1beta1,
		ns:                namespace,
	}
}

// FakeEvictions extends fakepolicyv1beta1.FakeEvictions to override the
// Policy implementation. This is because the default Policy fake implementation
// does not propagate the eviction name.
type FakeEvictions struct {
	*fakepolicyv1beta1.FakePolicyV1beta1
	ns string
}

// Evict overrides the fakepolicyv1beta1.FakeEvictions to override the
// Policy implementation. This is because the default Policy fake implementation
// does not propagate the eviction name.
func (c *FakeEvictions) Evict(_ context.Context, eviction *apipolicyv1beta1.Eviction) error {
	action := k8stesting.GetActionImpl{}
	action.Name = eviction.Name
	action.Verb = "post"
	action.Namespace = c.ns
	action.Resource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	action.Subresource = "eviction"
	_, err := c.Fake.Invokes(action, eviction)
	return err
}
