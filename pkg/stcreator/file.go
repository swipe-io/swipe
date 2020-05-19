package stcreator

// ContentExtractor content extractor.
type ContentExtractor func() (string, error)

// F file type.
type F struct {
	Name     string
	Template interface{}
	Data     interface{}
}

// Accept implement interface I.
func (f *F) Accept(v Visitor, path string) error {
	return v.VisitFile(f, path)
}
