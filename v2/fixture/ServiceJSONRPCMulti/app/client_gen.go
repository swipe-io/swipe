//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/l-vitaly/go-kit/transport/http/jsonrpc"
	"github.com/pquerna/ffjson/ffjson"
)

func NewClientJSONRPCA(tgt string, options ...ClientOption) (InterfaceA, error) {
	opts := &clientOpts{}
	c := &clientA{}
	for _, o := range options {
		o(opts)
	}
	if strings.HasPrefix(tgt, "[") {
		host, port, err := net.SplitHostPort(tgt)
		if err != nil {
			return nil, err
		}
		tgt = host + ":" + port
	}
	u, err := url.Parse(tgt)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	opts.aTestMethodClientOption = append(
		opts.aTestMethodClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			return nil, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, aTestMethodErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			return nil, nil
		}),
	)
	c.aTestMethodEndpoint = jsonrpc.NewClient(
		u,
		"a.aTestMethod",
		append(opts.genericClientOption, opts.aTestMethodClientOption...)...,
	).Endpoint()
	c.aTestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.aTestMethodEndpointMiddleware...))(c.aTestMethodEndpoint)
	return c, nil
}
func NewClientJSONRPCB(tgt string, options ...ClientOption) (InterfaceB, error) {
	opts := &clientOpts{}
	c := &clientB{}
	for _, o := range options {
		o(opts)
	}
	if strings.HasPrefix(tgt, "[") {
		host, port, err := net.SplitHostPort(tgt)
		if err != nil {
			return nil, err
		}
		tgt = host + ":" + port
	}
	u, err := url.Parse(tgt)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	opts.bCreateClientOption = append(
		opts.bCreateClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			req, ok := obj.(BCreateCreateRequest)
			if !ok {
				return nil, fmt.Errorf("couldn't assert request as BCreateCreateRequest, got %T", obj)
			}
			b, err := ffjson.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request %T: %s", obj, err)
			}
			return b, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, bCreateErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			return nil, nil
		}),
	)
	c.bCreateEndpoint = jsonrpc.NewClient(
		u,
		"b.bCreate",
		append(opts.genericClientOption, opts.bCreateClientOption...)...,
	).Endpoint()
	c.bCreateEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bCreateEndpointMiddleware...))(c.bCreateEndpoint)
	opts.bDeleteClientOption = append(
		opts.bDeleteClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			req, ok := obj.(BDeleteDeleteRequest)
			if !ok {
				return nil, fmt.Errorf("couldn't assert request as BDeleteDeleteRequest, got %T", obj)
			}
			b, err := ffjson.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request %T: %s", obj, err)
			}
			return b, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, bDeleteErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			var resp BDeleteDeleteResponse
			err := ffjson.Unmarshal(response.Result, &resp)
			if err != nil {
				return nil, fmt.Errorf("couldn't unmarshal body to BDeleteDeleteResponse: %s", err)
			}
			return resp, nil
		}),
	)
	c.bDeleteEndpoint = jsonrpc.NewClient(
		u,
		"b.bDelete",
		append(opts.genericClientOption, opts.bDeleteClientOption...)...,
	).Endpoint()
	c.bDeleteEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bDeleteEndpointMiddleware...))(c.bDeleteEndpoint)
	opts.bGetClientOption = append(
		opts.bGetClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			req, ok := obj.(BGetGetRequest)
			if !ok {
				return nil, fmt.Errorf("couldn't assert request as BGetGetRequest, got %T", obj)
			}
			b, err := ffjson.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request %T: %s", obj, err)
			}
			return b, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, bGetErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			var resp User
			err := ffjson.Unmarshal(response.Result, &resp)
			if err != nil {
				return nil, fmt.Errorf("couldn't unmarshal body to BGetGetResponse: %s", err)
			}
			return resp, nil
		}),
	)
	c.bGetEndpoint = jsonrpc.NewClient(
		u,
		"b.bGet",
		append(opts.genericClientOption, opts.bGetClientOption...)...,
	).Endpoint()
	c.bGetEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bGetEndpointMiddleware...))(c.bGetEndpoint)
	opts.bGetAllClientOption = append(
		opts.bGetAllClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			req, ok := obj.(BGetAllGetAllRequest)
			if !ok {
				return nil, fmt.Errorf("couldn't assert request as BGetAllGetAllRequest, got %T", obj)
			}
			b, err := ffjson.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request %T: %s", obj, err)
			}
			return b, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, bGetAllErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			var resp []*User
			err := ffjson.Unmarshal(response.Result, &resp)
			if err != nil {
				return nil, fmt.Errorf("couldn't unmarshal body to BGetAllGetAllResponse: %s", err)
			}
			return resp, nil
		}),
	)
	c.bGetAllEndpoint = jsonrpc.NewClient(
		u,
		"b.bGetAll",
		append(opts.genericClientOption, opts.bGetAllClientOption...)...,
	).Endpoint()
	c.bGetAllEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bGetAllEndpointMiddleware...))(c.bGetAllEndpoint)
	opts.bTestMethodClientOption = append(
		opts.bTestMethodClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			req, ok := obj.(BTestMethodTestMethodRequest)
			if !ok {
				return nil, fmt.Errorf("couldn't assert request as BTestMethodTestMethodRequest, got %T", obj)
			}
			b, err := ffjson.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request %T: %s", obj, err)
			}
			return b, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, bTestMethodErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			var resp map[string]map[int][]string
			err := ffjson.Unmarshal(response.Result, &resp)
			if err != nil {
				return nil, fmt.Errorf("couldn't unmarshal body to BTestMethodTestMethodResponse: %s", err)
			}
			return resp, nil
		}),
	)
	c.bTestMethodEndpoint = jsonrpc.NewClient(
		u,
		"b.bTestMethod",
		append(opts.genericClientOption, opts.bTestMethodClientOption...)...,
	).Endpoint()
	c.bTestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bTestMethodEndpointMiddleware...))(c.bTestMethodEndpoint)
	opts.bTestMethod2ClientOption = append(
		opts.bTestMethod2ClientOption,
		jsonrpc.ClientRequestEncoder(func(_ context.Context, obj interface{}) (json.RawMessage, error) {
			req, ok := obj.(BTestMethod2TestMethod2Request)
			if !ok {
				return nil, fmt.Errorf("couldn't assert request as BTestMethod2TestMethod2Request, got %T", obj)
			}
			b, err := ffjson.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request %T: %s", obj, err)
			}
			return b, nil
		}),
		jsonrpc.ClientResponseDecoder(func(_ context.Context, response jsonrpc.Response) (interface{}, error) {
			if response.Error != nil {
				return nil, bTestMethod2ErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)
			}
			return nil, nil
		}),
	)
	c.bTestMethod2Endpoint = jsonrpc.NewClient(
		u,
		"b.bTestMethod2",
		append(opts.genericClientOption, opts.bTestMethod2ClientOption...)...,
	).Endpoint()
	c.bTestMethod2Endpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bTestMethod2EndpointMiddleware...))(c.bTestMethod2Endpoint)
	return c, nil
}
