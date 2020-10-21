package generator

import (
	"context"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/swipe-io/swipe/v2/internal/git"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

const defaultTemplate = `# {{.ServiceName}} : A short description of the service. <code>{{ .GIT.LastTag.Name }}</code>
A complete description of the service and what it does.

## Example

<code>
go run ./cmd/service
</code>

## Docs

ToDo.

## Contributing

ToDo.

## Contributors

ToDo.

## Author

ToDo.

## Changelog

ToDo.

## Versions

{{range $index, $tag := .GIT.Tags -}}
   {{if gt $index 0 -}}, {{end -}}
   [{{$tag.Name}}](https://{{$.RootPkgPath}}/tree/{{$tag.Name}})
{{end -}}
`

type readmeGeneratorOptionsGateway interface {
	AppID() string
	AppName() string
	JSONRPCDocOutput() string
	ReadmeOutput() string
	ReadmeTemplatePath() string
}

type readmeGenerator struct {
	writer.BaseWriter
	options        readmeGeneratorOptionsGateway
	basePkgPath    string
	outputDir      string
	workDir        string
	gitTags        []git.Tag
	markdownOutput string
	tpl            *template.Template
}

func (g *readmeGenerator) Prepare(ctx context.Context) (err error) {
	g.outputDir, err = filepath.Abs(filepath.Join(g.workDir, g.options.ReadmeOutput()))
	if err != nil {
		return err
	}
	g.markdownOutput, err = filepath.Abs(filepath.Join(g.workDir, g.options.JSONRPCDocOutput()))
	if err != nil {
		return err
	}
	var templatePath string
	if templatePath == "" {
		templatePath, err = filepath.Abs(filepath.Join(g.workDir, ".swipe"))
		if err != nil {
			return err
		}
		if _, err := os.Stat(templatePath); err != nil {
			if err = os.MkdirAll(templatePath, 0755); err != nil {
				return err
			}
		}
	} else {
		templatePath, err = filepath.Abs(filepath.Join(g.workDir, g.options.ReadmeTemplatePath()))
		if err != nil {
			return err
		}
	}
	templateFilepath := filepath.Join(templatePath, "README.md.tpl")
	if _, err := os.Stat(templateFilepath); err != nil {
		err = ioutil.WriteFile(templateFilepath, []byte(defaultTemplate), 0755)
		if err != nil {
			return err
		}
	}
	data, err := ioutil.ReadFile(templateFilepath)
	if err != nil {
		return err
	}
	t, err := template.New("readmeGenerator").Parse(string(data))
	if err != nil {
		return err
	}
	g.tpl = t
	return nil
}

func (g *readmeGenerator) Process(ctx context.Context) (err error) {
	return g.tpl.Execute(g, map[string]interface{}{
		"ID":          g.options.AppID(),
		"ServiceName": g.options.AppName(),
		"RootPkgPath": g.basePkgPath,
		"GIT": map[string]interface{}{
			"Tags": g.gitTags,
		},
	})
}

func (g *readmeGenerator) PkgName() string {
	return ""
}

func (g *readmeGenerator) OutputDir() string {
	return g.outputDir
}

func (g *readmeGenerator) Filename() string {
	return "README.md"
}

func NewReadme(
	options readmeGeneratorOptionsGateway,
	basePkgPath string,
	workDir string,
	gitTags []git.Tag,
) generator.Generator {
	return &readmeGenerator{
		options:     options,
		basePkgPath: basePkgPath,
		workDir:     workDir,
		gitTags:     gitTags,
	}
}
