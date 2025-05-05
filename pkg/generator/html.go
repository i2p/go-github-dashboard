package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/russross/blackfriday/v2"
)

// HTMLGenerator handles the generation of HTML files
type HTMLGenerator struct {
	outputDir string
	template  *template.Template
	verbose   bool
}

// NewHTMLGenerator creates a new HTMLGenerator
func NewHTMLGenerator(config *types.Config) (*HTMLGenerator, error) {
	// Create the template
	indexTmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GitHub Dashboard {{if .Username}}for @{{.Username}}{{else}}for {{.Organization}}{{end}}</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <header>
        <h1>GitHub Dashboard {{if .Username}}for <a href="https://github.com/{{.Username}}">@{{.Username}}</a>{{else}}for <a href="https://github.com/{{.Organization}}">{{.Organization}}</a>{{end}}</h1>
        <div class="dashboard-stats">
            <span>{{len .Repositories}} repositories</span>
            <span>{{.TotalPRs}} open pull requests</span>
            <span>{{.TotalIssues}} open issues</span>
            <span>{{.TotalDiscussions}} recent discussions</span>
			<span>{{.TotalWorkflowRuns}} recent workflow runs</span>
        </div>
        <p class="generated-at">Generated on {{.GeneratedAt.Format "January 2, 2006 at 15:04"}}</p>
    </header>

    <main>
        <div class="repositories">
            <h2>Repositories</h2>
            
            {{range .Repositories}}
            <div class="repository">
                <div class="collapsible">
                    <input type="checkbox" id="repo-{{.Name}}" class="toggle">
                    <label for="repo-{{.Name}}" class="toggle-label">
                        <span class="repo-name">{{.Name}}</span>
                        <div class="repo-stats">
                            <span class="stat">{{len .PullRequests}} PRs</span>
                            <span class="stat">{{len .Issues}} issues</span>
                            <span class="stat">{{len .Discussions}} discussions</span>
							<span class="stat">{{len .WorkflowRuns}} workflows</span>
                        </div>
                    </label>
                    <div class="collapsible-content">
                        <div class="repo-details">
                            <p class="repo-description">{{if .Description}}{{.Description}}{{else}}No description provided.{{end}}</p>
                            <div class="repo-meta">
                                <a href="{{.URL}}" target="_blank">View on GitHub</a>
                                <span>⭐ {{.Stars}}</span>
                                <span>🍴 {{.Forks}}</span>
                                <span>Updated: {{.LastUpdated.Format "2006-01-02"}}</span>
                            </div>
                        </div>

                        {{if .PullRequests}}
                        <div class="collapsible">
                            <input type="checkbox" id="prs-{{.Name}}" class="toggle">
                            <label for="prs-{{.Name}}" class="toggle-label section-label pr-label">
                                Open Pull Requests ({{len .PullRequests}})
                            </label>
                            <div class="collapsible-content">
                                <table class="data-table">
                                    <thead>
                                        <tr>
                                            <th>Title</th>
                                            <th>Author</th>
                                            <th>Updated</th>
                                            <th>Labels</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {{range .PullRequests}}
                                        <tr>
                                            <td><a href="{{.URL}}" target="_blank">{{.Title}}</a></td>
                                            <td><a href="{{.AuthorURL}}" target="_blank">@{{.Author}}</a></td>
                                            <td>{{.UpdatedAt.Format "2006-01-02"}}</td>
                                            <td>{{range $i, $label := .Labels}}{{if $i}}, {{end}}{{$label.Name}}{{else}}<em>none</em>{{end}}</td>
                                        </tr>
                                        {{end}}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        {{end}}

                        {{if .Issues}}
                        <div class="collapsible">
                            <input type="checkbox" id="issues-{{.Name}}" class="toggle">
                            <label for="issues-{{.Name}}" class="toggle-label section-label issue-label">
                                Open Issues ({{len .Issues}})
                            </label>
                            <div class="collapsible-content">
                                <table class="data-table">
                                    <thead>
                                        <tr>
                                            <th>Title</th>
                                            <th>Author</th>
                                            <th>Updated</th>
                                            <th>Labels</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {{range .Issues}}
                                        <tr>
                                            <td><a href="{{.URL}}" target="_blank">{{.Title}}</a></td>
                                            <td><a href="{{.AuthorURL}}" target="_blank">@{{.Author}}</a></td>
                                            <td>{{.UpdatedAt.Format "2006-01-02"}}</td>
                                            <td>{{range $i, $label := .Labels}}{{if $i}}, {{end}}{{$label.Name}}{{else}}<em>none</em>{{end}}</td>
                                        </tr>
                                        {{end}}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        {{end}}

                        {{if .Discussions}}
                        <div class="collapsible">
                            <input type="checkbox" id="discussions-{{.Name}}" class="toggle">
                            <label for="discussions-{{.Name}}" class="toggle-label section-label discussion-label">
                                Recent Discussions ({{len .Discussions}})
                            </label>
                            <div class="collapsible-content">
                                <table class="data-table">
                                    <thead>
                                        <tr>
                                            <th>Title</th>
                                            <th>Started By</th>
                                            <th>Last Activity</th>
                                            <th>Category</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {{range .Discussions}}
                                        <tr>
                                            <td><a href="{{.URL}}" target="_blank">{{.Title}}</a></td>
                                            <td><a href="{{.AuthorURL}}" target="_blank">@{{.Author}}</a></td>
                                            <td>{{.LastUpdated.Format "2006-01-02"}}</td>
                                            <td>{{.Category}}</td>
                                        </tr>
                                        {{end}}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        {{end}}
						{{if .WorkflowRuns}}
							<div class="collapsible">
								<input type="checkbox" id="workflows-{{.Name}}" class="toggle">
								<label for="workflows-{{.Name}}" class="toggle-label section-label workflow-label">
									Recent Workflow Runs ({{len .WorkflowRuns}})
								</label>
								<div class="collapsible-content">
									<table class="data-table">
										<thead>
											<tr>
												<th>Workflow</th>
												<th>Branch</th>
												<th>Status</th>
												<th>Run #</th>
												<th>Created</th>
											</tr>
										</thead>
										<tbody>
											{{range .WorkflowRuns}}
											<tr>
												<td><a href="{{.URL}}" target="_blank">{{.Name}}</a></td>
												<td>{{.Branch}}</td>
												<td class="workflow-status workflow-status-{{.Status}} workflow-conclusion-{{.Conclusion}}">
													{{if eq .Status "completed"}}
														{{if eq .Conclusion "success"}}✅ Success{{end}}
														{{if eq .Conclusion "failure"}}❌ Failure{{end}}
														{{if eq .Conclusion "cancelled"}}⚪ Cancelled{{end}}
														{{if eq .Conclusion "skipped"}}⏭️ Skipped{{end}}
														{{if eq .Conclusion "timed_out"}}⏱️ Timed Out{{end}}
														{{if eq .Conclusion ""}}⚪ {{.Status}}{{end}}
													{{else}}
														{{if eq .Status "in_progress"}}🔄 In Progress{{end}}
														{{if eq .Status "queued"}}⏳ Queued{{end}}
														{{if eq .Status ""}}⚪ Unknown{{end}}
													{{end}}
												</td>
												<td>{{.RunNumber}}</td>
												<td>{{.CreatedAt.Format "2006-01-02 15:04"}}</td>
											</tr>
											{{end}}
										</tbody>
									</table>
								</div>
							</div>
						{{end}}

                        <div class="repo-links">
                            <a href="repositories/{{.Name}}.md">View as Markdown</a>
                        </div>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
    </main>

    <footer>
        <p>Generated with <a href="https://github.com/yourusername/go-github-dashboard">go-github-dashboard</a></p>
    </footer>
</body>
</html>`

	tmpl, err := template.New("index.html").Parse(indexTmpl)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML template: %w", err)
	}

	return &HTMLGenerator{
		outputDir: config.OutputDir,
		template:  tmpl,
		verbose:   config.Verbose,
	}, nil
}

// GenerateCSS generates the CSS file for the dashboard
func (g *HTMLGenerator) GenerateCSS() error {
	css := `/* Base styles */
:root {
    --primary-color: #0366d6;
    --secondary-color: #586069;
    --background-color: #ffffff;
    --border-color: #e1e4e8;
    --pr-color: #28a745;
    --issue-color: #d73a49;
    --discussion-color: #6f42c1;
    --hover-color: #f6f8fa;
    --font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
}

* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: var(--font-family);
    line-height: 1.5;
    color: #24292e;
    background-color: var(--background-color);
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
}

/* Header styles */
header {
    margin-bottom: 30px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border-color);
}

header h1 {
    margin-bottom: 10px;
}

.dashboard-stats {
    display: flex;
    flex-wrap: wrap;
    gap: 15px;
    margin-bottom: 10px;
}

.dashboard-stats span {
    background-color: #f1f8ff;
    border-radius: 20px;
    padding: 5px 12px;
    font-size: 14px;
}

.generated-at {
    font-size: 14px;
    color: var(--secondary-color);
}

/* Repository styles */
.repositories {
    margin-bottom: 30px;
}

.repositories h2 {
    margin-bottom: 20px;
}

.repository {
    margin-bottom: 15px;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    overflow: hidden;
}

.repo-details {
    padding: 15px;
    border-bottom: 1px solid var(--border-color);
}

.repo-description {
    margin-bottom: 10px;
}

.repo-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 15px;
    font-size: 14px;
    color: var(--secondary-color);
}

.repo-links {
    padding: 10px 15px;
    font-size: 14px;
    border-top: 1px solid var(--border-color);
}

/* Collapsible sections */
.collapsible {
    width: 100%;
}

.toggle {
    position: absolute;
    opacity: 0;
    z-index: -1;
}

.toggle-label {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 15px;
    font-weight: 600;
    cursor: pointer;
    background-color: #f6f8fa;
    position: relative;
}

.section-label {
    border-top: 1px solid var(--border-color);
    font-weight: 500;
}

.pr-label {
    color: var(--pr-color);
}

.issue-label {
    color: var(--issue-color);
}

.discussion-label {
    color: var(--discussion-color);
}

.toggle-label::after {
    content: '+';
    font-size: 18px;
    transition: transform 0.3s ease;
}

.toggle:checked ~ .toggle-label::after {
    content: '−';
}

.collapsible-content {
    max-height: 0;
    overflow: hidden;
    transition: max-height 0.35s ease;
}

.toggle:checked ~ .collapsible-content {
    max-height: 100vh;
}

/* Table styles */
.data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
}

.data-table th,
.data-table td {
    padding: 8px 15px;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
}

.data-table th {
    background-color: #f6f8fa;
    font-weight: 600;
}

.data-table tr:hover {
    background-color: var(--hover-color);
}

/* Links */
a {
    color: var(--primary-color);
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

/* Repository name and stats */
.repo-name {
    font-size: 16px;
}

.repo-stats {
    display: flex;
    gap: 10px;
}

.stat {
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 12px;
    background-color: #f1f8ff;
    color: var(--primary-color);
}

/* Footer */
footer {
    margin-top: 40px;
    padding-top: 20px;
    border-top: 1px solid var(--border-color);
    font-size: 14px;
    color: var(--secondary-color);
    text-align: center;
}

/* Responsive adjustments */
@media (max-width: 768px) {
    .toggle-label {
        flex-direction: column;
        align-items: flex-start;
        gap: 5px;
    }
    
    .repo-stats {
        align-self: flex-start;
    }
    
    .data-table {
        display: block;
        overflow-x: auto;
    }
    
    .dashboard-stats {
        flex-direction: column;
        align-items: flex-start;
        gap: 5px;
    }

	.workflow-label {
		color: #2088ff;
	}

	.workflow-status {
		font-weight: 500;
	}

	.workflow-status-completed.workflow-conclusion-success {
		color: #22863a;
	}

	.workflow-status-completed.workflow-conclusion-failure {
		color: #cb2431;
	}

	.workflow-status-in_progress {
		color: #dbab09;
	}

	.workflow-status-queued {
		color: #6f42c1;
	}
}`

	err := os.WriteFile(filepath.Join(g.outputDir, "style.css"), []byte(css), 0644)
	if err != nil {
		return fmt.Errorf("error writing CSS file: %w", err)
	}

	return nil
}

// GenerateHTML generates the main HTML dashboard
func (g *HTMLGenerator) GenerateHTML(dashboard types.Dashboard) error {
	if g.verbose {
		log.Println("Generating HTML dashboard")
	}

	// Render the template
	var buf bytes.Buffer
	err := g.template.Execute(&buf, dashboard)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write the file
	outputPath := filepath.Join(g.outputDir, "index.html")
	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing HTML file: %w", err)
	}

	// Generate the CSS file
	err = g.GenerateCSS()
	if err != nil {
		return err
	}

	return nil
}

// ConvertMarkdownToHTML converts a markdown file to HTML
func (g *HTMLGenerator) ConvertMarkdownToHTML(markdownPath string) (string, error) {
	if g.verbose {
		log.Printf("Converting markdown to HTML: %s", markdownPath)
	}

	// Read the markdown file
	markdownContent, err := os.ReadFile(markdownPath)
	if err != nil {
		return "", fmt.Errorf("error reading markdown file: %w", err)
	}

	// Convert the markdown to HTML
	htmlContent := blackfriday.Run(markdownContent)

	// Determine the output filename
	baseName := filepath.Base(markdownPath)
	htmlFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".html"
	htmlPath := filepath.Join(g.outputDir, "repositories", htmlFileName)

	// Create a simple HTML wrapper
	htmlPage := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="../style.css">
    <style>
        .markdown-body {
            padding: 20px;
            max-width: 1000px;
            margin: 0 auto;
        }
        .markdown-body table {
            width: 100%%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        .markdown-body th, .markdown-body td {
            padding: 8px 15px;
            text-align: left;
            border: 1px solid var(--border-color);
        }
        .markdown-body th {
            background-color: #f6f8fa;
        }
        .back-link {
            display: inline-block;
            margin: 20px;
        }
    </style>
</head>
<body>
    <a href="../index.html" class="back-link">← Back to Dashboard</a>
    <div class="markdown-body">
        %s
    </div>
</body>
</html>`, strings.TrimSuffix(baseName, filepath.Ext(baseName)), string(htmlContent))

	// Write the HTML file
	err = os.WriteFile(htmlPath, []byte(htmlPage), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing HTML file: %w", err)
	}

	return htmlPath, nil
}

// ConvertAllMarkdownToHTML converts all markdown files to HTML
func (g *HTMLGenerator) ConvertAllMarkdownToHTML(markdownPaths []string) ([]string, error) {
	var htmlPaths []string

	for _, markdownPath := range markdownPaths {
		htmlPath, err := g.ConvertMarkdownToHTML(markdownPath)
		if err != nil {
			return htmlPaths, err
		}
		htmlPaths = append(htmlPaths, htmlPath)
	}

	return htmlPaths, nil
}
