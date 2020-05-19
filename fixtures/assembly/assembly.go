package assembly

import (
	"github.com/swipe-io/swipe/fixtures/model"
	"github.com/swipe-io/swipe/fixtures/user"
	"github.com/swipe-io/swipe/pkg/swipe"
)

func Swipe() {
	swipe.Build(
		swipe.Assembly(model.User{}, user.User{},
			swipe.AssemblyMapping([]string{
				".Point.T", ".Point.Type",
			}),
			swipe.AssemblyExclude([]string{"Password"}, []string{}),
			swipe.AssemblyFormatter(".Name",
				nil,
				func(from user.User) string {
					return "OK"
				},
			),
		),
	)
}
