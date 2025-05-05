package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/google/go-github/v58/github"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

// GitHubClient wraps the GitHub API client with additional functionality
type GitHubClient struct {
	client      *github.Client
	cache       *Cache
	rateLimited bool
	config      *types.Config
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(config *types.Config, cache *Cache) *GitHubClient {
	var httpClient *http.Client

	// Create a retry client
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = nil // Disable logging from the retry client

	if config.GithubToken != "" {
		// If token is provided, use it for authentication
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GithubToken},
		)
		httpClient = oauth2.NewClient(context.Background(), ts)
		retryClient.HTTPClient = httpClient
	}

	client := github.NewClient(retryClient.StandardClient())

	return &GitHubClient{
		client:      client,
		cache:       cache,
		rateLimited: false,
		config:      config,
	}
}

// GetRepositories fetches repositories for a user or organization
func (g *GitHubClient) GetRepositories(ctx context.Context) ([]types.Repository, error) {
	var allRepos []types.Repository

	cacheKey := "repos_"
	if g.config.User != "" {
		cacheKey += g.config.User
	} else {
		cacheKey += g.config.Organization
	}

	// Try to get from cache first
	if cachedRepos, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Println("Using cached repositories")
		}
		return cachedRepos.([]types.Repository), nil
	}

	if g.config.Verbose {
		log.Println("Fetching repositories from GitHub API")
	}

	for {
		if g.config.User != "" {
			opts := &github.RepositoryListOptions{
				ListOptions: github.ListOptions{PerPage: 100},
				Sort:        "updated",
			}
			repos, resp, err := g.client.Repositories.List(ctx, g.config.User, opts)
			if err != nil {
				return nil, fmt.Errorf("error fetching repositories: %w", err)
			}

			for _, repo := range repos {
				allRepos = append(allRepos, convertRepository(repo))
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		} else {
			opts := &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{PerPage: 100},
				Sort:        "updated",
			}
			repos, resp, err := g.client.Repositories.ListByOrg(ctx, g.config.Organization, opts)
			if err != nil {
				return nil, fmt.Errorf("error fetching repositories: %w", err)
			}

			for _, repo := range repos {
				allRepos = append(allRepos, convertRepository(repo))
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}

	// Cache the results
	g.cache.Set(cacheKey, allRepos)

	return allRepos, nil
}

// GetPullRequests fetches open pull requests for a repository
func (g *GitHubClient) GetPullRequests(ctx context.Context, owner, repo string) ([]types.PullRequest, error) {
	var allPRs []types.PullRequest
	cacheKey := fmt.Sprintf("prs_%s_%s", owner, repo)

	// Try to get from cache first
	if cachedPRs, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Printf("Using cached pull requests for %s/%s", owner, repo)
		}
		return cachedPRs.([]types.PullRequest), nil
	}

	if g.config.Verbose {
		log.Printf("Fetching pull requests for %s/%s", owner, repo)
	}

	opts := &github.PullRequestListOptions{
		State:       "open",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		prs, resp, err := g.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("error fetching pull requests: %w", err)
		}

		for _, pr := range prs {
			allPRs = append(allPRs, convertPullRequest(pr))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Cache the results
	g.cache.Set(cacheKey, allPRs)

	return allPRs, nil
}

// GetIssues fetches open issues for a repository
func (g *GitHubClient) GetIssues(ctx context.Context, owner, repo string) ([]types.Issue, error) {
	var allIssues []types.Issue
	cacheKey := fmt.Sprintf("issues_%s_%s", owner, repo)

	// Try to get from cache first
	if cachedIssues, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Printf("Using cached issues for %s/%s", owner, repo)
		}
		return cachedIssues.([]types.Issue), nil
	}

	if g.config.Verbose {
		log.Printf("Fetching issues for %s/%s", owner, repo)
	}

	opts := &github.IssueListByRepoOptions{
		State:       "open",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		issues, resp, err := g.client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("error fetching issues: %w", err)
		}

		for _, issue := range issues {
			// Skip pull requests (they appear in the issues API)
			if issue.PullRequestLinks != nil {
				continue
			}
			allIssues = append(allIssues, convertIssue(issue))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Cache the results
	g.cache.Set(cacheKey, allIssues)

	return allIssues, nil
}

// GetDiscussions fetches recent discussions for a repository
func (g *GitHubClient) GetDiscussions(ctx context.Context, owner, repo string) ([]types.Discussion, error) {
	// Note: The GitHub API v3 doesn't have a direct endpoint for discussions
	// We'll simulate this functionality by retrieving discussions via RSS feed
	// This will be implemented in the RSS parser
	return []types.Discussion{}, nil
}

// Helper functions to convert GitHub API types to our domain types
func convertRepository(repo *github.Repository) types.Repository {
	r := types.Repository{
		Name:        repo.GetName(),
		FullName:    repo.GetFullName(),
		Description: repo.GetDescription(),
		URL:         repo.GetHTMLURL(),
		Owner:       repo.GetOwner().GetLogin(),
		Stars:       repo.GetStargazersCount(),
		Forks:       repo.GetForksCount(),
	}

	if repo.UpdatedAt != nil {
		r.LastUpdated = repo.UpdatedAt.Time
	}

	return r
}

func convertPullRequest(pr *github.PullRequest) types.PullRequest {
	pullRequest := types.PullRequest{
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		URL:       pr.GetHTMLURL(),
		Author:    pr.GetUser().GetLogin(),
		AuthorURL: pr.GetUser().GetHTMLURL(),
		Status:    pr.GetState(),
	}

	if pr.CreatedAt != nil {
		pullRequest.CreatedAt = pr.CreatedAt.Time
	}

	if pr.UpdatedAt != nil {
		pullRequest.UpdatedAt = pr.UpdatedAt.Time
	}

	for _, label := range pr.Labels {
		pullRequest.Labels = append(pullRequest.Labels, types.Label{
			Name:  label.GetName(),
			Color: label.GetColor(),
		})
	}

	return pullRequest
}

func convertIssue(issue *github.Issue) types.Issue {
	i := types.Issue{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		URL:       issue.GetHTMLURL(),
		Author:    issue.GetUser().GetLogin(),
		AuthorURL: issue.GetUser().GetHTMLURL(),
	}

	if issue.CreatedAt != nil {
		i.CreatedAt = issue.CreatedAt.Time
	}

	if issue.UpdatedAt != nil {
		i.UpdatedAt = issue.UpdatedAt.Time
	}

	for _, label := range issue.Labels {
		i.Labels = append(i.Labels, types.Label{
			Name:  label.GetName(),
			Color: label.GetColor(),
		})
	}

	return i
}
