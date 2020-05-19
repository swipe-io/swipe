package stcreator

// D dir type.
type D struct {
	Name     string
	Children []I
}

// AddChildren add I to children.
func (d *D) AddChildren(i I) {
	d.Children = append(d.Children, i)
}

// Accept implement interface I.
func (d *D) Accept(v Visitor, path string) error {
	return v.VisitDir(d, path)
}
