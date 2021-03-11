//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/l-vitaly/go-kit/transport/http/jsonrpc"
)

type httpError struct {
	code    int
	data    interface{}
	message string
}

func (e *httpError) Error() string {
	return e.message
}
func (e *httpError) StatusCode() int {
	return e.code
}
func (e *httpError) ErrorData() interface{} {
	return e.data
}
func (e *httpError) SetErrorData(data interface{}) {
	e.data = data
}
func (e *httpError) SetErrorMessage(message string) {
	e.message = message
}
func interfaceBCreateErrorDecode(code int, message string, data interface{}) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	case -32001:
		err = ErrUnauthorized{}
	}
	if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {
		err.SetErrorData(data)
	}
	if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {
		err.SetErrorMessage(message)
	}
	return
}

func interfaceBDeleteErrorDecode(code int, message string, data interface{}) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {
		err.SetErrorData(data)
	}
	if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {
		err.SetErrorMessage(message)
	}
	return
}

func interfaceBGetErrorDecode(code int, message string, data interface{}) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {
		err.SetErrorData(data)
	}
	if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {
		err.SetErrorMessage(message)
	}
	return
}

func interfaceBGetAllErrorDecode(code int, message string, data interface{}) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {
		err.SetErrorData(data)
	}
	if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {
		err.SetErrorMessage(message)
	}
	return
}

func interfaceBTestMethodErrorDecode(code int, message string, data interface{}) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {
		err.SetErrorData(data)
	}
	if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {
		err.SetErrorMessage(message)
	}
	return
}

func interfaceBTestMethod2ErrorDecode(code int, message string, data interface{}) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {
		err.SetErrorData(data)
	}
	if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {
		err.SetErrorMessage(message)
	}
	return
}

func middlewareChain(middlewares []endpoint.Middleware) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		if len(middlewares) == 0 {
			return next
		}
		outer := middlewares[0]
		others := middlewares[1:]
		for i := len(others) - 1; i >= 0; i-- {
			next = others[i](next)
		}
		return outer(next)
	}
}
func GenericServerOptions(v ...jsonrpc.ServerOption) ServerOption {
	return func(o *serverOpts) { o.genericServerOption = v }
}

func GenericServerEndpointMiddlewares(v ...endpoint.Middleware) ServerOption {
	return func(o *serverOpts) { o.genericEndpointMiddleware = v }
}

type ServerOption func(*serverOpts)
type serverOpts struct {
	genericServerOption                     []jsonrpc.ServerOption
	genericEndpointMiddleware               []endpoint.Middleware
	interfaceBCreateServerOption            []jsonrpc.ServerOption
	interfaceBCreateEndpointMiddleware      []endpoint.Middleware
	interfaceBDeleteServerOption            []jsonrpc.ServerOption
	interfaceBDeleteEndpointMiddleware      []endpoint.Middleware
	interfaceBGetServerOption               []jsonrpc.ServerOption
	interfaceBGetEndpointMiddleware         []endpoint.Middleware
	interfaceBGetAllServerOption            []jsonrpc.ServerOption
	interfaceBGetAllEndpointMiddleware      []endpoint.Middleware
	interfaceBTestMethodServerOption        []jsonrpc.ServerOption
	interfaceBTestMethodEndpointMiddleware  []endpoint.Middleware
	interfaceBTestMethod2ServerOption       []jsonrpc.ServerOption
	interfaceBTestMethod2EndpointMiddleware []endpoint.Middleware
}

func InterfaceBCreateServerOptions(opt ...jsonrpc.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBCreateServerOption = opt }
}

func InterfaceBCreateServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBCreateEndpointMiddleware = opt }
}

func InterfaceBDeleteServerOptions(opt ...jsonrpc.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBDeleteServerOption = opt }
}

func InterfaceBDeleteServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBDeleteEndpointMiddleware = opt }
}

func InterfaceBGetServerOptions(opt ...jsonrpc.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetServerOption = opt }
}

func InterfaceBGetServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetEndpointMiddleware = opt }
}

func InterfaceBGetAllServerOptions(opt ...jsonrpc.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetAllServerOption = opt }
}

func InterfaceBGetAllServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetAllEndpointMiddleware = opt }
}

func InterfaceBTestMethodServerOptions(opt ...jsonrpc.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethodServerOption = opt }
}

func InterfaceBTestMethodServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethodEndpointMiddleware = opt }
}

func InterfaceBTestMethod2ServerOptions(opt ...jsonrpc.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethod2ServerOption = opt }
}

func InterfaceBTestMethod2ServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethod2EndpointMiddleware = opt }
}