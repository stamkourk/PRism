package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type PullRequest struct {
	Title       string `json:"title"`
	URL         string `json:"html_url"`
	Repo        string
	APIURL      string // Add API URL for getting PR details
	Description string // PR description
}

type GitHubItem struct {
	pr PullRequest
}

func (i GitHubItem) Title() string       { return i.pr.Title }
func (i GitHubItem) Description() string { return i.pr.URL }
func (i GitHubItem) FilterValue() string { return i.pr.Title }

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
	scrollPos int  // for scrolling in detail view (vertical)
	scrollCol int  // for horizontal scrolling in detail view
	gPressed  bool // for detecting double-g (gg)
}

func getCurrentUser(token string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Login, nil
}

func getReviewRequests(user, token string) ([]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/search/issues?q=review-requested:%s+type:pr+state:open", user)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Title   string `json:"title"`
			HTMLURL string `json:"html_url"`
			RepoURL string `json:"repository_url"`
			PullReq *struct {
				URL string `json:"url"`
			} `json:"pull_request"`
			Number int `json:"number"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var prs []PullRequest
	for _, item := range result.Items {
		if item.PullReq == nil {
			continue // skip issues that are not PRs
		}
		repoPath := item.RepoURL[len("https://api.github.com/repos/"):]
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d", repoPath, item.Number)
		repo := repoPath[strings.LastIndex(repoPath, "/")+1:]
		prs = append(prs, PullRequest{
			Title:  item.Title,
			URL:    item.HTMLURL,
			Repo:   repo,
			APIURL: apiURL,
			// Description will be filled in filterDirectReviewRequests
		})
	}
	return prs, nil
}

// filterDirectReviewRequests filters PRs to only those where the user is directly requested for review.
// Returns the filtered PRs and the number of PRs that could not be fetched.
func filterDirectReviewRequests(prs []PullRequest, user, token string) ([]PullRequest, int, error) {
	var filtered []PullRequest
	var skipped int
	for _, pr := range prs {
		req, _ := http.NewRequest("GET", pr.APIURL, nil)
		req.Header.Set("Authorization", "token "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			skipped++
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			skipped++
			continue
		}

		var prDetail struct {
			Body               string `json:"body"`
			RequestedReviewers []struct {
				Login string `json:"login"`
			} `json:"requested_reviewers"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&prDetail); err != nil {
			skipped++
			continue
		}
		for _, reviewer := range prDetail.RequestedReviewers {
			if reviewer.Login == user {
				pr.Description = prDetail.Body
				filtered = append(filtered, pr)
				break
			}
		}
	}
	return filtered, skipped, nil
}

func initialModel(items []list.Item, username string) model {
	// Dynamically set list width and height to terminal size minus some overhead
	width, height := 120, 20
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		if w > 10 {
			width = w - 5
		}
		if h > 5 {
			height = h - 5
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = fmt.Sprintf("Pull Requests requesting a review from @%s", username)

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
				m.scrollPos = 0 // reset scroll position
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
		m.gPressed = false
		return m, nil
	case "j", "down":
		m.scrollPos++
		m.gPressed = false
		return m, nil
	case "k", "up":
		if m.scrollPos > 0 {
			m.scrollPos--
		}
		m.gPressed = false
		return m, nil
	case "h", "left":
		if m.scrollCol > 0 {
			m.scrollCol--
		}
		m.gPressed = false
		return m, nil
	case "l", "right":
		m.scrollCol++
		m.gPressed = false
		return m, nil
	case "G":
		m.scrollToBottom()
		m.gPressed = false
		return m, nil
	case "g":
		if m.gPressed {
			m.scrollPos = 0
			m.gPressed = false
			return m, nil
		}
		m.gPressed = true
		return m, nil
	default:
		m.gPressed = false
	}
	return m, nil
}

func (m *model) scrollToBottom() {
	if m.selected != nil {
		descLines := strings.Split(m.selected.pr.Description, "\n")
		_, height := 120, 20
		if _, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			if h > 5 {
				height = h - 5
			}
		}
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

	// Pad with blank lines so the footer is always at the bottom
	bodyLines := strings.Count(body, "\n") + 1
	footerLines := 1
	totalLines := bodyLines + footerLines
	if totalLines < height {
		body += strings.Repeat("\n", height-totalLines)
	}

	return body + footer
}

func (m model) renderDetailHeader() string {
	return fmt.Sprintf("\n%s\n\n%s\n\n--- Description ---\n", m.selected.pr.Title, m.selected.pr.URL)
}

func (m model) renderDetailFooter() string {
	rawFooter := "↑/k up • ↓/j down • ←/h left • →/l right • gg top • shift+g bottom • q/esc back"
	return list.DefaultStyles().HelpStyle.Render(rawFooter)
}

func (m model) renderVisibleDescription(width, height int) string {
	descLines := strings.Split(m.selected.pr.Description, "\n")
	visibleLines := height - 8 // header + footer + padding
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

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Error: GH_TOKEN environment variable is not set")
	}

	username, err := getCurrentUser(token)
	if err != nil {
		log.Fatal("Error fetching current user: ", err)
	}

	prs, err := getReviewRequests(username, token)
	if err != nil {
		log.Fatal("Error fetching review requests: ", err)
	}

	prs, skipped, err := filterDirectReviewRequests(prs, username, token)
	if err != nil {
		log.Fatal("Error filtering direct review requests: ", err)
	}

	if len(prs) == 0 {
		fmt.Println("✅ No PRs need your review!")
		if skipped > 0 {
			fmt.Printf("⚠️  Could not fetch details for %d PR(s)\n", skipped)
		}
		return
	}

	var items []list.Item
	for _, pr := range prs {
		items = append(items, GitHubItem{pr})
	}

	if skipped > 0 {
		fmt.Printf("⚠️  Could not fetch details for %d PR(s)\n", skipped)
	}

	p := tea.NewProgram(initialModel(items, username))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
