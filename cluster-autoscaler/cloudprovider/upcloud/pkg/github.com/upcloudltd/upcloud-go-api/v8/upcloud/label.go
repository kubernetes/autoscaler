package upcloud

// Label represents a key-value pair label in a response.
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
