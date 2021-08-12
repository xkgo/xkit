package xconv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToBool(t *testing.T) {

	v, err := ToBool("True")
	assert.Nil(t, err)
	assert.True(t, v)

}
