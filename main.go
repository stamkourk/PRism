package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

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
