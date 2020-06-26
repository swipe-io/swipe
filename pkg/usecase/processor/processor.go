package processor

import (
	"github.com/swipe-io/swipe/pkg/usecase/generator"
)

type Processor interface {
	SetOption(option interface{}) bool
	Generators() []generator.Generator
}
