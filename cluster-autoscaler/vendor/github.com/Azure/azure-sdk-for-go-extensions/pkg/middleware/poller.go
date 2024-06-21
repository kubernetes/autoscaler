/*
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

package middleware

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// AggressivePollingOptions is a very aggressive set of poller options
func AggressivePollingOptions() *runtime.PollUntilDoneOptions {
	return &runtime.PollUntilDoneOptions{
		// Frequency is the time to wait between polling intervals in absence of a Retry-After header.
		//Allowed minimum is one second.
		Frequency: time.Second * 1,
	}
}
