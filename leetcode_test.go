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

func TestManacherFallback(t *testing.T) {
	assert.Assert(t, ManacherFallback("") == "")
	assert.Assert(t, ManacherFallback("abcd") == "a")
	assert.Assert(t, ManacherFallback("babcd") == "bab")
	assert.Assert(t, ManacherFallback("cbabcd") == "cbabc")
	assert.Assert(t, ManacherFallback("cbaabcd") == "cbaabc")

	assert.Assert(t, ManacherWithFallback("") == "")
	assert.Assert(t, ManacherWithFallback("abcd") == "a")
	assert.Assert(t, ManacherWithFallback("babcd") == "bab")
	assert.Assert(t, ManacherWithFallback("cbabcd") == "cbabc")
	assert.Assert(t, ManacherWithFallback("cbaabcd") == "cbaabc")
}
