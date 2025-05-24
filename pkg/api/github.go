package api

import (
	"context"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/google/go-github/v58/github"
)

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

// Helper function to convert GitHub API types to our domain types
func convertWorkflowRun(run *github.WorkflowRun) types.WorkflowRun {
	workflowRun := types.WorkflowRun{
		ID:         run.GetID(),
		Name:       run.GetName(),
		URL:        run.GetHTMLURL(),
		Status:     run.GetStatus(),
		Conclusion: run.GetConclusion(),
		RunNumber:  run.GetRunNumber(),
		Branch:     run.GetHeadBranch(),
	}

	if run.CreatedAt != nil {
		workflowRun.CreatedAt = run.CreatedAt.Time
	}

	if run.UpdatedAt != nil {
		workflowRun.UpdatedAt = run.UpdatedAt.Time
	}

	return workflowRun
}
