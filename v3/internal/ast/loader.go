package ast

import (
	"context"
	"errors"
	"go/ast"
	"go/types"
	stdstrings "strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type CommentFields struct {
	fields map[uint32]map[string]string
	h      typeutil.Hasher
}

func (c *CommentFields) GetByFieldName(t types.Type, name string) string {
	comments := c.fields[c.h.Hash(t)]
	return comments[name]
}

func (c *CommentFields) Add(t types.Type, comments map[string]string) {
	c.fields[c.h.Hash(t)] = comments
}

type Loader struct {
	ctx           context.Context
	wd            string
	env           []string
	patterns      []string
	module        *packages.Module
	commentFuncs  map[string][]string
	commentFields *CommentFields
	pkgs          []*packages.Package
	enums         *typeutil.Map
}

func (l *Loader) CommentFields() *CommentFields {
	return l.commentFields
}

func (l *Loader) CommentFuncs() map[string][]string {
	return l.commentFuncs
}

func (l *Loader) Module() *packages.Module {
	return l.module
}

func (l *Loader) Pkgs() []*packages.Package {
	return l.pkgs
}

func (l *Loader) Patterns() []string {
	return l.patterns
}

func (l *Loader) Env() []string {
	return l.env
}

func (l *Loader) WorkDir() string {
	return l.wd
}

func (l *Loader) run() (errs []error) {
	var (
		err error
	)

	l.commentFuncs = map[string][]string{}
	l.commentFields = &CommentFields{fields: map[uint32]map[string]string{}, h: typeutil.MakeHasher()}
	l.enums = new(typeutil.Map)

	cfg := &packages.Config{
		Context: l.ctx,
		Mode: packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedImports |
			packages.NeedName |
			packages.NeedModule |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles,
		Dir:        l.wd,
		Env:        l.env,
		BuildFlags: []string{"-tags=swipe"},
	}
	escaped := make([]string, len(l.patterns))
	for i := range l.patterns {
		escaped[i] = "pattern=" + l.patterns[i]
	}
	l.pkgs, err = packages.Load(cfg, escaped...)
	if err != nil {
		return []error{err}
	}
	for _, p := range l.pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	for _, pkg := range l.pkgs {
		if pkg.Module == nil {
			continue
		}
		if l.module == nil && l.wd == pkg.Module.Dir {
			l.module = pkg.Module
			break
		}
	}
	if l.module == nil {
		errs = append(errs, errors.New("go mod not found, run go mod init"))
		return
	}

	inspect(l.pkgs, func(p *packages.Package, n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			obj := p.TypesInfo.ObjectOf(ts.Name)
			if st, ok := ts.Type.(*ast.StructType); ok {
				comments := map[string]string{}
				for _, field := range st.Fields.List {
					if field.Comment != nil {
						if len(field.Comment.List) > 0 {
							for _, name := range field.Names {
								comments[name.Name] = stdstrings.TrimLeft(field.Comment.List[0].Text, "/")
							}
						}
					}
				}
				if len(comments) > 0 {
					l.commentFields.Add(obj.Type().Underlying(), comments)
				}
			}
		} else if spec, ok := n.(*ast.Field); ok {
			t := p.TypesInfo.TypeOf(spec.Type)
			if t != nil {
				var comments []string
				if spec.Doc != nil {
					for _, comment := range spec.Doc.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if spec.Comment != nil {
					for _, comment := range spec.Comment.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if len(comments) > 0 {
					for _, name := range spec.Names {
						if obj := p.TypesInfo.ObjectOf(name); obj != nil {
							l.commentFuncs[obj.String()] = comments
						}
					}
				}
			}
		} else if spec, ok := n.(*ast.FuncDecl); ok {
			obj := p.TypesInfo.ObjectOf(spec.Name)
			if obj != nil {
				var comments []string
				if spec.Doc != nil {
					for _, comment := range spec.Doc.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if len(comments) > 0 {
					l.commentFuncs[obj.String()] = comments
				}
			}
		}
		return true
	})
	return
}

func NewLoader(wd string, env []string, patterns []string) (*Loader, []error) {
	l := &Loader{
		wd:       wd,
		env:      env,
		patterns: patterns,
	}
	errs := l.run()
	if len(errs) > 0 {
		return nil, errs
	}
	return l, nil
}
