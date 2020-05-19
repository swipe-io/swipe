package writer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	stdtypes "go/types"
	"sort"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/pkg/astcopy"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type ImportInfo struct {
	Name    string
	Differs bool
}

type Writer struct {
	pkg         *packages.Package
	allPkgs     []*packages.Package
	basePath    string
	imports     map[string]ImportInfo
	anonImports map[string]bool
	buf         bytes.Buffer
}

func (w *Writer) Inspect(f func(p *packages.Package, n ast.Node) bool) {
	for _, p := range w.allPkgs {
		for _, pp := range p.Imports {
			for _, file := range pp.Syntax {
				ast.Inspect(file, func(n ast.Node) bool {
					return f(pp, n)
				})
			}
		}
	}
}

func (w *Writer) BasePath() string {
	return w.basePath
}

func (w *Writer) WriteCheckErr(body func()) {
	w.Write("if err != nil {\n")
	body()
	w.Write("}\n")
}

func (w *Writer) WriteType(name string) {
	w.Write("type %s ", name)
}

func (w *Writer) WriteFunc(name, recv string, params, results []string, body func()) {
	w.Write("func")

	if recv != "" {
		w.Write(" (%s)", recv)
	}

	w.Write(" %s(", name)
	w.WriteSignature(params)
	w.Write(") ")

	if len(results) > 0 {
		w.Write("( ")
		w.WriteSignature(results)
		w.Write(") ")
	}

	w.Write("{\n")
	body()
	w.Write("}\n\n")
}

func (w *Writer) WriteDefer(params []string, calls []string, body func()) {
	w.Write("defer func(")
	w.Write(stdstrings.Join(params, ","))
	w.Write(") {\n")
	body()
	w.Write("}(")
	w.Write(stdstrings.Join(calls, ","))
	w.Write(")\n")
}

func (w *Writer) WriteSignature(keyvals []string) {
	if len(keyvals) == 0 {
		return
	}
	if len(keyvals)%2 != 0 {
		panic("missing Value")
	}
	for i := 0; i < len(keyvals); i += 2 {
		if i > 0 {
			w.Write(", ")
		}
		name := "_"
		if keyvals[i] != "" {
			name = keyvals[i]
		}
		w.Write("%s %s", name, keyvals[i+1])
	}
}

func (w *Writer) WriteTypeStruct(name string, keyvals []string) {
	w.WriteType(name)
	w.WriteStruct(keyvals, false)
	w.WriteLn()
	w.WriteLn()
}

func (w *Writer) WriteStruct(keyvals []string, assign bool) {
	w.Write(" struct ")
	if assign {
		w.WriteStructAssign(keyvals)
	} else {
		w.WriteStructDefined(keyvals)
	}
}

func (w *Writer) WriteStructDefined(keyvals []string) {
	if len(keyvals)%2 != 0 {
		panic("missing Value")
	}
	w.Write("{\n")
	for i := 0; i < len(keyvals); i += 2 {
		w.Write("%s %s\n", keyvals[i], keyvals[i+1])
		continue
	}
	w.Write("}")
}

func (w *Writer) WriteStructAssign(keyvals []string) {
	if len(keyvals)%2 != 0 {
		panic("missing Value")
	}
	w.Write("{")
	for i := 0; i < len(keyvals); i += 2 {
		if i > 0 {
			w.Write(", ")
		}
		w.Write("%s: %s", keyvals[i], keyvals[i+1])
	}
	w.Write("}")
}

func (w *Writer) WriteFuncCall(id, name string, params []string) {
	w.Write(id + "." + name + "(")
	w.Write(stdstrings.Join(params, ","))
	w.Write(")\n")
}

func (w *Writer) WriteLn() {
	w.Write("\n")
}

func (w *Writer) Write(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&w.buf, format, args...)
}

func (w *Writer) WriteAST(node ast.Node) {
	node = w.RewritePkgRefs(node)
	if err := printer.Fprint(&w.buf, w.pkg.Fset, node); err != nil {
		panic(err)
	}
}

func (w *Writer) getConvertFunc(kind stdtypes.BasicKind, tmpId, valueId string) string {
	funcName := w.getConvertFuncName(kind)
	if funcName == "" {
		return ""
	}
	return fmt.Sprintf("%s, err := %s.%s", tmpId, w.Import("strconv", "strconv"), fmt.Sprintf(funcName, valueId))
}

func (w *Writer) getConvertFuncName(kind stdtypes.BasicKind) string {
	switch kind {
	case stdtypes.Int, stdtypes.Int8, stdtypes.Int16, stdtypes.Int32, stdtypes.Int64:
		return "Atoi(%s)"
	case stdtypes.Float32, stdtypes.Float64:
		return "ParseFloat(%s, " + types.GetBitSize(kind) + ")"
	case stdtypes.Uint, stdtypes.Uint8, stdtypes.Uint16, stdtypes.Uint32, stdtypes.Uint64:
		return "ParseUint(%s, 10, " + types.GetBitSize(kind) + ")"
	case stdtypes.Bool:
		return "ParseBool(%s)"
	default:
		return ""
	}
}

func (w *Writer) getFormatFunc(kind stdtypes.BasicKind, valueId string) string {
	funcName := w.getFormatFuncName(kind)
	if funcName == "" {
		return valueId
	}
	return fmt.Sprintf("%s.%s", w.Import("strconv", "strconv"), fmt.Sprintf(funcName, valueId))
}

func (w *Writer) getFormatFuncName(kind stdtypes.BasicKind) string {
	switch kind {
	case stdtypes.Int, stdtypes.Int8, stdtypes.Int16, stdtypes.Int32, stdtypes.Int64:
		return "FormatInt(int64(%s), 10)"
	case stdtypes.Float32, stdtypes.Float64:
		return "FormatFloat(float64(%s), 'w', 2, " + types.GetBitSize(kind) + ")"
	case stdtypes.Uint, stdtypes.Uint8, stdtypes.Uint16, stdtypes.Uint32, stdtypes.Uint64:
		return "FormatUint(uint64(%s), 10)"
	case stdtypes.Bool:
		return "FormatBool(%s)"
	default:
		return ""
	}
}

func (w *Writer) GetFormatType(valueId string, f *stdtypes.Var) string {
	switch t := f.Type().(type) {
	case *stdtypes.Basic:
		return w.getFormatFunc(t.Kind(), valueId)
	}
	return valueId
}

func (w *Writer) writeConvertBasicType(name, assignId, valueId string, t *stdtypes.Basic, sliceErr string, declareVar bool) {
	useCheckErr := true

	fmtPkg := w.Import("fmt", "fmt")
	tmpId := stdstrings.ToLower(name) + strings.UcFirst(t.String())

	convertFunc := w.getConvertFunc(t.Kind(), tmpId, valueId)
	if convertFunc != "" {
		w.Write("%s\n", convertFunc)
	} else {
		useCheckErr = false
		tmpId = valueId
	}
	if useCheckErr {
		errMsg := strconv.Quote("convert error " + name + ": %w")
		w.Write("if err != nil {\n")
		if sliceErr == "" {
			w.Write("return nil, %s.Errorf(%s, err)\n", fmtPkg, errMsg)
		} else {
			w.Write("%[1]s = append(%[1]s, %s.Errorf(%s, err))\n", sliceErr, fmtPkg, errMsg)
		}
		w.Write("}\n")
	}

	if declareVar {
		w.Write("var ")
	}

	w.Write("%s = ", assignId)
	if t.Kind() != stdtypes.String {
		w.Write("%s(%s)", t.String(), tmpId)
	} else {
		w.Write("%s", tmpId)
	}
	w.Write("\n")
}

func (w *Writer) WriteConvertType(assignId, valueId string, f *stdtypes.Var, sliceErr string, declareVar bool) {
	var tmpId string

	switch t := f.Type().(type) {
	case *stdtypes.Basic:
		w.writeConvertBasicType(f.Name(), assignId, valueId, t, sliceErr, declareVar)
	case *stdtypes.Slice:
		stringsPkg := w.Import("strings", "strings")
		switch t := t.Elem().(type) {
		case *stdtypes.Basic:
			tmpId = "parts" + stdstrings.ToLower(f.Name()) + strings.UcFirst(t.String())
			w.Write("%s := %s.Split(%s, \",\")\n", tmpId, stringsPkg, valueId)
			switch t.Kind() {
			case stdtypes.Int:
				if declareVar {
					w.Write("var ")
				}
				w.Write("%s = make([]%s, len(%s))\n", assignId, t.String(), tmpId)
				w.Write("for i, s := range %s {\n", tmpId)
				w.writeConvertBasicType("tmp", assignId+"[i]", "s", t, sliceErr, false)
				w.Write("}\n")
			}
		}
	case *stdtypes.Pointer:
		if t.Elem().String() == "net/url.URL" {
			urlPkg := w.Import("url", "net/url")
			tmpId := stdstrings.ToLower(f.Name()) + "URL"
			w.Write("%s, err := %s.Parse(%s)\n", tmpId, urlPkg, valueId)
			w.Write("if err != nil {\n")
			if sliceErr == "" {
				w.Write("return nil, err\n")
			} else {
				w.Write("%[1]s = append(%[1]s, err)\n", sliceErr)
			}
			w.Write("}\n")
			if declareVar {
				w.Write("var ")
			}
			w.Write("%s = %s\n", assignId, tmpId)
		}
	case *stdtypes.Named:
		if t.Obj().Pkg().Path() == "github.com/satori/go.uuid" {
			uuidPkg := w.Import("", t.Obj().Pkg().Path())
			if declareVar {
				w.Write("var ")
			}
			w.Write("%s, err = %s.FromString(%s)\n", assignId, uuidPkg, valueId)
			w.Write("if err != nil {\n")
			if sliceErr == "" {
				w.Write("return nil, err\n")
			} else {
				w.Write("%[1]s = append(%[1]s, err)\n", sliceErr)
			}
			w.Write("}\n")
		}
	}
}

func (w *Writer) Frame() []byte {
	var buf bytes.Buffer
	buf.WriteString("// Code generated by Swipe. DO NOT EDIT.\n\n")
	buf.WriteString("//go:generate swipe\n")
	buf.WriteString("package ")
	buf.WriteString(w.pkg.Name)
	buf.WriteString("\n\n")

	if w.HasImports() {
		buf.WriteString("import (\n")
		for _, impPath := range w.GetSortedImports() {
			info := w.imports[impPath]
			if info.Differs {
				_, _ = fmt.Fprintf(&buf, "\t%s %q\n", info.Name, impPath)
			} else {
				_, _ = fmt.Fprintf(&buf, "\t%q\n", impPath)
			}
		}
		buf.WriteString(")\n\n")
	}
	if len(w.anonImports) > 0 {
		buf.WriteString("import (\n")
		anonImps := make([]string, 0, len(w.anonImports))
		for impPath := range w.anonImports {
			anonImps = append(anonImps, impPath)
		}
		sort.Strings(anonImps)

		for _, impPath := range anonImps {
			_, _ = fmt.Fprintf(&buf, "\t_ %s\n", impPath)
		}
		buf.WriteString(")\n\n")
	}
	buf.Write(w.buf.Bytes())
	return buf.Bytes()
}

func (w *Writer) qualifyPkg(pkg *stdtypes.Package) string {
	return w.Import(pkg.Name(), pkg.Path())
}

func (w *Writer) Import(name, path string) string {
	if path == w.pkg.PkgPath {
		return ""
	}
	const vendorPart = "vendor/"
	unvendored := path
	if i := stdstrings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		unvendored = path[i+len(vendorPart):]
	}
	if info, ok := w.imports[unvendored]; ok {
		return info.Name
	}
	newName := disambiguate(name, func(n string) bool {
		return n == "err" || w.nameInFileScope(n)
	})
	w.imports[unvendored] = ImportInfo{
		Name:    newName,
		Differs: newName != name,
	}
	return newName
}

func (w *Writer) RewritePkgRefs(node ast.Node) ast.Node {
	start, end := node.Pos(), node.End()

	node = astcopy.CopyAST(node)

	node = astutil.Apply(node, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.Ident:
			obj := w.pkg.TypesInfo.ObjectOf(node)
			if obj == nil {
				return false
			}
			if pkg := obj.Pkg(); pkg != nil && obj.Parent() == pkg.Scope() && pkg.Path() != w.pkg.PkgPath {
				newPkgID := w.Import(pkg.Name(), pkg.Path())
				c.Replace(&ast.SelectorExpr{
					X:   ast.NewIdent(newPkgID),
					Sel: ast.NewIdent(node.Name),
				})
				return false
			}
			return true
		case *ast.SelectorExpr:
			pkgIdent, ok := node.X.(*ast.Ident)
			if !ok {
				return true
			}
			pkgName, ok := w.pkg.TypesInfo.ObjectOf(pkgIdent).(*stdtypes.PkgName)
			if !ok {
				return true
			}
			imported := pkgName.Imported()
			newPkgID := w.Import(imported.Name(), imported.Path())
			c.Replace(&ast.SelectorExpr{
				X:   ast.NewIdent(newPkgID),
				Sel: ast.NewIdent(node.Sel.Name),
			})
			return false
		default:
			return true
		}
	}, nil)
	newNames := make(map[stdtypes.Object]string)
	inNewNames := func(n string) bool {
		for _, other := range newNames {
			if other == n {
				return true
			}
		}
		return false
	}
	var scopeStack []*stdtypes.Scope
	pkgScope := w.pkg.Types.Scope()
	node = astutil.Apply(node, func(c *astutil.Cursor) bool {
		if scope := w.pkg.TypesInfo.Scopes[c.Node()]; scope != nil {
			scopeStack = append(scopeStack, scope)
		}
		id, ok := c.Node().(*ast.Ident)
		if !ok {
			return true
		}
		obj := w.pkg.TypesInfo.ObjectOf(id)
		if obj == nil {
			return true
		}
		if n, ok := newNames[obj]; ok {
			c.Replace(ast.NewIdent(n))
			return false
		}
		if par := obj.Parent(); par == nil || par == pkgScope {
			return true
		}
		objName := obj.Name()
		if pos := obj.Pos(); pos < start || end <= pos || !(w.nameInFileScope(objName) || inNewNames(objName)) {
			return true
		}
		newName := disambiguate(objName, func(n string) bool {
			if w.nameInFileScope(n) || inNewNames(n) {
				return true
			}
			if len(scopeStack) > 0 {
				_, obj := scopeStack[len(scopeStack)-1].LookupParent(n, token.NoPos)
				if obj != nil {
					return true
				}
			}
			return false
		})
		newNames[obj] = newName
		c.Replace(ast.NewIdent(newName))
		return false
	}, func(c *astutil.Cursor) bool {
		if w.pkg.TypesInfo.Scopes[c.Node()] != nil {
			scopeStack = scopeStack[:len(scopeStack)-1]
		}
		return true
	})
	return node
}

func (w *Writer) nameInFileScope(name string) bool {
	for _, other := range w.imports {
		if other.Name == name {
			return true
		}
	}
	_, obj := w.pkg.Types.Scope().LookupParent(name, token.NoPos)
	return obj != nil
}

func (w *Writer) HasImports() bool {
	return len(w.imports) > 0
}

func (w *Writer) GetSortedImports() []string {
	imps := make([]string, 0, len(w.imports))
	for impPath := range w.imports {
		imps = append(imps, impPath)
	}
	sort.Strings(imps)
	return imps
}

func (w *Writer) AddAnonImport(path string) {
	w.anonImports[path] = true
}

func (w *Writer) TypeString(t stdtypes.Type) string {
	return stdtypes.TypeString(t, w.qualifyPkg)
}

func (w *Writer) TypeOf(expr ast.Expr) stdtypes.Type {
	return w.pkg.TypesInfo.TypeOf(expr)
}

func (w *Writer) ZeroValue(t stdtypes.Type) string {
	switch u := t.Underlying().(type) {
	case *stdtypes.Array, *stdtypes.Struct:
		return w.TypeString(t) + "{}"
	case *stdtypes.Basic:
		info := u.Info()
		switch {
		case info&stdtypes.IsBoolean != 0:
			return "false"
		case info&(stdtypes.IsInteger|stdtypes.IsFloat|stdtypes.IsComplex) != 0:
			return "0"
		case info&stdtypes.IsString != 0:
			return `""`
		default:
			panic("unreachable")
		}
	case *stdtypes.Chan, *stdtypes.Interface, *stdtypes.Map, *stdtypes.Pointer, *stdtypes.Signature, *stdtypes.Slice:
		return "nil"
	default:
		panic("unreachable")
	}
}

func NewWriter(pkg *packages.Package, allPkgs []*packages.Package, basePath string) *Writer {
	return &Writer{
		pkg:         pkg,
		allPkgs:     allPkgs,
		basePath:    basePath,
		imports:     map[string]ImportInfo{},
		anonImports: map[string]bool{},
	}
}
