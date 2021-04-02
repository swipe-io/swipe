package option

import (
	"fmt"
	goast "go/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/swipe-io/swipe/v2/internal/ast"
)

type Interface struct {
	Type      *IfaceType `mapstructure:"iface"`
	Namespace string     `mapstructure:"ns"`
}

type OpenapiTag struct {
	Methods []goast.SelectorExpr `mapstructure:"methods"`
	Tags    []string             `mapstructure:"tags"`
}

type ServiceOptions struct {
	HTTPEnable  bool         `mapstructure:"HTTPServer"`
	Interfaces  []Interface  `mapstructure:"Interface"`
	OpenapiTags []OpenapiTag `mapstructure:"OpenapiTags"`
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

	d := NewDecoder(astLoader)

	result, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}

	for _, build := range result {

		for _, b := range build.Builds {

			fmt.Println(b)
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
