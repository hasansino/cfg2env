package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultilineTag(t *testing.T) {
	tag := StructTag(`
			env:"A"
			 default:"def_value_of_a"
			desc:"just a dummy text"`)

	assert.Equal(t, "A", tag.Get("env"))
	assert.Equal(t, "def_value_of_a", tag.Get("default"))
	assert.Equal(t, "just a dummy text", tag.Get("desc"))

	tag = StructTag(`
	desc:"A is just a dummy value for purpose of this test
	and should not be used as real example, this text is 
	just here for placeholder ... testing testing"`)

	assert.Equal(t, "A is just a dummy value for purpose of this test\n\tand should not be used as real example, this text is \n\tjust here for placeholder ... testing testing", tag.Get("desc"))
}
