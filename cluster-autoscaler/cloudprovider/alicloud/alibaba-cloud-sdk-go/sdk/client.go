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

package sdk

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/endpoints"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/responses"
	"net"
	"net/http"
	"strconv"
	"sync"
)

// Version value will be replaced while build: -ldflags="-X sdk.version=x.x.x"
var Version = "0.0.1"

// Client is common SDK client
type Client struct {
	regionId       string
	config         *Config
	signer         auth.Signer
	httpClient     *http.Client
	asyncTaskQueue chan func()

	debug     bool
	isRunning bool
	// void "panic(write to close channel)" cause of addAsync() after Shutdown()
	asyncChanLock *sync.RWMutex
}

// Init not support yet
func (client *Client) Init() (err error) {
	panic("not support yet")
}

// InitWithOptions provide options such as regionId and auth
func (client *Client) InitWithOptions(regionId string, config *Config, credential auth.Credential) (err error) {
	client.isRunning = true
	client.asyncChanLock = new(sync.RWMutex)
	client.regionId = regionId
	client.config = config
	if err != nil {
		return
	}
	client.httpClient = &http.Client{}

	if config.HttpTransport != nil {
		client.httpClient.Transport = config.HttpTransport
	}

	if config.Timeout > 0 {
		client.httpClient.Timeout = config.Timeout
	}

	if config.EnableAsync {
		client.EnableAsync(config.GoRoutinePoolSize, config.MaxTaskQueueSize)
	}

	client.signer, err = auth.NewSignerWithCredential(credential, client.ProcessCommonRequestWithSigner)

	return
}

// EnableAsync enable async task queue
func (client *Client) EnableAsync(routinePoolSize, maxTaskQueueSize int) {
	client.asyncTaskQueue = make(chan func(), maxTaskQueueSize)
	for i := 0; i < routinePoolSize; i++ {
		go func() {
			for client.isRunning {
				select {
				case task, notClosed := <-client.asyncTaskQueue:
					if notClosed {
						task()
					}
				}
			}
		}()
	}
}

// InitWithAccessKey need accessKeyId and accessKeySecret
func (client *Client) InitWithAccessKey(regionId, accessKeyId, accessKeySecret string) (err error) {
	config := client.InitClientConfig()
	credential := &credentials.BaseCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	return client.InitWithOptions(regionId, config, credential)
}

// InitWithStsToken need regionId,accessKeyId,accessKeySecret and securityToken
func (client *Client) InitWithStsToken(regionId, accessKeyId, accessKeySecret, securityToken string) (err error) {
	config := client.InitClientConfig()
	credential := &credentials.StsTokenCredential{
		AccessKeyId:       accessKeyId,
		AccessKeySecret:   accessKeySecret,
		AccessKeyStsToken: securityToken,
	}
	return client.InitWithOptions(regionId, config, credential)
}

// InitWithRamRoleArn need regionId,accessKeyId,accessKeySecret,roleArn and roleSessionName
func (client *Client) InitWithRamRoleArn(regionId, accessKeyId, accessKeySecret, roleArn, roleSessionName string) (err error) {
	config := client.InitClientConfig()
	credential := &credentials.RamRoleArnCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		RoleArn:         roleArn,
		RoleSessionName: roleSessionName,
	}
	return client.InitWithOptions(regionId, config, credential)
}

// InitWithRsaKeyPair need key pair
func (client *Client) InitWithRsaKeyPair(regionId, publicKeyId, privateKey string, sessionExpiration int) (err error) {
	config := client.InitClientConfig()
	credential := &credentials.RsaKeyPairCredential{
		PrivateKey:        privateKey,
		PublicKeyId:       publicKeyId,
		SessionExpiration: sessionExpiration,
	}
	return client.InitWithOptions(regionId, config, credential)
}

// InitWithEcsRamRole need regionId and roleName
func (client *Client) InitWithEcsRamRole(regionId, roleName string) (err error) {
	config := client.InitClientConfig()
	credential := &credentials.EcsRamRoleCredential{
		RoleName: roleName,
	}
	return client.InitWithOptions(regionId, config, credential)
}

// InitClientConfig init client config
func (client *Client) InitClientConfig() (config *Config) {
	if client.config != nil {
		return client.config
	}
	return NewConfig()
}

// DoAction do request to open api
func (client *Client) DoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return client.DoActionWithSigner(request, response, nil)
}

// BuildRequestWithSigner build request signer
func (client *Client) BuildRequestWithSigner(request requests.AcsRequest, signer auth.Signer) (err error) {
	// add clientVersion
	request.GetHeaders()["x-sdk-core-version"] = Version

	regionId := client.regionId
	if len(request.GetRegionId()) > 0 {
		regionId = request.GetRegionId()
	}

	// resolve endpoint
	resolveParam := &endpoints.ResolveParam{
		Domain:               request.GetDomain(),
		Product:              request.GetProduct(),
		RegionId:             regionId,
		LocationProduct:      request.GetLocationServiceCode(),
		LocationEndpointType: request.GetLocationEndpointType(),
		CommonApi:            client.ProcessCommonRequest,
	}
	endpoint, err := endpoints.Resolve(resolveParam)
	if err != nil {
		return
	}
	request.SetDomain(endpoint)

	// init request params
	err = requests.InitParams(request)
	if err != nil {
		return
	}

	// signature
	var finalSigner auth.Signer
	if signer != nil {
		finalSigner = signer
	} else {
		finalSigner = client.signer
	}
	httpRequest, err := buildHttpRequest(request, finalSigner, regionId)
	if client.config.UserAgent != "" {
		httpRequest.Header.Set("User-Agent", client.config.UserAgent)
	}
	return err
}

// DoActionWithSigner do action with signer
func (client *Client) DoActionWithSigner(request requests.AcsRequest, response responses.AcsResponse, signer auth.Signer) (err error) {

	// add clientVersion
	request.GetHeaders()["x-sdk-core-version"] = Version

	regionId := client.regionId
	if len(request.GetRegionId()) > 0 {
		regionId = request.GetRegionId()
	}

	// resolve endpoint
	resolveParam := &endpoints.ResolveParam{
		Domain:               request.GetDomain(),
		Product:              request.GetProduct(),
		RegionId:             regionId,
		LocationProduct:      request.GetLocationServiceCode(),
		LocationEndpointType: request.GetLocationEndpointType(),
		CommonApi:            client.ProcessCommonRequest,
	}
	endpoint, err := endpoints.Resolve(resolveParam)
	if err != nil {
		return
	}
	request.SetDomain(endpoint)

	if request.GetScheme() == "" {
		request.SetScheme(client.config.Scheme)
	}
	// init request params
	err = requests.InitParams(request)
	if err != nil {
		return
	}

	// signature
	var finalSigner auth.Signer
	if signer != nil {
		finalSigner = signer
	} else {
		finalSigner = client.signer
	}
	httpRequest, err := buildHttpRequest(request, finalSigner, regionId)
	if client.config.UserAgent != "" {
		httpRequest.Header.Set("User-Agent", client.config.UserAgent)
	}
	if err != nil {
		return
	}
	var httpResponse *http.Response
	for retryTimes := 0; retryTimes <= client.config.MaxRetryTime; retryTimes++ {
		httpResponse, err = client.httpClient.Do(httpRequest)

		var timeout bool
		// receive error
		if err != nil {
			if !client.config.AutoRetry {
				return
			} else if timeout = isTimeout(err); !timeout {
				// if not timeout error, return
				return
			} else if retryTimes >= client.config.MaxRetryTime {
				// timeout but reached the max retry times, return
				timeoutErrorMsg := fmt.Sprintf(errors.TimeoutErrorMessage, strconv.Itoa(retryTimes+1), strconv.Itoa(retryTimes+1))
				err = errors.NewClientError(errors.TimeoutErrorCode, timeoutErrorMsg, err)
				return
			}
		}
		//  if status code >= 500 or timeout, will trigger retry
		if client.config.AutoRetry && (timeout || isServerError(httpResponse)) {
			// rewrite signatureNonce and signature
			httpRequest, err = buildHttpRequest(request, finalSigner, regionId)
			if err != nil {
				return
			}
			continue
		}
		break
	}
	err = responses.Unmarshal(response, httpResponse, request.GetAcceptFormat())
	// wrap server errors
	if serverErr, ok := err.(*errors.ServerError); ok {
		var wrapInfo = map[string]string{}
		wrapInfo["StringToSign"] = request.GetStringToSign()
		err = errors.WrapServerError(serverErr, wrapInfo)
	}
	return
}

func buildHttpRequest(request requests.AcsRequest, singer auth.Signer, regionId string) (httpRequest *http.Request, err error) {
	err = auth.Sign(request, singer, regionId)
	if err != nil {
		return
	}
	requestMethod := request.GetMethod()
	requestUrl := request.BuildUrl()
	body := request.GetBodyReader()
	httpRequest, err = http.NewRequest(requestMethod, requestUrl, body)
	if err != nil {
		return
	}
	for key, value := range request.GetHeaders() {
		httpRequest.Header[key] = []string{value}
	}
	// host is a special case
	if host, containsHost := request.GetHeaders()["Host"]; containsHost {
		httpRequest.Host = host
	}
	return
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	netErr, isNetError := err.(net.Error)
	return isNetError && netErr.Timeout()
}

func isServerError(httpResponse *http.Response) bool {
	return httpResponse.StatusCode >= http.StatusInternalServerError
}

// AddAsyncTask create async task
// only block when any one of the following occurs:
// 1. the asyncTaskQueue is full, increase the queue size to avoid this
// 2. Shutdown() in progressing, the client is being closed
func (client *Client) AddAsyncTask(task func()) (err error) {
	if client.asyncTaskQueue != nil {
		client.asyncChanLock.RLock()
		defer client.asyncChanLock.RUnlock()
		if client.isRunning {
			client.asyncTaskQueue <- task
		}
	} else {
		err = errors.NewClientError(errors.AsyncFunctionNotEnabledCode, errors.AsyncFunctionNotEnabledMessage, nil)
	}
	return
}

// GetConfig return client config
func (client *Client) GetConfig() *Config {
	return client.config
}

// NewClient return SDK client
func NewClient() (client *Client, err error) {
	client = &Client{}
	err = client.Init()
	return
}

// NewClientWithOptions create client with options
func NewClientWithOptions(regionId string, config *Config, credential auth.Credential) (client *Client, err error) {
	client = &Client{}
	err = client.InitWithOptions(regionId, config, credential)
	return
}

// NewClientWithAccessKey create client with accessKeyId and accessKeySecret
func NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret string) (client *Client, err error) {
	client = &Client{}
	err = client.InitWithAccessKey(regionId, accessKeyId, accessKeySecret)
	return
}

// NewClientWithStsToken create client with stsToken
func NewClientWithStsToken(regionId, stsAccessKeyId, stsAccessKeySecret, stsToken string) (client *Client, err error) {
	client = &Client{}
	err = client.InitWithStsToken(regionId, stsAccessKeyId, stsAccessKeySecret, stsToken)
	return
}

// NewClientWithRamRoleArn create client with ramRoleArn
func NewClientWithRamRoleArn(regionId string, accessKeyId, accessKeySecret, roleArn, roleSessionName string) (client *Client, err error) {
	client = &Client{}
	err = client.InitWithRamRoleArn(regionId, accessKeyId, accessKeySecret, roleArn, roleSessionName)
	return
}

// NewClientWithEcsRamRole create client with ramRole on ECS
func NewClientWithEcsRamRole(regionId string, roleName string) (client *Client, err error) {
	client = &Client{}
	err = client.InitWithEcsRamRole(regionId, roleName)
	return
}

// NewClientWithRsaKeyPair create client with key-pair
func NewClientWithRsaKeyPair(regionId string, publicKeyId, privateKey string, sessionExpiration int) (client *Client, err error) {
	client = &Client{}
	err = client.InitWithRsaKeyPair(regionId, publicKeyId, privateKey, sessionExpiration)
	return
}

// NewClientWithStsRoleArn is deprecated: Use NewClientWithRamRoleArn in this package instead.
func NewClientWithStsRoleArn(regionId string, accessKeyId, accessKeySecret, roleArn, roleSessionName string) (client *Client, err error) {
	return NewClientWithRamRoleArn(regionId, accessKeyId, accessKeySecret, roleArn, roleSessionName)
}

// NewClientWithStsRoleNameOnEcs is deprecated: Use NewClientWithEcsRamRole in this package instead.
func NewClientWithStsRoleNameOnEcs(regionId string, roleName string) (client *Client, err error) {
	return NewClientWithEcsRamRole(regionId, roleName)
}

// ProcessCommonRequest do action with common request
func (client *Client) ProcessCommonRequest(request *requests.CommonRequest) (response *responses.CommonResponse, err error) {
	request.TransToAcsRequest()
	response = responses.NewCommonResponse()
	err = client.DoAction(request, response)
	return
}

// ProcessCommonRequestWithSigner do action with common request and singer
func (client *Client) ProcessCommonRequestWithSigner(request *requests.CommonRequest, signerInterface interface{}) (response *responses.CommonResponse, err error) {
	if signer, isSigner := signerInterface.(auth.Signer); isSigner {
		request.TransToAcsRequest()
		response = responses.NewCommonResponse()
		err = client.DoActionWithSigner(request, response, signer)
		return
	}
	panic("should not be here")
}

// Shutdown destruction of client
func (client *Client) Shutdown() {
	client.signer.Shutdown()
	// lock the addAsync()
	client.asyncChanLock.Lock()
	defer client.asyncChanLock.Unlock()
	client.isRunning = false
	close(client.asyncTaskQueue)
}
