package assembly

import (
	"fmt"
	"go/ast"
	stdtypes "go/types"
	"strings"

	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/writer"
)

var nilType = stdtypes.Universe.Lookup("nil").Type()

type scope string

const (
	dtoScope   scope = "dto"
	modelScope       = "model"
)

type formatter struct {
	fn   *stdtypes.Func
	expr ast.Expr
}

type mapping struct {
	from, to *stdtypes.Var
	fields   []mapping
}

type mappingStruct struct {
	v      *stdtypes.Var
	fields map[string]mappingStruct
}

type Assembly struct {
	w *writer.Writer
}

func (w *Assembly) Write(opt *parser.Option) error {
	var (
		excludeDTO   = []string{}
		excludeModel = []string{}
	)

	dtoOpt := parser.MustOption(opt.Get("dto"))
	modelOpt := parser.MustOption(opt.Get("model"))

	dtoType := dtoOpt.Value.Type()
	modelType := modelOpt.Value.Type()

	dtoNamed, ok := dtoType.(*stdtypes.Named)
	if !ok {
		return errors.NotePosition(dtoOpt.Position,
			fmt.Errorf("the %s must be a struct type; found %s", dtoOpt.Name, w.w.TypeString(dtoOpt.Value.Type())))
	}
	modelNamed, ok := modelType.(*stdtypes.Named)
	if !ok {
		return errors.NotePosition(modelOpt.Position,
			fmt.Errorf("the %s must be a struct type; found %s", modelOpt.Name, w.w.TypeString(modelOpt.Value.Type())))
	}
	if opt, ok := opt.Get("AssemblyExclude"); ok {
		excludeDTO = parser.MustOption(opt.Get("dto")).Value.StringSlice()
		excludeModel = parser.MustOption(opt.Get("model")).Value.StringSlice()
	}

	dtoStructType := dtoType.Underlying().(*stdtypes.Struct)
	modelStructType := modelType.Underlying().(*stdtypes.Struct)

	fromattersDTO := map[string]parser.Value{}
	fromattersModel := map[string]parser.Value{}

	if opts, ok := opt.GetSlice("AssemblyFormatter"); ok {
		for _, opt := range opts {
			fldOpt := parser.MustOption(opt.Get("fieldName"))
			fieldName := fldOpt.Value.String()

			fnDTOOpt := parser.MustOption(opt.Get("formatterDTO"))
			fnType := fnDTOOpt.Value.Type()

			if !stdtypes.Identical(fnType, nilType) {
				_, ok := fnType.(*stdtypes.Signature)
				if !ok {
					return errors.NotePosition(opt.Position, fmt.Errorf("the %s param %s must be func", opt.Name, fnDTOOpt.Name))
				}
				fromattersDTO[fieldName] = fnDTOOpt.Value
			}

			fnModelOpt := parser.MustOption(opt.Get("formatterModel"))
			fnType = fnModelOpt.Value.Type()

			if !stdtypes.Identical(fnType, nilType) {
				_, ok = fnType.(*stdtypes.Signature)
				if !ok {
					return errors.NotePosition(opt.Position, fmt.Errorf("the %s param %s must be func", opt.Name, fnModelOpt.Name))
				}
				fromattersModel[fieldName] = fnModelOpt.Value
			}
		}
	}

	fieldsMapping := map[string]string{}
	if opt, ok := opt.Get("AssemblyMapping"); ok {
		mappingValues := opt.Value.StringSlice()
		for i := 0; i < len(mappingValues); i += 2 {
			fieldsMapping[mappingValues[i]] = mappingValues[i+1]
		}
	}

	fromDTOToModelFuncName := fmt.Sprintf("AssemblyFromDTO%sToModel%s", dtoNamed.Obj().Name(), modelNamed.Obj().Name())
	fromModelToDTOFuncName := fmt.Sprintf("AssemblyFromModel%sToDTO%s", modelNamed.Obj().Name(), dtoNamed.Obj().Name())

	dtoTypeStr := w.w.TypeString(dtoType)
	modelTypeStr := w.w.TypeString(modelType)

	maps := w.makeMapping(modelStructType, dtoStructType, fieldsMapping, "")

	w.w.WriteFunc(
		fromDTOToModelFuncName,
		"",
		[]string{"from", dtoTypeStr},
		[]string{"to", modelTypeStr},
		func() {
			w.w.Write("to = %s", modelTypeStr)
			w.writeMapping(maps, checkForExclude(excludeModel), fromattersModel, "", false)
			w.w.Write("\n")
			w.w.Write("return\n")
		},
	)

	w.w.WriteFunc(
		fromModelToDTOFuncName,
		"",
		[]string{"from", modelTypeStr},
		[]string{"to", dtoTypeStr},
		func() {
			w.w.Write("to = %s", dtoTypeStr)
			w.writeMapping(maps, checkForExclude(excludeDTO), fromattersDTO, "", true)
			w.w.Write("\n")
			w.w.Write("return\n")
		},
	)

	return nil
}

func (w *Assembly) writeMapping(maps []mapping, checkFn func(string) bool, formatters map[string]parser.Value, prefix string, revert bool) error {
	w.w.Write("{")
	for _, m := range maps {
		if m.to == nil {
			continue
		}

		var from, to = m.from, m.to
		if revert {
			from, to = m.to, m.from
		}

		fieldPath := prefix + "." + from.Name()

		if checkFn != nil && !checkFn(fieldPath) {
			continue
		}

		if f, ok := formatters[fieldPath]; ok {
			w.w.Write("%s: ", to.Name())
			w.w.WriteAST(f.Expr())
			w.w.Write("(from),")
		} else {

			if len(m.fields) == 0 {
				w.w.Write("%s: from%s.%s,", to.Name(), prefix, from.Name())
			} else {
				w.w.Write("%s: %s", to.Name(), w.w.TypeString(from.Type()))
				if err := w.writeMapping(m.fields, checkFn, formatters, fieldPath, !revert); err != nil {
					return err
				}
			}
		}
	}
	w.w.Write("}")

	return nil
}

func (w *Assembly) makeMapping(from, to *stdtypes.Struct, fieldsMapping map[string]string, prefix string) (result []mapping) {
	for i := 0; i < from.NumFields(); i++ {
		m := mapping{}

		fromField := from.Field(i)

		fromFieldName := prefix + "." + fromField.Name()

		toField := findFieldByName(fromFieldName, to, prefix)
		if toFieldName, ok := fieldsMapping[fromFieldName]; ok && toField == nil {
			toField = findFieldByName(toFieldName, to, prefix)
		}

		var fields []mapping

		if toField != nil {
			if fromField.Type().Underlying().String() != "time.Time" {
				if nestedFrom, ok := fromField.Type().Underlying().(*stdtypes.Struct); ok {
					if nestedTo, ok := toField.Type().Underlying().(*stdtypes.Struct); ok {
						fields = w.makeMapping(nestedFrom, nestedTo, fieldsMapping, fromFieldName)
					}
				}
			}
		}

		m.from = fromField
		m.to = toField

		m.fields = fields

		result = append(result, m)

		// for j := 0; j < to.NumFields(); j++ {

		// toField := to.Field(j)

		// if name, ok := fieldsMapping[prefix+"."+fa.Name()]; ok {
		// fbName = name
		// }

		// fmt.Println(fromField, toField)

		// fbName := fb.Name()

		// if fa.Name() == name {
		// 	m := mapping{name: fa.Name(), from: fa, to: fb}

		// 	fas, fasOK := fa.Type().Underlying().(*stdtypes.Struct)
		// 	fbs, fbsOK := fb.Type().Underlying().(*stdtypes.Struct)

		// 	if fasOK && fbsOK {
		// 		m.fields = w.makeMapping(fas, fbs, fieldsMapping, prefix+"."+fa.Name())
		// 	}

		// 	result = append(result, m)
		// }
		// }
	}
	return
}

func findFieldByName(name string, s *stdtypes.Struct, prefix string) *stdtypes.Var {
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if prefix+"."+f.Name() == name {
			return f
		}
	}
	return nil
}

func checkForExclude(excludes []string) func(name string) bool {
	return func(name string) bool {
		found := false
		for _, s := range excludes {
			if strings.Contains("."+s, name) {
				found = true
				break
			}
		}
		return !found
	}
}

func New(w *writer.Writer) *Assembly {
	return &Assembly{w: w}
}
