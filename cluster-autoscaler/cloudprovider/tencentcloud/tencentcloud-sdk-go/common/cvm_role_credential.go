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

const ExpiredTimeout = 300

type CvmRoleCredential struct {
	roleName     string
	expiredTime  int64
	tmpSecretId  string
	tmpSecretKey string
	token        string
	source       *CvmRoleProvider
}

func (c *CvmRoleCredential) GetSecretId() string {
	if c.needRefresh() {
		c.refresh()
	}
	return c.tmpSecretId
}

func (c *CvmRoleCredential) GetToken() string {
	if c.needRefresh() {
		c.refresh()
	}
	return c.token
}

func (c *CvmRoleCredential) GetSecretKey() string {
	if c.needRefresh() {
		c.refresh()
	}
	return c.tmpSecretKey
}

func (c *CvmRoleCredential) needRefresh() bool {
	if c.tmpSecretId == "" || c.tmpSecretKey == "" || c.token == "" || c.expiredTime-ExpiredTimeout <= time.Now().Unix() {
		return true
	}
	return false
}

func (c *CvmRoleCredential) refresh() {
	newCre, err := c.source.GetCredential()
	if err != nil {
		log.Println(err)
		return
	}
	*c = *newCre.(*CvmRoleCredential)
}
