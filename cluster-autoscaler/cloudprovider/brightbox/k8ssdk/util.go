// Copyright 2018 Brightbox Systems Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8ssdk

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	ProviderName   = "brightbox"
	ProviderPrefix = ProviderName + "://"
)

// Parse the provider id string and return a string that should be a server id
// Should be no need for  error checking here, since the input string
// is constrained in format by the k8s process
func MapProviderIDToServerID(providerID string) string {
	if strings.HasPrefix(providerID, ProviderPrefix) {
		return strings.TrimPrefix(providerID, ProviderPrefix)
	}
	return providerID
}

// Add the provider prefix to the server ID
func MapServerIDToProviderID(serverID string) string {
	return ProviderPrefix + serverID
}

// Parse the zone handle and return the embedded region id
// Zone names are of the form: ${region-name}-${ix}
// So we look for the last '-' and trim just before that
func MapZoneHandleToRegion(zoneHandle string) (string, error) {
	ix := strings.LastIndex(zoneHandle, "-")
	if ix == -1 {
		return "", fmt.Errorf("unexpected zone: %s", zoneHandle)
	}
	return zoneHandle[:ix], nil
}

// getenvWithDefault retrieves the value of the environment variable
// named by the key. If the variable is not present, return the default
// value instead.
func getenvWithDefault(key string, defaultValue string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultValue
}

// get a list of inserts and deletes that changes oldList into newList
func getSyncLists(oldList []string, newList []string) ([]string, []string) {
	sort.Strings(oldList)
	sort.Strings(newList)
	var x, y int
	var insList, delList []string
	for x < len(oldList) || y < len(newList) {
		switch {
		case y >= len(newList):
			delList = append(delList, oldList[x])
			x++
		case x >= len(oldList):
			insList = append(insList, newList[y])
			y++
		case oldList[x] < newList[y]:
			delList = append(delList, oldList[x])
			x++
		case oldList[x] > newList[y]:
			insList = append(insList, newList[y])
			y++
		default:
			y++
			x++
		}
	}
	return insList, delList
}

func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}
