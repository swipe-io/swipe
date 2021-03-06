# Swipe Installation

Install the **Swipe** console utility:

````bash
go get github.com/swipe-io/swipe/cmd/swipe
````

**Swipe** requires a Go version with [Golang Modules](https://github.com/golang/go/wiki/Modules) support. So don't forget to initialize the Go module:

````bash
go mod init github.com/my/repo
````

and then install the Swipe package:

````bash
go get github.com/swipe-io/swipe/v2
````

## 🔥 WARNING

**Swipe** generates code using an option: a function that calls functions that define the generation parameters.

To describe the generation parameters, create a .go file and add a function, add a `swipe.Build` call to the function body. You also need to add the build tag `// + build swipe` so that Golang will ignore the file when you build your application.

Below is a simple example of a JSON RPC transport generation configuration file:

````bash
// +build swipe

package transport

import (
    "github.com/example/app/pkg/controller"

    . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            HTTPServer(),
      
            Interface((*controller.ExampleController)(nil), ""),

            ClientsEnable([]string{"go"}),

            JSONRPCEnable(),  

            OpenapiEnable(),
            OpenapiOutput("./docs"),
            OpenapiInfo("Service", "Example description.", "v1.0.0"),

            MethodDefaultOptions(
                Logging(true),
                Instrumenting(true),
            ),
        ),
    )
}
````

If you want to generate code, you can run:

````bash
swipe ./pkg/...
````

The above command will search for all functions containing `swipe.Build` and generate code in `_gen.`.


> To view the original source for this documentation [**click here**](https://swipeio.dev/docs/installation) Original Format -*Russian*


[Next ==>](rest-api.md.md) **REST API**

[<== Home](README.md) 🏠
