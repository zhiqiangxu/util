package math

import (
	"testing"

	"gotest.tools/assert"
)

func TestDivCount(t *testing.T) {
	assert.Equal(t, DivCount(10), 4)
	assert.Equal(t, DivCount(144), 15)

	assert.Equal(t, AbelGroups(72).Uint64(), uint64(6))
}
