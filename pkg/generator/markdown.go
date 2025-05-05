package generator

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
)

// MarkdownGenerator handles the generation of markdown files
type MarkdownGenerator struct {
	outputDir string
	template  *template.Template
	verbose   bool
}

// NewMarkdownGenerator creates a new MarkdownGenerator
func NewMarkdownGenerator(config *types.Config) (*MarkdownGenerator, error) {
	// Create the template
	tmpl, err := template.New("repository.md.tmpl").Parse(`# Repository: {{.Name}}

{{if .Description}}{{.Description}}{{else}}No description provided.{{end}}

## Open Pull Requests

{{if .PullRequests}}
| Title | Author | Updated | Labels |
|-------|--------|---------|--------|
{{range .PullRequests}}| [{{.Title}}]({{.URL}}) | [{{.Author}}]({{.AuthorURL}}) | {{.UpdatedAt.Format "2006-01-02"}} | {{range $i, $label := .Labels}}{{if $i}}, {{end}}{{$label.Name}}{{else}}*none*{{end}} |
{{end}}
{{else}}
*No open pull requests*
{{end}}

## Open Issues

{{if .Issues}}
| Title | Author | Updated | Labels |
|-------|--------|---------|--------|
{{range .Issues}}| [{{.Title}}]({{.URL}}) | [{{.Author}}]({{.AuthorURL}}) | {{.UpdatedAt.Format "2006-01-02"}} | {{range $i, $label := .Labels}}{{if $i}}, {{end}}{{$label.Name}}{{else}}*none*{{end}} |
{{end}}
{{else}}
*No open issues*
{{end}}

## Recent Discussions

{{if .Discussions}}
| Title | Started By | Last Activity | Category |
|-------|------------|---------------|----------|
{{range .Discussions}}| [{{.Title}}]({{.URL}}) | [{{.Author}}]({{.AuthorURL}}) | {{.LastUpdated.Format "2006-01-02"}} | {{.Category}} |
{{end}}
{{else}}
*No recent discussions*
{{end}}

---
*Generated at {{.GeneratedAt.Format "2006-01-02 15:04:05"}}*
`)
	if err != nil {
		return nil, fmt.Errorf("error parsing markdown template: %w", err)
	}

	return &MarkdownGenerator{
		outputDir: config.OutputDir,
		template:  tmpl,
		verbose:   config.Verbose,
	}, nil
}

// GenerateRepositoryMarkdown generates a markdown file for a repository
func (g *MarkdownGenerator) GenerateRepositoryMarkdown(repo types.Repository) (string, error) {
	if g.verbose {
		log.Printf("Generating markdown for repository %s", repo.FullName)
	}

	// Prepare the template data
	data := struct {
		types.Repository
		GeneratedAt time.Time
	}{
		Repository:  repo,
		GeneratedAt: time.Now(),
	}

	// Render the template
	var buf bytes.Buffer
	err := g.template.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	// Write the file
	outputPath := filepath.Join(g.outputDir, "repositories", fmt.Sprintf("%s.md", repo.Name))
	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing markdown file: %w", err)
	}

	return outputPath, nil
}

// GenerateAllRepositoriesMarkdown generates markdown files for all repositories
func (g *MarkdownGenerator) GenerateAllRepositoriesMarkdown(dashboard types.Dashboard) ([]string, error) {
	var filePaths []string

	for _, repo := range dashboard.Repositories {
		path, err := g.GenerateRepositoryMarkdown(repo)
		if err != nil {
			return filePaths, err
		}
		filePaths = append(filePaths, path)
	}

	return filePaths, nil
}
