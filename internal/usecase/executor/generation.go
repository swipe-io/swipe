package executor

type GenerateResult struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

type GenerationExecutor interface {
	Execute() (results []GenerateResult, errs []error)
}
