package swipe

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swipe-io/swipe/v2/internal/astloader"

	"github.com/google/go-cmp/cmp"

	"github.com/swipe-io/swipe/v2/internal/interface/executor"
	"github.com/swipe-io/swipe/v2/internal/interface/factory"
	"github.com/swipe-io/swipe/v2/internal/interface/frame"
	"github.com/swipe-io/swipe/v2/internal/interface/registry"
	"github.com/swipe-io/swipe/v2/internal/option"
	ue "github.com/swipe-io/swipe/v2/internal/usecase/executor"
)

var record = flag.Bool("record", false, "write expected result without running tests")
var onlyDiff = flag.Bool("only-diff", false, "show only diff")

func newGeneratorExecutor(wd string) ue.GenerationExecutor {
	astl := astloader.NewLoader(wd, os.Environ(), []string{"."}, nil)
	l := option.NewLoader(astl)
	r := registry.NewRegistry(l)
	i := factory.NewImporterFactory()
	ff := frame.NewFrameFactory(Version)
	return executor.NewGenerationExecutor(r, i, ff, l)
}

func TestSwipe(t *testing.T) {
	const testRoot = "fixtures"

	testdataEnts, err := ioutil.ReadDir(testRoot)
	if err != nil {
		t.Fatal(err)
	}
	tests := make([]*testCase, 0, len(testdataEnts))
	for _, ent := range testdataEnts {
		name := ent.Name()
		if !ent.IsDir() || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}
		test, err := loadTestCase(filepath.Join(testRoot, name))
		if err != nil {
			t.Error(err)
			continue
		}
		tests = append(tests, test)
	}
	for _, test := range tests {
		ge := newGeneratorExecutor(test.testCasePath)

		test := test
		t.Run(test.name, func(t *testing.T) {
			results, errs := ge.Execute()
			if len(errs) > 0 {
				for _, e := range errs {
					t.Error(e)
				}
			}

			// clear all before generated files.
			_ = filepath.Walk(test.testCasePath, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					if strings.Contains(info.Name(), "_gen") {
						_ = os.Remove(path)
					}
				}
				return nil
			})

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
			//if !*record && len(test.expectedOutput) > 0 {
			//for _, expectedContent := range test.expectedOutput {
			//t.Errorf("there are expected results which are not.\n*** expected:\n%s\n\n***", string(expectedContent))
			//}
			//}
		})
	}
}

type testCase struct {
	name                 string
	expectedOutput       map[string][]byte
	expectedError        bool
	expectedErrorStrings []string
	testCasePath         string
}

func loadTestCase(root string) (*testCase, error) {
	name := filepath.Base(root)
	testCasePath, err := filepath.Abs(filepath.Join(root, "app"))
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
	}, nil

}
