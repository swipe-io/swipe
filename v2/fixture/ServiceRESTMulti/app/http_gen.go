//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	http2 "net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
)

type httpError struct {
	code int
}

func (e *httpError) Error() string {
	return http2.StatusText(e.code)
}
func (e *httpError) StatusCode() int {
	return e.code
}
func interfaceATestMethodErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	return
}

func interfaceBCreateErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	return
}

func interfaceBDeleteErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	return
}

func interfaceBGetErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	return
}

func interfaceBGetAllErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	return
}

func interfaceBTestMethodErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
	}
	return
}

func interfaceBTestMethod2ErrorDecode(code int) (err error) {
	switch code {
	default:
		err = &httpError{code: code}
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
func GenericServerOptions(v ...http.ServerOption) ServerOption {
	return func(o *serverOpts) { o.genericServerOption = v }
}

func GenericServerEndpointMiddlewares(v ...endpoint.Middleware) ServerOption {
	return func(o *serverOpts) { o.genericEndpointMiddleware = v }
}

type ServerOption func(*serverOpts)
type serverOpts struct {
	genericServerOption                     []http.ServerOption
	genericEndpointMiddleware               []endpoint.Middleware
	interfaceATestMethodServerOption        []http.ServerOption
	interfaceATestMethodEndpointMiddleware  []endpoint.Middleware
	interfaceBCreateServerOption            []http.ServerOption
	interfaceBCreateEndpointMiddleware      []endpoint.Middleware
	interfaceBDeleteServerOption            []http.ServerOption
	interfaceBDeleteEndpointMiddleware      []endpoint.Middleware
	interfaceBGetServerOption               []http.ServerOption
	interfaceBGetEndpointMiddleware         []endpoint.Middleware
	interfaceBGetAllServerOption            []http.ServerOption
	interfaceBGetAllEndpointMiddleware      []endpoint.Middleware
	interfaceBTestMethodServerOption        []http.ServerOption
	interfaceBTestMethodEndpointMiddleware  []endpoint.Middleware
	interfaceBTestMethod2ServerOption       []http.ServerOption
	interfaceBTestMethod2EndpointMiddleware []endpoint.Middleware
}

func InterfaceATestMethodServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceATestMethodServerOption = opt }
}

func InterfaceATestMethodServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceATestMethodEndpointMiddleware = opt }
}

func InterfaceBCreateServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBCreateServerOption = opt }
}

func InterfaceBCreateServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBCreateEndpointMiddleware = opt }
}

func InterfaceBDeleteServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBDeleteServerOption = opt }
}

func InterfaceBDeleteServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBDeleteEndpointMiddleware = opt }
}

func InterfaceBGetServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetServerOption = opt }
}

func InterfaceBGetServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetEndpointMiddleware = opt }
}

func InterfaceBGetAllServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetAllServerOption = opt }
}

func InterfaceBGetAllServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBGetAllEndpointMiddleware = opt }
}

func InterfaceBTestMethodServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethodServerOption = opt }
}

func InterfaceBTestMethodServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethodEndpointMiddleware = opt }
}

func InterfaceBTestMethod2ServerOptions(opt ...http.ServerOption) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethod2ServerOption = opt }
}

func InterfaceBTestMethod2ServerEndpointMiddlewares(opt ...endpoint.Middleware) ServerOption {
	return func(c *serverOpts) { c.interfaceBTestMethod2EndpointMiddleware = opt }
}