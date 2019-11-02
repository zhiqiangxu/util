package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestUUID(t *testing.T) {
	assert.Assert(t, PoorManUUID(true)%2 == 1)
	assert.Assert(t, PoorManUUID(false)%2 == 0)
}
