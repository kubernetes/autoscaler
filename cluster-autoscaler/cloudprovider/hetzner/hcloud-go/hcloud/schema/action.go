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

package schema

import "time"

// Action defines the schema of an action.
type Action struct {
	ID        int                       `json:"id"`
	Status    string                    `json:"status"`
	Command   string                    `json:"command"`
	Progress  int                       `json:"progress"`
	Started   time.Time                 `json:"started"`
	Finished  *time.Time                `json:"finished"`
	Error     *ActionError              `json:"error"`
	Resources []ActionResourceReference `json:"resources"`
}

// ActionResourceReference defines the schema of an action resource reference.
type ActionResourceReference struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

// ActionError defines the schema of an error embedded
// in an action.
type ActionError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ActionGetResponse is the schema of the response when
// retrieving a single action.
type ActionGetResponse struct {
	Action Action `json:"action"`
}

// ActionListResponse defines the schema of the response when listing actions.
type ActionListResponse struct {
	Actions []Action `json:"actions"`
}
