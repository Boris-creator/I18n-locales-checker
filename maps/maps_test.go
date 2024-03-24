package maps

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGet(t *testing.T) {
	m := map[string]any{
		"foo": "bar",
		"baz": map[string]string{
			"foo": "123",
		},
	}
	assert.Equal(t, "default", Get[string](m, "bar", "default"))
	assert.Equal(t, "123", Get[string](m, "baz.foo", ""))
}
