package corehandlers

import (
	"net/http"
	"os"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/internal/sdktesting"
)

func TestAddHostExecEnvUserAgentHander(t *testing.T) {
	cases := []struct {
		ExecEnv string
		Expect  string
	}{
		{ExecEnv: "Lambda", Expect: execEnvUAKey + "/Lambda"},
		{ExecEnv: "", Expect: ""},
		{ExecEnv: "someThingCool", Expect: execEnvUAKey + "/someThingCool"},
	}

	for i, c := range cases {
		sdktesting.StashEnv()
		os.Setenv(execEnvVar, c.ExecEnv)

		req := &request.Request{
			HTTPRequest: &http.Request{
				Header: http.Header{},
			},
		}
		AddHostExecEnvUserAgentHander.Fn(req)

		if err := req.Error; err != nil {
			t.Fatalf("%d, expect no error, got %v", i, err)
		}

		if e, a := c.Expect, req.HTTPRequest.Header.Get("User-Agent"); e != a {
			t.Errorf("%d, expect %v user agent, got %v", i, e, a)
		}
	}
}
