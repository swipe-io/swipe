package types

import (
	"errors"
	"fmt"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type FilterFn func(p *types.Var) bool

func Params(vars []*types.Var, fn func(p *types.Var) []string, filterFn func(p *types.Var) bool) (results []string) {
	for _, p := range vars {
		if filterFn != nil && !filterFn(p) {
			continue
		}
		results = append(results, fn(p)...)
	}
	return
}

func NameTypeParams(vars []*types.Var, qf types.Qualifier, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{p.Name(), types.TypeString(p.Type(), qf)}
	}, filterFn)
}

func NameType(vars []*types.Var, qf types.Qualifier, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{"", types.TypeString(p.Type(), qf)}
	}, filterFn)
}

//func DetectAppPath(pkg *packages.Package) (string, error) {
//	basePath, err := DetectBasePath(pkg)
//	if err != nil {
//		return "", err
//	}
//
//	srcPath := filepath.Join(build.Default.GOPATH, "src") + "/"
//	index := strings.Index(basePath, srcPath)
//
//	fmt.Println(basePath, srcPath, pkg.PkgPath)
//
//	if index != -1 {
//		return basePath[index+len(srcPath):], nil
//	}
//	return "", errors.New("fail detected app path")
//}
//
func DetectBasePath(pkg *packages.Package) (string, error) {
	paths := pkg.GoFiles
	if len(paths) == 0 {
		return "", errors.New("no files to derive output directory from")
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if dir2 := filepath.Dir(p); dir2 != dir {
			return "", fmt.Errorf("found conflicting directories %q and %q", dir, dir2)
		}
	}
	return dir, nil
}
