package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	listView viewState = iota
	detailView
)

type model struct {
	list      list.Model
	err       error
	username  string
	state     viewState
	selected  *GitHubItem
	scrollPos int
	scrollCol int
	gPressed  bool
}

func initialModel(items []list.Item, username string) model {
	width, height := getTerminalSize()
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "Pull Requests requesting a review from @" + username
	return model{list: l, username: username, state: listView}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case listView:
		return m.updateListView(msg)
	case detailView:
		return m.updateDetailView(msg)
	default:
		return m, nil
	}
}

func (m model) updateListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if sel, ok := m.list.SelectedItem().(GitHubItem); ok {
				m.selected = &sel
				m.state = detailView
				m.scrollPos = 0
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateDetailView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleDetailKeyMsg(msg)
	default:
		m.gPressed = false
	}
	return m, nil
}

func (m model) handleDetailKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = listView
	case "j", "down":
		m.scrollPos++
	case "k", "up":
		if m.scrollPos > 0 {
			m.scrollPos--
		}
	case "h", "left":
		if m.scrollCol > 0 {
			m.scrollCol--
		}
	case "l", "right":
		m.scrollCol++
	case "G":
		m.scrollToBottom()
	case "g":
		if m.gPressed {
			m.scrollPos = 0
		} else {
			m.gPressed = true

			return m, nil
		}
	}

	m.gPressed = false

	return m, nil
}

func (m *model) scrollToBottom() {
	if m.selected != nil {
		descLines := splitLines(m.selected.pr.Description)
		_, height := getTerminalSize()
		visibleLines := height - 8
		if visibleLines < 1 {
			visibleLines = 1
		}
		m.scrollPos = len(descLines) - visibleLines
		if m.scrollPos < 0 {
			m.scrollPos = 0
		}
	}
}
