package config

import (
	stdstrings "strings"

	"github.com/fatih/structtag"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/option"
)

type Bool bool

func (r Bool) String() string {
	if r {
		return "yes"
	}
	return "no"
}

type fldOpts struct {
	desc      string
	name      string
	fieldPath string
	required  Bool
	useZero   Bool
	isFlag    bool
	t         interface{}
}

func (o fldOpts) tagName() string {
	if o.isFlag {
		return "flag"
	}
	return "env"
}

func getFieldOpts(f *option.VarType, tags *structtag.Tags) (result fldOpts) {
	result.name = strcase.ToScreamingSnake(f.Name.Upper())
	result.fieldPath = f.Name.Value
	result.t = f.Type

	if tag, err := tags.Get("env"); err == nil {
		for _, o := range tag.Options {
			switch o {
			case "use_zero":
				result.useZero = true
			case "required":
				result.required = true
			case "use_flag":
				result.name = strcase.ToKebab(f.Name.Upper())
				result.isFlag = true
			default:
				if stdstrings.HasPrefix(o, "desc:") {
					descParts := stdstrings.Split(o, "desc:")
					if len(descParts) == 2 {
						result.desc = descParts[1]
					}
				}
			}
		}
		if tag.Name != "" {
			result.name = tag.Name
		}
	}

	return
}

type callbackFn func(f, parent *option.VarType, opts fldOpts)

func walkStructRecursive(st *option.StructType, parent *option.VarType, fPOpts fldOpts, fn callbackFn) {
	for _, field := range st.Fields {
		fOpts := getFieldOpts(field.Var, field.Tags)
		if fPOpts.name != "" && parent != nil {
			fOpts.name = fPOpts.name + "_" + fOpts.name
			fOpts.fieldPath = fPOpts.fieldPath + "." + fOpts.fieldPath
		}

		if !isExclusionStructType(field.Var) {
			walkRecursive(field.Var.Type, field.Var, fOpts, fn)
			continue
		}

		fn(field.Var, parent, fOpts)
	}
}

func walk(t interface{}, fn callbackFn) {
	walkRecursive(t, nil, fldOpts{}, fn)
}

func walkRecursive(t interface{}, parent *option.VarType, fPOpts fldOpts, fn callbackFn) {
	switch t := t.(type) {
	case *option.StructType:
		walkStructRecursive(t, parent, fPOpts, fn)
	case *option.NamedType:
		walkRecursive(t.Type, parent, fPOpts, fn)
	}
}

func isExclusionStructType(v *option.VarType) bool {
	switch t := v.Type.(type) {
	case *option.SliceType, *option.ArrayType, *option.MapType:
		return true
	case *option.BasicType:
		return true
	case *option.NamedType:
		switch t.Pkg.Path {
		case "github.com/google/uuid":
			switch t.Name.Value {
			case "UUID":
				return true
			}
		case "time":
			switch t.Name.Value {
			case "Time", "Duration":
				return true
			}
		case "net/url":
			switch t.Name.Value {
			case "URL":
				return true
			}
		}
	}
	return false
}
