package configenv

import (
	stdstrings "strings"

	option2 "github.com/swipe-io/swipe/v2/internal/option"

	"github.com/fatih/structtag"
	"github.com/swipe-io/strcase"
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
	typeStr   string
}

func (o fldOpts) tagName() string {
	if o.isFlag {
		return "flag"
	}
	return "env"
}

func getFieldOpts(f *option2.VarType, tags *structtag.Tags) (result fldOpts) {
	result.name = strcase.ToScreamingSnake(f.Name.UpperCase)
	result.fieldPath = f.Name.UpperCase

	if tag, err := tags.Get("env"); err == nil {
		for _, o := range tag.Options {
			switch o {
			case "use_zero":
				result.useZero = true
			case "required":
				result.required = true
			case "use_flag":
				result.name = strcase.ToKebab(f.Name.UpperCase)
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
