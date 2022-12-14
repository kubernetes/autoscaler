package polly

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/private/protocol/rest"
)

func init() {
	initRequest = func(r *request.Request) {
		if r.Operation.Name == opSynthesizeSpeech {
			r.Operation.BeforePresignFn = restGETPresignStrategy
		}
	}
}

// restGETPresignStrategy will prepare the request from a POST to a GET request.
// Enabling the presigner to sign the request as a GET.
func restGETPresignStrategy(r *request.Request) error {
	r.Handlers.Build.Clear()
	r.Handlers.Build.PushBack(rest.BuildAsGET)
	r.Operation.HTTPMethod = "GET"
	r.HTTPRequest.Method = "GET"
	return nil
}
