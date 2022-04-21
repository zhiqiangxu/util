package math

import (
	"testing"

	"gotest.tools/assert"
)

func TestLambda(t *testing.T) {
	assert.Equal(t, Lambda(35), 12)
	assert.Equal(t, Lambda(34), 16)
	assert.Equal(t, Lambda(33), 10)
	assert.Equal(t, Lambda(16), 4)
}
