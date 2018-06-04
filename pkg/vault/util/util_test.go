package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveImageTag(t *testing.T) {
	data := []struct {
		name     string
		image    string
		expected string
	}{
		{
			"with tag",
			"a:2324",
			"a",
		},
		{
			"without tag",
			"a",
			"a",
		},
	}

	for _, test := range data {
		t.Run(test.name, func(t *testing.T) {
			resp := RemoveImageTag(test.image)
			assert.Equal(t, test.expected, resp)
		})
	}
}
