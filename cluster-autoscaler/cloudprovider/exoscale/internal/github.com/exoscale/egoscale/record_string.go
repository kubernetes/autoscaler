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
	_ = x[A-0]
	_ = x[AAAA-1]
	_ = x[ALIAS-2]
	_ = x[CNAME-3]
	_ = x[HINFO-4]
	_ = x[MX-5]
	_ = x[NAPTR-6]
	_ = x[NS-7]
	_ = x[POOL-8]
	_ = x[SPF-9]
	_ = x[SRV-10]
	_ = x[SSHFP-11]
	_ = x[TXT-12]
	_ = x[URL-13]
}

const _Record_name = "AAAAAALIASCNAMEHINFOMXNAPTRNSPOOLSPFSRVSSHFPTXTURL"

var _Record_index = [...]uint8{0, 1, 5, 10, 15, 20, 22, 27, 29, 33, 36, 39, 44, 47, 50}

func (i Record) String() string {
	if i < 0 || i >= Record(len(_Record_index)-1) {
		return "Record(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Record_name[_Record_index[i]:_Record_index[i+1]]
}
