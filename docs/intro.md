# Getting started

Swipe requires latest Go version with
[Modules](https://github.com/golang/go/wiki/Modules) support and uses import versioning. 
So make sure to initialize a Go module:

```shell
go mod init github.com/my/repo
```

and then install last swipe package:

```shell
go get github.com/swipe-io/swipe
```

and then install last cli swipe: 

```shell
go get github.com/swipe-io/swipe/cmd/swipe/...
```

::: danger
The package version must match the cli version.
:::

::: tip
Find out cli version do `swipe version`
:::

swipe generates code using an option: a function that calls functions that define the generation parameters.

To describe the generation parameters, add the `swipe.Build` function to the function body, also add the build tag `+build swipe` so that Golang does not try to compile the settings file.

Example file:

```go
// +build swipe

package jsonrpc

import (
	"github.com/swipe-io/swipe/fixtures/service"

 	. "github.com/swipe-io/swipe/pkg/swipe"
)

func Swipe() {    
 	Build(
 		Service(
 			(*service.Interface)(nil),
 			Transport("http",
 				ClientEnable(),
 				Openapi(
 					OpenapiOutput("/../../docs"),
 					OpenapiVersion("1.0.0"),
 				),
 			),
 			Logging(),
 			Instrumenting(),
 		),
 	)
}
```

If you want the generate code, you can run:

```shell
swipe ./pkg/...
```

the command above will search for all functions containing `swipe.Build` and generate code in the file `swipe_gen.go`.