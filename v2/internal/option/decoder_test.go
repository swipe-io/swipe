package option_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/swipe-io/swipe/v2/internal/option"

	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/swipe/v2/internal/ast"
)

type Interface struct {
	Type      *option.IfaceType `mapstructure:"iface"`
	Namespace string            `mapstructure:"ns"`
}

type OpenapiTag struct {
	Methods []*option.SelectorType `mapstructure:"methods"`
	Tags    []string               `mapstructure:"tags"`
}

type ServiceOptions struct {
	HTTPServer  *struct{}
	Interfaces  []*Interface `mapstructure:"Interface"`
	OpenapiTags *OpenapiTag  `mapstructure:"Tags"`
}

func TestParser_Parse(t *testing.T) {
	wd, err := filepath.Abs("./fixtures")
	if err != nil {
		t.Fatal(err)
	}
	astLoader, errs := ast.NewLoader(wd, os.Environ(), []string{"./fixtures/..", "github.pie.apple.com/ISS-Tools/zeus-service/pkg/..."})
	if len(errs) > 0 {
		for _, err := range errs {
			t.Log(err)
		}
		t.Fatal("AST loader failed")
	}

	modules, err := option.Decode(astLoader.Pkg(), astLoader.Pkgs(), astLoader.CommentFuncs(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, module := range modules {
		for _, build := range module.Builds {
			if opts, ok := build.Option["Service"]; ok {
				var o ServiceOptions
				err = mapstructure.Decode(opts, &o)
				fmt.Println(o, err)
			}
		}
	}

	//p := NewParser()
	//p.Parse(nil)

	//type args struct {
	//	s interface{}
	//}
	//tests := []struct {
	//	name string
	//	args args
	//	want interface{}
	//}{
	//	{
	//		name: "",
	//		args: args{s: &TestStruct{}},
	//		want: nil,
	//	},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		p := NewParser()
	//		if got := p.Parse(tt.args.s, data); !reflect.DeepEqual(got, tt.want) {
	//			t.Errorf("Parse() = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}
