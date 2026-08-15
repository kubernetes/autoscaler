/*
Copyright 2021 The Kubernetes Authors.

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

package common

import (
	"log"
	"time"
)

type RoleArnCredential struct {
	roleArn         string
	roleSessionName string
	durationSeconds int64
	expiredTime     int64
	token           string
	tmpSecretId     string
	tmpSecretKey    string
	source          Provider
}

func (c *RoleArnCredential) GetSecretId() string {
	if c.needRefresh() {
		c.refresh()
	}
	return c.tmpSecretId
}

func (c *RoleArnCredential) GetSecretKey() string {
	if c.needRefresh() {
		c.refresh()
	}
	return c.tmpSecretKey

}

func (c *RoleArnCredential) GetToken() string {
	if c.needRefresh() {
		c.refresh()
	}
	return c.token
}

func (c *RoleArnCredential) needRefresh() bool {
	if c.tmpSecretKey == "" || c.tmpSecretId == "" || c.token == "" || c.expiredTime <= time.Now().Unix() {
		return true
	}
	return false
}

func (c *RoleArnCredential) refresh() {
	newCre, err := c.source.GetCredential()
	if err != nil {
		log.Println(err)
	}
	*c = *newCre.(*RoleArnCredential)
}
