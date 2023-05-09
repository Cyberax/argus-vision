package utils

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var SomeMutableString = NewMutableContextKey[string]("SomeMutableString")
var AnotherMutableString = NewMutableContextKey[string]("AnotherMutableString")

func TestMutableContext(t *testing.T) {
	root := context.Background()

	mc := tryGetMutableContext(root)
	assert.Nil(t, mc)
	assert.Panics(t, func() {
		mustGetMutableContext(root)
	})

	root = ForkMutableContext(root)
	child := ForkMutableContext(root)

	SetMutableContextValue(root, SomeMutableString, "root")
	SetMutableContextValue(root, AnotherMutableString, "anotherRoot")
	SetMutableContextValue(child, SomeMutableString, "child")

	var s string
	s = MustGetMutableContextValue(child, SomeMutableString)
	assert.Equal(t, "child", s)
	s = MustGetMutableContextValue(child, AnotherMutableString)
	assert.Equal(t, "anotherRoot", s)

	s = MustGetMutableContextValue(root, SomeMutableString)
	assert.Equal(t, "root", s)

	EditMutableContextValue(root, SomeMutableString, func(value string, present bool) (string, bool) {
		assert.Equal(t, "root", value)
		assert.True(t, present)
		return "root_2", true
	})
	assert.Equal(t, "root_2", MustGetMutableContextValue(root, SomeMutableString))

	EditMutableContextValue(root, SomeMutableString, func(value string, present bool) (string, bool) {
		assert.True(t, present)
		return "", false
	})

	EditMutableContextValue(root, SomeMutableString, func(value string, present bool) (string, bool) {
		assert.False(t, present)
		return "", false
	})
}
