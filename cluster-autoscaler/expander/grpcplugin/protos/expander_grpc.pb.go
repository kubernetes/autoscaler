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

package protos

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ExpanderClient is the client API for Expander service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExpanderClient interface {
	BestOptions(ctx context.Context, in *BestOptionsRequest, opts ...grpc.CallOption) (*BestOptionsResponse, error)
}

type expanderClient struct {
	cc grpc.ClientConnInterface
}

func NewExpanderClient(cc grpc.ClientConnInterface) ExpanderClient {
	return &expanderClient{cc}
}

func (c *expanderClient) BestOptions(ctx context.Context, in *BestOptionsRequest, opts ...grpc.CallOption) (*BestOptionsResponse, error) {
	out := new(BestOptionsResponse)
	err := c.cc.Invoke(ctx, "/grpcplugin.Expander/BestOptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExpanderServer is the server API for Expander service.
// All implementations should embed UnimplementedExpanderServer
// for forward compatibility
type ExpanderServer interface {
	BestOptions(context.Context, *BestOptionsRequest) (*BestOptionsResponse, error)
}

// UnimplementedExpanderServer should be embedded to have forward compatible implementations.
type UnimplementedExpanderServer struct {
}

func (UnimplementedExpanderServer) BestOptions(context.Context, *BestOptionsRequest) (*BestOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BestOptions not implemented")
}

// UnsafeExpanderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExpanderServer will
// result in compilation errors.
type UnsafeExpanderServer interface {
	mustEmbedUnimplementedExpanderServer()
}

func RegisterExpanderServer(s grpc.ServiceRegistrar, srv ExpanderServer) {
	s.RegisterService(&Expander_ServiceDesc, srv)
}

func _Expander_BestOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BestOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExpanderServer).BestOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcplugin.Expander/BestOptions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExpanderServer).BestOptions(ctx, req.(*BestOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Expander_ServiceDesc is the grpc.ServiceDesc for Expander service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Expander_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpcplugin.Expander",
	HandlerType: (*ExpanderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BestOptions",
			Handler:    _Expander_BestOptions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "expander.proto",
}
