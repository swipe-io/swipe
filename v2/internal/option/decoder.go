package option

import (
	"errors"
	goast "go/ast"
	"go/constant"
	"go/token"
	stdtypes "go/types"

	"github.com/fatih/structtag"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/annotation"
	"github.com/swipe-io/swipe/v2/internal/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type namedObject interface {
	Exported() bool
	Name() string
	Pkg() *stdtypes.Package
}

type Build struct {
	Pkg      *PackageType
	BasePath string
	Option   map[string]interface{}
}

type Module struct {
	Path     string
	External bool
	Builds   []*Build
}

type Result struct {
	Modules map[string][]Module
}

type Decoder struct {
	pkg          *packages.Package
	pkgs         []*packages.Package
	commentFuncs map[string][]string
}

func normalizeName(v namedObject) NameType {
	nt := NameType{}

	nt.Origin = v.Name()

	if v.Exported() {
		nt.UpperCase = v.Name()
		nt.Var = strcase.ToLowerCamel(v.Name())
		nt.LowerCase = nt.Var
	} else {
		nt.UpperCase = strcase.ToCamel(v.Name())
		nt.Var = v.Name()
		nt.LowerCase = nt.Var
	}
	return nt
}

func (d *Decoder) normalizeVar(t *stdtypes.Var, comment string) *VarType {
	if t == nil {
		return nil
	}
	var pkg *PackageType
	if t.Pkg() != nil {
		pkg = &PackageType{
			Name: t.Pkg().Name(),
			Path: t.Pkg().Path(),
		}
	}
	return &VarType{
		Name:     normalizeName(t),
		Embedded: t.Embedded(),
		Exported: t.Exported(),
		IsField:  t.IsField(),
		Type:     d.normalizeType(t.Type().Underlying(), pkg, false),
		Zero:     zeroValue(t.Type().Underlying()),
		Comment:  comment,
	}
}

func (d *Decoder) normalizeStruct(obj *stdtypes.TypeName, t *stdtypes.Struct, pkg *PackageType, isPointer bool) *StructType {
	if t == nil {
		return nil
	}
	var name NameType
	if obj != nil {
		name = normalizeName(obj)
	}
	result := &StructType{
		Name:      name,
		Pkg:       pkg,
		IsPointer: isPointer,
	}
	for i := 0; i < t.NumFields(); i++ {
		f := &StructFieldType{
			Var: d.normalizeVar(t.Field(i), ""),
		}
		if tags, err := structtag.Parse(t.Tag(i)); err == nil {
			f.Tags = tags
		}
		result.Fields = append(result.Fields, f)
	}
	return result
}

func (d *Decoder) normalizeType(t stdtypes.Type, pkg *PackageType, isPointer bool) interface{} {
	switch t := t.(type) {
	case *stdtypes.Map:
		return d.normalizeMap(t.Key(), t.Elem(), pkg, isPointer)
	case *stdtypes.Slice:
		return d.normalizeSlice(t.Elem(), pkg, isPointer)
	case *stdtypes.Array:
		return d.normalizeArray(t.Elem(), t.Len(), pkg, isPointer)
	case *stdtypes.Pointer:
		return d.normalizeType(t.Elem(), pkg, true)
	case *stdtypes.Struct:
		return d.normalizeStruct(nil, t, pkg, isPointer)
	case *stdtypes.Signature:
		return d.normalizeSignature(t, nil)
	case *stdtypes.Named:
		switch tt := t.Obj().Type().Underlying().(type) {
		case *stdtypes.Interface:
			return d.normalizeInterface(t.Obj(), tt, pkg)
		case *stdtypes.Struct:
			return d.normalizeStruct(t.Obj(), tt, pkg, isPointer)
		}
	case *stdtypes.Basic:
		return d.normalizeBasic(t, isPointer)
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

func (d *Decoder) normalizeInterface(obj *stdtypes.TypeName, t *stdtypes.Interface, pkg *PackageType) *IfaceType {
	it := &IfaceType{
		Name: normalizeName(obj),
		Pkg:  pkg,
	}
	for i := 0; i < t.NumMethods(); i++ {
		it.Methods = append(it.Methods, d.normalizeFunc(t.Method(i)))
	}
	for i := 0; i < t.NumEmbeddeds(); i++ {
		it.Embeddeds = append(it.Embeddeds, d.normalizeType(t.EmbeddedType(i), nil, false))
	}
	for i := 0; i < t.NumExplicitMethods(); i++ {
		it.ExplicitMethods = append(it.ExplicitMethods, d.normalizeFunc(t.ExplicitMethod(i)))
	}
	return it
}

func (d *Decoder) normalizeFunc(t *stdtypes.Func) *FuncType {
	comments := d.commentFuncs[t.String()]

	comment, paramsComment := parseMethodComments(comments)

	return &FuncType{
		FullName: t.FullName(),
		Name:     normalizeName(t),
		Exported: t.Exported(),
		Sig:      d.normalizeSignature(t.Type().(*stdtypes.Signature), paramsComment),
		Comment:  comment,
	}
}

func (d *Decoder) normalizeSignature(t *stdtypes.Signature, comments map[string]string) *SignType {
	if t == nil {
		return nil
	}
	st := &SignType{
		Recv: d.normalizeVar(t.Recv(), ""),
	}
	var paramOffset int
	if t.Variadic() {
		st.Variadic = d.normalizeVar(t.Params().At(t.Params().Len()-1), "")
		paramOffset = 1
	}
	for i := 0; i < t.Params().Len()-paramOffset; i++ {
		v := t.Params().At(i)
		if types.IsContext(v.Type()) {
			st.Ctx = d.normalizeVar(v, "")
			continue
		}
		st.Params = append(st.Params, d.normalizeVar(v, comments[v.Name()]))
	}
	for i := 0; i < t.Results().Len(); i++ {
		v := t.Results().At(i)
		if i == 0 && v.Name() != "" {
			st.ResultNamed = true
		}
		if types.IsError(v.Type()) {
			st.Err = d.normalizeVar(v, "")
			continue
		}
		st.Results = append(st.Results, d.normalizeVar(v, ""))
	}
	return st
}

func (d *Decoder) normalizeMap(key stdtypes.Type, val stdtypes.Type, pkg *PackageType, isPointer bool) *MapType {
	mapType := &MapType{IsPointer: isPointer}
	mapType.Key = d.normalizeType(key, pkg, false)
	mapType.Value = d.normalizeType(val, pkg, false)
	return mapType
}

func (d *Decoder) normalizeSlice(val stdtypes.Type, pkg *PackageType, isPointer bool) *SliceType {
	return &SliceType{
		Value:     d.normalizeType(val, pkg, false),
		IsPointer: isPointer,
	}
}

func (d *Decoder) normalizeArray(val stdtypes.Type, len int64, pkg *PackageType, isPointer bool) *ArrayType {
	return &ArrayType{
		Value:     d.normalizeType(val, pkg, false),
		Len:       len,
		IsPointer: isPointer,
	}
}

func (d *Decoder) normalizeObject(obj stdtypes.Object) interface{} {
	var pkg *PackageType
	if obj.Pkg() != nil {
		pkg = &PackageType{}
		fndPkg := findPkgByID(d.pkgs, obj.Pkg().Path())
		if fndPkg != nil {
			pkg.Name = fndPkg.Name
			pkg.Path = fndPkg.PkgPath
			if fndPkg.Module != nil {
				pkg.Module = &ModuleType{
					Version:  fndPkg.Module.Version,
					Path:     fndPkg.Module.Path,
					External: fndPkg.Module.Path != d.pkg.Module.Path,
				}
			}
		}
	}
	return d.normalizeType(obj.Type(), pkg, false)
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
			return d.normalizeObject(pkg.TypesInfo.Uses[vt.Sel]), nil
		case *goast.Ident:
			return d.normalizeObject(pkg.TypesInfo.Uses[vt]), nil
		case *goast.ArrayType:
			switch elt := vt.Elt.(type) {
			default:
				var value []*SelectorType
				for _, expr := range e.Elts {
					if selExpr, ok := expr.(*goast.SelectorExpr); ok {
						var (
							selObj stdtypes.Object
							xObj   stdtypes.Object
						)
						selObj = pkg.TypesInfo.Uses[selExpr.Sel]
						if xIdent, ok := selExpr.X.(*goast.Ident); ok {
							xObj = pkg.TypesInfo.Uses[xIdent]
						}

						value = append(value, &SelectorType{
							Sel: d.normalizeObject(selObj),
							X:   d.normalizeObject(xObj),
						})
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
		return d.normalizeObject(pkg.TypesInfo.Uses[e]), nil
	case *goast.StarExpr:
		return d.decodeRecursive(pkg, e.X)
	case *goast.UnaryExpr:
		return d.decodeRecursive(pkg, e.X)
	case *goast.ParenExpr:
		return d.decodeRecursive(pkg, e.X)
	case *goast.SelectorExpr:
		return d.normalizeObject(pkg.TypesInfo.Uses[e.Sel]), nil
	case *goast.CallExpr:
		return d.decodeRecursive(pkg, e.Fun)
	}
	return nil, nil
}

func (d *Decoder) callDecodeArgs(pkg *packages.Package, obj stdtypes.Object, args []goast.Expr) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	sig := obj.Type().(*stdtypes.Signature)

	for i, arg := range args {
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
				comments := d.commentFuncs[obj.String()]
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
	for _, pkg := range d.pkgs {
		for expr := range pkg.TypesInfo.Types {
			expr = astutil.Unparen(expr)
			callExpr, ok := expr.(*goast.CallExpr)
			if !ok {
				continue
			}
			fun := callExpr.Fun
			if selExpr, ok := fun.(*goast.SelectorExpr); ok {
				fun = selExpr.Sel
			}
			callIdent, ok := fun.(*goast.Ident)
			if !ok {
				continue
			}
			obj := pkg.TypesInfo.Uses[callIdent]
			if obj == nil {
				continue
			}
			if obj.Name() == "Build" {
				if _, ok := result[pkg.Module.Path]; !ok {
					result[pkg.Module.Path] = &Module{
						Path:     pkg.Module.Path,
						External: d.pkg.Module.Path != pkg.Module.Path,
					}
				}
				option, err := d.callDecodeArgs(pkg, obj, callExpr.Args)
				if err != nil {
					return nil, err
				}
				basePath, err := detectBasePath(pkg)
				if err != nil {
					return nil, err
				}
				build := &Build{
					Pkg: &PackageType{
						Name: pkg.Name,
						Path: pkg.PkgPath,
					},
					BasePath: basePath,
					Option:   option,
				}
				result[pkg.Module.Path].Builds = append(result[pkg.Module.Path].Builds, build)
			}
		}
	}
	return
}

func Decode(pkg *packages.Package, pkgs []*packages.Package, commentFuncs map[string][]string) (result map[string]*Module, err error) {
	return (&Decoder{
		pkg:          pkg,
		pkgs:         pkgs,
		commentFuncs: commentFuncs,
	}).decode()
}
