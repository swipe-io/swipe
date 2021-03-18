package option

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/swipe-io/swipe/v2/internal/astloader"
)

type Interface struct {
}

type TestStruct struct {
	Interfaces []Interface `swipe:"Interface"`
}

func TestParser_Parse(t *testing.T) {
	wd, err := filepath.Abs("./fixtures")
	if err != nil {
		t.Fatal(err)
	}
	astLoader := astloader.NewLoader(wd, os.Environ(), []string{"./fixtures/.."})
	data, errs := astLoader.Process()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Log(err)
		}
		t.Fatal("AST loader failed")
	}

	type args struct {
		s interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "",
			args: args{s: &TestStruct{}},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			if got := p.Parse(tt.args.s, data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
