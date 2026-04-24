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

// Package sdk provides a minimal Nebius API client for the cluster-autoscaler.
// It uses raw gRPC with hand-written protobuf wire encoding to avoid adding
// the Nebius SDK (github.com/nebius/gosdk) as a dependency. Only the proto
// fields actually consumed by the autoscaler are decoded; unknown fields are
// silently skipped so the codec tolerates API additions.
package sdk

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"
)

// Proto field numbers are derived from the Nebius API proto definitions at
// https://github.com/nebius/api. Only fields used by the autoscaler are listed.
//
// nebius.common.v1.ResourceMetadata:
//   id=1, parent_id=2, name=3, resource_version=4, labels=7
//
// nebius.mk8s.v1.NodeGroup:
//   metadata=1, spec=2, status=3
//
// nebius.mk8s.v1.NodeGroupSpec:
//   version=1, fixed_node_count=2(oneof), template=3, strategy=4, autoscaling=5(oneof), auto_repair=6
//
// nebius.mk8s.v1.NodeGroupAutoscalingSpec:
//   min_node_count=1, max_node_count=2
//
// nebius.mk8s.v1.NodeGroupStatus:
//   target_node_count=3
//
// nebius.mk8s.v1.ListNodeGroupsRequest:
//   parent_id=1, page_size=2, page_token=3
//
// nebius.mk8s.v1.ListNodeGroupsResponse:
//   items=1, next_page_token=2
//
// nebius.mk8s.v1.GetNodeGroupRequest:
//   id=1
//
// nebius.mk8s.v1.UpdateNodeGroupRequest:
//   metadata=1, spec=2
//
// nebius.compute.v1.Instance:
//   metadata=1
//
// nebius.compute.v1.ListInstancesRequest:
//   parent_id=1, page_size=2, page_token=3
//
// nebius.compute.v1.ListInstancesResponse:
//   items=1, next_page_token=2
//
// nebius.compute.v1.DeleteInstanceRequest:
//   id=1

// nebiusCodec implements grpc encoding.Codec for the local Nebius types.
// It registers with content-type "proto" so that gRPC servers accept it
// just like standard protobuf.
type nebiusCodec struct{}

func (nebiusCodec) Name() string { return "proto" }

func (nebiusCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case *ListNodeGroupsRequest:
		return marshalListNodeGroupsRequest(m)
	case *GetNodeGroupRequest:
		return marshalGetNodeGroupRequest(m)
	case *UpdateNodeGroupRequest:
		return marshalUpdateNodeGroupRequest(m)
	case *ListInstancesRequest:
		return marshalListInstancesRequest(m)
	case *DeleteInstanceRequest:
		return marshalDeleteInstanceRequest(m)
	// Response types: return empty for gRPC internal bookkeeping.
	case *ListNodeGroupsResponse:
		return nil, nil
	case *ListInstancesResponse:
		return nil, nil
	case *NodeGroup:
		return nil, nil
	case *rawBytes:
		return nil, nil
	default:
		return nil, fmt.Errorf("nebiusCodec: unsupported marshal type %T", v)
	}
}

func (nebiusCodec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case *ListNodeGroupsResponse:
		return unmarshalListNodeGroupsResponse(data, m)
	case *ListInstancesResponse:
		return unmarshalListInstancesResponse(data, m)
	case *NodeGroup:
		return unmarshalNodeGroup(data, m)
	// Request types: no-op unmarshal.
	case *ListNodeGroupsRequest:
		return nil
	case *GetNodeGroupRequest:
		return nil
	case *UpdateNodeGroupRequest:
		return nil
	case *ListInstancesRequest:
		return nil
	case *DeleteInstanceRequest:
		return nil
	case *rawBytes:
		m.data = append([]byte(nil), data...)
		return nil
	default:
		return fmt.Errorf("nebiusCodec: unsupported unmarshal type %T", v)
	}
}

// --- Marshalers ---

func marshalListNodeGroupsRequest(m *ListNodeGroupsRequest) ([]byte, error) {
	var b []byte
	b = appendString(b, 1, m.ParentID)
	b = appendString(b, 3, m.PageToken)
	return b, nil
}

func marshalGetNodeGroupRequest(m *GetNodeGroupRequest) ([]byte, error) {
	var b []byte
	b = appendString(b, 1, m.ID)
	return b, nil
}

func marshalUpdateNodeGroupRequest(m *UpdateNodeGroupRequest) ([]byte, error) {
	var b []byte
	if m.Metadata != nil {
		metaBytes := marshalResourceMetadata(m.Metadata)
		b = appendBytes(b, 1, metaBytes)
	}
	if m.Spec != nil {
		specBytes := marshalNodeGroupSpec(m.Spec)
		b = appendBytes(b, 2, specBytes)
	}
	return b, nil
}

func marshalListInstancesRequest(m *ListInstancesRequest) ([]byte, error) {
	var b []byte
	b = appendString(b, 1, m.ParentID)
	b = appendString(b, 3, m.PageToken)
	return b, nil
}

func marshalDeleteInstanceRequest(m *DeleteInstanceRequest) ([]byte, error) {
	var b []byte
	b = appendString(b, 1, m.ID)
	return b, nil
}

func marshalResourceMetadata(m *ResourceMetadata) []byte {
	// If we have the raw bytes from a previous decode, use them for round-trip fidelity.
	if m.raw != nil {
		return m.raw
	}
	var b []byte
	b = appendString(b, 1, m.ID)
	b = appendString(b, 2, m.ParentID)
	b = appendString(b, 3, m.Name)
	if m.ResourceVersion != 0 {
		b = appendVarint(b, 4, uint64(m.ResourceVersion))
	}
	for k, v := range m.Labels {
		entry := appendString(nil, 1, k)
		entry = appendString(entry, 2, v)
		b = appendBytes(b, 7, entry)
	}
	return b
}

func marshalNodeGroupSpec(m *NodeGroupSpec) []byte {
	var b []byte
	b = appendString(b, 1, m.Version)
	if m.FixedNodeCount != nil {
		b = appendVarint(b, 2, uint64(*m.FixedNodeCount))
	}
	// Preserve opaque template (field 3) for round-trip.
	if m.templateRaw != nil {
		b = appendBytes(b, 3, m.templateRaw)
	}
	// Preserve opaque strategy (field 4) for round-trip.
	if m.strategyRaw != nil {
		b = appendBytes(b, 4, m.strategyRaw)
	}
	if m.Autoscaling != nil {
		asBytes := marshalAutoscaling(m.Autoscaling)
		b = appendBytes(b, 5, asBytes)
	}
	// Preserve opaque auto_repair (field 6) for round-trip.
	if m.autoRepairRaw != nil {
		b = appendBytes(b, 6, m.autoRepairRaw)
	}
	return b
}

func marshalAutoscaling(m *NodeGroupAutoscalingSpec) []byte {
	var b []byte
	if m.MinNodeCount != 0 {
		b = appendVarint(b, 1, uint64(m.MinNodeCount))
	}
	if m.MaxNodeCount != 0 {
		b = appendVarint(b, 2, uint64(m.MaxNodeCount))
	}
	return b
}

// --- Unmarshalers ---

func unmarshalListNodeGroupsResponse(data []byte, m *ListNodeGroupsResponse) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag at offset %d", len(data))
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType: // items
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for field 1")
			}
			ng := &NodeGroup{}
			if err := unmarshalNodeGroup(v, ng); err != nil {
				return err
			}
			m.Items = append(m.Items, ng)
			data = data[n:]
		case num == 2 && wtyp == protowire.BytesType: // next_page_token
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string for field 2")
			}
			m.NextPageToken = v
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalListInstancesResponse(data []byte, m *ListInstancesResponse) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType: // items
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for field 1")
			}
			inst := &Instance{}
			if err := unmarshalInstance(v, inst); err != nil {
				return err
			}
			m.Items = append(m.Items, inst)
			data = data[n:]
		case num == 2 && wtyp == protowire.BytesType: // next_page_token
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string for field 2")
			}
			m.NextPageToken = v
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalNodeGroup(data []byte, m *NodeGroup) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType: // metadata
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for metadata")
			}
			m.Metadata = &ResourceMetadata{}
			if err := unmarshalResourceMetadata(v, m.Metadata); err != nil {
				return err
			}
			data = data[n:]
		case num == 2 && wtyp == protowire.BytesType: // spec
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for spec")
			}
			m.Spec = &NodeGroupSpec{}
			if err := unmarshalNodeGroupSpec(v, m.Spec); err != nil {
				return err
			}
			data = data[n:]
		case num == 3 && wtyp == protowire.BytesType: // status
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for status")
			}
			m.Status = &NodeGroupStatus{}
			if err := unmarshalNodeGroupStatus(v, m.Status); err != nil {
				return err
			}
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalResourceMetadata(data []byte, m *ResourceMetadata) error {
	// Save raw bytes for round-trip during update requests.
	m.raw = append([]byte(nil), data...)
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType: // id
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string for id")
			}
			m.ID = v
			data = data[n:]
		case num == 2 && wtyp == protowire.BytesType: // parent_id
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string for parent_id")
			}
			m.ParentID = v
			data = data[n:]
		case num == 3 && wtyp == protowire.BytesType: // name
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string for name")
			}
			m.Name = v
			data = data[n:]
		case num == 4 && wtyp == protowire.VarintType: // resource_version
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return fmt.Errorf("invalid varint for resource_version")
			}
			m.ResourceVersion = int64(v)
			data = data[n:]
		case num == 7 && wtyp == protowire.BytesType: // labels (map)
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for labels")
			}
			k, val, err := unmarshalMapEntry(v)
			if err != nil {
				return err
			}
			if m.Labels == nil {
				m.Labels = make(map[string]string)
			}
			m.Labels[k] = val
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalNodeGroupSpec(data []byte, m *NodeGroupSpec) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType: // version
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string for version")
			}
			m.Version = v
			data = data[n:]
		case num == 2 && wtyp == protowire.VarintType: // fixed_node_count (oneof)
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return fmt.Errorf("invalid varint for fixed_node_count")
			}
			count := int64(v)
			m.FixedNodeCount = &count
			data = data[n:]
		case num == 3 && wtyp == protowire.BytesType: // template (opaque)
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for template")
			}
			m.templateRaw = append([]byte(nil), v...)
			data = data[n:]
		case num == 4 && wtyp == protowire.BytesType: // strategy (opaque)
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for strategy")
			}
			m.strategyRaw = append([]byte(nil), v...)
			data = data[n:]
		case num == 5 && wtyp == protowire.BytesType: // autoscaling (oneof)
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for autoscaling")
			}
			m.Autoscaling = &NodeGroupAutoscalingSpec{}
			if err := unmarshalAutoscaling(v, m.Autoscaling); err != nil {
				return err
			}
			data = data[n:]
		case num == 6 && wtyp == protowire.BytesType: // auto_repair (opaque)
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for auto_repair")
			}
			m.autoRepairRaw = append([]byte(nil), v...)
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalAutoscaling(data []byte, m *NodeGroupAutoscalingSpec) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.VarintType: // min_node_count
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return fmt.Errorf("invalid varint for min_node_count")
			}
			m.MinNodeCount = int64(v)
			data = data[n:]
		case num == 2 && wtyp == protowire.VarintType: // max_node_count
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return fmt.Errorf("invalid varint for max_node_count")
			}
			m.MaxNodeCount = int64(v)
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalNodeGroupStatus(data []byte, m *NodeGroupStatus) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 3 && wtyp == protowire.VarintType: // target_node_count
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return fmt.Errorf("invalid varint for target_node_count")
			}
			m.TargetNodeCount = int64(v)
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalInstance(data []byte, m *Instance) error {
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch {
		case num == 1 && wtyp == protowire.BytesType: // metadata
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return fmt.Errorf("invalid bytes for metadata")
			}
			m.Metadata = &ResourceMetadata{}
			if err := unmarshalResourceMetadata(v, m.Metadata); err != nil {
				return err
			}
			data = data[n:]
		default:
			n, err := skipField(wtyp, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

// --- Map entry helpers ---

func unmarshalMapEntry(data []byte) (string, string, error) {
	var key, value string
	for len(data) > 0 {
		num, wtyp, n := protowire.ConsumeTag(data)
		if n < 0 {
			return "", "", fmt.Errorf("invalid map entry tag")
		}
		data = data[n:]
		if wtyp != protowire.BytesType {
			nn, err := skipField(wtyp, data)
			if err != nil {
				return "", "", err
			}
			data = data[nn:]
			continue
		}
		v, n := protowire.ConsumeString(data)
		if n < 0 {
			return "", "", fmt.Errorf("invalid map entry string")
		}
		switch num {
		case 1:
			key = v
		case 2:
			value = v
		}
		data = data[n:]
	}
	return key, value, nil
}

// --- Wire encoding helpers ---

func appendString(b []byte, num protowire.Number, s string) []byte {
	if s == "" {
		return b
	}
	b = protowire.AppendTag(b, num, protowire.BytesType)
	b = protowire.AppendString(b, s)
	return b
}

func appendVarint(b []byte, num protowire.Number, v uint64) []byte {
	b = protowire.AppendTag(b, num, protowire.VarintType)
	b = protowire.AppendVarint(b, v)
	return b
}

func appendBytes(b []byte, num protowire.Number, data []byte) []byte {
	if len(data) == 0 {
		return b
	}
	b = protowire.AppendTag(b, num, protowire.BytesType)
	b = protowire.AppendBytes(b, data)
	return b
}

func skipField(wtyp protowire.Type, data []byte) (int, error) {
	var n int
	switch wtyp {
	case protowire.VarintType:
		_, n = protowire.ConsumeVarint(data)
	case protowire.Fixed32Type:
		_, n = protowire.ConsumeFixed32(data)
	case protowire.Fixed64Type:
		_, n = protowire.ConsumeFixed64(data)
	case protowire.BytesType:
		_, n = protowire.ConsumeBytes(data)
	case protowire.StartGroupType:
		_, n = protowire.ConsumeGroup(protowire.Number(0), data)
	default:
		return 0, fmt.Errorf("unknown wire type %d", wtyp)
	}
	if n < 0 {
		return 0, fmt.Errorf("failed to skip field with wire type %d", wtyp)
	}
	return n, nil
}
