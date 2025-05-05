// pkg/cmd/generate.go
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/go-i2p/go-github-dashboard/pkg/api"
	"github.com/go-i2p/go-github-dashboard/pkg/config"
	"github.com/go-i2p/go-github-dashboard/pkg/generator"
	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the GitHub dashboard",
	Long:  `Fetches GitHub repository data and generates a static dashboard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate() error {
	// Get configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error with configuration: %w", err)
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interruptions gracefully
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Initialize the cache
	cache := api.NewCache(cfg)

	// Initialize clients
	githubClient := api.NewGitHubClient(cfg, cache)
	rssClient := api.NewRSSClient(cfg, cache)

	// Initialize generators
	mdGenerator, err := generator.NewMarkdownGenerator(cfg)
	if err != nil {
		return fmt.Errorf("error creating markdown generator: %w", err)
	}

	htmlGenerator, err := generator.NewHTMLGenerator(cfg)
	if err != nil {
		return fmt.Errorf("error creating HTML generator: %w", err)
	}

	// Create dashboard data structure
	dashboard := types.Dashboard{
		Username:     cfg.User,
		Organization: cfg.Organization,
		GeneratedAt:  time.Now(),
	}

	// Fetch repositories
	fmt.Println("Fetching repositories...")
	repositories, err := githubClient.GetRepositories(ctx)
	if err != nil {
		return fmt.Errorf("error fetching repositories: %w", err)
	}

	fmt.Printf("Found %d repositories\n", len(repositories))

	// Create a wait group for parallel processing
	var wg sync.WaitGroup
	reposChan := make(chan types.Repository, len(repositories))

	// Process each repository in parallel
	for i := range repositories {
		wg.Add(1)
		go func(repo types.Repository) {
			defer wg.Done()

			// Check if context is canceled
			if ctx.Err() != nil {
				return
			}

			owner := repo.Owner
			repoName := repo.Name

			fmt.Printf("Processing repository: %s/%s\n", owner, repoName)

			// Fetch pull requests from RSS first, fall back to API
			pullRequests, err := rssClient.GetPullRequestsFromRSS(ctx, owner, repoName)
			if err != nil || len(pullRequests) == 0 {
				if cfg.Verbose && err != nil {
					log.Printf("Error fetching pull requests from RSS for %s/%s: %v, falling back to API", owner, repoName, err)
				}
				pullRequests, err = githubClient.GetPullRequests(ctx, owner, repoName)
				if err != nil {
					log.Printf("Error fetching pull requests for %s/%s: %v", owner, repoName, err)
				}
			}

			// Fetch issues from RSS first, fall back to API
			issues, err := rssClient.GetIssuesFromRSS(ctx, owner, repoName)
			if err != nil || len(issues) == 0 {
				if cfg.Verbose && err != nil {
					log.Printf("Error fetching issues from RSS for %s/%s: %v, falling back to API", owner, repoName, err)
				}
				issues, err = githubClient.GetIssues(ctx, owner, repoName)
				if err != nil {
					log.Printf("Error fetching issues for %s/%s: %v", owner, repoName, err)
				}
			}

			// Fetch discussions (only available through RSS)
			discussions, err := rssClient.GetDiscussionsFromRSS(ctx, owner, repoName)
			if err != nil {
				log.Printf("Error fetching discussions for %s/%s: %v", owner, repoName, err)
			}

			// Update the repository with the fetched data
			repo.PullRequests = pullRequests
			repo.Issues = issues
			repo.Discussions = discussions

			// Send the updated repository to the channel
			reposChan <- repo
		}(repositories[i])
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(reposChan)
	}()

	// Collect results
	var processedRepos []types.Repository
	for repo := range reposChan {
		processedRepos = append(processedRepos, repo)
	}

	// Sort repositories by name (for consistent output)
	dashboard.Repositories = processedRepos

	// Count totals
	for _, repo := range dashboard.Repositories {
		dashboard.TotalPRs += len(repo.PullRequests)
		dashboard.TotalIssues += len(repo.Issues)
		dashboard.TotalDiscussions += len(repo.Discussions)
	}

	// Generate markdown files
	fmt.Println("Generating markdown files...")
	markdownPaths, err := mdGenerator.GenerateAllRepositoriesMarkdown(dashboard)
	if err != nil {
		return fmt.Errorf("error generating markdown files: %w", err)
	}

	// Convert markdown to HTML
	fmt.Println("Converting markdown to HTML...")
	_, err = htmlGenerator.ConvertAllMarkdownToHTML(markdownPaths)
	if err != nil {
		return fmt.Errorf("error converting markdown to HTML: %w", err)
	}

	// Generate the main HTML dashboard
	fmt.Println("Generating HTML dashboard...")
	err = htmlGenerator.GenerateHTML(dashboard)
	if err != nil {
		return fmt.Errorf("error generating HTML dashboard: %w", err)
	}

	// Create a README in the output directory
	readmePath := filepath.Join(cfg.OutputDir, "README.md")
	var targetName string
	if cfg.User != "" {
		targetName = "@" + cfg.User
	} else {
		targetName = cfg.Organization
	}

	readme := fmt.Sprintf("# GitHub Dashboard\n\nThis dashboard was generated for %s on %s.\n\n",
		targetName,
		dashboard.GeneratedAt.Format("January 2, 2006"))
	readme += fmt.Sprintf("- Total repositories: %d\n", len(dashboard.Repositories))
	readme += fmt.Sprintf("- Total open pull requests: %d\n", dashboard.TotalPRs)
	readme += fmt.Sprintf("- Total open issues: %d\n", dashboard.TotalIssues)
	readme += fmt.Sprintf("- Total recent discussions: %d\n\n", dashboard.TotalDiscussions)
	readme += "To view the dashboard, open `index.html` in your browser.\n"

	err = os.WriteFile(readmePath, []byte(readme), 0644)
	if err != nil {
		log.Printf("Error writing README file: %v", err)
	}

	fmt.Printf("\nDashboard generated successfully in %s\n", cfg.OutputDir)
	fmt.Printf("Open %s/index.html in your browser to view the dashboard\n", cfg.OutputDir)

	return nil
}
