/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package simulator

import (
	"time"
)

const (
	maxUsageRecorded = 50
)

// UsageRecord records which node was considered helpful to which node during pod rescheduling analysis.
type UsageRecord struct {
	usingTooMany  bool
	using         map[string]time.Time
	usedByTooMany bool
	usedBy        map[string]time.Time
}

// UsageTracker track usage relationship between nodes in pod rescheduling calculations.
type UsageTracker struct {
	usage map[string]*UsageRecord
}

// NewUsageTracker builds new usage tracker.
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{
		usage: make(map[string]*UsageRecord),
	}
}

// Get gets the given node UsageRecord, if present
func (tracker *UsageTracker) Get(node string) (data *UsageRecord, found bool) {
	data, found = tracker.usage[node]
	return data, found
}

// RegisterUsage registers that node A uses nodeB during usage calculations at time timestamp.
func (tracker *UsageTracker) RegisterUsage(nodeA string, nodeB string, timestamp time.Time) {
	if record, found := tracker.usage[nodeA]; found {
		if len(record.using) >= maxUsageRecorded {
			record.usingTooMany = true
		} else {
			record.using[nodeB] = timestamp
		}
	} else {
		record := UsageRecord{
			using:  make(map[string]time.Time),
			usedBy: make(map[string]time.Time),
		}
		record.using[nodeB] = timestamp
		tracker.usage[nodeA] = &record
	}

	if record, found := tracker.usage[nodeB]; found {
		if len(record.usedBy) >= maxUsageRecorded {
			record.usedByTooMany = true
		} else {
			record.usedBy[nodeA] = timestamp
		}
	} else {
		record := UsageRecord{
			using:  make(map[string]time.Time),
			usedBy: make(map[string]time.Time),
		}
		record.usedBy[nodeA] = timestamp
		tracker.usage[nodeB] = &record
	}
}

// Unregister removes the given node from all usage records
func (tracker *UsageTracker) Unregister(node string) {
	if record, found := tracker.usage[node]; found {
		for using := range record.using {
			if record2, found := tracker.usage[using]; found {
				delete(record2.usedBy, node)
			}
		}
		for usedBy := range record.usedBy {
			if record2, found := tracker.usage[usedBy]; found {
				delete(record2.using, node)
			}
		}
		delete(tracker.usage, node)
	}
}

func filterOutOld(timestampMap map[string]time.Time, cutoff time.Time) {
	toRemove := make([]string, 0)
	for key, timestamp := range timestampMap {
		if timestamp.Before(cutoff) {
			toRemove = append(toRemove, key)
		}
	}
	for _, key := range toRemove {
		delete(timestampMap, key)
	}
}

// CleanUp removes all relations updated before the cutoff time.
func (tracker *UsageTracker) CleanUp(cutoff time.Time) {
	toDelete := make([]string, 0)
	for key, usageRecord := range tracker.usage {
		if !usageRecord.usingTooMany {
			filterOutOld(usageRecord.using, cutoff)
		}
		if !usageRecord.usedByTooMany {
			filterOutOld(usageRecord.usedBy, cutoff)
		}
		if !usageRecord.usingTooMany && !usageRecord.usedByTooMany && len(usageRecord.using) == 0 && len(usageRecord.usedBy) == 0 {
			toDelete = append(toDelete, key)
		}
	}
	for _, key := range toDelete {
		delete(tracker.usage, key)
	}
}

// RemoveNodeFromTracker removes node from tracker and also cleans the passed utilization map.
func RemoveNodeFromTracker(tracker *UsageTracker, node string, utilization map[string]time.Time) {
	keysToRemove := make([]string, 0)
	if mainRecord, found := tracker.Get(node); found {
		if mainRecord.usingTooMany {
			keysToRemove = getAllKeys(utilization)
		} else {
		usingloop:
			for usedNode := range mainRecord.using {
				if usedNodeRecord, found := tracker.Get(usedNode); found {
					if usedNodeRecord.usedByTooMany {
						keysToRemove = getAllKeys(utilization)
						break usingloop
					} else {
						for anotherNode := range usedNodeRecord.usedBy {
							keysToRemove = append(keysToRemove, anotherNode)
						}
					}
				}
			}
		}
	}
	tracker.Unregister(node)
	delete(utilization, node)
	for _, key := range keysToRemove {
		delete(utilization, key)
	}
}

func getAllKeys(m map[string]time.Time) []string {
	result := make([]string, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	return result
}
