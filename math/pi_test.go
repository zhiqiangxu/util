package math

import (
	"testing"

	"gotest.tools/assert"
)

func TestPi(t *testing.T) {
	assert.Equal(t, Pi(5), 3)
	assert.Equal(t, Pi(4), 2)
	assert.Equal(t, Pi(6), 3)
}
