package main

import "testing"

func TestGetVisibleLineRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		scrollPos     int
		visibleLines  int
		totalLines    int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "should succeed with normal range",
			scrollPos:     2,
			visibleLines:  3,
			totalLines:    10,
			expectedStart: 2,
			expectedEnd:   5,
		},
		{
			name:          "should handle scroll at end",
			scrollPos:     9,
			visibleLines:  3,
			totalLines:    10,
			expectedStart: 7,
			expectedEnd:   10,
		},
		{
			name:          "should handle visibleLines > totalLines",
			scrollPos:     0,
			visibleLines:  5,
			totalLines:    3,
			expectedStart: 0,
			expectedEnd:   3,
		},
		{
			name:          "should handle scroll near end",
			scrollPos:     5,
			visibleLines:  2,
			totalLines:    6,
			expectedStart: 4,
			expectedEnd:   6,
		},
		{
			name:          "should handle negative scroll",
			scrollPos:     -2,
			visibleLines:  3,
			totalLines:    10,
			expectedStart: 0,
			expectedEnd:   3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start, end := getVisibleLineRange(tt.scrollPos, tt.visibleLines, tt.totalLines)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf(
					"getVisibleLineRange(scrollPos=%d, visibleLines=%d, totalLines=%d) = (%d,%d), expected (%d,%d)",
					tt.scrollPos, tt.visibleLines, tt.totalLines, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}

func TestGetVisibleLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		line      string
		scrollCol int
		width     int
		expected  string
	}{
		{
			name:      "should return middle slice",
			line:      "abcdef",
			scrollCol: 2,
			width:     3,
			expected:  "cde",
		},
		{
			name:      "should return start slice",
			line:      "abcdef",
			scrollCol: 0,
			width:     2,
			expected:  "ab",
		},
		{
			name:      "should return empty when scrollCol past end",
			line:      "abcdef",
			scrollCol: 10,
			width:     3,
			expected:  "",
		},
		{
			name:      "should return wide at end",
			line:      "abcdef",
			scrollCol: 4,
			width:     10,
			expected:  "ef",
		},
		{
			name:      "should handle empty line",
			line:      "",
			scrollCol: 0,
			width:     5,
			expected:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getVisibleLine(tt.line, tt.scrollCol, tt.width)
			if got != tt.expected {
				t.Errorf(
					"getVisibleLine(line=%q, scrollCol=%d, width=%d) = %q, expected %q",
					tt.line, tt.scrollCol, tt.width, got, tt.expected)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "should split multi-line string",
			input:    "a\nb\nc",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "should handle empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "should handle single line",
			input:    "one",
			expected: []string{"one"},
		},
		{
			name:     "should split two lines",
			input:    "x\ny",
			expected: []string{"x", "y"},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitLines(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitLines(%q) length = %d, expected %d", tt.input, len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitLines(%q)[%d] = %q, expected %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}
