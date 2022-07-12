//go:build go1.7 && integration
// +build go1.7,integration

package lexmodelsv2

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/awstesting/integration"
)

func TestInteg_ListBots(t *testing.T) {
	sess := integration.SessionWithDefaultRegion("us-west-2")

	client := New(sess)

	_, err := client.ListBots(&ListBotsInput{})
	if err != nil {
		t.Fatalf("expect API call, got %v", err)
	}
}
