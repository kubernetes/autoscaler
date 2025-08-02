package testing

import "strings"

// errors provides batching errors
type errors []error

func (es errors) Error() string {
	var b strings.Builder
	for _, e := range es {
		b.WriteString(e.Error())
	}

	return b.String()
}
