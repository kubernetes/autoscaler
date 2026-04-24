/*
Copyright The Kubernetes Authors.

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

package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

// TestCodecRoundTrip_ListNodeGroupsRequest verifies marshal → unmarshal fidelity.
func TestCodecRoundTrip_ListNodeGroupsRequest(t *testing.T) {
	codec := nebiusCodec{}

	req := &ListNodeGroupsRequest{
		ParentID:  "cluster-abc",
		PageToken: "page2",
	}

	data, err := codec.Marshal(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Verify wire format manually: field 1 = "cluster-abc", field 3 = "page2"
	got := &ListNodeGroupsRequest{}
	remaining := data
	for len(remaining) > 0 {
		num, wtyp, n := protowire.ConsumeTag(remaining)
		require.Greater(t, n, 0)
		remaining = remaining[n:]
		require.Equal(t, protowire.BytesType, wtyp)
		v, n := protowire.ConsumeString(remaining)
		require.Greater(t, n, 0)
		remaining = remaining[n:]
		switch num {
		case 1:
			got.ParentID = v
		case 3:
			got.PageToken = v
		}
	}
	assert.Equal(t, req.ParentID, got.ParentID)
	assert.Equal(t, req.PageToken, got.PageToken)
}

// TestCodecRoundTrip_NodeGroup verifies unmarshal of a hand-encoded NodeGroup.
func TestCodecRoundTrip_NodeGroup(t *testing.T) {
	codec := nebiusCodec{}

	// Build a NodeGroup in wire format:
	// metadata { id="ng-1" (field 1), name="my-group" (field 3), labels { "env":"prod" } (field 7) }
	// spec { autoscaling { min=2 (field 1), max=10 (field 2) } (field 5), version="1.31" (field 1) }
	// status { target_node_count=5 (field 3) }

	// Build metadata
	var metaBytes []byte
	metaBytes = protowire.AppendTag(metaBytes, 1, protowire.BytesType)
	metaBytes = protowire.AppendString(metaBytes, "ng-1")
	metaBytes = protowire.AppendTag(metaBytes, 3, protowire.BytesType)
	metaBytes = protowire.AppendString(metaBytes, "my-group")
	// labels map entry: key=1 "env", value=2 "prod"
	var mapEntry []byte
	mapEntry = protowire.AppendTag(mapEntry, 1, protowire.BytesType)
	mapEntry = protowire.AppendString(mapEntry, "env")
	mapEntry = protowire.AppendTag(mapEntry, 2, protowire.BytesType)
	mapEntry = protowire.AppendString(mapEntry, "prod")
	metaBytes = protowire.AppendTag(metaBytes, 7, protowire.BytesType)
	metaBytes = protowire.AppendBytes(metaBytes, mapEntry)

	// Build autoscaling spec
	var asBytes []byte
	asBytes = protowire.AppendTag(asBytes, 1, protowire.VarintType)
	asBytes = protowire.AppendVarint(asBytes, 2)
	asBytes = protowire.AppendTag(asBytes, 2, protowire.VarintType)
	asBytes = protowire.AppendVarint(asBytes, 10)

	// Build spec
	var specBytes []byte
	specBytes = protowire.AppendTag(specBytes, 1, protowire.BytesType)
	specBytes = protowire.AppendString(specBytes, "1.31")
	specBytes = protowire.AppendTag(specBytes, 5, protowire.BytesType)
	specBytes = protowire.AppendBytes(specBytes, asBytes)

	// Build status
	var statusBytes []byte
	statusBytes = protowire.AppendTag(statusBytes, 3, protowire.VarintType)
	statusBytes = protowire.AppendVarint(statusBytes, 5)

	// Build full NodeGroup
	var ngBytes []byte
	ngBytes = protowire.AppendTag(ngBytes, 1, protowire.BytesType)
	ngBytes = protowire.AppendBytes(ngBytes, metaBytes)
	ngBytes = protowire.AppendTag(ngBytes, 2, protowire.BytesType)
	ngBytes = protowire.AppendBytes(ngBytes, specBytes)
	ngBytes = protowire.AppendTag(ngBytes, 3, protowire.BytesType)
	ngBytes = protowire.AppendBytes(ngBytes, statusBytes)

	ng := &NodeGroup{}
	err := codec.Unmarshal(ngBytes, ng)
	require.NoError(t, err)

	assert.Equal(t, "ng-1", ng.Metadata.ID)
	assert.Equal(t, "my-group", ng.Metadata.Name)
	assert.Equal(t, map[string]string{"env": "prod"}, ng.Metadata.Labels)
	assert.Equal(t, "1.31", ng.Spec.Version)
	require.NotNil(t, ng.Spec.Autoscaling)
	assert.Equal(t, int64(2), ng.Spec.Autoscaling.MinNodeCount)
	assert.Equal(t, int64(10), ng.Spec.Autoscaling.MaxNodeCount)
	assert.Nil(t, ng.Spec.FixedNodeCount)
	assert.Equal(t, int64(5), ng.Status.TargetNodeCount)
}

// TestCodecRoundTrip_NodeGroupFixedCount verifies the fixed_node_count oneof path.
func TestCodecRoundTrip_NodeGroupFixedCount(t *testing.T) {
	codec := nebiusCodec{}

	// Build spec with fixed_node_count=3 (field 2, varint)
	var specBytes []byte
	specBytes = protowire.AppendTag(specBytes, 2, protowire.VarintType)
	specBytes = protowire.AppendVarint(specBytes, 3)

	var ngBytes []byte
	ngBytes = protowire.AppendTag(ngBytes, 2, protowire.BytesType)
	ngBytes = protowire.AppendBytes(ngBytes, specBytes)

	ng := &NodeGroup{}
	err := codec.Unmarshal(ngBytes, ng)
	require.NoError(t, err)

	require.NotNil(t, ng.Spec)
	require.NotNil(t, ng.Spec.FixedNodeCount)
	assert.Equal(t, int64(3), *ng.Spec.FixedNodeCount)
	assert.Nil(t, ng.Spec.Autoscaling)
}

// TestCodecRoundTrip_UpdateRequest verifies marshal of UpdateNodeGroupRequest
// including metadata round-trip and fixed node count.
func TestCodecRoundTrip_UpdateRequest(t *testing.T) {
	codec := nebiusCodec{}

	fixedCount := int64(7)
	req := &UpdateNodeGroupRequest{
		Metadata: &ResourceMetadata{
			ID:   "ng-1",
			Name: "my-group",
		},
		Spec: &NodeGroupSpec{
			Version:        "1.31",
			FixedNodeCount: &fixedCount,
		},
	}

	data, err := codec.Marshal(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Parse the wire format to verify structure.
	remaining := data
	var gotMetadata, gotSpec []byte
	for len(remaining) > 0 {
		num, wtyp, n := protowire.ConsumeTag(remaining)
		require.Greater(t, n, 0)
		remaining = remaining[n:]
		require.Equal(t, protowire.BytesType, wtyp)
		v, n := protowire.ConsumeBytes(remaining)
		require.Greater(t, n, 0)
		remaining = remaining[n:]
		switch num {
		case 1:
			gotMetadata = v
		case 2:
			gotSpec = v
		}
	}

	require.NotEmpty(t, gotMetadata, "metadata should be present")
	require.NotEmpty(t, gotSpec, "spec should be present")

	// Verify spec contains fixed_node_count=7 (field 2, varint)
	specRemaining := gotSpec
	var foundVersion string
	var foundFixed int64
	for len(specRemaining) > 0 {
		num, wtyp, n := protowire.ConsumeTag(specRemaining)
		require.Greater(t, n, 0)
		specRemaining = specRemaining[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType:
			v, n := protowire.ConsumeString(specRemaining)
			require.Greater(t, n, 0)
			foundVersion = v
			specRemaining = specRemaining[n:]
		case num == 2 && wtyp == protowire.VarintType:
			v, n := protowire.ConsumeVarint(specRemaining)
			require.Greater(t, n, 0)
			foundFixed = int64(v)
			specRemaining = specRemaining[n:]
		default:
			t.Fatalf("unexpected field %d with wire type %d in spec", num, wtyp)
		}
	}
	assert.Equal(t, "1.31", foundVersion)
	assert.Equal(t, int64(7), foundFixed)
}

// TestCodecRoundTrip_ListNodeGroupsResponse verifies unmarshal of a response
// with multiple items and pagination.
func TestCodecRoundTrip_ListNodeGroupsResponse(t *testing.T) {
	codec := nebiusCodec{}

	// Build two minimal NodeGroups
	buildNG := func(id string) []byte {
		var metaBytes []byte
		metaBytes = protowire.AppendTag(metaBytes, 1, protowire.BytesType)
		metaBytes = protowire.AppendString(metaBytes, id)
		var ngBytes []byte
		ngBytes = protowire.AppendTag(ngBytes, 1, protowire.BytesType)
		ngBytes = protowire.AppendBytes(ngBytes, metaBytes)
		return ngBytes
	}

	var data []byte
	// item 1
	data = protowire.AppendTag(data, 1, protowire.BytesType)
	data = protowire.AppendBytes(data, buildNG("ng-1"))
	// item 2
	data = protowire.AppendTag(data, 1, protowire.BytesType)
	data = protowire.AppendBytes(data, buildNG("ng-2"))
	// next_page_token
	data = protowire.AppendTag(data, 2, protowire.BytesType)
	data = protowire.AppendString(data, "next-page")

	resp := &ListNodeGroupsResponse{}
	err := codec.Unmarshal(data, resp)
	require.NoError(t, err)

	require.Len(t, resp.Items, 2)
	assert.Equal(t, "ng-1", resp.Items[0].Metadata.ID)
	assert.Equal(t, "ng-2", resp.Items[1].Metadata.ID)
	assert.Equal(t, "next-page", resp.NextPageToken)
}

// TestCodecRoundTrip_ListInstancesResponse verifies instance unmarshal with labels.
func TestCodecRoundTrip_ListInstancesResponse(t *testing.T) {
	codec := nebiusCodec{}

	// Build an instance with metadata { id="inst-1", labels { "nebius.com/node-group-id": "ng-1" } }
	var mapEntry []byte
	mapEntry = protowire.AppendTag(mapEntry, 1, protowire.BytesType)
	mapEntry = protowire.AppendString(mapEntry, "nebius.com/node-group-id")
	mapEntry = protowire.AppendTag(mapEntry, 2, protowire.BytesType)
	mapEntry = protowire.AppendString(mapEntry, "ng-1")

	var metaBytes []byte
	metaBytes = protowire.AppendTag(metaBytes, 1, protowire.BytesType)
	metaBytes = protowire.AppendString(metaBytes, "inst-1")
	metaBytes = protowire.AppendTag(metaBytes, 7, protowire.BytesType)
	metaBytes = protowire.AppendBytes(metaBytes, mapEntry)

	var instBytes []byte
	instBytes = protowire.AppendTag(instBytes, 1, protowire.BytesType)
	instBytes = protowire.AppendBytes(instBytes, metaBytes)

	var data []byte
	data = protowire.AppendTag(data, 1, protowire.BytesType)
	data = protowire.AppendBytes(data, instBytes)

	resp := &ListInstancesResponse{}
	err := codec.Unmarshal(data, resp)
	require.NoError(t, err)

	require.Len(t, resp.Items, 1)
	assert.Equal(t, "inst-1", resp.Items[0].Metadata.ID)
	assert.Equal(t, "ng-1", resp.Items[0].Metadata.Labels["nebius.com/node-group-id"])
}

// TestCodecSkipsUnknownFields verifies the codec doesn't choke on extra fields.
func TestCodecSkipsUnknownFields(t *testing.T) {
	codec := nebiusCodec{}

	var data []byte
	// field 1 (id) - known
	data = protowire.AppendTag(data, 1, protowire.BytesType)
	data = protowire.AppendString(data, "ng-1")
	// field 99 (unknown varint)
	data = protowire.AppendTag(data, 99, protowire.VarintType)
	data = protowire.AppendVarint(data, 42)
	// field 100 (unknown bytes)
	data = protowire.AppendTag(data, 100, protowire.BytesType)
	data = protowire.AppendString(data, "ignored")

	req := &GetNodeGroupRequest{}
	// GetNodeGroupRequest only has field 1 (id), but we wrapped it in a
	// NodeGroup unmarshal context to test skip behavior.
	ng := &NodeGroup{}
	// Wrap the data as metadata (field 1) inside a NodeGroup
	var ngData []byte
	ngData = protowire.AppendTag(ngData, 1, protowire.BytesType)
	ngData = protowire.AppendBytes(ngData, data)
	// Add an unknown field at NodeGroup level
	ngData = protowire.AppendTag(ngData, 50, protowire.VarintType)
	ngData = protowire.AppendVarint(ngData, 999)

	err := codec.Unmarshal(ngData, ng)
	require.NoError(t, err)
	assert.Equal(t, "ng-1", ng.Metadata.ID)

	_ = req // just verifying the codec doesn't panic
}

// TestCodecRoundTrip_MetadataPreservesRaw verifies that metadata round-trips
// faithfully through marshal → unmarshal → marshal.
func TestCodecRoundTrip_MetadataPreservesRaw(t *testing.T) {
	// Build metadata with resource_version (field 4) which we read but also
	// preserve in raw bytes for update round-trip.
	var metaBytes []byte
	metaBytes = protowire.AppendTag(metaBytes, 1, protowire.BytesType)
	metaBytes = protowire.AppendString(metaBytes, "ng-1")
	metaBytes = protowire.AppendTag(metaBytes, 4, protowire.VarintType)
	metaBytes = protowire.AppendVarint(metaBytes, 42)

	meta := &ResourceMetadata{}
	err := unmarshalResourceMetadata(metaBytes, meta)
	require.NoError(t, err)
	assert.Equal(t, "ng-1", meta.ID)
	assert.Equal(t, int64(42), meta.ResourceVersion)

	// Re-marshal should produce the same raw bytes (used in UpdateNodeGroupRequest).
	reEncoded := marshalResourceMetadata(meta)
	assert.Equal(t, metaBytes, reEncoded)
}
