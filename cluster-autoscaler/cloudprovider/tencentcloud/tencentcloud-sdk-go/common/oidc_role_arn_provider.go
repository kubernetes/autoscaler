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

package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/errors"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/profile"
)

type OIDCRoleArnProvider struct {
	region           string
	providerId       string
	webIdentityToken string
	roleArn          string
	roleSessionName  string
	durationSeconds  int64
}

type oidcStsRsp struct {
	Response struct {
		Credentials struct {
			Token        string `json:"Token"`
			TmpSecretId  string `json:"TmpSecretId"`
			TmpSecretKey string `json:"TmpSecretKey"`
		} `json:"Credentials"`
		ExpiredTime int       `json:"ExpiredTime"`
		Expiration  time.Time `json:"Expiration"`
		RequestId   string    `json:"RequestId"`
	} `json:"Response"`
}

func NewOIDCRoleArnProvider(region, providerId, webIdentityToken, roleArn, roleSessionName string, durationSeconds int64) *OIDCRoleArnProvider {
	return &OIDCRoleArnProvider{
		region:           region,
		providerId:       providerId,
		webIdentityToken: webIdentityToken,
		roleArn:          roleArn,
		roleSessionName:  roleSessionName,
		durationSeconds:  durationSeconds,
	}
}

// DefaultTkeOIDCRoleArnProvider returns a OIDCRoleArnProvider with some default options:
//  1. providerId will be environment var TKE_PROVIDER_ID
//  2. webIdentityToken will be the content of file specified by env TKE_WEB_IDENTITY_TOKEN_FILE
//  3. roleArn will be env TKE_ROLE_ARN
//  4. roleSessionName will be "tencentcloud-go-sdk-" + timestamp
//  5. durationSeconds will be 7200s
func DefaultTkeOIDCRoleArnProvider() (*OIDCRoleArnProvider, error) {
	reg := os.Getenv("TKE_REGION")
	if reg == "" {
		return nil, errors.New("env TKE_REGION not exist")
	}

	providerId := os.Getenv("TKE_PROVIDER_ID")
	if providerId == "" {
		return nil, errors.New("env TKE_PROVIDER_ID not exist")
	}

	tokenFile := os.Getenv("TKE_WEB_IDENTITY_TOKEN_FILE")
	if tokenFile == "" {
		return nil, errors.New("env TKE_WEB_IDENTITY_TOKEN_FILE not exist")
	}
	tokenBytes, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	roleArn := os.Getenv("TKE_ROLE_ARN")
	if roleArn == "" {
		return nil, errors.New("env TKE_ROLE_ARN not exist")
	}

	sessionName := defaultSessionName + strconv.FormatInt(time.Now().UnixNano()/1000, 10)

	return NewOIDCRoleArnProvider(reg, providerId, string(tokenBytes), roleArn, sessionName, defaultDurationSeconds), nil
}

func (r *OIDCRoleArnProvider) GetCredential() (CredentialIface, error) {
	const (
		service = "sts"
		version = "2018-08-13"
		action  = "AssumeRoleWithWebIdentity"
	)
	if r.durationSeconds > 43200 || r.durationSeconds <= 0 {
		return nil, tcerr.NewTencentCloudSDKError(creErr, "AssumeRoleWithWebIdentity durationSeconds should be in the range of 0~43200s", "")
	}
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = endpoint
	cpf.HttpProfile.ReqMethod = "POST"

	client := NewCommonClient(nil, r.region, cpf)
	request := tchttp.NewCommonRequest(service, version, action)
	request.SetSkipSign(true)

	params := map[string]interface{}{
		"ProviderId":       r.providerId,
		"WebIdentityToken": r.webIdentityToken,
		"RoleArn":          r.roleArn,
		"RoleSessionName":  r.roleSessionName,
		"DurationSeconds":  r.durationSeconds,
	}
	err := request.SetActionParameters(params)
	if err != nil {
		return nil, err
	}

	response := tchttp.NewCommonResponse()
	err = client.Send(request, response)
	if err != nil {
		return nil, err
	}
	rspSt := new(oidcStsRsp)

	if err = json.Unmarshal(response.GetBody(), rspSt); err != nil {
		return nil, tcerr.NewTencentCloudSDKError(creErr, err.Error(), "")
	}

	return &RoleArnCredential{
		roleArn:         r.roleArn,
		roleSessionName: r.roleSessionName,
		durationSeconds: r.durationSeconds,
		expiredTime:     int64(rspSt.Response.ExpiredTime),
		token:           rspSt.Response.Credentials.Token,
		tmpSecretId:     rspSt.Response.Credentials.TmpSecretId,
		tmpSecretKey:    rspSt.Response.Credentials.TmpSecretKey,
		source:          r,
	}, nil
}
