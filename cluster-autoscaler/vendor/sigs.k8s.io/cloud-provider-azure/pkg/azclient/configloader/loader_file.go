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
	"os"
)

type FileLoader[Type any] struct {
	filePath string
	configLoader[Type]
	decoderFactory[Type]
}

// newFileLoader creates a FileLoader with the specified file path and loader.
// decoderFactory is a function that creates a new loader from the content of the file. it should never be nil.
func newFileLoader[Type any](filePath string, loader configLoader[Type], decoder decoderFactory[Type]) configLoader[Type] {
	return &FileLoader[Type]{
		filePath:       filePath,
		configLoader:   loader,
		decoderFactory: decoder,
	}
}

func (f *FileLoader[Type]) Load(ctx context.Context) (*Type, error) {
	if f.configLoader == nil {
		f.configLoader = newEmptyLoader[Type](nil)
	}

	content, err := os.ReadFile(f.filePath)
	if err != nil {
		return nil, err
	}

	loader := f.decoderFactory(content, f.configLoader)
	return loader.Load(ctx)
}
