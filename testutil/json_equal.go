package testutil

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
)

func JSONDiffs(t *testing.T, expected, actual string) []string {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(expected), &o1)
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal([]byte(actual), &o2)
	if err != nil {
		t.Error(err)
	}

	return deep.Equal(o1, o2)
}
