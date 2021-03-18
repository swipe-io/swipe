package v2

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/swipe-io/swipe/v2/internal/astloader"
	"github.com/swipe-io/swipe/v2/internal/interface/executor"
	"github.com/swipe-io/swipe/v2/internal/interface/factory"
	"github.com/swipe-io/swipe/v2/internal/interface/frame"
	"github.com/swipe-io/swipe/v2/internal/interface/registry"
	"github.com/swipe-io/swipe/v2/internal/option"
	ue "github.com/swipe-io/swipe/v2/internal/usecase/executor"
)

var record = flag.Bool("record", false, "write expected result without running tests")
var onlyDiff = flag.Bool("only-diff", false, "show only diff")

func newGeneratorExecutor(wd string, pkgs []string) ue.GenerationExecutor {
	patterns := []string{"."}
	patterns = append(patterns, pkgs...)

	astl := astloader.NewLoader(wd, os.Environ(), patterns)
	l := option.NewLoader(astl)
	r := registry.NewRegistry(l)
	i := factory.NewImporterFactory()
	ff := frame.NewFrameFactory(Version)
	return executor.NewGenerationExecutor(r, i, ff, l)
}

func TestSwipe(t *testing.T) {
	testdataEnts, err := filepath.Glob("../../swipe-test*")
	if err != nil {
		t.Fatal(err)
	}
	tests := make([]*testCase, 0, len(testdataEnts))
	for _, name := range testdataEnts {
		var pkgs []string
		if data, err := ioutil.ReadFile(filepath.Join(name, "pkgs")); err == nil {
			pkgs = strings.Split(string(data), "\n")
		}
		test, err := loadTestCase(name, pkgs)
		if err != nil {
			t.Error(err)
			continue
		}
		tests = append(tests, test)
	}
	for _, test := range tests {
		ge := newGeneratorExecutor(test.testCasePath, test.pkgs)

		test := test
		t.Run(test.name, func(t *testing.T) {
			results, errs := ge.Execute()
			if len(errs) > 0 {
				for _, e := range errs {
					t.Error(e)
				}
			}

			if *record {
				// clear all before generated files.
				_ = filepath.Walk(test.testCasePath, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						if strings.Contains(info.Name(), "_gen") {
							_ = os.Remove(path)
						}
					}
					return nil
				})
			}

			for _, result := range results {
				if len(result.Errs) > 0 {
					t.Fatalf("result errors: %v", result.Errs)
				}
				if *record {
					if err := ioutil.WriteFile(result.OutputPath, result.Content, 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					if expectedContent, ok := test.expectedOutput[result.OutputPath]; ok {
						if !bytes.Equal(expectedContent, result.Content) {
							actual, expected := string(result.Content), string(expectedContent)
							diff := cmp.Diff(strings.Split(expected, "\n"), strings.Split(actual, "\n"))
							buf := new(bytes.Buffer)
							buf.WriteString(fmt.Sprintf("swipe output differs from expected file %s.\n", result.OutputPath))
							if !*onlyDiff {
								buf.WriteString(fmt.Sprintf("*** actual:\n%s\n\n*** expected:\n%s\n\n", actual, expected))
							}
							buf.WriteString(fmt.Sprintf("*** diff:\n%s", diff))
							t.Fatal(buf.String())
						}
						delete(test.expectedOutput, result.OutputPath)
					}
				}
			}
			if !*record && len(test.expectedOutput) > 0 {
				for _, expectedContent := range test.expectedOutput {
					t.Errorf("there are expected results which are not.\n*** expected:\n%s\n\n***", string(expectedContent))
				}
			}
		})
	}
}

type testCase struct {
	name                 string
	expectedOutput       map[string][]byte
	expectedError        bool
	expectedErrorStrings []string
	testCasePath         string
	pkgs                 []string
}

func loadTestCase(root string, pkgs []string) (*testCase, error) {
	name := filepath.Base(root)
	testCasePath, err := filepath.Abs(filepath.Join(root, "pkg", "transport"))
	if err != nil {
		return nil, err
	}
	expectedFiles, err := ioutil.ReadDir(testCasePath)
	if err != nil {
		return nil, err
	}
	expectedOutput := make(map[string][]byte)
	for _, file := range expectedFiles {
		if !file.IsDir() && strings.Contains(file.Name(), "_gen") {
			expectedFile, err := filepath.Abs(filepath.Join(testCasePath, file.Name()))
			if err != nil {
				return nil, err
			}
			data, err := ioutil.ReadFile(expectedFile)
			if err != nil {
				return nil, err
			}
			expectedOutput[expectedFile] = data
		}
	}

	return &testCase{
		name:                 name,
		testCasePath:         testCasePath,
		expectedOutput:       expectedOutput,
		expectedError:        false,
		expectedErrorStrings: nil,
		pkgs:                 pkgs,
	}, nil

}
