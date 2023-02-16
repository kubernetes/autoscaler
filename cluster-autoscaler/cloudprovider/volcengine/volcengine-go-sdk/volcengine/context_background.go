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

package volcengine

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import "context"

// BackgroundContext returns a context that will never be canceled, has no
// values, and no deadline. This context is used by the SDK to provide
// backwards compatibility with non-context API operations and functionality.
// See https://golang.org/pkg/context for more information on Contexts.
func BackgroundContext() Context {
	return context.Background()
}
