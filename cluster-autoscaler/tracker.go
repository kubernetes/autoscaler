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

package main

import (
	"time"
)

const (
	maxUsageRecorded = 50
)

// NodeRelation tells which node is related to the given one and when this relation was checked for the last time.
type NodeRelation struct {
	node      string
	timestamp time.Time
}

// UsageRecord records which node was considered helpful to which node during pod rescheduling analysis.
type UsageRecord struct {
	usingTooMany  bool
	using         []NodeRelation
	usedByTooMany bool
	usedBy        []NodeRelation
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
		updated := false
	nodeloop1:
		for _, using := range record.using {
			if using.node == nodeB {
				if using.timestamp.Before(timestamp) {
					using.timestamp = timestamp
				}
				updated = true
				break nodeloop1
			}
		}
		if !updated {
			if len(record.using) >= maxUsageRecorded {
				record.usingTooMany = true
			} else {
				record.using = append(record.using, NodeRelation{nodeB, timestamp})
			}
		}
	} else {
		tracker.usage[nodeA] = &UsageRecord{
			using: []NodeRelation{{nodeB, timestamp}},
		}
	}

	if record, found := tracker.usage[nodeB]; found {
		updated := false
	nodeloop2:
		for _, usedby := range record.usedBy {
			if usedby.node == nodeA {
				if usedby.timestamp.Before(timestamp) {
					usedby.timestamp = timestamp
				}
				updated = true
				break nodeloop2
			}
		}
		if !updated {
			if len(record.usedBy) >= maxUsageRecorded {
				record.usedByTooMany = true
			} else {
				record.usedBy = append(record.usedBy, NodeRelation{nodeA, timestamp})
			}
		}
	} else {
		tracker.usage[nodeB] = &UsageRecord{
			usedBy: []NodeRelation{{nodeA, timestamp}},
		}
	}
}

func filterOutOld(relations []NodeRelation, cutoff time.Time) []NodeRelation {
	result := make([]NodeRelation, 0, len(relations))
	for _, relation := range relations {
		if relation.timestamp.After(cutoff) {
			result = append(result, relation)
		}
	}
	return result
}

// CleanUp removes all relations updated before the cutoff time.
func (tracker *UsageTracker) CleanUp(cutoff time.Time) {
	toDelete := make([]string, 0)
	for key, usageRecord := range tracker.usage {
		if !usageRecord.usingTooMany {
			usageRecord.using = filterOutOld(usageRecord.using, cutoff)
		}
		if !usageRecord.usedByTooMany {
			usageRecord.usedBy = filterOutOld(usageRecord.usedBy, cutoff)
		}
		if !usageRecord.usingTooMany && !usageRecord.usedByTooMany && len(usageRecord.using) == 0 && len(usageRecord.usedBy) == 0 {
			toDelete = append(toDelete, key)
		}
	}
	for _, key := range toDelete {
		delete(tracker.usage, key)
	}
}
