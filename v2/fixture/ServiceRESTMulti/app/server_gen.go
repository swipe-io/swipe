//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	http2 "net/http"
	"strconv"

	"github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pquerna/ffjson/ffjson"
)

type errorWrapper struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}

func defaultErrorEncoder(ctx context.Context, err error, w http2.ResponseWriter) {
	var errData interface{}
	if e, ok := err.(interface{ ErrorData() interface{} }); ok {
		errData = e.ErrorData()
	}
	data, merr := ffjson.Marshal(errorWrapper{Error: err.Error(), Data: errData})
	if merr != nil {
		_, _ = w.Write([]byte("unexpected error"))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := err.(http.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	code := http2.StatusInternalServerError
	if sc, ok := err.(http.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func encodeResponseHTTP(ctx context.Context, w http2.ResponseWriter, response interface{}) (err error) {
	contentType := "application/json; charset=utf-8"
	statusCode := 200
	h := w.Header()
	var data []byte
	if response != nil {
		data, err = ffjson.Marshal(response)
		if err != nil {
			return err
		}
	} else {
		contentType = "text/plain; charset=utf-8"
		statusCode = 201
	}
	h.Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(data)
	return nil
}

// MakeHandlerREST HTTP REST Transport
func MakeHandlerREST(svcInterfaceA InterfaceA, svcInterfaceB InterfaceB, options ...ServerOption) (http2.Handler, error) {
	opts := &serverOpts{}
	for _, o := range options {
		o(opts)
	}
	opts.genericServerOption = append(opts.genericServerOption, http.ServerErrorEncoder(defaultErrorEncoder))
	epSetA := MakeInterfaceAEndpointSet(svcInterfaceA)
	epSetB := MakeInterfaceBEndpointSet(svcInterfaceB)
	epSetA.TestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceATestMethodEndpointMiddleware...))(epSetA.TestMethodEndpoint)
	epSetB.CreateEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBCreateEndpointMiddleware...))(epSetB.CreateEndpoint)
	epSetB.DeleteEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBDeleteEndpointMiddleware...))(epSetB.DeleteEndpoint)
	epSetB.GetEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBGetEndpointMiddleware...))(epSetB.GetEndpoint)
	epSetB.GetAllEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBGetAllEndpointMiddleware...))(epSetB.GetAllEndpoint)
	epSetB.TestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBTestMethodEndpointMiddleware...))(epSetB.TestMethodEndpoint)
	epSetB.TestMethod2Endpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBTestMethod2EndpointMiddleware...))(epSetB.TestMethod2Endpoint)
	r := mux.NewRouter()
	r.Methods("GET").Path("/a/testmethod").Handler(http.NewServer(
		epSetA.TestMethodEndpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			return nil, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceATestMethodServerOption...)...,
	))
	r.Methods(http2.MethodPost).Path("/b/create").Handler(http.NewServer(
		epSetB.CreateEndpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req InterfaceBCreateRequest
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for InterfaceBCreateRequest: %w", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to InterfaceBCreateRequest: %w", err)
			}
			return req, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceBCreateServerOption...)...,
	))
	r.Methods(http2.MethodPost).Path("/b/delete").Handler(http.NewServer(
		epSetB.DeleteEndpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req InterfaceBDeleteRequest
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for InterfaceBDeleteRequest: %w", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to InterfaceBDeleteRequest: %w", err)
			}
			return req, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceBDeleteServerOption...)...,
	))
	r.Methods(http2.MethodPost).Path("/b/get-test").Handler(http.NewServer(
		epSetB.GetEndpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req InterfaceBGetRequest
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for InterfaceBGetRequest: %w", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to InterfaceBGetRequest: %w", err)
			}
			q := r.URL.Query()
			tmpcc := q.Get("cc")
			if tmpcc != "" {
				ccInt, err := strconv.Atoi(tmpcc)
				if err != nil {
					return nil, err
				}
				req.Cc = int(ccInt)
			}
			return req, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceBGetServerOption...)...,
	))
	r.Methods(http2.MethodPost).Path("/b/getall").Handler(http.NewServer(
		epSetB.GetAllEndpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req InterfaceBGetAllRequest
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for InterfaceBGetAllRequest: %w", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to InterfaceBGetAllRequest: %w", err)
			}
			return req, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceBGetAllServerOption...)...,
	))
	r.Methods(http2.MethodPost).Path("/b/testmethod").Handler(http.NewServer(
		epSetB.TestMethodEndpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req InterfaceBTestMethodRequest
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for InterfaceBTestMethodRequest: %w", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to InterfaceBTestMethodRequest: %w", err)
			}
			return req, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceBTestMethodServerOption...)...,
	))
	r.Methods(http2.MethodPost).Path("/b/testmethod2").Handler(http.NewServer(
		epSetB.TestMethod2Endpoint,
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req InterfaceBTestMethod2Request
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for InterfaceBTestMethod2Request: %w", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to InterfaceBTestMethod2Request: %w", err)
			}
			return req, nil
		},
		encodeResponseHTTP,
		append(opts.genericServerOption, opts.interfaceBTestMethod2ServerOption...)...,
	))
	return r, nil
}
