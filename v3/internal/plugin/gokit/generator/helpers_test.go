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
		{
			"success test 2",
			`map[string]interface{}{"data": response }`,
			args{[]string{"data"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wrapDataServer(tt.args.parts); got != tt.want {
				t.Errorf("wrapDataServer() = %v, want %v", got, tt.want)
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
		name           string
		args           args
		wantResult     string
		wantStructPath string
	}{
		{
			"success 1",
			args{[]string{"data"}, "string"},
			"struct { Data string `json:\"data\"`\n}",
			"Data",
		},
		{
			"success 2",
			args{[]string{"data", "user"}, "string"},
			"struct { Data struct {\nUser string `json:\"user\"`\n} `json:\"data\"`}",
			"Data.User",
		},
		{
			"success 2",
			args{[]string{"data", "user", "foo"}, "string"},
			"struct { Data struct {\nUser struct {\nFoo string `json:\"foo\"`\n} `json:\"user\"`} `json:\"data\"`}",
			"Data.User.Foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotStructPath := wrapDataClient(tt.args.parts, tt.args.responseType)
			if gotResult != tt.wantResult {
				t.Errorf("wrapDataClient() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
			if gotStructPath != tt.wantStructPath {
				t.Errorf("wrapDataClient() gotStructPath = %v, want %v", gotStructPath, tt.wantStructPath)
			}
		})
	}
}
