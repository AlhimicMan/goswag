package wrapper

import (
	"encoding/json"
	"fmt"
)

type ErrorResult struct {
	Status  int
	Message interface{}
}

func (e ErrorResult) Error() string {
	msgRes, err := json.Marshal(e.Message)
	if err != nil {
		return fmt.Sprintf("internal error processing error message: %v", err)
	}
	return fmt.Sprintf("Status: %d, Message: %s", e.Status, string(msgRes))
}
