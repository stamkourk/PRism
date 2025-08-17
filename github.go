package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type PullRequest struct {
	Title       string `json:"title"`
	URL         string `json:"html_url"`
	Repo        string
	APIURL      string
	Description string
}

type GitHubItem struct {
	pr PullRequest
}

func (i GitHubItem) Title() string       { return i.pr.Title }
func (i GitHubItem) Description() string { return i.pr.URL }
func (i GitHubItem) FilterValue() string { return i.pr.Title }

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
	var prs []PullRequest
	client := &http.Client{}

	perPage := 100
	for page := 1; page <= 10; page++ { // GitHub caps search results at 1000 (10 * 100)
		url := fmt.Sprintf(
			"https://api.github.com/search/issues?q=review-requested:%s+type:pr+state:open&per_page=%d&page=%d",
			user, perPage, page,
		)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "token "+token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

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

		if len(result.Items) == 0 {
			break // no more pages
		}

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
			})
		}

		// Stop early if less than a full page
		if len(result.Items) < perPage {
			break
		}
	}

	return prs, nil
}

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
