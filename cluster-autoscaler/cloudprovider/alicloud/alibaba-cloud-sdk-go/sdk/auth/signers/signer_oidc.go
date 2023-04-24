/*
Copyright 2018 The Kubernetes Authors.

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

package signers

import (
	"encoding/json"
	"fmt"
	"github.com/jmespath/go-jmespath"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/responses"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/utils"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultOIDCDurationSeconds = 3600
)

// OIDCSigner is kind of signer
type OIDCSigner struct {
	*credentialUpdater
	roleSessionName   string
	sessionCredential *SessionCredential
	credential        *credentials.OIDCCredential
}

// NewOIDCSigner returns OIDCSigner
func NewOIDCSigner(credential *credentials.OIDCCredential) (signer *OIDCSigner, err error) {
	signer = &OIDCSigner{
		credential: credential,
	}

	signer.credentialUpdater = &credentialUpdater{
		credentialExpiration: credential.RoleSessionExpiration,
		buildRequestMethod:   signer.buildCommonRequest,
		responseCallBack:     signer.refreshCredential,
		refreshApi:           signer.refreshApi,
	}

	if len(credential.RoleSessionName) > 0 {
		signer.roleSessionName = credential.RoleSessionName
	} else {
		signer.roleSessionName = "kubernetes-cluster-autoscaler-" + strconv.FormatInt(time.Now().UnixNano()/1000, 10)
	}
	if credential.RoleSessionExpiration > 0 {
		if credential.RoleSessionExpiration >= 900 && credential.RoleSessionExpiration <= 3600 {
			signer.credentialExpiration = credential.RoleSessionExpiration
		} else {
			err = errors.NewClientError(errors.InvalidParamErrorCode, "Assume Role session duration should be in the range of 15min - 1Hr", nil)
		}
	} else {
		signer.credentialExpiration = defaultOIDCDurationSeconds
	}
	return
}

// GetName returns "HMAC-SHA1"
func (*OIDCSigner) GetName() string {
	return "HMAC-SHA1"
}

// GetType returns ""
func (*OIDCSigner) GetType() string {
	return ""
}

// GetVersion returns "1.0"
func (*OIDCSigner) GetVersion() string {
	return "1.0"
}

// GetAccessKeyId returns accessKeyId
func (signer *OIDCSigner) GetAccessKeyId() (accessKeyId string, err error) {
	if signer.sessionCredential == nil || signer.needUpdateCredential() {
		err = signer.updateCredential()
	}
	if err != nil && (signer.sessionCredential == nil || len(signer.sessionCredential.AccessKeyId) <= 0) {
		return "", err
	}

	return signer.sessionCredential.AccessKeyId, nil
}

// GetExtraParam returns params
func (signer *OIDCSigner) GetExtraParam() map[string]string {
	if signer.sessionCredential == nil || signer.needUpdateCredential() {
		signer.updateCredential()
	}
	if signer.sessionCredential == nil || len(signer.sessionCredential.StsToken) <= 0 {
		return make(map[string]string)
	}

	return map[string]string{"SecurityToken": signer.sessionCredential.StsToken}
}

// Sign create signer
func (signer *OIDCSigner) Sign(stringToSign, secretSuffix string) string {
	secret := signer.sessionCredential.AccessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

func (signer *OIDCSigner) buildCommonRequest() (request *requests.CommonRequest, err error) {
	const endpoint = "sts.aliyuncs.com"
	const stsApiVersion = "2015-04-01"
	const action = "AssumeRoleWithOIDC"
	request = requests.NewCommonRequest()
	request.Scheme = requests.HTTPS
	request.Domain = endpoint
	request.Method = requests.POST
	request.QueryParams["Action"] = action
	request.QueryParams["Version"] = stsApiVersion
	request.QueryParams["Format"] = "JSON"
	request.QueryParams["Timestamp"] = utils.GetTimeInFormatISO8601()
	request.QueryParams["SignatureNonce"] = utils.GetUUIDV4()
	request.FormParams["RoleArn"] = signer.credential.RoleArn
	request.FormParams["OIDCProviderArn"] = signer.credential.OIDCProviderArn
	request.FormParams["OIDCToken"] = signer.getOIDCToken(signer.credential.OIDCTokenFilePath)
	request.QueryParams["RoleSessionName"] = signer.credential.RoleSessionName
	request.Headers["host"] = endpoint
	request.Headers["Accept-Encoding"] = "identity"
	request.Headers["content-type"] = "application/x-www-form-urlencoded"
	request.Headers["user-agent"] = fmt.Sprintf("AlibabaCloud (%s; %s) Golang/%s Core/%s TeaDSL/1 kubernetes-cluster-autoscaler", runtime.GOOS, runtime.GOARCH, strings.Trim(runtime.Version(), "go"), "0.01")
	return
}

func (signer *OIDCSigner) getOIDCToken(OIDCTokenFilePath string) string {
	tokenPath := OIDCTokenFilePath
	_, err := os.Stat(tokenPath)
	if os.IsNotExist(err) {
		tokenPath = os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")
		if tokenPath == "" {
			klog.Error("oidc token file path is missing")
			return ""
		}
	}

	token, err := os.ReadFile(tokenPath)
	if err != nil {
		klog.Errorf("get oidc token from file %s failed: %s", tokenPath, err)
		return ""
	}
	return string(token)
}

func (signer *OIDCSigner) refreshApi(request *requests.CommonRequest) (response *responses.CommonResponse, err error) {
	body := utils.GetUrlFormedMap(request.FormParams)
	httpRequest, err := http.NewRequest(request.Method, fmt.Sprintf("%s://%s/?%s", strings.ToLower(request.Scheme), request.Domain, utils.GetUrlFormedMap(request.QueryParams)), strings.NewReader(body))
	if err != nil {
		klog.Errorf("refresh RRSA token failed: %s", err)
		return
	}

	httpRequest.Proto = "HTTP/1.1"
	httpRequest.Host = request.Domain
	for k, v := range request.Headers {
		httpRequest.Header.Add(k, v)
	}

	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		klog.Errorf("refresh RRSA token failed: %s", err)
		return
	}

	response = responses.NewCommonResponse()
	err = responses.Unmarshal(response, httpResponse, "")

	return
}

func (signer *OIDCSigner) refreshCredential(response *responses.CommonResponse) (err error) {
	if response.GetHttpStatus() != http.StatusOK {
		message := "refresh RRSA failed"
		err = errors.NewServerError(response.GetHttpStatus(), response.GetHttpContentString(), message)
		return
	}

	var data interface{}
	err = json.Unmarshal(response.GetHttpContentBytes(), &data)
	if err != nil {
		klog.Errorf("refresh RRSA token err, json.Unmarshal fail: %s", err)
		return
	}
	accessKeyId, err := jmespath.Search("Credentials.AccessKeyId", data)
	if err != nil {
		klog.Errorf("refresh RRSA token err, fail to get AccessKeyId: %s", err)
		return
	}
	accessKeySecret, err := jmespath.Search("Credentials.AccessKeySecret", data)
	if err != nil {
		klog.Errorf("refresh RRSA token err, fail to get AccessKeySecret: %s", err)
		return
	}
	securityToken, err := jmespath.Search("Credentials.SecurityToken", data)
	if err != nil {
		klog.Errorf("refresh RRSA token err, fail to get SecurityToken: %s", err)
		return
	}
	expiration, err := jmespath.Search("Credentials.Expiration", data)
	if err != nil {
		klog.Errorf("refresh RRSA token err, fail to get Expiration: %s", err)
		return
	}

	if accessKeyId == nil || accessKeySecret == nil || securityToken == nil {
		return
	}

	expirationTime, err := time.Parse("2006-01-02T15:04:05Z", expiration.(string))
	signer.credentialExpiration = int(expirationTime.Unix() - time.Now().Unix())
	signer.sessionCredential = &SessionCredential{
		AccessKeyId:     accessKeyId.(string),
		AccessKeySecret: accessKeySecret.(string),
		StsToken:        securityToken.(string),
	}

	return
}

// GetSessionCredential returns SessionCredential
func (signer *OIDCSigner) GetSessionCredential() *SessionCredential {
	return signer.sessionCredential
}

// Shutdown doesn't implement
func (signer *OIDCSigner) Shutdown() {}
