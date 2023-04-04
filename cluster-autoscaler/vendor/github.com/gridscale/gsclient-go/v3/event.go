package gsclient

import (
	"context"
	"net/http"
)

// EventOperator provides an interface for operations on events.
type EventOperator interface {
	GetEventList(ctx context.Context) ([]Event, error)
}

// EventList holds a list of events.
type EventList struct {
	// Array of events.
	List []EventProperties `json:"events"`
}

// Event represent a single event.
type Event struct {
	// Properties of an event.
	Properties EventProperties `json:"event"`
}

// EventProperties holds the properties of an event.
type EventProperties struct {
	// Type of object (server, storage, IP) etc.
	ObjectType string `json:"object_type"`

	// The UUID of the event.
	RequestUUID string `json:"request_uuid"`

	// The UUID of the objects the event was executed on.
	ObjectUUID string `json:"object_uuid"`

	// The type of change.
	Activity string `json:"activity"`

	// The type of request.
	RequestType string `json:"request_type"`

	// True or false, whether the request was successful or not.
	RequestStatus string `json:"request_status"`

	// A detailed description of the change.
	Change string `json:"change"`

	// Time the event was triggered.
	Timestamp GSTime `json:"timestamp"`

	// The UUID of the user that triggered the event.
	UserUUID string `json:"user_uuid"`

	// The user that triggered the event.
	// Usually the user's email if the event was triggered by request of a user,
	// otherwise a short descriptive name of the system component responsible.
	Initiator string `json:"initiator"`
}

// GetEventList gets a list of events.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/EventGetAll
func (c *Client) GetEventList(ctx context.Context) ([]Event, error) {
	r := gsRequest{
		uri:                 apiEventBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var events []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		events = append(events, Event{Properties: properties})
	}
	return events, err
}
