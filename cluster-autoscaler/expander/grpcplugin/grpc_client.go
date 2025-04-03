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

package grpcplugin

import (
	"context"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/grpcplugin/protos"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	gRPCTimeout        = 5 * time.Second
	gRPCMaxRecvMsgSize = 128 << 20
)

type grpcclientstrategy struct {
	grpcClient protos.ExpanderClient
}

// NewFilter returns an expansion filter that creates a gRPC client, and calls out to a gRPC server
func NewFilter(expanderCert string, expanderUrl string) expander.Filter {
	client := createGRPCClient(expanderCert, expanderUrl)
	if client == nil {
		return &grpcclientstrategy{grpcClient: nil}
	}
	return &grpcclientstrategy{grpcClient: client}
}

func createGRPCClient(expanderCert string, expanderUrl string) protos.ExpanderClient {
	if expanderCert == "" {
		log.Fatalf("GRPC Expander Cert not specified, insecure connections not allowed")
		return nil
	}
	creds, err := credentials.NewClientTLSFromFile(expanderCert, "")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
		return nil
	}
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(gRPCMaxRecvMsgSize)),
	}
	klog.V(2).Infof("Dialing: %s with dialopt: %v", expanderUrl, dialOpts)
	conn, err := grpc.Dial(expanderUrl, dialOpts...)
	if err != nil {
		log.Fatalf("Fail to dial server: %v", err)
		return nil
	}
	return protos.NewExpanderClient(conn)
}

func (g *grpcclientstrategy) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) []expander.Option {
	if g.grpcClient == nil {
		klog.Errorf("Incorrect gRPC client config, filtering no options")
		return expansionOptions
	}

	// Transform inputs to gRPC inputs
	grpcOptionsSlice, nodeGroupIDOptionMap := populateOptionsForGRPC(expansionOptions)
	grpcNodeMap := populateNodeInfoForGRPC(nodeInfo)

	// call gRPC server to get BestOption
	klog.V(2).Infof("GPRC call of best options to server with %v options", len(nodeGroupIDOptionMap))
	ctx, cancel := context.WithTimeout(context.Background(), gRPCTimeout)
	defer cancel()
	bestOptionsResponse, err := g.grpcClient.BestOptions(ctx, &protos.BestOptionsRequest{Options: grpcOptionsSlice, NodeMap: grpcNodeMap})
	if err != nil {
		klog.V(4).Infof("GRPC call failed, no options filtered: %v", err)
		return expansionOptions
	}

	if bestOptionsResponse == nil || len(bestOptionsResponse.Options) == 0 {
		klog.V(4).Info("GRPC returned nil bestOptions")
		return nil
	}
	// Transform back options slice
	options := transformAndSanitizeOptionsFromGRPC(bestOptionsResponse.Options, nodeGroupIDOptionMap)
	if options == nil {
		klog.V(4).Info("Unable to sanitize GPRC returned bestOptions, no options filtered")
		return expansionOptions
	}
	return options
}

// populateOptionsForGRPC creates a map of nodegroup ID and options, as well as a slice of Options objects for the gRPC call
func populateOptionsForGRPC(expansionOptions []expander.Option) ([]*protos.Option, map[string]expander.Option) {
	grpcOptionsSlice := []*protos.Option{}
	nodeGroupIDOptionMap := make(map[string]expander.Option)
	for _, option := range expansionOptions {
		nodeGroupIDOptionMap[option.NodeGroup.Id()] = option
		grpcOptionsSlice = append(grpcOptionsSlice, newOptionMessage(option.NodeGroup.Id(), int32(option.NodeCount), option.Debug, option.Pods))
	}
	return grpcOptionsSlice, nodeGroupIDOptionMap
}

// populateNodeInfoForGRPC looks at the corresponding v1.Node object per NodeInfo object, and populates the grpcNodeInfoMap with these to pass over grpc
func populateNodeInfoForGRPC(nodeInfos map[string]*framework.NodeInfo) map[string]*v1.Node {
	grpcNodeInfoMap := make(map[string]*v1.Node)
	for nodeId, nodeInfo := range nodeInfos {
		grpcNodeInfoMap[nodeId] = nodeInfo.Node()
	}
	return grpcNodeInfoMap
}

func transformAndSanitizeOptionsFromGRPC(bestOptionsResponseOptions []*protos.Option, nodeGroupIDOptionMap map[string]expander.Option) []expander.Option {
	var options []expander.Option
	for _, option := range bestOptionsResponseOptions {
		if option == nil {
			klog.Error("GRPC server returned nil Option")
			continue
		}
		if _, ok := nodeGroupIDOptionMap[option.NodeGroupId]; ok {
			options = append(options, nodeGroupIDOptionMap[option.NodeGroupId])
		} else {
			klog.Errorf("GRPC server returned invalid nodeGroup ID: %s", option.NodeGroupId)
			continue
		}
	}
	return options
}

func newOptionMessage(nodeGroupId string, nodeCount int32, debug string, pods []*v1.Pod) *protos.Option {
	return &protos.Option{NodeGroupId: nodeGroupId, NodeCount: nodeCount, Debug: debug, Pod: pods}
}
