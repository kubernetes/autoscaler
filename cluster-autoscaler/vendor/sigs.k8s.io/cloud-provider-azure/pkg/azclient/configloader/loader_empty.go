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

package configloader

import (
	"context"
)

type emptyLoader[Type any] struct {
	Data *Type
}

func (e *emptyLoader[Type]) Load(ctx context.Context) (*Type, error) {
	return e.Data, nil
}

func newEmptyLoader[Type any](config *Type) configLoader[Type] {
	if config == nil {
		config = new(Type)
	}
	return &emptyLoader[Type]{
		Data: config,
	}
}
