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

package errors

import "strings"

/* const code and msg prefix */
const (
	SignatureDostNotMatchErrorCode = "SignatureDoesNotMatch"
	MessagePrefix                  = "Specified signature is not matched with our calculation. server string to sign is:"
)

// SignatureDostNotMatchWrapper implements tryWrap interface
type SignatureDostNotMatchWrapper struct{}

func (*SignatureDostNotMatchWrapper) tryWrap(error *ServerError, wrapInfo map[string]string) (bool, *ServerError) {
	clientStringToSign := wrapInfo["StringToSign"]
	if error.errorCode == SignatureDostNotMatchErrorCode && clientStringToSign != "" {
		message := error.message
		if strings.HasPrefix(message, MessagePrefix) {
			serverStringToSign := message[len(MessagePrefix):]
			if clientStringToSign == serverStringToSign {
				// user secret is error
				error.recommend = "Please check you AccessKeySecret"
			} else {
				error.recommend = "This may be a bug with the SDK and we hope you can submit this question in the " +
					"github issue(https://github.com/aliyun/alibaba-cloud-sdk-go/issues), thanks very much"
			}
		}
		return true, error
	}
	return false, nil
}
