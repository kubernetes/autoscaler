package paperspace

type Filter struct {
	Where map[string]interface{} `json:"where,omitempty"`
	Limit int64                  `json:"limit,omitempty"`
	Order string                 `json:"order,omitempty"`
}
