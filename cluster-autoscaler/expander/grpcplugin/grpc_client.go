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
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	var dialOpt grpc.DialOption

	// if no Cert file specified, use insecure
	if expanderCert == "" {
		dialOpt = grpc.WithInsecure()
	} else {
		creds, err := credentials.NewClientTLSFromFile(expanderCert, "")
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
			return nil
		}
		dialOpt = grpc.WithTransportCredentials(creds)
	}
	klog.V(2).Info("Dialing ", expanderUrl, " dialopt: ", dialOpt)
	conn, err := grpc.Dial(expanderUrl, dialOpt)
	if err != nil {
		log.Fatalf("fail to dial server: %v", err)
		return nil
	}
	return protos.NewExpanderClient(conn)
}

func (g *grpcclientstrategy) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) []expander.Option {
	if g.grpcClient == nil {
		log.Fatalf("Incorrect gRPC client config, filtering no options")
		return expansionOptions
	}

	// Transform inputs to gRPC inputs
	nodeGroupIDOptionMap := make(map[string]expander.Option)
	grpcOptionsSlice := []*protos.Option{}
	populateOptionsForGRPC(expansionOptions, nodeGroupIDOptionMap, &grpcOptionsSlice)
	grpcNodeInfoMap := make(map[string]*v1.Node)
	populateNodeInfoForGRPC(nodeInfo, grpcNodeInfoMap)

	// call gRPC server to get BestOption
	klog.V(2).Info("GPRC call of best options to server with ", len(nodeGroupIDOptionMap), " options")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bestOptionsResponse, err := g.grpcClient.BestOptions(ctx, &protos.BestOptionsRequest{Options: grpcOptionsSlice, NodeInfoMap: grpcNodeInfoMap})
	if err != nil {
		klog.V(2).Info("GRPC call timed out, no options filtered")
		return expansionOptions
	}

	if bestOptionsResponse == nil || bestOptionsResponse.Options == nil {
		klog.V(2).Info("GRPC returned nil bestOptions, no options filtered")
		return expansionOptions
	}
	// Transform back options slice
	options := transformAndSanitizeOptionsFromGRPC(bestOptionsResponse.Options, nodeGroupIDOptionMap)
	if options == nil {
		klog.V(2).Info("Unable to sanitize GPRC returned bestOptions, no options filtered")
		return expansionOptions
	}
	return options
}

// populateOptionsForGRPC creates a map of nodegroup ID and options, as well as a slice of Options objects for the gRPC call
func populateOptionsForGRPC(expansionOptions []expander.Option, nodeGroupIDOptionMap map[string]expander.Option, grpcOptionsSlice *[]*protos.Option) {
	for _, option := range expansionOptions {
		nodeGroupIDOptionMap[option.NodeGroup.Id()] = option
		*grpcOptionsSlice = append(*grpcOptionsSlice, newOptionMessage(option.NodeGroup.Id(), int32(option.NodeCount), option.Debug, option.Pods))
	}
}

// populateNodeInfoForGRPC modifies the nodeInfo object, and replaces it with the v1.Node to pass through grpc
func populateNodeInfoForGRPC(nodeInfos map[string]*schedulerframework.NodeInfo, grpcNodeInfoMap map[string]*v1.Node) {
	for nodeId, nodeInfo := range nodeInfos {
		grpcNodeInfoMap[nodeId] = nodeInfo.Node()
	}
}

func transformAndSanitizeOptionsFromGRPC(bestOptionsResponseOptions []*protos.Option, nodeGroupIDOptionMap map[string]expander.Option) []expander.Option {
	var options []expander.Option
	for _, option := range bestOptionsResponseOptions {
		if option == nil {
			klog.Errorf("gRPC server returned nil Option")
			return nil
		}
		if _, ok := nodeGroupIDOptionMap[option.NodeGroupId]; ok {
			options = append(options, nodeGroupIDOptionMap[option.NodeGroupId])
		} else {
			klog.Errorf("gRPC server returned invalid nodeGroup ID: ", option.NodeGroupId)
			return nil
		}
	}
	return options
}

func newOptionMessage(nodeGroupId string, nodeCount int32, debug string, pods []*v1.Pod) *protos.Option {
	return &protos.Option{NodeGroupId: nodeGroupId, NodeCount: nodeCount, Debug: debug, Pod: pods}
}
