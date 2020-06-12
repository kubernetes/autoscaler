/*
Copyright 2020 The Kubernetes Authors.

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

package egoscale

import (
	"context"
	"fmt"
	"time"
)

//RunstatusEvent is a runstatus event
type RunstatusEvent struct {
	Created *time.Time `json:"created,omitempty"`
	State   string     `json:"state,omitempty"`
	Status  string     `json:"status"`
	Text    string     `json:"text"`
}

// UpdateRunstatusIncident create runstatus incident event
// Events can be updates or final message with status completed.
func (client *Client) UpdateRunstatusIncident(ctx context.Context, incident RunstatusIncident, event RunstatusEvent) error {
	if incident.EventsURL == "" {
		return fmt.Errorf("empty Events URL for %#v", incident)
	}

	_, err := client.runstatusRequest(ctx, incident.EventsURL, event, "POST")
	return err
}

// UpdateRunstatusMaintenance adds a event to a maintenance.
// Events can be updates or final message with status completed.
func (client *Client) UpdateRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance, event RunstatusEvent) error {
	if maintenance.EventsURL == "" {
		return fmt.Errorf("empty Events URL for %#v", maintenance)
	}

	_, err := client.runstatusRequest(ctx, maintenance.EventsURL, event, "POST")
	return err
}
