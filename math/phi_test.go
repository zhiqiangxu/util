package math

import (
	"testing"

	"gotest.tools/assert"
)

func TestPhi(t *testing.T) {
	assert.Equal(t, Phi(10), 4)
	assert.Equal(t, Phi(11), 10)
	assert.Equal(t, Phi(12), 4)
	assert.Equal(t, Phi(14), 6)
	assert.Equal(t, Phi(15), 8)
	assert.Equal(t, Phi(229), 228)
	assert.Equal(t, Phi(242), 110)
}
