package generator

import (
	"context"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

## Versions

{{range $index, $tag := .GIT.Tags -}}
    {{if gt $index 0 -}}, {{end -}}
    [{{$tag.Name}}](https://{{$.BasePkgPath}}/tree/{{$tag.Name}})
{{end -}}
`

type readme struct {
	writer.BaseWriter
	info           model.GenerateInfo
	o              model.ServiceOption
	outputDir      string
	markdownOutput string
	t              *template.Template
}

func (g *readme) Prepare(ctx context.Context) (err error) {
	g.outputDir, err = filepath.Abs(filepath.Join(g.info.BasePath, g.o.Readme.OutputDir))
	if err != nil {
		return err
	}
	g.markdownOutput, err = filepath.Abs(filepath.Join(g.info.BasePath, g.o.Transport.MarkdownDoc.OutputDir))
	if err != nil {
		return err
	}
	var templatePath string
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
	} else {
		templatePath, err = filepath.Abs(filepath.Join(g.info.BasePath, g.o.Readme.TemplatePath))
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
		"BasePkgPath": g.info.BasePkgPath,
		"JSONRPCDoc": map[string]interface{}{
			"Enabled": g.o.Transport.MarkdownDoc.Enable,
			"Path":    filepath.Join(relPath, "jsonrpc_"+strings.ToLower(g.o.ID)+"_doc.md"),
		},
		"GIT": map[string]interface{}{
			"Tags": g.info.GitTags,
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
