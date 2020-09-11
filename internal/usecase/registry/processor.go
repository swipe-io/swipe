package registry

import (
	"github.com/swipe-io/swipe/v2/internal/astloader"
	"github.com/swipe-io/swipe/v2/internal/option"
	up "github.com/swipe-io/swipe/v2/internal/usecase/processor"
)

type ProcessorRegistry interface {
	NewProcessor(o *option.ResultOption, data *astloader.Data) (up.Processor, error)
}
