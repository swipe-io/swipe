package executor

type GenerateResult struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

type GenerationExecutor interface {
	Execute(wd string, env []string, patterns []string) (results []GenerateResult, errs []error)
	Cleanup(wd string)
}
