package generator

import "testing"

func Test_wrapData(t *testing.T) {
	type args struct {
		parts []string
	}
	tests := []struct {
		name string
		want string
		args args
	}{
		{
			"success test 1",
			`map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": response }}}`,
			args{[]string{"a", "b", "c"}},
		},
		{
			"success test 2",
			`map[string]interface{}{"a": response }`,
			args{[]string{"a"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wrapDataServer(tt.args.parts); got != tt.want {
				t.Errorf("wrapDataClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wrapDataClient(t *testing.T) {
	type args struct {
		parts        []string
		responseType string
	}
	tests := []struct {
		name     string
		want     string
		wantPath string
		args     args
	}{
		{
			"success test 1",
			"struct { A struct { B struct {\nData User `json:\"c\"`\n} `json:\"b\"`} `json:\"a\"`}",
			"A.B",
			args{
				[]string{"a", "b", "c"},
				"User",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, path := wrapDataClient(tt.args.parts, tt.args.responseType)
			if got != tt.want {
				t.Errorf("wrapDataClient() = %v, want %v", got, tt.want)
			}
			if path != tt.wantPath {
				t.Errorf("wrapDataClient() = %v, want %v", path, tt.wantPath)
			}
		})
	}
}
