package iotdataplane_test

import (
	"fmt"
	"log"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/session"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/iot"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/iotdataplane"
)

func ExampleIoTDataPlane_describeEndpoint() {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		log.Fatal("Failed to create aws session", err)
	}

	// we need to use an IoT control plane client to get an endpoint address
	ctrlSvc := iot.New(sess)
	descResp, err := ctrlSvc.DescribeEndpoint(&iot.DescribeEndpointInput{})
	if err != nil {
		log.Fatal("failed to get dataplane endpoint", err)
	}

	// create a IoT data plane client using the endpoint address we retrieved
	dataSvc := iotdataplane.New(sess, &aws.Config{
		Endpoint: descResp.EndpointAddress,
	})
	output, err := dataSvc.GetThingShadow(&iotdataplane.GetThingShadowInput{
		// specify a ThingName
		ThingName: aws.String("fake-thing"),
	})
	// prints the string representation of GetThingShadowOutput
	fmt.Println(output.GoString())
}
