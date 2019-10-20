package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestMaxLengthOfUniqueSubslice(t *testing.T) {
	{
		s1 := []byte("abcdee")

		assert.Assert(t, MaxLengthOfUniqueSubslice(s1) == 5)
	}

	{
		s2 := []byte("abcdaef")

		assert.Assert(t, MaxLengthOfUniqueSubslice(s2) == 6)
	}
}
