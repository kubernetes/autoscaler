/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	"fmt"
)

// KeyError is returned when the mapping key was not found.
type KeyError struct {
	key interface{}
}

// NewKeyError returns a new KeyError.
func NewKeyError(key interface{}) KeyError {
	return KeyError{key}
}

func (e KeyError) Error() string {
	return fmt.Sprintf("KeyError: %s", e.key)
}
