package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DisableFilteringTeamPRs bool `yaml:"disableFilteringTeamPRs"`
}

func main() {
	// Load config.yaml from current directory
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: could not load config.yaml: %v (using defaults)", err)
	}

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

	skipped := 0
	if !cfg.DisableFilteringTeamPRs {
		prs, skipped, err = filterDirectReviewRequests(prs, username, token)
		if err != nil {
			log.Fatal("Error filtering direct review requests: ", err)
		}
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

func loadConfig(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return cfg, err
}
