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

package requests

import (
	"fmt"
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"reflect"
	"strconv"
)

/* const vars */
const (
	RPC = "RPC"
	ROA = "ROA"

	HTTP  = "HTTP"
	HTTPS = "HTTPS"

	DefaultHttpPort = "80"

	GET     = "GET"
	PUT     = "PUT"
	POST    = "POST"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"

	Json = "application/json"
	Xml  = "application/xml"
	Raw  = "application/octet-stream"
	Form = "application/x-www-form-urlencoded"

	Header = "Header"
	Query  = "Query"
	Body   = "Body"
	Path   = "Path"

	HeaderSeparator = "\n"
)

// AcsRequest interface
type AcsRequest interface {
	GetScheme() string
	GetMethod() string
	GetDomain() string
	GetPort() string
	GetRegionId() string
	GetUrl() string
	GetQueries() string
	GetHeaders() map[string]string
	GetQueryParams() map[string]string
	GetFormParams() map[string]string
	GetContent() []byte
	GetBodyReader() io.Reader
	GetStyle() string
	GetProduct() string
	GetVersion() string
	GetActionName() string
	GetAcceptFormat() string
	GetLocationServiceCode() string
	GetLocationEndpointType() string

	SetStringToSign(stringToSign string)
	GetStringToSign() string

	SetDomain(domain string)
	SetContent(content []byte)
	SetScheme(scheme string)
	BuildUrl() string
	BuildQueries() string

	addHeaderParam(key, value string)
	addQueryParam(key, value string)
	addFormParam(key, value string)
	addPathParam(key, value string)
}

// base class
type baseRequest struct {
	Scheme   string
	Method   string
	Domain   string
	Port     string
	RegionId string

	product string
	version string

	actionName string

	AcceptFormat string

	QueryParams map[string]string
	Headers     map[string]string
	FormParams  map[string]string
	Content     []byte

	locationServiceCode  string
	locationEndpointType string

	queries string

	stringToSign string
}

// GetQueryParams returns QueryParams
func (request *baseRequest) GetQueryParams() map[string]string {
	return request.QueryParams
}

// GetFormParams returns FormParams
func (request *baseRequest) GetFormParams() map[string]string {
	return request.FormParams
}

// GetContent returns Content
func (request *baseRequest) GetContent() []byte {
	return request.Content
}

// GetVersion returns version
func (request *baseRequest) GetVersion() string {
	return request.version
}

// GetActionName returns actionName
func (request *baseRequest) GetActionName() string {
	return request.actionName
}

// SetContent returns content
func (request *baseRequest) SetContent(content []byte) {
	request.Content = content
}

func (request *baseRequest) addHeaderParam(key, value string) {
	request.Headers[key] = value
}

func (request *baseRequest) addQueryParam(key, value string) {
	request.QueryParams[key] = value
}

func (request *baseRequest) addFormParam(key, value string) {
	request.FormParams[key] = value
}

// GetAcceptFormat returns AcceptFormat
func (request *baseRequest) GetAcceptFormat() string {
	return request.AcceptFormat
}

// GetLocationServiceCode returns locationServiceCode
func (request *baseRequest) GetLocationServiceCode() string {
	return request.locationServiceCode
}

// GetLocationEndpointType returns locationEndpointType
func (request *baseRequest) GetLocationEndpointType() string {
	return request.locationEndpointType
}

// GetProduct returns product
func (request *baseRequest) GetProduct() string {
	return request.product
}

// GetScheme returns scheme
func (request *baseRequest) GetScheme() string {
	return request.Scheme
}

// SetScheme sets scheme
func (request *baseRequest) SetScheme(scheme string) {
	request.Scheme = scheme
}

// GetMethod returns Method
func (request *baseRequest) GetMethod() string {
	return request.Method
}

// GetDomain returns Domain
func (request *baseRequest) GetDomain() string {
	return request.Domain
}

// SetDomain sets host
func (request *baseRequest) SetDomain(host string) {
	request.Domain = host
}

// GetPort returns port
func (request *baseRequest) GetPort() string {
	return request.Port
}

// GetRegionId returns regionId
func (request *baseRequest) GetRegionId() string {
	return request.RegionId
}

// GetHeaders returns headers
func (request *baseRequest) GetHeaders() map[string]string {
	return request.Headers
}

// SetContentType sets content type
func (request *baseRequest) SetContentType(contentType string) {
	request.Headers["Content-Type"] = contentType
}

// GetContentType returns content type
func (request *baseRequest) GetContentType() (contentType string, contains bool) {
	contentType, contains = request.Headers["Content-Type"]
	return
}

// SetStringToSign sets stringToSign
func (request *baseRequest) SetStringToSign(stringToSign string) {
	request.stringToSign = stringToSign
}

// GetStringToSign returns stringToSign
func (request *baseRequest) GetStringToSign() string {
	return request.stringToSign
}

func defaultBaseRequest() (request *baseRequest) {
	request = &baseRequest{
		Scheme:       "",
		AcceptFormat: "JSON",
		Method:       GET,
		QueryParams:  make(map[string]string),
		Headers: map[string]string{
			"x-sdk-client":      "golang/1.0.0",
			"x-sdk-invoke-type": "normal",
			"Accept-Encoding":   "identity",
		},
		FormParams: make(map[string]string),
	}
	return
}

// InitParams returns params
func InitParams(request AcsRequest) (err error) {
	requestValue := reflect.ValueOf(request).Elem()
	err = flatRepeatedList(requestValue, request, "", "")
	return
}

func flatRepeatedList(dataValue reflect.Value, request AcsRequest, position, prefix string) (err error) {
	dataType := dataValue.Type()
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		name, containsNameTag := field.Tag.Lookup("name")
		fieldPosition := position
		if fieldPosition == "" {
			fieldPosition, _ = field.Tag.Lookup("position")
		}
		typeTag, containsTypeTag := field.Tag.Lookup("type")
		if containsNameTag {
			if !containsTypeTag {
				// simple param
				key := prefix + name
				value := dataValue.Field(i).String()
				err = addParam(request, fieldPosition, key, value)
				if err != nil {
					return
				}
			} else if typeTag == "Repeated" {
				// repeated param
				repeatedFieldValue := dataValue.Field(i)
				if repeatedFieldValue.Kind() != reflect.Slice {
					// possible value: {"[]string", "*[]struct"}, we must call Elem() in the last condition
					repeatedFieldValue = repeatedFieldValue.Elem()
				}
				if repeatedFieldValue.IsValid() && !repeatedFieldValue.IsNil() {
					for m := 0; m < repeatedFieldValue.Len(); m++ {
						elementValue := repeatedFieldValue.Index(m)
						key := prefix + name + "." + strconv.Itoa(m+1)
						if elementValue.Type().String() == "string" {
							value := elementValue.String()
							err = addParam(request, fieldPosition, key, value)
							if err != nil {
								return
							}
						} else {
							err = flatRepeatedList(elementValue, request, fieldPosition, key+".")
							if err != nil {
								return
							}
						}
					}
				}
			}
		}
	}
	return
}

func addParam(request AcsRequest, position, name, value string) (err error) {
	if len(value) > 0 {
		switch position {
		case Header:
			request.addHeaderParam(name, value)
		case Query:
			request.addQueryParam(name, value)
		case Path:
			request.addPathParam(name, value)
		case Body:
			request.addFormParam(name, value)
		default:
			errMsg := fmt.Sprintf(errors.UnsupportedParamPositionErrorMessage, position)
			err = errors.NewClientError(errors.UnsupportedParamPositionErrorCode, errMsg, nil)
		}
	}
	return
}
