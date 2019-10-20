package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestMaxLengthOfUniqueSubslice(t *testing.T) {
	{
		s1 := "abcdee"

		assert.Assert(t, MaxLengthOfUniqueSubstring(s1) == 5)
	}

	{
		s2 := "abcdaef"

		assert.Assert(t, MaxLengthOfUniqueSubstring(s2) == 6)
	}
}
