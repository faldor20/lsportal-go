package lsportal

import "testing"

func TestGetOnlyInclusions(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		text           string
		inclusionRegex string
		exclusionRegex string
		expected       string
	}{
		{
			name:           "Simple inclusion",
			text:           "Hello, world! This is a test.",
			inclusionRegex: `(Hello|test)`,
			exclusionRegex: "",
			expected:       "Hello                   test ",
		},
		{
			name:           "Inclusion with newlines",
			text:           "Line 1\nLine 2\nLine 3",
			inclusionRegex: `(Line \d)`,
			exclusionRegex: "",
			expected:       "Line 1\nLine 2\nLine 3",
		},
		{
			name:           "Inclusion and exclusion",
			text:           "~Hello,\n ;world;!~\n This is a test.",
			inclusionRegex: `(~[\s\S]*~)`,
			exclusionRegex: `(;[\s\S]*;)`,
			expected:       "~Hello,\n        !~\n                ",
		},
		{
			name:           "Inclusion and exclusion",
			text:           "~Hello,\n ;world;!~\n This is a test.",
			inclusionRegex: `~([\s\S]*)~`,
			exclusionRegex: `(;[\s\S]*;)`,
			expected:       " Hello,\n        ! \n                ",
		},
		{
			name:           "No matches",
			text:           "Hello, world!",
			inclusionRegex: `(nothing)`,
			exclusionRegex: "",
			expected:       "             ",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := whitespaceExceptInclusions(tc.text, tc.inclusionRegex, tc.exclusionRegex)
			validateChanges(t, tc.text, result)

			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}

func TestGetOnlyInclusionsWithRanges(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		text           string
		inclusionRegex string
		exclusionRegex string
		expected       string
		expectedRanges [][2]int
	}{
		{
			name:           "Simple inclusion",
			text:           "~23456~89",
			inclusionRegex: `~([\s\S]*?)~`,
			exclusionRegex: "",
			expected:       " 23456   ",
			expectedRanges: [][2]int{{1, 6}},
		},
		{
			name:           "Simple inclusion",
			text:           "~!23456~89\n~!123456~",
			inclusionRegex: `~!([\s\S]*?)~`,
			exclusionRegex: "",
			expected:       "  23456   \n  123456 ",
			expectedRanges: [][2]int{{2, 7}, {13, 19}},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, ranges := whitespaceExceptInclusions(tc.text, tc.inclusionRegex, tc.exclusionRegex)
			validateChanges(t, tc.text, result)

			if result != tc.expected {
				t.Errorf("Expected result: %q, Got: %q", tc.expected, result)
			}

			if len(ranges) != len(tc.expectedRanges) {
				t.Errorf("Expected %d ranges, Got %d ranges", len(tc.expectedRanges), len(ranges))
			} else {
				for i, expectedRange := range tc.expectedRanges {
					start, end := ranges[i].IndexesIn([]byte(tc.text))
					actualRanges := [2]int{start, end}

					if actualRanges != expectedRange {
						t.Errorf("Expected range %d: %v, Got: %v", i, expectedRange, ranges[i])
					}
				}
			}
		})
	}
}
