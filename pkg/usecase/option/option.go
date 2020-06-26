package option

import (
	"github.com/swipe-io/swipe/pkg/parser"
)

type Option interface {
	Parse(option *parser.Option) (interface{}, error)
}
