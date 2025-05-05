# Go GitHub Dashboard

A pure Go command-line application that generates a static GitHub dashboard by aggregating repository data from GitHub API and RSS feeds.

## Features

- Generate a comprehensive dashboard for any GitHub user or organization
- View open pull requests, issues, and recent discussions for each repository
- Clean, responsive UI with collapsible sections (no JavaScript required)
- Hybrid approach using both GitHub API and RSS feeds for efficient data collection
- Intelligent caching to reduce API calls and handle rate limits
- Fully static output that can be deployed to any static hosting service

## Installation

### Using Go

```bash
go install github.com/yourusername/go-github-dashboard/cmd/go-github-dashboard@latest
```

### From Source

```bash
git clone https://github.com/yourusername/go-github-dashboard.git
cd go-github-dashboard
go build ./cmd/go-github-dashboard
```

## Usage

```bash
# Generate dashboard for a user
go-github-dashboard generate --user octocat --output ./dashboard

# Generate dashboard for an organization
go-github-dashboard generate --org kubernetes --output ./k8s-dashboard

# Use with authentication token (recommended for large organizations)
go-github-dashboard generate --user developername --token $GITHUB_TOKEN --output ./my-dashboard

# Specify cache duration
go-github-dashboard generate --user octocat --cache-ttl 2h --output ./dashboard

# Show version information
go-github-dashboard version
```

### Command-line Options

- `--user` or `-u`: GitHub username to generate dashboard for
- `--org` or `-o`: GitHub organization to generate dashboard for
- `--output` or `-d`: Output directory for the dashboard (default: `./dashboard`)
- `--token` or `-t`: GitHub API token (optional, increases rate limits)
- `--cache-dir`: Directory for caching API responses (default: `./.cache`)
- `--cache-ttl`: Cache time-to-live duration (default: `1h`)
- `--verbose` or `-v`: Enable verbose output

### Environment Variables

You can also set configuration using environment variables:

- `GITHUB_DASHBOARD_USER`: GitHub username
- `GITHUB_DASHBOARD_ORG`: GitHub organization
- `GITHUB_DASHBOARD_OUTPUT`: Output directory
- `GITHUB_DASHBOARD_TOKEN` or `GITHUB_TOKEN`: GitHub API token
- `GITHUB_DASHBOARD_CACHE_DIR`: Cache directory
- `GITHUB_DASHBOARD_CACHE_TTL`: Cache TTL duration
- `GITHUB_DASHBOARD_VERBOSE`: Enable verbose output (set to `true`)

## Output Structure

The generated dashboard follows this structure:

```
output/
├── index.html            # Main HTML dashboard
├── style.css             # CSS styling
├── README.md             # Dashboard information
├── repositories/         # Directory containing markdown files
│   ├── repo1.md          # Markdown version of repository data
│   ├── repo1.html        # HTML version of repository data
│   ├── repo2.md
│   ├── repo2.html
│   └── ...
```

## Development

The project is structured as follows:

```
go-github-dashboard/
├── cmd/
│   └── go-github-dashboard/
│       └── main.go       # Application entry point
├── pkg/
│   ├── api/              # API clients for GitHub and RSS
│   │   ├── github.go     # GitHub API client
│   │   ├── rss.go        # RSS feed parser
│   │   └── cache.go      # Caching implementation
│   ├── cmd/              # Cobra commands
│   │   ├── root.go       # Root command and flag definition
│   │   ├── generate.go   # Generate dashboard command
│   │   └── version.go    # Version information command
│   ├── config/           # Configuration handling with Viper
│   │   └── config.go     # Config validation and processing
│   ├── generator/        # HTML and markdown generators
│   │   ├── markdown.go   # Markdown file generation
│   │   └── html.go       # HTML dashboard generation
│   └── types/            # Type definitions
│       └── types.go      # Core data structures
└── README.md
```

## Key Features

- **Concurrent Repository Processing**: Uses Go's concurrency features to fetch and process multiple repositories in parallel
- **Intelligent Data Sourcing**: Prioritizes RSS feeds to avoid API rate limits, falling back to the GitHub API when needed
- **Resilient API Handling**: Implements retries, timeouts, and graceful error handling for network requests
- **Efficient Caching**: Stores API responses to reduce duplicate requests and speed up subsequent runs
- **Pure Go Implementation**: Uses only Go standard library and well-maintained third-party packages
- **No JavaScript Requirement**: Dashboard uses CSS-only techniques for interactivity (collapsible sections)
- **Clean Architecture**: Separation of concerns with distinct packages for different responsibilities

## Example Dashboard

Once generated, the dashboard presents:

- An overview of all repositories with key metrics
- Expandable sections for each repository showing:
  - Open pull requests with author and label information
  - Open issues organized by priority and label
  - Recent discussions categorized by topic
- Responsive design that works on both desktop and mobile browsers
- Both HTML and Markdown versions of all content

## License

[MIT License](LICENSE)