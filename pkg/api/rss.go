package api

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/mmcdole/gofeed"
)

// RSSClient handles fetching data from GitHub RSS feeds
type RSSClient struct {
	parser *gofeed.Parser
	cache  *Cache
	config *types.Config
}

// NewRSSClient creates a new RSS client
func NewRSSClient(config *types.Config, cache *Cache) *RSSClient {
	return &RSSClient{
		parser: gofeed.NewParser(),
		cache:  cache,
		config: config,
	}
}

// GetDiscussionsFromRSS fetches recent discussions from GitHub RSS feed
func (r *RSSClient) GetDiscussionsFromRSS(ctx context.Context, owner, repo string) ([]types.Discussion, error) {
	var discussions []types.Discussion
	cacheKey := fmt.Sprintf("discussions_rss_%s_%s", owner, repo)

	// Try to get from cache first
	if cachedDiscussions, found := r.cache.Get(cacheKey); found {
		if r.config.Verbose {
			log.Printf("Using cached discussions for %s/%s", owner, repo)
		}
		return cachedDiscussions.([]types.Discussion), nil
	}

	if r.config.Verbose {
		log.Printf("Fetching discussions from RSS for %s/%s", owner, repo)
	}

	// GitHub discussions RSS feed URL
	feedURL := fmt.Sprintf("https://github.com/%s/%s/discussions.atom", owner, repo)

	feed, err := r.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		// If we can't fetch the feed, just return an empty slice rather than failing
		if r.config.Verbose {
			log.Printf("Error fetching discussions feed for %s/%s: %v", owner, repo, err)
		}
		return []types.Discussion{}, nil
	}

	cutoffDate := time.Now().AddDate(0, 0, -30) // Last 30 days

	for _, item := range feed.Items {
		// Skip items older than 30 days
		if item.PublishedParsed != nil && item.PublishedParsed.Before(cutoffDate) {
			continue
		}

		// Parse the author info
		author := ""
		authorURL := ""
		if item.Author != nil {
			author = item.Author.Name
			// Extract the GitHub username from the author name if possible
			if strings.Contains(author, "(") {
				parts := strings.Split(author, "(")
				if len(parts) > 1 {
					username := strings.TrimSuffix(strings.TrimPrefix(parts[1], "@"), ")")
					author = username
					authorURL = fmt.Sprintf("https://github.com/%s", username)
				}
			}
		}

		// Parse the category
		category := "Discussion"
		if len(item.Categories) > 0 {
			category = item.Categories[0]
		}

		discussion := types.Discussion{
			Title:     item.Title,
			URL:       item.Link,
			Author:    author,
			AuthorURL: authorURL,
			Category:  category,
		}

		if item.PublishedParsed != nil {
			discussion.CreatedAt = *item.PublishedParsed
		}

		if item.UpdatedParsed != nil {
			discussion.LastUpdated = *item.UpdatedParsed
		} else if item.PublishedParsed != nil {
			discussion.LastUpdated = *item.PublishedParsed
		}

		discussions = append(discussions, discussion)
	}

	// Cache the results
	r.cache.Set(cacheKey, discussions)

	return discussions, nil
}

// GetIssuesFromRSS fetches recent issues from GitHub RSS feed
func (r *RSSClient) GetIssuesFromRSS(ctx context.Context, owner, repo string) ([]types.Issue, error) {
	var issues []types.Issue
	cacheKey := fmt.Sprintf("issues_rss_%s_%s", owner, repo)

	// Try to get from cache first
	if cachedIssues, found := r.cache.Get(cacheKey); found {
		if r.config.Verbose {
			log.Printf("Using cached issues from RSS for %s/%s", owner, repo)
		}
		return cachedIssues.([]types.Issue), nil
	}

	if r.config.Verbose {
		log.Printf("Fetching issues from RSS for %s/%s", owner, repo)
	}

	// GitHub issues RSS feed URL
	feedURL := fmt.Sprintf("https://github.com/%s/%s/issues.atom", owner, repo)

	feed, err := r.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		// If we can't fetch the feed, just return an empty slice
		if r.config.Verbose {
			log.Printf("Error fetching issues feed for %s/%s: %v", owner, repo, err)
		}
		return []types.Issue{}, nil
	}

	for _, item := range feed.Items {
		// Skip pull requests (they appear in the issues feed)
		if strings.Contains(item.Link, "/pull/") {
			continue
		}

		// Parse the issue number from the URL
		number := 0
		parts := strings.Split(item.Link, "/issues/")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &number)
		}

		// Parse the author info
		author := ""
		authorURL := ""
		if item.Author != nil {
			author = item.Author.Name
			// Extract the GitHub username from the author name if possible
			if strings.Contains(author, "(") {
				parts := strings.Split(author, "(")
				if len(parts) > 1 {
					username := strings.TrimSuffix(strings.TrimPrefix(parts[1], "@"), ")")
					author = username
					authorURL = fmt.Sprintf("https://github.com/%s", username)
				}
			}
		}

		issue := types.Issue{
			Number:    number,
			Title:     item.Title,
			URL:       item.Link,
			Author:    author,
			AuthorURL: authorURL,
		}

		if item.PublishedParsed != nil {
			issue.CreatedAt = *item.PublishedParsed
		}

		if item.UpdatedParsed != nil {
			issue.UpdatedAt = *item.UpdatedParsed
		} else if item.PublishedParsed != nil {
			issue.UpdatedAt = *item.PublishedParsed
		}

		// Note: RSS doesn't include labels, so we'll have an empty labels slice

		issues = append(issues, issue)
	}

	// Cache the results
	r.cache.Set(cacheKey, issues)

	return issues, nil
}

// GetPullRequestsFromRSS fetches recent pull requests from GitHub RSS feed
func (r *RSSClient) GetPullRequestsFromRSS(ctx context.Context, owner, repo string) ([]types.PullRequest, error) {
	var pullRequests []types.PullRequest
	cacheKey := fmt.Sprintf("prs_rss_%s_%s", owner, repo)

	// Try to get from cache first
	if cachedPRs, found := r.cache.Get(cacheKey); found {
		if r.config.Verbose {
			log.Printf("Using cached pull requests from RSS for %s/%s", owner, repo)
		}
		return cachedPRs.([]types.PullRequest), nil
	}

	if r.config.Verbose {
		log.Printf("Fetching pull requests from RSS for %s/%s", owner, repo)
	}

	// GitHub pull requests RSS feed URL
	feedURL := fmt.Sprintf("https://github.com/%s/%s/pulls.atom", owner, repo)

	feed, err := r.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		// If we can't fetch the feed, just return an empty slice
		if r.config.Verbose {
			log.Printf("Error fetching pull requests feed for %s/%s: %v", owner, repo, err)
		}
		return []types.PullRequest{}, nil
	}

	for _, item := range feed.Items {
		// Parse the PR number from the URL
		number := 0
		parts := strings.Split(item.Link, "/pull/")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &number)
		}

		// Parse the author info
		author := ""
		authorURL := ""
		if item.Author != nil {
			author = item.Author.Name
			// Extract the GitHub username from the author name if possible
			if strings.Contains(author, "(") {
				parts := strings.Split(author, "(")
				if len(parts) > 1 {
					username := strings.TrimSuffix(strings.TrimPrefix(parts[1], "@"), ")")
					author = username
					authorURL = fmt.Sprintf("https://github.com/%s", username)
				}
			}
		}

		pr := types.PullRequest{
			Number:    number,
			Title:     item.Title,
			URL:       item.Link,
			Author:    author,
			AuthorURL: authorURL,
			Status:    "open", // All items in the feed are open
		}

		if item.PublishedParsed != nil {
			pr.CreatedAt = *item.PublishedParsed
		}

		if item.UpdatedParsed != nil {
			pr.UpdatedAt = *item.UpdatedParsed
		} else if item.PublishedParsed != nil {
			pr.UpdatedAt = *item.PublishedParsed
		}

		// Note: RSS doesn't include labels, so we'll have an empty labels slice

		pullRequests = append(pullRequests, pr)
	}

	// Cache the results
	r.cache.Set(cacheKey, pullRequests)

	return pullRequests, nil
}
