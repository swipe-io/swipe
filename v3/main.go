package v3

import (
	"github.com/swipe-io/swipe/v3/cmd"
	_ "github.com/swipe-io/swipe/v3/internal/plugin/config"
	_ "github.com/swipe-io/swipe/v3/internal/plugin/es"
	_ "github.com/swipe-io/swipe/v3/internal/plugin/gokit"
)

func Main() {
	cmd.Execute()
}
