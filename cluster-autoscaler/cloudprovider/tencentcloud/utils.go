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

package tencentcloud

import (
	"regexp"
	"time"
)

func getInstanceIdsFromMessage(instances []string, msg string) ([]string, []string) {
	errInstance := make([]string, 0)
	// 为了防止修改instanceIds中的值
	regx := regexp.MustCompile(`ins-([0-9]|[a-z])*`)
	res := regx.FindAll([]byte(msg), -1)
	for _, r := range res {
		errInstance = append(errInstance, string(r))
	}
	for _, errIns := range errInstance {
		for i, okIns := range instances {
			if errIns == okIns {
				instances = append(instances[:i], instances[i+1:]...)
				break
			}
		}
	}
	return instances, errInstance
}

func retryDo(op func() (interface{}, error), checker func(interface{}, error) bool, timeout uint64, interval uint64) (ret interface{}, isTimeout bool, err error) {
	isTimeout = false
	var tm <-chan time.Time = time.After(time.Duration(timeout) * time.Second)

	times := 0
	for {
		times = times + 1
		select {
		case <-tm:
			isTimeout = true
			return
		default:
		}
		ret, err = op()
		if checker(ret, err) {
			return
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
