package model

import (
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type GenerateInfo struct {
	Pkg        *packages.Package
	Pkgs       []*packages.Package
	CommentMap *typeutil.Map
	BasePath   string
	Version    string
}
