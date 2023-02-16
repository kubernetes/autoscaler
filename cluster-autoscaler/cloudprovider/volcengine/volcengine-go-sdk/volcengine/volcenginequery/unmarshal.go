/*
Copyright 2023 The Kubernetes Authors.

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

package volcenginequery

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/response"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/special"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcengineerr"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcengineutil"
)

// UnmarshalHandler is a named request handler for unmarshaling volcenginequery protocol requests
var UnmarshalHandler = request.NamedHandler{Name: "volcenginesdk.volcenginequery.Unmarshal", Fn: Unmarshal}

// UnmarshalMetaHandler is a named request handler for unmarshaling volcenginequery protocol request metadata
var UnmarshalMetaHandler = request.NamedHandler{Name: "volcenginesdk.volcenginequery.UnmarshalMeta", Fn: UnmarshalMeta}

// Unmarshal unmarshals a response for an VOLCSTACK Query service.
func Unmarshal(r *request.Request) {
	defer r.HTTPResponse.Body.Close()
	if r.DataFilled() {
		body, err := ioutil.ReadAll(r.HTTPResponse.Body)
		if err != nil {
			fmt.Printf("read volcenginebody err, %v\n", err)
			r.Error = err
			return
		}

		var forceJsonNumberDecoder bool

		if r.Config.ForceJsonNumberDecode != nil {
			forceJsonNumberDecoder = r.Config.ForceJsonNumberDecode(r.Context(), r.MergeRequestInfo())
		}

		if reflect.TypeOf(r.Data) == reflect.TypeOf(&map[string]interface{}{}) {
			if err = json.Unmarshal(body, &r.Data); err != nil || forceJsonNumberDecoder {
				//try next
				decoder := json.NewDecoder(bytes.NewReader(body))
				decoder.UseNumber()
				if err = decoder.Decode(&r.Data); err != nil {
					fmt.Printf("Unmarshal err, %v\n", err)
					r.Error = err
					return
				}
			}
			var info interface{}

			ptr := r.Data.(*map[string]interface{})
			info, err = volcengineutil.ObtainSdkValue("ResponseMetadata.Error.Code", *ptr)
			if err != nil {
				r.Error = err
				return
			}
			if info != nil {
				if processBodyError(r, &response.VolcengineResponse{}, body, forceJsonNumberDecoder) {
					return
				}
			}

		} else {
			volcengineResponse := response.VolcengineResponse{}
			if processBodyError(r, &volcengineResponse, body, forceJsonNumberDecoder) {
				return
			}

			if _, ok := reflect.TypeOf(r.Data).Elem().FieldByName("Metadata"); ok {
				if volcengineResponse.ResponseMetadata != nil {
					volcengineResponse.ResponseMetadata.HTTPCode = r.HTTPResponse.StatusCode
				}
				r.Metadata = *(volcengineResponse.ResponseMetadata)
				reflect.ValueOf(r.Data).Elem().FieldByName("Metadata").Set(reflect.ValueOf(volcengineResponse.ResponseMetadata))
			}

			var (
				b      []byte
				source interface{}
			)

			if r.Config.CustomerUnmarshalData != nil {
				source = r.Config.CustomerUnmarshalData(r.Context(), r.MergeRequestInfo(), volcengineResponse)
			} else {
				if sp, ok := special.ResponseSpecialMapping()[r.ClientInfo.ServiceName]; ok {
					source = sp(volcengineResponse, r.Data)
				} else {
					source = volcengineResponse.Result
				}
			}

			if b, err = json.Marshal(source); err != nil {
				fmt.Printf("Unmarshal err, %v\n", err)
				r.Error = err
				return
			}
			if err = json.Unmarshal(b, &r.Data); err != nil || forceJsonNumberDecoder {
				decoder := json.NewDecoder(bytes.NewReader(b))
				decoder.UseNumber()
				if err = decoder.Decode(&r.Data); err != nil {
					fmt.Printf("Unmarshal err, %v\n", err)
					r.Error = err
					return
				}
			}
		}

	}
}

// UnmarshalMeta unmarshals header response values for an VOLCSTACK Query service.
func UnmarshalMeta(r *request.Request) {

}

func processBodyError(r *request.Request, volcengineResponse *response.VolcengineResponse, body []byte, forceJsonNumberDecoder bool) bool {
	if err := json.Unmarshal(body, &volcengineResponse); err != nil || forceJsonNumberDecoder {
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		if err = decoder.Decode(&r.Data); err != nil {
			fmt.Printf("Unmarshal err, %v\n", err)
			r.Error = err
			return true
		}
	}
	if volcengineResponse.ResponseMetadata.Error != nil && volcengineResponse.ResponseMetadata.Error.Code != "" {
		r.Error = volcengineerr.NewRequestFailure(
			volcengineerr.New(volcengineResponse.ResponseMetadata.Error.Code, volcengineResponse.ResponseMetadata.Error.Message, nil),
			http.StatusBadRequest,
			volcengineResponse.ResponseMetadata.RequestId,
		)
		processUnmarshalError(unmarshalErrorInfo{
			Request:  r,
			Response: volcengineResponse,
			Body:     body,
			Err:      r.Error,
		})
		return true
	}
	return false
}
