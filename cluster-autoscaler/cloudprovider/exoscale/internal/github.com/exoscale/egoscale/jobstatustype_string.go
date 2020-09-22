/*
Copyright 2020 The Kubernetes Authors.

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

package egoscale

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Pending-0]
	_ = x[Success-1]
	_ = x[Failure-2]
}

const _JobStatusType_name = "PendingSuccessFailure"

var _JobStatusType_index = [...]uint8{0, 7, 14, 21}

func (i JobStatusType) String() string {
	if i < 0 || i >= JobStatusType(len(_JobStatusType_index)-1) {
		return "JobStatusType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _JobStatusType_name[_JobStatusType_index[i]:_JobStatusType_index[i+1]]
}
