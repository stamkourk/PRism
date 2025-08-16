package main

import (
	"os"
	"strings"

	"golang.org/x/term"
)

func getTerminalSize() (int, int) {
	width, height := 120, 20
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		if w > 10 {
			width = w - 5
		}
		if h > 5 {
			height = h - 5
		}
	}

	return width, height
}

func getVisibleLineRange(scrollPos, visibleLines, totalLines int) (int, int) {
	start := scrollPos
	if start > totalLines-visibleLines {
		start = totalLines - visibleLines
	}
	if start < 0 {
		start = 0
	}
	end := start + visibleLines
	if end > totalLines {
		end = totalLines
	}

	return start, end
}

func getVisibleLine(line string, scrollCol, width int) string {
	runes := []rune(line)
	col := scrollCol
	if col > len(runes) {
		col = len(runes)
	}
	maxWidth := width
	if maxWidth < 1 {
		maxWidth = 1
	}
	endCol := col + maxWidth
	if endCol > len(runes) {
		endCol = len(runes)
	}

	return string(runes[col:endCol])
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
