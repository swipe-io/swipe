package generator

import (
	"context"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tsuyoshiwada/go-gitcmd"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/writer"
)

const defaultTemplate = `# {{.ServiceName}} : A short description of the service. <code>{{ .GIT.LastTag.Name }}</code>
A complete description of the service and what it does.

## Example

<code>
go run ./cmd/service
</code>

## Docs

ToDo.

{{ if .JSONRPCDoc.Enabled }}
[JSON RPC Client Doc]({{ .JSONRPCDoc.Path }})
{{ end }}

## Contributing

ToDo.

## Contributors

ToDo.

## Author

ToDo.

## Changelog

ToDo.
`

type Tag struct {
	Name    string
	Subject string
	Date    time.Time
}

type readme struct {
	writer.BaseWriter
	info           model.GenerateInfo
	o              model.ServiceOption
	outputDir      string
	markdownOutput string
	t              *template.Template
	tags           []Tag
	lastTag        Tag
}

func (g *readme) prepareGitTags() ([]Tag, error) {
	git := gitcmd.New(nil)

	separator := "@@__SWIPE__@@"

	out, err := git.Exec("for-each-ref",
		"--format",
		"%(refname)"+separator+"%(subject)"+separator+"%(taggerdate)"+separator+"%(authordate)",
		"refs/tags",
	)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	var tags []Tag
	for _, line := range lines {
		tokens := strings.Split(line, separator)
		if len(tokens) != 4 {
			continue
		}
		name := strings.Replace(tokens[0], "refs/tags/", "", 1)
		subject := strings.TrimSpace(tokens[1])
		date, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", tokens[2])
		if err != nil {
			date, err = time.Parse("Mon Jan 2 15:04:05 2006 -0700", tokens[2])
			if err != nil {
				return nil, err
			}
		}
		tags = append(tags, Tag{
			Name:    name,
			Subject: subject,
			Date:    date,
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		return !tags[i].Date.Before(tags[j].Date)
	})
	return tags, nil
}

func (g *readme) Prepare(ctx context.Context) error {
	tags, err := g.prepareGitTags()
	if err != nil {
		return err
	}
	g.tags = tags
	if len(g.tags) > 0 {
		g.lastTag = g.tags[0]
	}
	g.outputDir, err = filepath.Abs(filepath.Join(g.info.BasePath, g.o.Readme.OutputDir))
	if err != nil {
		return err
	}
	g.markdownOutput, err = filepath.Abs(filepath.Join(g.info.BasePath, g.o.Transport.MarkdownDoc.OutputDir))
	if err != nil {
		return err
	}
	templatePath := g.o.Readme.TemplatePath
	if templatePath == "" {
		templatePath, err = filepath.Abs(filepath.Join(g.info.BasePath, ".swipe"))
		if err != nil {
			return err
		}
		if _, err := os.Stat(templatePath); err != nil {
			if err = os.MkdirAll(templatePath, 0755); err != nil {
				return err
			}
		}
	}
	templateFilepath := filepath.Join(templatePath, "README.tpl.md")
	if _, err := os.Stat(templateFilepath); err != nil {
		err = ioutil.WriteFile(templateFilepath, []byte(defaultTemplate), 0755)
		if err != nil {
			return err
		}
	}
	data, err := ioutil.ReadFile(filepath.Join(templatePath, "README.tpl.md"))
	if err != nil {
		return err
	}
	t, err := template.New("readme").Parse(string(data))
	if err != nil {
		return err
	}
	g.t = t
	return nil
}

func (g *readme) Process(ctx context.Context) (err error) {
	var relPath string
	if g.o.Transport.MarkdownDoc.Enable {
		relPath, err = filepath.Rel(g.outputDir, g.markdownOutput)
		if err != nil {
			return err
		}
	}
	return g.t.Execute(g, map[string]interface{}{
		"ID":          g.o.RawID,
		"ServiceName": g.o.ID,
		"JSONRPCDoc": map[string]interface{}{
			"Enabled": g.o.Transport.MarkdownDoc.Enable,
			"Path":    filepath.Join(relPath, "jsonrpc_"+strings.ToLower(g.o.ID)+"_doc.md"),
		},
		"GIT": map[string]interface{}{
			"Tags":    g.tags,
			"LastTag": g.lastTag,
		},
	})
}

func (g *readme) PkgName() string {
	return ""
}

func (g *readme) OutputDir() string {
	return g.outputDir
}

func (g *readme) Filename() string {
	return "README.md"
}

func NewReadme(info model.GenerateInfo, o model.ServiceOption) Generator {
	return &readme{info: info, o: o}
}
