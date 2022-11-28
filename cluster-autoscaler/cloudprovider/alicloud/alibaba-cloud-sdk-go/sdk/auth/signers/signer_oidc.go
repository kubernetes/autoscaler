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
	"github.com/google/uuid"
	"github.com/jmespath/go-jmespath"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/responses"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/utils"
	"net/http"
	"os"
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
	commonApi         func(request *requests.CommonRequest, signer interface{}) (response *responses.CommonResponse, err error)
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
		signer.roleSessionName = "aliyun-go-sdk-" + strconv.FormatInt(time.Now().UnixNano()/1000, 10)
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
	request = requests.NewCommonRequest()
	request.Domain = "sts.aliyuncs.com"
	request.Scheme = requests.HTTPS
	request.Method = "POST"
	request.QueryParams["Timestamp"] = utils.GetTimeInFormatISO8601()
	request.QueryParams["Action"] = "AssumeRoleWithOIDC"
	request.QueryParams["Format"] = "JSON"
	request.FormParams["RoleArn"] = signer.credential.RoleArn
	request.FormParams["OIDCProviderArn"] = signer.credential.OIDCProviderArn
	request.FormParams["OIDCToken"] = signer.getOIDCToken(signer.credential.OIDCTokenFilePath)
	request.QueryParams["RoleSessionName"] = signer.credential.RoleSessionName
	request.QueryParams["Version"] = "2015-04-01"
	request.QueryParams["SignatureNonce"] = uuid.New().String()
	request.Headers["Host"] = request.Domain
	request.Headers["Accept-Encoding"] = "identity"
	request.Headers["content-type"] = "application/x-www-form-urlencoded"
	return
}

func (signer *OIDCSigner) getOIDCToken(OIDCTokenFilePath string) string {
	tokenPath := OIDCTokenFilePath
	_, err := os.Stat(tokenPath)
	if os.IsNotExist(err) {
		tokenPath = os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")
		if tokenPath == "" {
			return ""
		}
	}

	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return ""
	}
	return string(token)
}

func (signer *OIDCSigner) refreshApi(request *requests.CommonRequest) (response *responses.CommonResponse, err error) {
	requestUrl := request.BuildUrl()
	var urlEncoded string
	if request.FormParams != nil {
		urlEncoded = utils.GetUrlFormedMap(request.FormParams)
	}

	httpRequest, err := http.NewRequest(request.Method, requestUrl, strings.NewReader(urlEncoded))
	if err != nil {
		fmt.Println("refresh RRSA token err", err)
		return
	}

	httpRequest.Proto = "HTTP/1.1"
	httpRequest.Host = request.Domain
	for key, value := range request.Headers {
		if value != "" {
			httpRequest.Header[key] = []string{value}
		}
	}

	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		fmt.Println("refresh RRSA token err", err)
		return
	}

	response = responses.NewCommonResponse()
	err = responses.Unmarshal(response, httpResponse, "")
	return
}

func (signer *OIDCSigner) refreshCredential(response *responses.CommonResponse) (err error) {
	if response.GetHttpStatus() != http.StatusOK {
		message := "refresh session token failed"
		err = errors.NewServerError(response.GetHttpStatus(), response.GetHttpContentString(), message)
		return
	}
	var data interface{}
	err = json.Unmarshal(response.GetHttpContentBytes(), &data)
	if err != nil {
		fmt.Println("refresh RRSA token err, json.Unmarshal fail", err)
		return
	}
	accessKeyId, err := jmespath.Search("Credentials.AccessKeyId", data)
	if err != nil {
		fmt.Println("refresh RRSA token err, fail to get AccessKeyId", err)
		return
	}
	accessKeySecret, err := jmespath.Search("Credentials.AccessKeySecret", data)
	if err != nil {
		fmt.Println("refresh RRSA token err, fail to get AccessKeySecret", err)
		return
	}
	securityToken, err := jmespath.Search("Credentials.SecurityToken", data)
	if err != nil {
		fmt.Println("refresh RRSA token err, fail to get SecurityToken", err)
		return
	}
	expiration, err := jmespath.Search("Credentials.Expiration", data)
	if err != nil {
		fmt.Println("refresh RRSA token err, fail to get Expiration", err)
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
