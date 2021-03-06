# REST API

## Basic

Enabling REST transport generation by default, Swipe generates REST transport.

To generate, just add the following settings:

````bash
package example

import (
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
        ),
    )
}
````

Swipe will generate transport methods for all exported methods.

> ℹ️ **INFO**
>
> Swipe automatically detects the HTTP method, if the method parameters
>
> are not redirected to the request parameters, headers or to the REST path,
>
> then the POST method will be used otherwise GET.

Swipe can generate Openapi (Swagger) documentation for details, see the Openapi (Swagger) section of the documentation

## Settings

You can control the generation of methods, for example, exclude a method from generation in the transport, redirect parameters to request parameters, headers or REST path, change the HTTP method.

### Change method settings

In order to change the settings of a method, you must use the `MethodOptions` function

````bash
package example

import (
  "http"
  
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodOptions(controller.UserController.Get,
                 RESTMethod(http.MethodPost), // all Get methods will have an HTTP POST method.
            )
        ),
    )
}
````

### Change default settings for all methods

In order to change the default settings for all methods, you must use the `MethodDefaultOptions` function

````bash
package example

import (
  "http"
  
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodDefaultOptions(
                 RESTMethod(http.MethodPost), // all methods will have an HTTP methodPOST. 
            )
        ),
    )
}
````

### Changing the HTTP method

Sets the HTTP method.

````bash
package example

import (
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodOptions(controller.UserController.Get,
                 RESTMethod(http.MethodPost), // у всех метода Get будет HTTP метод POST.
            )
        ),
    )
}
````

### Changing the HTTP path

Sets the HTTP path, defaults to the lowercase interface method name, for example:

for Get will be the path `/ get`
for GetByID the path will be `/ getbyid`

````bash
package example

import (
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodOptions(controller.UserController.Get,
                 RESTPath("/users"), // the Get method will have the path / users.
            )
        ),
    )
}
````

### Binding values from the HTTP header to a method parameter.

Used to bind values from the HTTP header to a method parameter.

For example, if you need to use the `X-Request-ID` header in the `requestID` method parameter:

> ℹ️ **INFO**
> Parameters are specified as an array, where the value pair
>
> is the name of the HTTP header and method parameter.

````bash
package example

import (
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodOptions(controller.UserController.Get,
                RESTHeaderVars([]string{"X-Request-ID", "requestID"}),
            )
        ),
    )
}
````

### Binding values from HTTP request parameters to a method parameter

Used to bind values from HTTP request parameters to a method parameter.

For example, if you need to use the `sort` query parameter in the parameter of the `sort` method:

> ℹ️ **INFO**
>
> Parameters are specified as an array, where the pair of values
>
> is the name of the HTTP request parameter and method.

````bash
package example

import (
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodOptions(controller.UserController.Get,
                RESTHeaderVars([]string{"X-Request-ID", "requestID"}),
            )
        ),
    )
}
````

### Wrapping a REST response into an object

Used to wrap a response in an object.

For example, you have an interface method `Get () (User, error)` and you need the contents of the fields of the `User` structure inside the `data` field

````bash
package example

import (
  "github.com/example/app/pkg/controller"
  
  . "github.com/swipe-io/swipe/v2"
)

func Swipe() {
    Build(
        Service(
            Interface((*controller.UserController)(nil), ""),
      
            HTTPServer(),
      
            MethodOptions(controller.UserController.Get,
                RESTWrapResponse("data"),
            )
        ),
    )
}
````

The response will be JSON like this:

````bash
{
  "data": {
    "firstName": "",
    "lastName": ""
  }
}
````

> To view the original source for this documentation [**click here**](https://swipeio.dev/docs/installation) Original Format -*Russian*

**Installation**[<== Previous](installation.md.md)  [Next ==>](json-rpc.md.md) **JSON RPC**

[<== Home](README.md) 🏠
