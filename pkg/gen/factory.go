package gen

import (
	"github.com/swipe-io/swipe/pkg/gen/assembly"
	"github.com/swipe-io/swipe/pkg/gen/config"
	"github.com/swipe-io/swipe/pkg/gen/service"
	"github.com/swipe-io/swipe/pkg/writer"
)

var factory = map[string]func(*writer.Writer) Generator{
	"ConfigEnv": func(w *writer.Writer) Generator {
		return config.New(w)
	},
	"Service": func(w *writer.Writer) Generator {
		return service.New(w)
	},
	"Assembly": func(w *writer.Writer) Generator {
		return assembly.New(w)
	},
}
