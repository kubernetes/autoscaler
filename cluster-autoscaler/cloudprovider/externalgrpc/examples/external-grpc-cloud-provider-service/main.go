/*
Copyright 2022 The Kubernetes Authors.

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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	cloudBuilder "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/examples/external-grpc-cloud-provider-service/wrapper"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	kube_flag "k8s.io/component-base/cli/flag"
	klog "k8s.io/klog/v2"
)

// MultiStringFlag is a flag for passing multiple parameters using same flag
type MultiStringFlag []string

// String returns string representation of the node groups.
func (flag *MultiStringFlag) String() string {
	return "[" + strings.Join(*flag, " ") + "]"
}

// Set adds a new configuration.
func (flag *MultiStringFlag) Set(value string) error {
	*flag = append(*flag, value)
	return nil
}

func multiStringFlag(name string, usage string) *MultiStringFlag {
	value := new(MultiStringFlag)
	flag.Var(value, name, usage)
	return value
}

var (
	// flags needed by the external grpc provider service
	address = flag.String("address", ":8086", "The address to expose the grpc service.")
	keyCert = flag.String("key-cert", "", "The path to the certificate key file. Empty string for insecure communication.")
	cert    = flag.String("cert", "", "The path to the certificate file. Empty string for insecure communication.")
	cacert  = flag.String("ca-cert", "", "The path to the ca certificate file. Empty string for insecure communication.")

	// flags needed by the specific cloud provider
	cloudProviderFlag = flag.String("cloud-provider", cloudBuilder.DefaultCloudProvider,
		"Cloud provider type. Available values: ["+strings.Join(cloudBuilder.AvailableCloudProviders, ",")+"]")
	cloudConfig    = flag.String("cloud-config", "", "The path to the cloud provider configuration file.  Empty string for no configuration file.")
	clusterName    = flag.String("cluster-name", "", "Autoscaled cluster name, if available")
	nodeGroupsFlag = multiStringFlag(
		"nodes",
		"sets min,max size and other configuration data for a node group in a format accepted by cloud provider. Can be used multiple times. Format: <min>:<max>:<other...>")
	nodeGroupAutoDiscoveryFlag = multiStringFlag(
		"node-group-auto-discovery",
		"One or more definition(s) of node group auto-discovery. "+
			"A definition is expressed `<name of discoverer>:[<key>[=<value>]]`. "+
			"The `aws` and `gce` cloud providers are currently supported. AWS matches by ASG tags, e.g. `asg:tag=tagKey,anotherTagKey`. "+
			"GCE matches by IG name prefix, and requires you to specify min and max nodes per IG, e.g. `mig:namePrefix=pfx,min=0,max=10` "+
			"Can be used multiple times.")
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()

	var s *grpc.Server

	// tls config
	var serverOpt grpc.ServerOption
	if *keyCert == "" || *cert == "" || *cacert == "" {
		klog.V(1).Info("no cert specified, using insecure")
		s = grpc.NewServer()
	} else {

		certificate, err := tls.LoadX509KeyPair(*cert, *keyCert)
		if err != nil {
			klog.Fatalf("failed to read certificate files: %s", err)
		}
		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(*cacert)
		if err != nil {
			klog.Fatalf("failed to read client ca cert: %s", err)
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			klog.Fatal("failed to append client certs")
		}
		transportCreds := credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{certificate},
			ClientCAs:    certPool,
		})
		serverOpt = grpc.Creds(transportCreds)
		s = grpc.NewServer(serverOpt)
	}

	//cloud provider config
	autoscalingOptions := config.AutoscalingOptions{
		CloudProviderName:      *cloudProviderFlag,
		CloudConfig:            *cloudConfig,
		NodeGroupAutoDiscovery: *nodeGroupAutoDiscoveryFlag,
		NodeGroups:             *nodeGroupsFlag,
		ClusterName:            *clusterName,
		GCEOptions: config.GCEOptions{
			ConcurrentRefreshes: 1,
		},
		UserAgent: "user-agent",
	}
	cloudProvider := cloudBuilder.NewCloudProvider(autoscalingOptions)
	srv := wrapper.NewCloudProviderGrpcWrapper(cloudProvider)

	// listen
	lis, err := net.Listen("tcp", *address)
	if err != nil {
		klog.Fatalf("failed to listen: %s", err)
	}

	// serve
	protos.RegisterCloudProviderServer(s, srv)
	klog.V(1).Infof("Server ready at: %s\n", *address)
	if err := s.Serve(lis); err != nil {
		klog.Fatalf("failed to serve: %v", err)
	}

}
