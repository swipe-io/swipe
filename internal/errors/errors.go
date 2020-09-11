package errors

import "go/token"

type ErrorCollector struct {
	errors []error
}

func (ec *ErrorCollector) Errors() []error {
	return ec.errors
}

func (ec *ErrorCollector) Add(errs ...error) {
	for _, e := range errs {
		if e != nil {
			ec.errors = append(ec.errors, e)
		}
	}
}

func MapErrors(errs []error, f func(error) error) []error {
	if len(errs) == 0 {
		return nil
	}
	newErrs := make([]error, len(errs))
	for i := range errs {
		newErrs[i] = f(errs[i])
	}
	return newErrs
}

type GenErr struct {
	error    error
	position token.Position
}

func (w *GenErr) Error() string {
	if !w.position.IsValid() {
		return w.error.Error()
	}
	return w.position.String() + ": " + w.error.Error()
}

func NotePosition(p token.Position, e error) error {
	switch e.(type) {
	case nil:
		return nil
	case *GenErr:
		return e
	default:
		return &GenErr{error: e, position: p}
	}
}

func NotePositionAll(p token.Position, errs []error) []error {
	return MapErrors(errs, func(e error) error {
		return NotePosition(p, e)
	})
}
