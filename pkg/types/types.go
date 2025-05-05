package types

import "time"

// Repository represents a GitHub repository with its basic information
type Repository struct {
	Name        string
	FullName    string
	Description string
	URL         string
	Owner       string
	Stars       int
	Forks       int
	LastUpdated time.Time

	PullRequests      []PullRequest
	Issues            []Issue
	Discussions       []Discussion
	WorkflowRuns      []WorkflowRun
	TotalWorkflowRuns int
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int
	Title     string
	URL       string
	Author    string
	AuthorURL string
	CreatedAt time.Time
	UpdatedAt time.Time
	Labels    []Label
	Status    string
}

// Issue represents a GitHub issue
type Issue struct {
	Number    int
	Title     string
	URL       string
	Author    string
	AuthorURL string
	CreatedAt time.Time
	UpdatedAt time.Time
	Labels    []Label
}

// Discussion represents a GitHub discussion
type Discussion struct {
	Title       string
	URL         string
	Author      string
	AuthorURL   string
	CreatedAt   time.Time
	LastUpdated time.Time
	Category    string
}

// Label represents a GitHub label
type Label struct {
	Name  string
	Color string
}

// Dashboard represents the entire dashboard data
type Dashboard struct {
	Username          string
	Organization      string
	GeneratedAt       time.Time
	Repositories      []Repository
	TotalPRs          int
	TotalIssues       int
	TotalDiscussions  int
	TotalWorkflowRuns int
}

// Config holds the application configuration
type Config struct {
	User         string
	Organization string
	OutputDir    string
	GithubToken  string
	CacheDir     string
	CacheTTL     time.Duration
	Verbose      bool
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID         int64
	Name       string
	URL        string
	Status     string // "completed", "in_progress", "queued"
	Conclusion string // "success", "failure", "cancelled", "skipped", etc.
	CreatedAt  time.Time
	UpdatedAt  time.Time
	RunNumber  int
	Branch     string
}
