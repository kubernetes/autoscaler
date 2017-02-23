/*
Copyright 2017 The Kubernetes Authors.

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

package utils

import (
	"sync"
	"time"
)

// LogLevel describes severity of log entry.
type LogLevel string

const (
	// Debug log level.
	Debug LogLevel = "DEBUG"
	// Info log level.
	Info LogLevel = "INFO"
	// Warning log level.
	Warning LogLevel = "WARNING"
	// Error log level.
	Error LogLevel = "ERROR"
)

// LogItem is a single entry in log managed by LogCollector.
type LogItem struct {
	// Log is the logged message body.
	Log string
	// Level describes log severity.
	Level LogLevel
	// Timestamp when the Log was created.
	Timestamp time.Time
}

const (
	// DefaultLogCollectorMaxItems defines default maximum size of LogCollector.
	DefaultLogCollectorMaxItems = 50
	// DefaultLogCollectorItemLifetime is the default time after which LogItem will be removed from LogCollector.
	DefaultLogCollectorItemLifetime = 15 * time.Minute
)

// LogCollector keeps recent log events. It is automatically truncated on each access based on predefined set of conditions.
type LogCollector struct {
	sync.Mutex
	maxItems     int
	itemLifetime time.Duration
	store        []LogItem
}

// NewLogCollector creates new LogCollector.
func NewLogCollector() *LogCollector {
	return &LogCollector{
		maxItems:     DefaultLogCollectorMaxItems,
		itemLifetime: DefaultLogCollectorItemLifetime,
		store:        make([]LogItem, 0)}
}

// To be executed under a lock.
func (lc *LogCollector) compact(now time.Time) {
	firstIndex := 0
	if len(lc.store) > lc.maxItems {
		firstIndex = len(lc.store) - lc.maxItems
	}
	threshold := now.Add(-lc.itemLifetime)
	for ; firstIndex < len(lc.store); firstIndex++ {
		if lc.store[firstIndex].Timestamp.After(threshold) {
			break
		}
	}
	if firstIndex > 0 {
		updatedStore := make([]LogItem, len(lc.store)-firstIndex)
		copy(updatedStore, lc.store[firstIndex:])
		lc.store = updatedStore
	}
}

// Log logs a single provided message in LogCollector.
func (lc *LogCollector) Log(msg string, level LogLevel) {
	lc.Lock()
	defer lc.Unlock()
	now := time.Now()
	lc.store = append(lc.store, LogItem{Log: msg, Level: level, Timestamp: now})
	lc.compact(now)
}

// GetLogs returns a copy of messages in log. This is an actual copy, so it will not reflect any future changes in log.
func (lc *LogCollector) GetLogs() []LogItem {
	lc.Lock()
	defer lc.Unlock()
	lc.compact(time.Now())
	result := make([]LogItem, len(lc.store))
	copy(result, lc.store)
	return result
}
