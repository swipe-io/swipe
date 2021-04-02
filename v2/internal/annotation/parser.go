package annotation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	errAnnotationSyntax      = errors.New("bad syntax for struct annotation pair")
	errAnnotationKeySyntax   = errors.New("bad syntax for struct annotation key")
	errAnnotationValueSyntax = errors.New("bad syntax for struct annotation value")

	errKeyNotSet             = errors.New("annotation key does not exist")
	errAnnotationNotExist    = errors.New("annotation does not exist")
	errAnnotationKeyMismatch = errors.New("mismatch between key and annotation.key")
)

type Annotation struct {
	Key     string
	Name    string
	Options []string
}

func (a *Annotation) HasOption(opt string) bool {
	for _, annotationOpt := range a.Options {
		if annotationOpt == opt {
			return true
		}
	}
	return false
}

func (a *Annotation) Value() string {
	options := strings.Join(a.Options, ",")
	if options != "" {
		return fmt.Sprintf(`%s,%s`, a.Name, options)
	}
	return a.Name
}

type Annotations struct {
	annotations []*Annotation
}

func (a *Annotations) Get(key string) (*Annotation, error) {
	for _, annotation := range a.annotations {
		if annotation.Key == key {
			return annotation, nil
		}
	}
	return nil, errAnnotationNotExist
}

func Parse(annotation string) (*Annotations, error) {
	var annotations []*Annotation
	hasAnnotation := annotation != ""

	for annotation != "" {
		i := 0
		for i < len(annotation) && annotation[i] == ' ' {
			i++
		}
		annotation = annotation[i:]
		if annotation == "" {
			break
		}

		if annotation[0] != '@' {
			break
		}

		annotation = annotation[1:]

		i = 0
		for i < len(annotation) && annotation[i] > ' ' && annotation[i] != ':' && annotation[i] != '"' && annotation[i] != 0x7f {
			i++
		}
		if i == 0 {
			return nil, errAnnotationKeySyntax
		}
		if i+1 >= len(annotation) || annotation[i] != ':' {
			return nil, errAnnotationSyntax
		}
		if annotation[i+1] != '"' {
			return nil, errAnnotationValueSyntax
		}
		key := annotation[:i]
		annotation = annotation[i+1:]

		i = 1
		for i < len(annotation) && annotation[i] != '"' {
			if annotation[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(annotation) {
			return nil, errAnnotationValueSyntax
		}

		lvalue := annotation[:i+1]
		annotation = annotation[i+1:]

		value, err := strconv.Unquote(lvalue)
		if err != nil {
			return nil, errAnnotationValueSyntax
		}

		res := strings.Split(value, ",")
		name := res[0]
		options := res[1:]
		if len(options) == 0 {
			options = nil
		}

		annotations = append(annotations, &Annotation{
			Key:     key,
			Name:    name,
			Options: options,
		})
	}
	if hasAnnotation && len(annotations) == 0 {
		return nil, nil
	}
	return &Annotations{annotations: annotations}, nil
}
