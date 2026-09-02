package upcloud

import "encoding/json"

// Label represents a key-value pair label in a response.
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// LabelSlice is a slice of Labels
// It exists to allow for a custom unmarshaller.
type LabelSlice []Label

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (ls *LabelSlice) UnmarshalJSON(b []byte) error {
	wrapper := struct {
		Labels []Label `json:"label"`
	}{}
	err := json.Unmarshal(b, &wrapper)
	if err != nil {
		return err
	}

	(*ls) = wrapper.Labels

	return nil
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (ls LabelSlice) MarshalJSON() ([]byte, error) {
	wrapper := struct {
		Labels []Label `json:"label"`
	}{}
	if ls == nil {
		ls = make(LabelSlice, 0)
	}
	wrapper.Labels = ls

	return json.Marshal(wrapper)
}
