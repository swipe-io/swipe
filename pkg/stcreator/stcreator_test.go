package stcreator_test

import (
	"path/filepath"
	"testing"

	"github.com/swipe-io/swipe/pkg/stcreator"
)

func TestCreator(t *testing.T) {
	parent := stcreator.D{
		Name: "project",
		Children: []stcreator.I{
			&stcreator.F{
				Name:    "Dockerfile",
				Content: "FROM {{.Name}}\n",
				Data: struct {
					Name string
				}{"ubuntu"},
			},
			&stcreator.D{
				Name: "pkg",
				Children: []stcreator.I{
					&stcreator.F{
						Name: ".gitkeep",
					},
				},
			},
		},
	}

	basePath, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	v := stcreator.New()

	if err := parent.Accept(v, basePath); err != nil {
		t.Fatal(err)
	}
}
