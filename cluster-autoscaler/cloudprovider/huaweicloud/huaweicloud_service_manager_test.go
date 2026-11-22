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

package huaweicloud

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	huaweicloudsdkas "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	huaweicloudsdkasmodel "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"

	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
)

func Test_extractTaintsFromTags(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want []apiv1.Taint
	}{
		{
			name: "tag in right format",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "empty taint key should be ignored",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_": "bar:NoSchedule",
			},
			want: []apiv1.Taint{},
		},
		{
			name: "invalid tag key should be ignored",
			args: map[string]string{
				"invalidTagKey": "bar:NoSchedule",
			},
			want: []apiv1.Taint{},
		},
		{
			name: "invalid taint effect should be ignored",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:InvalidEffect",
			},
			want: []apiv1.Taint{},
		},
		{
			name: "empty taint value",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": ":NoSchedule",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "one tag with valid tag, one tag with invalid key, ignore the invalid one",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
				"invalidTagKey": ":NoSchedule",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "one tag with valid key/value, one tag with invalid value, ignore the invalid one",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
				"k8s.io_cluster-autoscaler_node-template_taint_bar": "invalidTagValue",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "one tag with valid key/value, one tag with invalid value length, ignore the invalid one",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
				"k8s.io_cluster-autoscaler_node-template_taint_bar": "foo:NoSchedule:more",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTaintsFromTags(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTaintsFromTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildGenericLabels(t *testing.T) {
	template := &asgTemplate{
		name:   "foo",
		region: "foo",
		zone:   "foo",
	}
	tests := []struct {
		name string
		tags map[string]string
		want map[string]string
	}{
		{
			name: "tags contain taints key, ignore it when extract labels",
			tags: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "true:PreferNoSchedule",
				"foo": "bar",
			},
			want: map[string]string{
				apiv1.LabelArchStable:         cloudprovider.DefaultArch,
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				apiv1.LabelInstanceTypeStable: template.name,
				apiv1.LabelTopologyRegion:     template.region,
				apiv1.LabelTopologyZone:       template.zone,
				apiv1.LabelHostname:           "foo",
				"foo":                         "bar",
			},
		},
		{
			name: "tags don't contain taints key",
			tags: map[string]string{
				"foo": "bar",
			},
			want: map[string]string{
				apiv1.LabelArchStable:         cloudprovider.DefaultArch,
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				apiv1.LabelInstanceTypeStable: template.name,
				apiv1.LabelTopologyRegion:     template.region,
				apiv1.LabelTopologyZone:       template.zone,
				apiv1.LabelHostname:           "foo",
				"foo":                         "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template.tags = tt.tags
			if got := buildGenericLabels(template, "foo"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildGenericLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The tests below cover the behaviour introduced by the HuaweiCloud provider
// changes on master (commit 4a95a56c9 "Refactor Huaweicloud service manager
// instance handling" and its follow-ups):
//
//   - listScalingInstancesPaged: Limit=100 pagination, InstanceId de-duplication,
//     termination on an empty page, termination on a short page, an absolute
//     maxPages cap, and an offset that advances by the returned page length.
//   - transformInstanceState: the ErrorCode values added for unhealthy and
//     unknown instances, plus the INITIALIZING health status rename.
//   - buildAsgFromSpec: matching a node group by groupID as well as groupName.
//   - AutoScalingGroup.DeleteNodes: the client-side membership pre-check is
//     gone, so every ProviderID is forwarded to the AS service.

const expectedPageSize = int32(100)
const expectedMaxPages = 100

// makeInstance builds a minimal ScalingGroupInstance carrying only an ID.
func makeInstance(id string) huaweicloudsdkasmodel.ScalingGroupInstance {
	i := id
	return huaweicloudsdkasmodel.ScalingGroupInstance{InstanceId: &i}
}

// makeInstances builds n instances named <prefix>-<zero padded index>.
func makeInstances(prefix string, from, n int) []huaweicloudsdkasmodel.ScalingGroupInstance {
	out := make([]huaweicloudsdkasmodel.ScalingGroupInstance, 0, n)
	for i := from; i < from+n; i++ {
		out = append(out, makeInstance(fmt.Sprintf("%s-%04d", prefix, i)))
	}
	return out
}

// instanceIDs returns the IDs of the given instances, skipping nil entries.
func instanceIDs(instances []huaweicloudsdkasmodel.ScalingGroupInstance) []string {
	ids := make([]string, 0, len(instances))
	for _, inst := range instances {
		if inst.InstanceId == nil {
			continue
		}
		ids = append(ids, *inst.InstanceId)
	}
	return ids
}

// recordedRequest is one ListScalingInstances call as the pager issued it.
type recordedRequest struct {
	groupID string
	start   int32
	limit   int32
}

// requestRecorder wraps a listScalingInstancesFunc, recording every request and
// counting the rows the fake handed back (so tests can prove that rows were
// dropped by de-duplication rather than never fetched).
type requestRecorder struct {
	calls   []recordedRequest
	rowsOut int
}

func (r *requestRecorder) wrap(fn listScalingInstancesFunc) listScalingInstancesFunc {
	return func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		rec := recordedRequest{groupID: opts.ScalingGroupId, limit: -1}
		if opts.StartNumber != nil {
			rec.start = *opts.StartNumber
		}
		if opts.Limit != nil {
			rec.limit = *opts.Limit
		}
		r.calls = append(r.calls, rec)
		resp, err := fn(opts)
		if err == nil && resp != nil && resp.ScalingGroupInstances != nil {
			r.rowsOut += len(*resp.ScalingGroupInstances)
		}
		return resp, err
	}
}

// starts returns the StartNumber of every recorded call.
func (r *requestRecorder) starts() []int32 {
	out := make([]int32, 0, len(r.calls))
	for _, call := range r.calls {
		out = append(out, call.start)
	}
	return out
}

// pagedFunc serves a fixed universe of `total` instances, honouring
// StartNumber/Limit the way the AS API does.
func pagedFunc(total int32) listScalingInstancesFunc {
	all := makeInstances("inst", 0, int(total))
	return func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		start := int32(0)
		if opts.StartNumber != nil {
			start = *opts.StartNumber
		}
		limit := expectedPageSize
		if opts.Limit != nil {
			limit = *opts.Limit
		}
		if start >= total {
			empty := []huaweicloudsdkasmodel.ScalingGroupInstance{}
			return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
				ScalingGroupInstances: &empty,
				TotalNumber:           &total,
			}, nil
		}
		end := start + limit
		if end > total {
			end = total
		}
		page := all[start:end]
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &page,
			TotalNumber:           &total,
		}, nil
	}
}

// staticPage returns a fake that always answers with the same page.
func staticPage(page []huaweicloudsdkasmodel.ScalingGroupInstance, total int32) listScalingInstancesFunc {
	return func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &page,
			TotalNumber:           &total,
		}, nil
	}
}

// assertRequests checks the request shape the pager must always produce.
func assertRequests(t *testing.T, rec *requestRecorder, wantStarts []int32) {
	t.Helper()
	if got := rec.starts(); !reflect.DeepEqual(got, wantStarts) {
		t.Fatalf("StartNumber sequence = %v, want %v", got, wantStarts)
	}
	for i, call := range rec.calls {
		if call.groupID != "test-asg" {
			t.Errorf("call %d: ScalingGroupId = %q, want %q", i, call.groupID, "test-asg")
		}
		if call.limit != expectedPageSize {
			t.Errorf("call %d: Limit = %d, want %d", i, call.limit, expectedPageSize)
		}
	}
}

// TestListScalingInstancesPaged_SinglePage covers the common small-group case:
// one full page followed by the empty page that ends the loop.
func TestListScalingInstancesPaged_SinglePage(t *testing.T) {
	rec := &requestRecorder{}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(pagedFunc(50)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 50 {
		t.Errorf("expected 50 instances, got %d", len(got))
	}
	// A short page does not end pagination: the offset advances by the 50 rows
	// that were returned, and the following empty page confirms the end.
	assertRequests(t, rec, []int32{0, 50})
}

// TestListScalingInstancesPaged_MultiplePages covers 500 instances: five full
// pages plus the empty page that terminates pagination.
func TestListScalingInstancesPaged_MultiplePages(t *testing.T) {
	rec := &requestRecorder{}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(pagedFunc(500)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 500 {
		t.Errorf("expected 500 instances, got %d", len(got))
	}
	if len(rec.calls) != 6 {
		t.Errorf("expected 6 calls (5 data + 1 empty), got %d", len(rec.calls))
	}
	// The offset must advance by the returned page length, never by the number
	// of uniquely kept instances.
	assertRequests(t, rec, []int32{0, 100, 200, 300, 400, 500})
}

// TestListScalingInstancesPaged_LargeGroup1000 covers a ~1000 instance group,
// the customer scenario that motivated the rewrite.
func TestListScalingInstancesPaged_LargeGroup1000(t *testing.T) {
	rec := &requestRecorder{}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(pagedFunc(1000)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1000 {
		t.Errorf("expected 1000 instances, got %d", len(got))
	}
	if len(rec.calls) != 11 {
		t.Errorf("expected 11 calls (10 data + 1 empty), got %d", len(rec.calls))
	}
	if len(rec.starts()) != 11 || rec.starts()[10] != 1000 {
		t.Errorf("expected last StartNumber to be 1000, got %v", rec.starts())
	}
}

// TestListScalingInstancesPaged_TotalNumberDrift pins the maxPages cap: the old
// code compared the collected count against TotalNumber and looped forever when
// the group kept growing during pagination.
func TestListScalingInstancesPaged_TotalNumberDrift(t *testing.T) {
	total := int32(1000)
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		start := *opts.StartNumber
		page := makeInstances("drift", int(start), int(expectedPageSize))
		// TotalNumber keeps moving, so it can never be used as a stop condition.
		total += 100
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &page,
			TotalNumber:           &total,
		}, nil
	}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(fn))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.calls) != expectedMaxPages {
		t.Errorf("expected pagination to stop at the %d page cap, got %d calls", expectedMaxPages, len(rec.calls))
	}
	if len(got) != int(expectedPageSize)*expectedMaxPages {
		t.Errorf("expected %d instances, got %d", int(expectedPageSize)*expectedMaxPages, len(got))
	}
}

// TestListScalingInstancesPaged_EmptyPageTerminates covers an AS reply with a
// non-nil but empty instance slice, which used to spin forever because the
// offset advanced by zero.
func TestListScalingInstancesPaged_EmptyPageTerminates(t *testing.T) {
	first := makeInstances("inst", 0, int(expectedPageSize))
	total := int32(100)
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		if *opts.StartNumber == 0 {
			return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
				ScalingGroupInstances: &first,
				TotalNumber:           &total,
			}, nil
		}
		empty := []huaweicloudsdkasmodel.ScalingGroupInstance{}
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &empty,
			TotalNumber:           &total,
		}, nil
	}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(fn))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 100 {
		t.Errorf("expected 100 instances, got %d", len(got))
	}
	assertRequests(t, rec, []int32{0, 100})
}

// TestListScalingInstancesPaged_RepeatedPageIsDedupedAndCapped covers an AS that
// ignores StartNumber and keeps returning the same page: de-duplication keeps a
// single copy and the page cap still terminates the loop.
func TestListScalingInstancesPaged_RepeatedPageIsDedupedAndCapped(t *testing.T) {
	page := makeInstances("inst", 0, int(expectedPageSize))
	rec := &requestRecorder{}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(staticPage(page, int32(expectedPageSize))))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != int(expectedPageSize) {
		t.Errorf("expected %d de-duplicated instances, got %d", expectedPageSize, len(got))
	}
	if len(rec.calls) != expectedMaxPages {
		t.Errorf("expected %d calls before hitting the cap, got %d", expectedMaxPages, len(rec.calls))
	}
	starts := rec.starts()
	if len(starts) < 3 || !reflect.DeepEqual(starts[:3], []int32{0, 100, 200}) {
		t.Errorf("offset must advance by the returned page length even when every row is a duplicate, starts = %v", starts)
	}
}

// TestListScalingInstancesPaged_DedupAcrossPages covers a page boundary that
// repeats instances already seen.
func TestListScalingInstancesPaged_DedupAcrossPages(t *testing.T) {
	newInstance := "inst-new"
	page0 := make([]huaweicloudsdkasmodel.ScalingGroupInstance, 0, int(expectedPageSize))
	page0 = append(page0, makeInstance(newInstance))
	for i := 1; i <= 99; i++ {
		page0 = append(page0, makeInstance(fmt.Sprintf("inst-%03d", i)))
	}
	page1 := make([]huaweicloudsdkasmodel.ScalingGroupInstance, 0, int(expectedPageSize))
	for i := 1; i <= 100; i++ {
		page1 = append(page1, makeInstance(fmt.Sprintf("inst-%03d", i)))
	}
	page2 := make([]huaweicloudsdkasmodel.ScalingGroupInstance, 0, 99)
	for i := 101; i <= 199; i++ {
		page2 = append(page2, makeInstance(fmt.Sprintf("inst-%03d", i)))
	}
	// Offsets 0, 100 and 200 serve the data; 299 is the 99 row page and the
	// empty page after it ends pagination.
	pages := map[int32][]huaweicloudsdkasmodel.ScalingGroupInstance{
		0:   page0,
		100: page1,
		200: page2,
		299: {},
	}
	total := int32(200)
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		page, ok := pages[*opts.StartNumber]
		if !ok {
			t.Fatalf("unexpected StartNumber %d", *opts.StartNumber)
		}
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &page,
			TotalNumber:           &total,
		}, nil
	}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(fn))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 200 {
		t.Errorf("expected 200 unique instances (100 duplicates dropped), got %d", len(got))
	}
	assertRequests(t, rec, []int32{0, 100, 200, 299})
	seen := make(map[string]struct{}, len(got))
	for _, inst := range got {
		if inst.InstanceId == nil {
			t.Fatal("result contains a nil InstanceId")
		}
		if _, dup := seen[*inst.InstanceId]; dup {
			t.Fatalf("duplicate instance %q survived de-duplication", *inst.InstanceId)
		}
		seen[*inst.InstanceId] = struct{}{}
	}
	if _, ok := seen[newInstance]; !ok {
		t.Errorf("instance %q missing from the result", newInstance)
	}
}

// TestListScalingInstancesPaged_DedupDuringScaleUp models the production
// scenario behind the fix: a 520 instance group receives 50 new instances at
// the head of the list while pagination is in flight, so later pages shift by
// 50 rows and re-serve part of the previous page.
func TestListScalingInstancesPaged_DedupDuringScaleUp(t *testing.T) {
	const instances = 520
	const inserted = 50

	// Page 0 is served from the list before the scale-up happened.
	beforeInsert := makeInstances("inst", 0, instances)
	// After the scale-up the same list with `inserted` new rows in front.
	shifted := make([]huaweicloudsdkasmodel.ScalingGroupInstance, 0, instances+inserted)
	shifted = append(shifted, makeInstances("new", 0, inserted)...)
	shifted = append(shifted, beforeInsert...)

	pages := map[int32][]huaweicloudsdkasmodel.ScalingGroupInstance{
		0:   beforeInsert[:expectedPageSize],
		100: shifted[100:200],
		200: shifted[200:300],
		300: shifted[300:400],
		400: shifted[400:500],
		500: shifted[500:], // 70 rows, the last short page
		570: {},            // empty page: the 70 row page above ends at offset 570
	}
	total := int32(instances)
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		page, ok := pages[*opts.StartNumber]
		if !ok {
			t.Fatalf("unexpected StartNumber %d", *opts.StartNumber)
		}
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &page,
			TotalNumber:           &total,
		}, nil
	}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(fn))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertRequests(t, rec, []int32{0, 100, 200, 300, 400, 500, 570})

	// 570 rows were handed back, 50 of them duplicates: the result must be the
	// 520 original instances, with no new-instance row ever paged in.
	if rec.rowsOut != instances+inserted {
		t.Fatalf("fake handed back %d rows, want %d", rec.rowsOut, instances+inserted)
	}
	if len(got) != instances {
		t.Errorf("expected %d unique instances after de-duplication, got %d", instances, len(got))
	}
	seen := make(map[string]struct{}, len(got))
	for _, inst := range got {
		if inst.InstanceId == nil {
			t.Fatal("result contains a nil InstanceId")
		}
		if _, dup := seen[*inst.InstanceId]; dup {
			t.Fatalf("duplicate instance %q survived de-duplication", *inst.InstanceId)
		}
		seen[*inst.InstanceId] = struct{}{}
	}
	wantIDs := instanceIDs(beforeInsert)
	if !reflect.DeepEqual(sortedKeys(seen), wantIDs) {
		t.Errorf("result IDs differ from the pre-scale-up list")
	}
}

// sortedKeys returns the map keys in ascending order.
func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestListScalingInstancesPaged_ServerCapsPageSize covers a service that honours
// Limit only up to a smaller internal page size (the SDK documents a default of
// 20 and no maximum). A short page must not be mistaken for the end of the
// list, otherwise the instance list is silently truncated and the autoscaler
// stops seeing part of the group.
func TestListScalingInstancesPaged_ServerCapsPageSize(t *testing.T) {
	const serverLimit int32 = 50
	total := int32(520)
	all := makeInstances("inst", 0, int(total))
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		start := *opts.StartNumber
		if start >= total {
			empty := []huaweicloudsdkasmodel.ScalingGroupInstance{}
			return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
				ScalingGroupInstances: &empty,
				TotalNumber:           &total,
			}, nil
		}
		end := start + serverLimit
		if end > total {
			end = total
		}
		page := all[start:end]
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{
			ScalingGroupInstances: &page,
			TotalNumber:           &total,
		}, nil
	}
	got, err := listScalingInstancesPaged("test-asg", rec.wrap(fn))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != int(total) {
		t.Errorf("expected %d instances despite the server side page cap, got %d", total, len(got))
	}
	// 10 full pages of 50, one 20 row page and the terminating empty page.
	if want := int(total/serverLimit) + 2; len(rec.calls) != want {
		t.Errorf("expected %d calls, got %d", want, len(rec.calls))
	}
}

// TestDeleteScalingInstances_NilClient covers an unavailable AS client. The
// method must report an error instead of dereferencing a nil client, which is
// the path DeleteNodes takes now that its instance pre-check is gone.
func TestDeleteScalingInstances_NilClient(t *testing.T) {
	csm := &cloudServiceManager{getASClientFunc: func() *huaweicloudsdkas.AsClient { return nil }}
	err := csm.DeleteScalingInstances("asg-a", []string{"inst-0"})
	if err == nil {
		t.Fatal("expected an error when the AS client is unavailable, got nil")
	}
	if !strings.Contains(err.Error(), "as client") {
		t.Errorf("error = %v, want it to mention the missing AS client", err)
	}
}

// TestListScalingInstancesPaged_ASError covers an AS error: it must surface
// immediately, without a second call.
func TestListScalingInstancesPaged_ASError(t *testing.T) {
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		return nil, fmt.Errorf("AS API timeout")
	}
	if _, err := listScalingInstancesPaged("test-asg", rec.wrap(fn)); err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(rec.calls) != 1 {
		t.Errorf("expected exactly 1 call on error, got %d", len(rec.calls))
	}
}

// TestListScalingInstancesPaged_NilResponse covers a nil response paired with a
// nil error, which must not panic.
func TestListScalingInstancesPaged_NilResponse(t *testing.T) {
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		return nil, nil
	}
	if _, err := listScalingInstancesPaged("test-asg", rec.wrap(fn)); err == nil {
		t.Fatal("expected error for nil response, got nil")
	}
	if len(rec.calls) != 1 {
		t.Errorf("expected exactly 1 call, got %d", len(rec.calls))
	}
}

// TestListScalingInstancesPaged_NilInstances covers a response whose
// ScalingGroupInstances field is nil.
func TestListScalingInstancesPaged_NilInstances(t *testing.T) {
	rec := &requestRecorder{}
	fn := func(opts *huaweicloudsdkasmodel.ListScalingInstancesRequest) (*huaweicloudsdkasmodel.ListScalingInstancesResponse, error) {
		return &huaweicloudsdkasmodel.ListScalingInstancesResponse{ScalingGroupInstances: nil}, nil
	}
	if _, err := listScalingInstancesPaged("test-asg", rec.wrap(fn)); err == nil {
		t.Fatal("expected error for nil ScalingGroupInstances, got nil")
	}
	if len(rec.calls) != 1 {
		t.Errorf("expected exactly 1 call, got %d", len(rec.calls))
	}
}

// TestTransformInstanceState covers the ErrorCode values added to the state
// conversion and the INITIALIZING health status rename.
func TestTransformInstanceState(t *testing.T) {
	csm := &cloudServiceManager{}
	lifeCycle := huaweicloudsdkasmodel.GetScalingGroupInstanceLifeCycleStateEnum()
	health := huaweicloudsdkasmodel.GetScalingGroupInstanceHealthStatusEnum()

	tests := []struct {
		name        string
		lifeCycle   huaweicloudsdkasmodel.ScalingGroupInstanceLifeCycleState
		health      huaweicloudsdkasmodel.ScalingGroupInstanceHealthStatus
		wantState   cloudprovider.InstanceState
		wantErrCode string
		wantErrMsg  string
	}{
		{
			name:      "inservice and normal is running",
			lifeCycle: lifeCycle.INSERVICE,
			health:    health.NORMAL,
			wantState: cloudprovider.InstanceRunning,
		},
		{
			name:      "inservice and initializing stays running",
			lifeCycle: lifeCycle.INSERVICE,
			health:    health.INITIALIZING,
			wantState: cloudprovider.InstanceRunning,
		},
		{
			name:      "pending is creating",
			lifeCycle: lifeCycle.PENDING,
			health:    health.NORMAL,
			wantState: cloudprovider.InstanceCreating,
		},
		{
			name:      "pending wait is creating",
			lifeCycle: lifeCycle.PENDING_WAIT,
			health:    health.NORMAL,
			wantState: cloudprovider.InstanceCreating,
		},
		{
			name:      "removing is deleting",
			lifeCycle: lifeCycle.REMOVING,
			health:    health.NORMAL,
			wantState: cloudprovider.InstanceDeleting,
		},
		{
			name:      "removing wait is deleting",
			lifeCycle: lifeCycle.REMOVING_WAIT,
			health:    health.NORMAL,
			wantState: cloudprovider.InstanceDeleting,
		},
		{
			name:        "unhealthy inservice instance reports healthStatusError",
			lifeCycle:   lifeCycle.INSERVICE,
			health:      health.ERROR,
			wantState:   cloudprovider.InstanceRunning,
			wantErrCode: "healthStatusError",
			wantErrMsg:  "ERROR",
		},
		{
			name:        "unknown health status reports invalidHealthState",
			lifeCycle:   lifeCycle.INSERVICE,
			health:      huaweicloudsdkasmodel.ScalingGroupInstanceHealthStatus{},
			wantState:   cloudprovider.InstanceRunning,
			wantErrCode: "invalidHealthState",
			wantErrMsg:  "invalid instance health state",
		},
		{
			name:       "unknown lifecycle state reports an error",
			lifeCycle:  huaweicloudsdkasmodel.ScalingGroupInstanceLifeCycleState{},
			health:     health.NORMAL,
			wantErrMsg: "invalid instance lifecycle state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := csm.transformInstanceState(tt.lifeCycle, tt.health)
			if got == nil {
				t.Fatal("transformInstanceState returned nil")
			}
			if got.State != tt.wantState {
				t.Errorf("State = %v, want %v", got.State, tt.wantState)
			}
			if tt.wantErrMsg == "" {
				if got.ErrorInfo != nil {
					t.Fatalf("unexpected ErrorInfo: %+v", got.ErrorInfo)
				}
				return
			}
			if got.ErrorInfo == nil {
				t.Fatal("expected ErrorInfo, got nil")
			}
			if got.ErrorInfo.ErrorClass != cloudprovider.OtherErrorClass {
				t.Errorf("ErrorClass = %v, want %v", got.ErrorInfo.ErrorClass, cloudprovider.OtherErrorClass)
			}
			if got.ErrorInfo.ErrorCode != tt.wantErrCode {
				t.Errorf("ErrorCode = %q, want %q", got.ErrorInfo.ErrorCode, tt.wantErrCode)
			}
			if !strings.Contains(got.ErrorInfo.ErrorMessage, tt.wantErrMsg) {
				t.Errorf("ErrorMessage = %q, want it to contain %q", got.ErrorInfo.ErrorMessage, tt.wantErrMsg)
			}
		})
	}
}

// TestBuildAsgFromSpec covers the groupID fallback added to node group lookup.
func TestBuildAsgFromSpec(t *testing.T) {
	asgs := []AutoScalingGroup{
		{groupName: "asg-a", groupID: "id-a", minInstanceNumber: 1, maxInstanceNumber: 5},
		{groupName: "asg-b", groupID: "id-b", minInstanceNumber: 2, maxInstanceNumber: 8},
	}

	tests := []struct {
		name          string
		spec          string
		wantErr       bool
		wantGroupName string
		wantGroupID   string
	}{
		{name: "matches by group name", spec: "1:10:asg-a", wantGroupName: "asg-a", wantGroupID: "id-a"},
		{name: "matches by group id", spec: "1:10:id-b", wantGroupName: "asg-b", wantGroupID: "id-b"},
		{name: "unknown name is an error", spec: "1:10:nope", wantErr: true},
		{name: "malformed spec is an error", spec: "not-a-spec", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAsgFromSpec(tt.spec, asgs, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got group %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.groupName != tt.wantGroupName || got.groupID != tt.wantGroupID {
				t.Errorf("got groupName=%q groupID=%q, want groupName=%q groupID=%q",
					got.groupName, got.groupID, tt.wantGroupName, tt.wantGroupID)
			}
			// The ASG's own size wins; the spec sizes are ignored by this builder.
			if got.minInstanceNumber != 1 && tt.wantGroupID == "id-a" {
				t.Errorf("minInstanceNumber = %d, want 1", got.minInstanceNumber)
			}
		})
	}
}

// fakeAutoScalingService records DeleteScalingInstances calls. It embeds the
// interface so only the methods a test exercises need a body.
type fakeAutoScalingService struct {
	CloudServiceManager
	deleteGroupID string
	deleteIDs     []string
	deleteErr     error
	getInstances  int
}

func (f *fakeAutoScalingService) DeleteScalingInstances(groupID string, instanceIDs []string) error {
	f.deleteGroupID = groupID
	f.deleteIDs = append([]string(nil), instanceIDs...)
	return f.deleteErr
}

func (f *fakeAutoScalingService) GetInstances(groupID string) ([]cloudprovider.Instance, error) {
	f.getInstances++
	return nil, fmt.Errorf("GetInstances must not be used by DeleteNodes")
}

// useMissingKubeconfig makes the post-AS cluster cleanup fail deterministically,
// so the assertions below stay focused on the AS request.
func useMissingKubeconfig(t *testing.T) {
	t.Helper()
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "missing-kubeconfig"))
}

// TestDeleteNodes_ForwardsEveryProviderID covers the removal of the client-side
// membership pre-check: DeleteNodes no longer lists instances first, it hands
// every node's ProviderID straight to the AS service.
func TestDeleteNodes_ForwardsEveryProviderID(t *testing.T) {
	useMissingKubeconfig(t)
	fake := &fakeAutoScalingService{}
	asg := &AutoScalingGroup{
		cloudServiceManager: fake,
		groupName:           "asg-a",
		groupID:             "id-a",
	}

	nodes := []*apiv1.Node{{}, {}, {}}
	for i, node := range nodes {
		node.Name = fmt.Sprintf("node-%d", i)
		node.Spec.ProviderID = fmt.Sprintf("inst-%d", i)
	}

	// The cluster cleanup step fails (no kubeconfig); that is out of scope here.
	_ = asg.DeleteNodes(context.Background(), nodes)

	if fake.getInstances != 0 {
		t.Errorf("DeleteNodes must not pre-list instances, GetInstances called %d time(s)", fake.getInstances)
	}
	if fake.deleteGroupID != "id-a" {
		t.Errorf("DeleteScalingInstances group = %q, want %q", fake.deleteGroupID, "id-a")
	}
	want := []string{"inst-0", "inst-1", "inst-2"}
	if !reflect.DeepEqual(fake.deleteIDs, want) {
		t.Errorf("DeleteScalingInstances ids = %v, want %v", fake.deleteIDs, want)
	}
}

// TestDeleteNodes_PropagatesASError covers AS failures short-circuiting the
// deletion before any cluster cleanup happens.
func TestDeleteNodes_PropagatesASError(t *testing.T) {
	useMissingKubeconfig(t)
	asErr := fmt.Errorf("AS rejected the request")
	fake := &fakeAutoScalingService{deleteErr: asErr}
	asg := &AutoScalingGroup{cloudServiceManager: fake, groupID: "id-a"}

	node := &apiv1.Node{}
	node.Name = "node-0"
	node.Spec.ProviderID = "inst-0"

	err := asg.DeleteNodes(context.Background(), []*apiv1.Node{node})
	if err == nil {
		t.Fatal("expected the AS error to be returned")
	}
	if !strings.Contains(err.Error(), asErr.Error()) {
		t.Errorf("error = %v, want it to wrap %v", err, asErr)
	}
	if !reflect.DeepEqual(fake.deleteIDs, []string{"inst-0"}) {
		t.Errorf("DeleteScalingInstances ids = %v, want [inst-0]", fake.deleteIDs)
	}
}
