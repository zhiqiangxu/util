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

func TestReverseDigits(t *testing.T) {
	assert.Assert(t, ReverseDigits(123) == 321)
	assert.Assert(t, ReverseDigits(-123) == -321)
}

func TestIsPalindrome(t *testing.T) {
	assert.Assert(t, IsPalindrome(121))
	assert.Assert(t, IsPalindrome(1122332211))
	assert.Assert(t, !IsPalindrome(123))
	assert.Assert(t, IsPalindrome(0))
}

func TestPatternMatchAllRec(t *testing.T) {
	assert.Assert(t, !PatternMatchAllRec("ss", "s"))
	assert.Assert(t, PatternMatchAllRec("ss", "s*"))
	assert.Assert(t, PatternMatchAllRec("ss", "s*s"))
	assert.Assert(t, PatternMatchAllRec("ss", ".*s"))
	assert.Assert(t, PatternMatchAllRec("ss", ".*"))
	assert.Assert(t, PatternMatchAllRec("aab", "c*a*b"))
	assert.Assert(t, !PatternMatchAllRec("mississippi", "mis*is*p*."))
	assert.Assert(t, PatternMatchAllRec("", ".*"))
	assert.Assert(t, PatternMatchAllRec("", ".*a*"))

}

func TestPatternMatchAllTD(t *testing.T) {
	assert.Assert(t, !PatternMatchAllTD("ss", "s"))
	assert.Assert(t, PatternMatchAllTD("ss", "s*"))
	assert.Assert(t, PatternMatchAllTD("ss", "s*s"))
	assert.Assert(t, PatternMatchAllTD("ss", ".*s"))
	assert.Assert(t, PatternMatchAllTD("ss", ".*"))
	assert.Assert(t, PatternMatchAllTD("aab", "c*a*b"))
	assert.Assert(t, !PatternMatchAllTD("mississippi", "mis*is*p*."))
	assert.Assert(t, PatternMatchAllTD("", ".*a*"))
}

func TestPatternMatchAllBU(t *testing.T) {
	assert.Assert(t, !PatternMatchAllBU("ss", "s"))
	assert.Assert(t, PatternMatchAllBU("ss", "s*"))
	assert.Assert(t, PatternMatchAllBU("ss", "s*s"))
	assert.Assert(t, PatternMatchAllBU("ss", ".*s"))
	assert.Assert(t, PatternMatchAllBU("ss", ".*"))
	assert.Assert(t, PatternMatchAllBU("aab", "c*a*b"))
	assert.Assert(t, !PatternMatchAllBU("mississippi", "mis*is*p*."))
	assert.Assert(t, PatternMatchAllBU("", ".*a*"))
}

func TestFindOnceNum(t *testing.T) {
	assert.Assert(t, FindOnceNum([]int{102, 101, 102}) == 101)
	assert.Assert(t, FindOnceNum([]int{999, 999, 102}) == 102)
}

func TestMinCoveringSubstr(t *testing.T) {
	assert.Assert(t, MinCoveringSubstr("ADOBECODEBANC", "ABC") == "BANC")
	assert.Assert(t, MinCoveringSubstr("ADOBECODEBAC", "ABC") == "BAC")
	assert.Assert(t, MinCoveringSubstr("", "ABC") == "")
}

func TestLongestConsecutive(t *testing.T) {
	n, len := LongestConsecutive([]int{100, 4, 200, 1, 3, 2})
	assert.Assert(t, n == 1 && len == 4)
}
