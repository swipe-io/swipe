package model

import (
	"github.com/swipe-io/swipe/pkg/git"
	"github.com/swipe-io/swipe/pkg/graph"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type GenerateInfo struct {
	Pkg         *packages.Package
	BasePkgPath string
	RootPath    string
	Pkgs        []*packages.Package
	CommentMap  *typeutil.Map
	ReturnTypes map[uint32][]interface{}
	BasePath    string
	Version     string
	GraphTypes  *graph.Graph
	MapTypes    map[uint32]*DeclType
	Enums       *typeutil.Map
	GitTags     []git.Tag
}

type Enum struct {
	Name  string
	Value string
}
