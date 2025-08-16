package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (m model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	switch m.state {
	case listView:
		return m.list.View()
	case detailView:
		return m.renderDetailView()
	default:
		return ""
	}
}

func (m model) renderDetailView() string {
	if m.selected == nil {
		return "[No PR selected]\n[Press q or esc to go back]"
	}
	width, height := getTerminalSize()
	header := m.renderDetailHeader()
	footer := m.renderDetailFooter()
	visibleDesc := m.renderVisibleDescription(width, height)
	body := header + visibleDesc

	bodyLines := strings.Count(body, "\n") + 1
	footerLines := 1
	totalLines := bodyLines + footerLines
	if totalLines < height {
		body += strings.Repeat("\n", height-totalLines)
	}

	return body + footer
}

func (m model) renderDetailHeader() string {
	return "\n" + m.selected.pr.Title + "\n\n" + m.selected.pr.URL + "\n\n--- Description ---\n"
}

func (m model) renderDetailFooter() string {
	rawFooter := "↑/k up • ↓/j down • ←/h left • →/l right • gg top • shift+g bottom • q/esc back"

	return list.DefaultStyles().HelpStyle.Render(rawFooter)
}

func (m model) renderVisibleDescription(width, height int) string {
	descLines := splitLines(m.selected.pr.Description)
	visibleLines := height - 8
	if visibleLines < 1 {
		visibleLines = 1
	}
	start, end := getVisibleLineRange(m.scrollPos, visibleLines, len(descLines))
	var visibleDesc []string
	for _, line := range descLines[start:end] {
		visibleDesc = append(visibleDesc, getVisibleLine(line, m.scrollCol, width))
	}

	return strings.Join(visibleDesc, "\n")
}
