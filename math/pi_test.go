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

func TestR2(t *testing.T) {
	assert.Equal(t, R2(5), 1)
	assert.Equal(t, R2(6), 1)
}
