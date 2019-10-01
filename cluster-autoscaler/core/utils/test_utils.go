/*
Copyright 2019 The Kubernetes Authors.

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

package utils

import (
	"time"
)

// MiB represents megabate (2^20) multiplier
const MiB = 1024 * 1024

// NothingReturned is a no-value marker returned by GetStringFromChan
// and GetStringFromChanImmediately functions
const NothingReturned = "Nothing returned"

// GetStringFromChan returns a value from channel or NothingReturned if no value in channel after timeout of 100 ms.
func GetStringFromChan(c chan string) string {
	select {
	case val := <-c:
		return val
	case <-time.After(100 * time.Millisecond):
		return NothingReturned
	}
}

// GetStringFromChanImmediately returns a value from channel or NothingReturned if no value in channel.
func GetStringFromChanImmediately(c chan string) string {
	select {
	case val := <-c:
		return val
	default:
		return NothingReturned
	}
}
