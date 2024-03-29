package option

import (
	"errors"
	"fmt"
	goast "go/ast"
	"go/constant"
	"go/token"
	stdtypes "go/types"
	"path/filepath"

	"github.com/swipe-io/swipe/v3/internal/ast"

	packages2 "github.com/swipe-io/swipe/v3/internal/packages"

	"github.com/fatih/structtag"
	"github.com/swipe-io/swipe/v3/internal/annotation"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Inject struct {
	Pkg      *PackageType
	BasePath string
	Option   map[string]interface{}
}

type Module struct {
	Path     string
	External bool
	Injects  []*Inject
}

type Result struct {
	Modules map[string][]Module
}

type Decoder struct {
	optionPkgs     map[string]string
	module         *packages.Module
	pkgs           *packages2.Packages
	commentFuncMap map[string][]string
	commentFields  *ast.CommentFields
}

func normalizeName(s string) String {
	return String{Value: s}
}

func (d *Decoder) normalizeVar(pkg *packages.Package, t *stdtypes.Var, comment string, visited map[string]interface{}) *VarType {
	if t == nil {
		return nil
	}
	varType := d.normalizeType(pkg, t.Type(), false, visited)

	var isContext bool
	if named, ok := varType.(*NamedType); ok {
		if _, ok := named.Type.(*IfaceType); ok {
			if named.Name.Value == "Context" && named.Pkg.Path == "context" {
				isContext = true
			}
		}
	}

	return &VarType{
		Name:       normalizeName(t.Name()),
		Embedded:   t.Embedded(),
		Exported:   t.Exported(),
		IsField:    t.IsField(),
		IsContext:  isContext,
		Type:       varType,
		originType: t.Type(),
		Zero:       zeroValue(t.Type().Underlying()),
		Comment:    comment,
	}
}

func (d *Decoder) normalizeStruct(pkg *packages.Package, t *stdtypes.Struct, isPointer bool, visited map[string]interface{}) *StructType {
	if t == nil {
		return nil
	}
	result := &StructType{
		IsPointer:  isPointer,
		originType: t,
	}
	for i := 0; i < t.NumFields(); i++ {
		field := t.Field(i)
		comment := d.commentFields.GetByFieldName(t, field.Name())
		f := &StructFieldType{
			Var: d.normalizeVar(pkg, field, comment, visited),
		}
		if tags, err := structtag.Parse(t.Tag(i)); err == nil {
			f.Tags = tags
		}
		result.Fields = append(result.Fields, f)
	}
	return result
}

func (d *Decoder) normalizeType(pkg *packages.Package, t interface{}, isPointer bool, visited map[string]interface{}) interface{} {
	switch t := t.(type) {
	case *stdtypes.Func:
		return d.normalizeFunc(pkg, t, visited)
	case *stdtypes.Map:
		return d.normalizeMap(pkg, t.Key(), t.Elem(), isPointer, visited)
	case *stdtypes.Slice:
		return d.normalizeSlice(pkg, t.Elem(), isPointer, visited)
	case *stdtypes.Array:
		return d.normalizeArray(pkg, t.Elem(), t.Len(), isPointer, visited)
	case *stdtypes.Pointer:
		return d.normalizeType(pkg, t.Elem(), true, visited)
	case *stdtypes.Struct:
		return d.normalizeStruct(pkg, t, isPointer, visited)
	case *stdtypes.Signature:
		return d.normalizeSignature(pkg, t, nil, visited)
	case *stdtypes.Interface:
		return d.normalizeInterface(pkg, t, visited)
	case *stdtypes.Named:
		return d.normalizeNamed(pkg, t, isPointer, visited)
	case *stdtypes.TypeName:
		return d.normalizeTypeName(pkg, t, isPointer, visited)
	case *stdtypes.Basic:
		return d.normalizeBasic(t, isPointer)
	}
	return nil
}

func (d *Decoder) normalizeTypeName(pkg *packages.Package, obj *stdtypes.TypeName, isPointer bool, visited map[string]interface{}) *NamedType {
	var prefix string
	if isPointer {
		prefix = "*"
	}
	if obj.Pkg() != nil {
		prefix += obj.Pkg().Path()
	}
	k := prefix + obj.Name()
	if v, ok := visited[k].(*NamedType); ok {
		return v
	}
	nt := &NamedType{
		Obj:       obj,
		Name:      normalizeName(obj.Name()),
		Pkg:       d.normalizePkg(obj.Pkg()),
		IsPointer: isPointer,
	}

	visited[k] = nt

	nt.Type = d.normalizeType(pkg, obj.Type().Underlying(), false, visited)

	return nt
}

func (d *Decoder) normalizeNamed(pkg *packages.Package, named *stdtypes.Named, isPointer bool, visited map[string]interface{}) *NamedType {
	var prefix string
	if isPointer {
		prefix = "*"
	}
	if named.Obj().Pkg() != nil {
		prefix += named.Obj().Pkg().Path()
	}
	k := prefix + named.Obj().Name()
	if v, ok := visited[k].(*NamedType); ok {
		return v
	}
	nt := &NamedType{
		Obj:       named.Obj(),
		Pkg:       d.normalizePkg(named.Obj().Pkg()),
		Name:      normalizeName(named.Obj().Name()),
		IsPointer: isPointer,
	}

	visited[k] = nt

	nt.Type = d.normalizeType(pkg, named.Obj().Type().Underlying(), false, visited)

	for i := 0; i < named.NumMethods(); i++ {
		nt.Methods = append(nt.Methods, d.normalizeFunc(pkg, named.Method(i), visited))
	}
	return nt
}

func (d *Decoder) normalizePkg(pkg *stdtypes.Package) *PackageType {
	if pkg != nil {
		var module *ModuleType
		fndPkg := d.pkgs.FindPkgByPath(pkg.Path())
		if fndPkg != nil {
			module = d.normalizeModule(fndPkg.Module)
		}
		return &PackageType{
			Name:   pkg.Name(),
			Path:   pkg.Path(),
			Module: module,
		}
	}
	return nil
}

func (d *Decoder) normalizeBasic(t *stdtypes.Basic, isPointer bool) *BasicType {
	return &BasicType{
		Name:      t.Name(),
		IsPointer: isPointer,
		kind:      t.Kind(),
	}
}

func (d *Decoder) normalizeInterface(pkg *packages.Package, t *stdtypes.Interface, visited map[string]interface{}) *IfaceType {
	it := &IfaceType{
		Origin: t,
	}
	for i := 0; i < t.NumMethods(); i++ {
		it.Methods = append(it.Methods, d.normalizeFunc(pkg, t.Method(i), visited))
	}
	for i := 0; i < t.NumEmbeddeds(); i++ {
		it.Embeddeds = append(it.Embeddeds, d.normalizeType(pkg, t.EmbeddedType(i), false, visited))
	}
	for i := 0; i < t.NumExplicitMethods(); i++ {
		it.ExplicitMethods = append(it.ExplicitMethods, d.normalizeFunc(pkg, t.ExplicitMethod(i), visited))
	}
	return it
}

func (d *Decoder) normalizeFunc(pkg *packages.Package, t *stdtypes.Func, visited map[string]interface{}) *FuncType {
	comments := d.commentFuncMap[t.String()]
	comment, paramsComment := parseMethodComments(comments)

	return &FuncType{
		Pkg:      d.normalizePkg(t.Pkg()),
		FullName: t.FullName(),
		Name:     normalizeName(t.Name()),
		Exported: t.Exported(),
		Sig:      d.normalizeSignature(pkg, t.Type().(*stdtypes.Signature), paramsComment, visited),
		Comment:  comment,
	}
}

func (d *Decoder) normalizeSignature(pkg *packages.Package, t *stdtypes.Signature, comments map[string]string, visited map[string]interface{}) *SignType {
	if t == nil {
		return nil
	}
	st := &SignType{
		IsVariadic: t.Variadic(),
	}
	if t.Recv() != nil {
		st.Recv = d.normalizeType(pkg, t.Recv().Type(), false, visited)
	}
	for i := 0; i < t.Params().Len(); i++ {
		v := t.Params().At(i)
		nv := d.normalizeVar(pkg, v, comments[v.Name()], visited)
		st.Params = append(st.Params, nv)
	}
	if t.Variadic() {
		st.Params[len(st.Params)-1].IsVariadic = true
	}
	for i := 0; i < t.Results().Len(); i++ {
		v := t.Results().At(i)
		nv := d.normalizeVar(pkg, v, comments[v.Name()], visited)
		if nv.Name.Value == "" {
			nv.Name = normalizeName(fmt.Sprintf("r%d", i+1))
		} else {
			st.IsNamed = true
		}
		st.Results = append(st.Results, nv)
	}
	return st
}

func (d *Decoder) normalizeMap(pkg *packages.Package, key stdtypes.Type, val stdtypes.Type, isPointer bool, visited map[string]interface{}) *MapType {
	mapType := &MapType{IsPointer: isPointer}
	mapType.Key = d.normalizeType(pkg, key, false, visited)
	mapType.Value = d.normalizeType(pkg, val, false, visited)
	return mapType
}

func (d *Decoder) normalizeSlice(pkg *packages.Package, val stdtypes.Type, isPointer bool, visited map[string]interface{}) *SliceType {
	return &SliceType{
		Value:     d.normalizeType(pkg, val, false, visited),
		IsPointer: isPointer,
	}
}

func (d *Decoder) normalizeArray(pkg *packages.Package, val stdtypes.Type, len int64, isPointer bool, visited map[string]interface{}) *ArrayType {
	return &ArrayType{
		Value:     d.normalizeType(pkg, val, false, visited),
		Len:       len,
		IsPointer: isPointer,
	}
}

func (d *Decoder) normalizeSelector(pkg *packages.Package, obj stdtypes.Object) interface{} {
	return &NamedType{
		Obj:  obj,
		Name: normalizeName(obj.Name()),
		Type: d.normalizeType(pkg, obj.Type().Underlying(), false, map[string]interface{}{}),
		Pkg:  d.normalizePkg(obj.Pkg()),
	}
}

func (d *Decoder) normalize(pkg *packages.Package, obj stdtypes.Object) interface{} {
	return d.normalizeType(pkg, obj, false, map[string]interface{}{})
}

func (d *Decoder) normalizeModule(module *packages.Module) *ModuleType {
	if module != nil {
		return &ModuleType{
			ID:       filepath.Base(module.Path),
			Version:  module.Version,
			Path:     module.Path,
			Dir:      module.Dir,
			External: module.Path != d.module.Path,
		}
	}
	return nil
}

func (d *Decoder) normalizePosition(pos token.Position) *PositionType {
	return &PositionType{
		Column:   pos.Column,
		Filename: pos.Filename,
		Line:     pos.Line,
		Offset:   pos.Offset,
		IsValid:  pos.IsValid(),
	}
}

func (d *Decoder) decodeRecursive(pkg *packages.Package, expr goast.Expr) (interface{}, error) {
	switch e := expr.(type) {
	case *goast.CompositeLit:
		switch vt := e.Type.(type) {
		case *goast.SelectorExpr:
			return d.normalize(pkg, pkg.TypesInfo.Uses[vt.Sel]), nil
		case *goast.Ident:
			return d.normalize(pkg, pkg.TypesInfo.Uses[vt]), nil
		case *goast.ArrayType:
			switch elt := vt.Elt.(type) {
			default:
				var value []interface{}
				for _, expr := range e.Elts {
					switch st := expr.(type) {
					case *goast.SelectorExpr:
						value = append(value, d.normalizeSelector(pkg, pkg.TypesInfo.Uses[st.Sel]))
					default:
						return nil, errors.New("fail")
					}
				}
				return value, nil
			case *goast.Ident:
				switch elt.Name {
				case "string":
					return makeStringSlice(e.Elts, pkg.TypesInfo), nil
				}
			}
		}
	case *goast.BasicLit:
		var value interface{}
		tv := pkg.TypesInfo.Types[e]
		if tv.IsValue() {
			value = constant.Val(tv.Value)
		}
		return value, nil
	case *goast.Ident:
		switch e.Name {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
		return d.normalize(nil, pkg.TypesInfo.Uses[e]), nil
	case *goast.StarExpr:
		return d.decodeRecursive(pkg, e.X)
	case *goast.UnaryExpr:
		return d.decodeRecursive(pkg, e.X)
	case *goast.ParenExpr:
		return d.decodeRecursive(pkg, e.X)
	case *goast.SelectorExpr:
		_, err := d.decodeRecursive(pkg, e.X)
		if err != nil {
			return nil, err
		}
		return d.normalizeSelector(pkg, pkg.TypesInfo.Uses[e.Sel]), nil
	case *goast.CallExpr:
		return d.decodeRecursive(pkg, e.Fun)
	}
	return nil, nil
}

func (d *Decoder) callDecodeArgs(pkg *packages.Package, obj stdtypes.Object, args []goast.Expr) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	sig := obj.Type().(*stdtypes.Signature)
	for i, arg := range args {
		//exprPos := pkg.Fset.Position(arg.Pos())
		if callExpr, ok := arg.(*goast.CallExpr); ok {
			fnExpr := astutil.Unparen(callExpr.Fun)
			if _, ok := fnExpr.(*goast.StarExpr); !ok {
				obj := qualifiedIdentObject(pkg.TypesInfo, fnExpr)
				if obj == nil {
					return nil, errors.New("failed get object")
				}
				val, err := d.callDecodeArgs(pkg, obj, callExpr.Args)
				if err != nil {
					return nil, err
				}
				var valueType string
				comments := d.commentFuncMap[obj.String()]
				for _, comment := range comments {
					if annotationOpts, _ := annotation.Parse(comment); annotationOpts != nil {
						if annotationOpt, err := annotationOpts.Get("type"); err == nil {
							valueType = annotationOpt.Value()
						}
					}
				}
				name := obj.Name()
				if valueType == "repeat" {
					if _, ok := result[name]; !ok {
						result[name] = []interface{}{}
					}
					v := result[name].([]interface{})
					v = append(v, val)
					result[name] = v

				} else {
					result[name] = val
				}
				continue
			}
		}
		vr := sigParamAt(sig, i)
		if vr.Name() == "" {
			return nil, errors.New("failed params name")
		}
		val, err := d.decodeRecursive(pkg, arg)
		if err != nil {
			return nil, err
		}
		result[vr.Name()] = val
	}
	return result, nil
}

func (d *Decoder) callDecode(pkg *packages.Package, e *goast.CallExpr) (map[string]interface{}, error) {
	obj := qualifiedIdentObject(pkg.TypesInfo, astutil.Unparen(e.Fun))
	if obj == nil {
		return nil, errors.New("failed get object")
	}
	result, err := d.callDecodeArgs(pkg, obj, e.Args)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{obj.Name(): result}, nil
}

func (d *Decoder) decode() (result map[string]*Module, err error) {
	result = map[string]*Module{}
	err = d.pkgs.TraverseTypes(func(pkg *packages.Package, expr goast.Expr, value stdtypes.TypeAndValue) (err error) {
		expr = astutil.Unparen(expr)
		callExpr, ok := expr.(*goast.CallExpr)
		if !ok {
			return
		}
		fun := callExpr.Fun
		if selExpr, ok := fun.(*goast.SelectorExpr); ok {
			fun = selExpr.Sel
		}
		callIdent, ok := fun.(*goast.Ident)
		if !ok {
			return
		}
		obj := d.pkgs.ObjectOf(callIdent)
		if obj == nil || obj.Pkg() == nil {
			return
		}
		if buildName, ok := d.optionPkgs[obj.Pkg().Name()]; ok && obj.Name() == buildName {
			if _, ok := result[pkg.Module.Path]; !ok {
				result[pkg.Module.Path] = &Module{
					Path:     pkg.Module.Path,
					External: d.module.Path != pkg.Module.Path,
				}
			}
			option, err := d.callDecodeArgs(pkg, obj, callExpr.Args)
			if err != nil {
				return err
			}
			basePath, err := detectBasePath(pkg)
			if err != nil {
				return err
			}
			build := &Inject{
				Pkg: &PackageType{
					Name:  pkg.Name,
					Path:  pkg.PkgPath,
					Types: pkg.Types,
				},
				BasePath: basePath,
				Option: map[string]interface{}{
					buildName: option,
				},
			}
			result[pkg.Module.Path].Injects = append(result[pkg.Module.Path].Injects, build)
		}
		return
	})
	return
}

func Decode(optionPkgs map[string]string, module *packages.Module, pkgs *packages2.Packages, commentFuncs map[string][]string, commentFields *ast.CommentFields) (result map[string]*Module, err error) {
	return (&Decoder{
		optionPkgs:     optionPkgs,
		module:         module,
		pkgs:           pkgs,
		commentFuncMap: commentFuncs,
		commentFields:  commentFields,
	}).decode()
}
