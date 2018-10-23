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

package metadata

import (
	"time"
)

// AttemptStrategy represents a strategy for waiting for an action
// to complete successfully. This is an internal type used by the
// implementation of other packages.
type AttemptStrategy struct {
	Total time.Duration // total duration of attempt.
	Delay time.Duration // interval between each try in the burst.
	Min   int           // minimum number of retries; overrides Total
}

// Attempt struct
type Attempt struct {
	strategy AttemptStrategy
	last     time.Time
	end      time.Time
	force    bool
	count    int
}

// Start begins a new sequence of attempts for the given strategy.
func (s AttemptStrategy) Start() *Attempt {
	now := time.Now()
	return &Attempt{
		strategy: s,
		last:     now,
		end:      now.Add(s.Total),
		force:    true,
	}
}

// Next waits until it is time to perform the next attempt or returns
// false if it is time to stop trying.
func (a *Attempt) Next() bool {
	now := time.Now()
	sleep := a.nextSleep(now)
	if !a.force && !now.Add(sleep).Before(a.end) && a.strategy.Min <= a.count {
		return false
	}
	a.force = false
	if sleep > 0 && a.count > 0 {
		time.Sleep(sleep)
		now = time.Now()
	}
	a.count++
	a.last = now
	return true
}

func (a *Attempt) nextSleep(now time.Time) time.Duration {
	sleep := a.strategy.Delay - now.Sub(a.last)
	if sleep < 0 {
		return 0
	}
	return sleep
}

// HasNext returns whether another attempt will be made if the current
// one fails. If it returns true, the following call to Next is
// guaranteed to return true.
func (a *Attempt) HasNext() bool {
	if a.force || a.strategy.Min > a.count {
		return true
	}
	now := time.Now()
	if now.Add(a.nextSleep(now)).Before(a.end) {
		a.force = true
		return true
	}
	return false
}
