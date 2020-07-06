package model

import (
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type GenerateInfo struct {
	Pkg         *packages.Package
	Pkgs        []*packages.Package
	CommentMap  *typeutil.Map
	ReturnTypes map[uint32][]interface{}
	BasePath    string
	Version     string
	MapTypes    map[uint32]*DeclType
}
