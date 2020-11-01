package fixcomment

import (
	"bytes"
	"context"
	"go/ast"
	"go/printer"

	"github.com/swipe-io/swipe/v2/internal/format"

	"golang.org/x/tools/go/packages"
)

type FixData struct {
	Filepath string
	Content  []byte
}

type FixComment struct {
	wd            string
	env, patterns []string
}

func NewFixComment(wd string, env []string, patterns []string) *FixComment {
	return &FixComment{wd: wd, env: env, patterns: patterns}
}

func (fc *FixComment) Execute() (result []FixData, err error) {
	cfg := &packages.Config{
		Context: context.TODO(),
		Mode:    packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedImports | packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:     fc.wd,
		Env:     fc.env,
	}
	escaped := make([]string, len(fc.patterns))
	for i := range fc.patterns {
		escaped[i] = "pattern=" + fc.patterns[i]
	}
	pkgs, err := packages.Load(cfg, escaped...)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for _, syntax := range pkg.Syntax {

			var comments []*ast.CommentGroup

			ast.Inspect(syntax, func(node ast.Node) bool {
				switch t := node.(type) {
				case *ast.CommentGroup:
					comments = append(comments, t)
				case *ast.FuncDecl:
					if t.Name.IsExported() && t.Doc.Text() == "" {
						//fmt.Printf("exported function declaration without documentation found on line %d: %s\n", pkg.Fset.Position(t.Pos()).Line, t.Name.Name)
						comment := &ast.Comment{
							Text:  "// " + t.Name.Name + " ...",
							Slash: t.Pos() - 1,
						}
						cg := &ast.CommentGroup{
							List: []*ast.Comment{comment},
						}
						t.Doc = cg
					}
				case *ast.GenDecl:
					if len(t.Specs) > 1 {
						for _, spec := range t.Specs {
							tp, ok := spec.(*ast.TypeSpec)
							if ok && tp.Name.IsExported() && tp.Doc.Text() == "" {
								if tp.Doc.Text() == "" {
									//fmt.Printf("exported type declaration without documentation found on line %d: %s\n", pkg.Fset.Position(t.Pos()).Line, tp.Name.Name)
									comment := &ast.Comment{
										Text:  "// " + tp.Name.Name + " ...",
										Slash: tp.Pos() - 1,
									}

									cg := &ast.CommentGroup{
										List: []*ast.Comment{comment},
									}
									tp.Doc = cg
								}
							}
						}
					} else if t.Doc.Text() == "" {
						tp, ok := t.Specs[0].(*ast.TypeSpec)
						if ok && tp.Name.IsExported() {
							//fmt.Printf("exported type declaration without documentation found on line %d: %s\n", pkg.Fset.Position(t.Pos()).Line, tp.Name.Name)
							comment := &ast.Comment{
								Text:  "// " + tp.Name.Name + " ...",
								Slash: t.Pos() - 1,
							}
							cg := &ast.CommentGroup{
								List: []*ast.Comment{comment},
							}
							t.Doc = cg
						}
					}
				}
				return true
			})

			syntax.Comments = comments
			buf := new(bytes.Buffer)
			if err := printer.Fprint(buf, pkg.Fset, syntax); err != nil {
				return nil, err
			}
			content, err := format.Source(buf.Bytes())
			if err != nil {
				return nil, err
			}
			result = append(result, FixData{
				Filepath: pkg.Fset.File(syntax.Pos()).Name(),
				Content:  content,
			})

		}
	}
	return
}
