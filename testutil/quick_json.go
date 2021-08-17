package testutil

import (
	"encoding/json"
	"testing"
)

func JSON(t *testing.T, value interface{}) string {
	b, err := json.Marshal(value)
	if err != nil {
		t.Error(err)
	}

	return string(b)
}
