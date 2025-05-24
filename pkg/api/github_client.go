package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/google/go-github/v58/github"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

// GitHubClient wraps the GitHub API client with additional functionality
type GitHubClient struct {
	client      *github.Client
	cache       *Cache
	config      *types.Config
	rateLimiter *RateLimiter
}

// RateLimiter manages GitHub API rate limiting
type RateLimiter struct {
	remaining   int
	resetTime   time.Time
	lastChecked time.Time
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(config *types.Config, cache *Cache) *GitHubClient {
	var httpClient *http.Client

	// Create a retry client with custom retry policy
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Logger = nil

	// Custom retry policy for rate limiting
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil && resp.StatusCode == 403 {
			// Check if it's a rate limit error
			if resp.Header.Get("X-RateLimit-Remaining") == "0" {
				resetTime := resp.Header.Get("X-RateLimit-Reset")
				if resetTime != "" {
					if resetTimestamp, parseErr := strconv.ParseInt(resetTime, 10, 64); parseErr == nil {
						waitTime := time.Unix(resetTimestamp, 0).Sub(time.Now())
						if waitTime > 0 && waitTime < 1*time.Hour {
							log.Printf("Rate limit exceeded. Waiting %v until reset...", waitTime)
							time.Sleep(waitTime + 5*time.Second) // Add 5 seconds buffer
							return true, nil
						}
					}
				}
			}
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	if config.GithubToken != "" {
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
		config:      config,
		rateLimiter: &RateLimiter{},
	}
}

// checkRateLimit updates rate limit information from response headers
func (g *GitHubClient) checkRateLimit(resp *github.Response) {
	if resp != nil && resp.Response != nil {
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
			if val, err := strconv.Atoi(remaining); err == nil {
				g.rateLimiter.remaining = val
			}
		}

		if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
			if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
				g.rateLimiter.resetTime = time.Unix(val, 0)
			}
		}

		g.rateLimiter.lastChecked = time.Now()

		if g.config.Verbose {
			log.Printf("Rate limit remaining: %d, resets at: %v",
				g.rateLimiter.remaining, g.rateLimiter.resetTime)
		}
	}
}

// waitForRateLimit waits if we're close to hitting rate limits
func (g *GitHubClient) waitForRateLimit(ctx context.Context) error {
	// If we have less than 100 requests remaining, wait for reset
	if g.rateLimiter.remaining < 100 && !g.rateLimiter.resetTime.IsZero() {
		waitTime := time.Until(g.rateLimiter.resetTime)
		if waitTime > 0 && waitTime < 1*time.Hour {
			log.Printf("Approaching rate limit (%d remaining). Waiting %v for reset...",
				g.rateLimiter.remaining, waitTime)

			select {
			case <-time.After(waitTime + 5*time.Second):
				log.Println("Rate limit reset. Continuing...")
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// GetRepositories fetches repositories for a user or organization with pagination
func (g *GitHubClient) GetRepositories(ctx context.Context) ([]types.Repository, error) {
	var allRepos []types.Repository

	cacheKey := "repos_"
	if g.config.User != "" {
		cacheKey += g.config.User
	} else {
		cacheKey += g.config.Organization
	}

	if cachedRepos, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Println("Using cached repositories")
		}
		return cachedRepos.([]types.Repository), nil
	}

	if g.config.Verbose {
		log.Println("Fetching repositories from GitHub API")
	}

	page := 1
	for {
		// Check rate limit before making request
		if err := g.waitForRateLimit(ctx); err != nil {
			return nil, err
		}

		var repos []*github.Repository
		var resp *github.Response
		var err error

		if g.config.User != "" {
			opts := &github.RepositoryListOptions{
				ListOptions: github.ListOptions{PerPage: 100, Page: page},
				Sort:        "updated",
			}
			repos, resp, err = g.client.Repositories.List(ctx, g.config.User, opts)
		} else {
			opts := &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{PerPage: 100, Page: page},
				Sort:        "updated",
			}
			repos, resp, err = g.client.Repositories.ListByOrg(ctx, g.config.Organization, opts)
		}

		if err != nil {
			return nil, fmt.Errorf("error fetching repositories (page %d): %w", page, err)
		}

		g.checkRateLimit(resp)

		for _, repo := range repos {
			allRepos = append(allRepos, convertRepository(repo))
		}

		if g.config.Verbose {
			log.Printf("Fetched page %d with %d repositories", page, len(repos))
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage

		// Add a small delay between requests to be respectful
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	g.cache.Set(cacheKey, allRepos)
	return allRepos, nil
}

// GetPullRequests fetches open pull requests with enhanced pagination
func (g *GitHubClient) GetPullRequests(ctx context.Context, owner, repo string) ([]types.PullRequest, error) {
	var allPRs []types.PullRequest
	cacheKey := fmt.Sprintf("prs_%s_%s", owner, repo)

	if cachedPRs, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Printf("Using cached pull requests for %s/%s", owner, repo)
		}
		return cachedPRs.([]types.PullRequest), nil
	}

	if g.config.Verbose {
		log.Printf("Fetching pull requests for %s/%s", owner, repo)
	}

	page := 1
	for {
		if err := g.waitForRateLimit(ctx); err != nil {
			return nil, err
		}

		opts := &github.PullRequestListOptions{
			State:       "open",
			Sort:        "updated",
			Direction:   "desc",
			ListOptions: github.ListOptions{PerPage: 100, Page: page},
		}

		prs, resp, err := g.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("error fetching pull requests (page %d): %w", page, err)
		}

		g.checkRateLimit(resp)

		for _, pr := range prs {
			allPRs = append(allPRs, convertPullRequest(pr))
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage

		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	g.cache.Set(cacheKey, allPRs)
	return allPRs, nil
}

// GetIssues fetches open issues with enhanced pagination
func (g *GitHubClient) GetIssues(ctx context.Context, owner, repo string) ([]types.Issue, error) {
	var allIssues []types.Issue
	cacheKey := fmt.Sprintf("issues_%s_%s", owner, repo)

	if cachedIssues, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Printf("Using cached issues for %s/%s", owner, repo)
		}
		return cachedIssues.([]types.Issue), nil
	}

	if g.config.Verbose {
		log.Printf("Fetching issues for %s/%s", owner, repo)
	}

	page := 1
	for {
		if err := g.waitForRateLimit(ctx); err != nil {
			return nil, err
		}

		opts := &github.IssueListByRepoOptions{
			State:       "open",
			Sort:        "updated",
			Direction:   "desc",
			ListOptions: github.ListOptions{PerPage: 100, Page: page},
		}

		issues, resp, err := g.client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("error fetching issues (page %d): %w", page, err)
		}

		g.checkRateLimit(resp)

		for _, issue := range issues {
			if issue.PullRequestLinks != nil {
				continue
			}
			allIssues = append(allIssues, convertIssue(issue))
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage

		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	g.cache.Set(cacheKey, allIssues)
	return allIssues, nil
}

// GetWorkflowRuns fetches recent workflow runs with rate limiting
func (g *GitHubClient) GetWorkflowRuns(ctx context.Context, owner, repo string) ([]types.WorkflowRun, error) {
	var allRuns []types.WorkflowRun
	cacheKey := fmt.Sprintf("workflow_runs_%s_%s", owner, repo)

	if cachedRuns, found := g.cache.Get(cacheKey); found {
		if g.config.Verbose {
			log.Printf("Using cached workflow runs for %s/%s", owner, repo)
		}
		return cachedRuns.([]types.WorkflowRun), nil
	}

	if g.config.Verbose {
		log.Printf("Fetching workflow runs for %s/%s", owner, repo)
	}

	if err := g.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	runs, resp, err := g.client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching workflow runs: %w", err)
	}

	g.checkRateLimit(resp)

	for _, run := range runs.WorkflowRuns {
		allRuns = append(allRuns, convertWorkflowRun(run))
	}

	g.cache.Set(cacheKey, allRuns)
	return allRuns, nil
}
