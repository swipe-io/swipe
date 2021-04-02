package option

import (
	goast "go/ast"
	stdtypes "go/types"
	"strings"

	"github.com/fatih/structtag"
	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/types"

	"github.com/swipe-io/swipe/v2/internal/annotation"
	"github.com/swipe-io/swipe/v2/internal/ast"

	"golang.org/x/tools/go/packages"
)

type namedObject interface {
	Exported() bool
	Name() string
}

type Build struct {
	PkgPath string
	PkgName string
	Option  interface{}
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
	loader *ast.Loader
}

func normalizeName(v namedObject) NameType {
	nt := NameType{}
	if v.Exported() {
		nt.Exported = v.Name()
		nt.Var = strcase.ToLowerCamel(v.Name())
		nt.NotExported = nt.Var
	} else {
		nt.Exported = strcase.ToCamel(v.Name())
		nt.Var = v.Name()
		nt.NotExported = nt.Var
	}
	return nt
}

func (d *Decoder) normalizeVar(t *stdtypes.Var, comment string) *VarType {
	if t == nil {
		return nil
	}
	return &VarType{
		Name:     normalizeName(t),
		Embedded: t.Embedded(),
		Exported: t.Exported(),
		IsField:  t.IsField(),
		Type:     d.normalizeType(t.Type().Underlying(), nil),
		Comment:  comment,
	}
}

func (d *Decoder) normalizeStruct(obj *stdtypes.TypeName, t *stdtypes.Struct, pkg *PackageType) *StructType {
	if t == nil {
		return nil
	}
	var name NameType
	if obj != nil {
		name = normalizeName(obj)
	}
	result := &StructType{
		Name: name,
		Pkg:  pkg,
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

func (d *Decoder) normalizeType(t stdtypes.Type, pkg *PackageType) interface{} {
	switch t := t.(type) {
	case *stdtypes.Pointer:
		return d.normalizeType(t.Elem(), pkg)
	case *stdtypes.Struct:
		return d.normalizeStruct(nil, t, nil)
	case *stdtypes.Signature:
		return d.normalizeSignature(t, nil)
	case *stdtypes.Named:
		switch tt := t.Obj().Type().Underlying().(type) {
		case *stdtypes.Interface:
			return d.normalizeInterface(t.Obj(), tt, pkg)
		case *stdtypes.Struct:
			return d.normalizeStruct(t.Obj(), tt, pkg)
		}
	case *stdtypes.Basic:
		return d.normalizeBasic(t)
	}
	return nil
}

func (d *Decoder) normalizeBasic(t *stdtypes.Basic) *BasicType {
	return &BasicType{
		Name: t.Name(),
		Zero: types.ZeroValue(t),
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
		it.Embeddeds = append(it.Embeddeds, d.normalizeType(t.EmbeddedType(i), nil))
	}
	for i := 0; i < t.NumExplicitMethods(); i++ {
		it.ExplicitMethods = append(it.ExplicitMethods, d.normalizeFunc(t.ExplicitMethod(i)))
	}
	return it
}

func (d *Decoder) normalizeFunc(t *stdtypes.Func) *FuncType {
	comments := d.loader.CommentFuncs()[t.String()]

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

func (d *Decoder) normalizeObject(obj stdtypes.Object) interface{} {
	var pkg *PackageType
	if obj.Pkg() != nil {
		pkg = &PackageType{}
		fndPkg := d.loader.FindPkgByID(obj.Pkg().Path())
		if fndPkg != nil {
			pkg.Name = fndPkg.Name
			pkg.Path = fndPkg.PkgPath
			if fndPkg.Module != nil {
				pkg.External = fndPkg.Module.Path != d.loader.Pkg().Module.Path
				pkg.Module = &ModuleType{
					Version: fndPkg.Module.Version,
					Path:    fndPkg.Module.Path,
				}
			}
		}
	}
	return d.normalizeType(obj.Type(), pkg)
}

func (d *Decoder) decode(pkg *packages.Package, expr goast.Expr) (interface{}, error) {
	switch e := expr.(type) {
	case *goast.CompositeLit:
		switch vt := e.Type.(type) {
		case *goast.Ident:
			return d.normalizeObject(pkg.TypesInfo.Uses[vt]), nil
		case *goast.ArrayType:
			switch elt := vt.Elt.(type) {
			case *goast.Ident:
				switch elt.Name {
				case "string":
					return makeStringSlice(e.Elts, pkg.TypesInfo), nil
				}
			}
			return e.Elts, nil
		}
	case *goast.BasicLit:
		var value interface{}
		tv := pkg.TypesInfo.Types[e]
		if tv.IsValue() {
			value = getValue(tv.Value)
		}
		return value, nil
	case *goast.Ident:
		return d.normalizeObject(pkg.TypesInfo.Uses[e]), nil
	case *goast.StarExpr:
		return d.decode(pkg, e.X)
	case *goast.UnaryExpr:
		return d.decode(pkg, e.X)
	case *goast.ParenExpr:
		return d.decode(pkg, e.X)
	case *goast.SelectorExpr:
		return d.normalizeObject(pkg.TypesInfo.Uses[e.Sel]), nil
	case *goast.CallExpr:
		switch v := e.Fun.(type) {
		default:
			return d.decode(pkg, e.Fun)
		case *goast.Ident:
			if fnObj := pkg.TypesInfo.Uses[v]; fnObj != nil {
				sig := fnObj.Type().(*stdtypes.Signature)
				if len(e.Args) > 0 {
					result := make(map[string]interface{}, len(e.Args))
					for i, arg := range e.Args {
						var key string
						var valueType = "value"

						if callArg, ok := arg.(*goast.CallExpr); ok {
							if callIdent, ok := callArg.Fun.(*goast.Ident); ok {
								callObj := pkg.TypesInfo.Uses[callIdent]
								if callObj == nil {
									continue
								}
								key = callObj.Name()

								comments := d.loader.CommentFuncs()[callObj.String()]
								annotationOpts, _ := annotation.Parse(strings.Join(comments, " "))
								if annotationOpts != nil {
									if annotationOpt, err := annotationOpts.Get("type"); err == nil {
										valueType = annotationOpt.Value()
									}
								}
							}
						}
						if key == "" {
							vr := sigParamAt(sig, i)
							if vr.Name() == "" {
								continue
							}
							key = vr.Name()
						}
						val, err := d.decode(pkg, arg)
						if err != nil {
							return nil, err
						}

						if valueType == "repeat" {
							if _, ok := result[key]; !ok {
								result[key] = []interface{}{}
							}

							v := result[key].([]interface{})
							v = append(v, val)
							result[key] = v

						} else {
							result[key] = val
						}
					}
					return map[string]interface{}{
						fnObj.Name(): result,
					}, nil
				}
				return true, nil
			}
		}
	}
	return nil, nil
}

func (d *Decoder) Decode() (result map[string]*Module, err error) {
	result = map[string]*Module{}
	for _, pkg := range d.loader.Pkgs() {

		for expr := range pkg.TypesInfo.Types {
			callExpr, ok := expr.(*goast.CallExpr)
			if !ok {
				continue
			}
			callIdent, ok := callExpr.Fun.(*goast.Ident)
			if !ok {
				continue
			}
			fnObj := pkg.TypesInfo.Uses[callIdent]
			if fnObj == nil {
				continue
			}
			if fnObj.Name() == "Build" {
				if _, ok := result[pkg.Module.Path]; !ok {
					result[pkg.Module.Path] = &Module{
						Path:     pkg.Module.Path,
						External: d.loader.Pkg().Module.Path != pkg.Module.Path,
					}
				}
				for _, arg := range callExpr.Args {
					val, err := d.decode(pkg, arg)
					if err != nil {
						return nil, err
					}
					result[pkg.Module.Path].Builds = append(result[pkg.Module.Path].Builds, &Build{
						PkgPath: pkg.PkgPath,
						PkgName: pkg.Name,
						Option:  val,
					})
				}
			}
		}
	}
	return
}

func NewDecoder(loader *ast.Loader) *Decoder {
	return &Decoder{loader: loader}
}
