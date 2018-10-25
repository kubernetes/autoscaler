package debug

import (
	"encoding/json"
	"fmt"
)

func Print(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		return fmt.Sprintf("error-marshalling[%v]", err)
	}
	return string(b)
}
