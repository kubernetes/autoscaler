package paperspace

type PaperspaceErrorResponse struct {
	Error *PaperspaceError `json:"error"`
}

type PaperspaceError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e PaperspaceError) Error() string {
	return e.Message
}
