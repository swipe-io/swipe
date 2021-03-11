//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/l-vitaly/go-kit/transport/http/jsonrpc"
)

type AppClient struct {
	AClient InterfaceA
	BClient InterfaceB
}

func NewClientJSONRPC(tgt string, opts ...ClientOption) (*AppClient, error) {
	aClient, err := NewClientJSONRPCA(tgt, opts...)
	if err != nil {
		return nil, err
	}
	bClient, err := NewClientJSONRPCB(tgt, opts...)
	if err != nil {
		return nil, err
	}
	return &AppClient{
		AClient: aClient,
		BClient: bClient,
	}, nil
}

type ClientOption func(*clientOpts)
type clientOpts struct {
	aTestMethodClientOption        []jsonrpc.ClientOption
	aTestMethodEndpointMiddleware  []endpoint.Middleware
	bCreateClientOption            []jsonrpc.ClientOption
	bCreateEndpointMiddleware      []endpoint.Middleware
	bDeleteClientOption            []jsonrpc.ClientOption
	bDeleteEndpointMiddleware      []endpoint.Middleware
	bGetClientOption               []jsonrpc.ClientOption
	bGetEndpointMiddleware         []endpoint.Middleware
	bGetAllClientOption            []jsonrpc.ClientOption
	bGetAllEndpointMiddleware      []endpoint.Middleware
	bTestMethodClientOption        []jsonrpc.ClientOption
	bTestMethodEndpointMiddleware  []endpoint.Middleware
	bTestMethod2ClientOption       []jsonrpc.ClientOption
	bTestMethod2EndpointMiddleware []endpoint.Middleware
	genericClientOption            []jsonrpc.ClientOption
	genericEndpointMiddleware      []endpoint.Middleware
}

func GenericClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.genericClientOption = opt }
}

func GenericClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.genericEndpointMiddleware = opt }
}

func ATestMethodClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.aTestMethodClientOption = opt }
}

func ATestMethodClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.aTestMethodEndpointMiddleware = opt }
}

func BCreateClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.bCreateClientOption = opt }
}

func BCreateClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.bCreateEndpointMiddleware = opt }
}

func BDeleteClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.bDeleteClientOption = opt }
}

func BDeleteClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.bDeleteEndpointMiddleware = opt }
}

func BGetClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.bGetClientOption = opt }
}

func BGetClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.bGetEndpointMiddleware = opt }
}

func BGetAllClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.bGetAllClientOption = opt }
}

func BGetAllClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.bGetAllEndpointMiddleware = opt }
}

func BTestMethodClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.bTestMethodClientOption = opt }
}

func BTestMethodClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.bTestMethodEndpointMiddleware = opt }
}

func BTestMethod2ClientOptions(opt ...jsonrpc.ClientOption) ClientOption {
	return func(c *clientOpts) { c.bTestMethod2ClientOption = opt }
}

func BTestMethod2ClientEndpointMiddlewares(opt ...endpoint.Middleware) ClientOption {
	return func(c *clientOpts) { c.bTestMethod2EndpointMiddleware = opt }
}

type clientA struct {
	aTestMethodEndpoint endpoint.Endpoint
}

func (c *clientA) TestMethod() {
	_, _ = c.aTestMethodEndpoint(context.Background(), nil)
	return
}

type clientB struct {
	bCreateEndpoint      endpoint.Endpoint
	bDeleteEndpoint      endpoint.Endpoint
	bGetEndpoint         endpoint.Endpoint
	bGetAllEndpoint      endpoint.Endpoint
	bTestMethodEndpoint  endpoint.Endpoint
	bTestMethod2Endpoint endpoint.Endpoint
}

func (c *clientB) Create(ctx context.Context, newData Data, name string, data []byte) error {
	_, err := c.bCreateEndpoint(ctx, BCreateCreateRequest{NewData: newData, Name: name, Data: data})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientB) Delete(ctx context.Context, id uint) (string, string, error) {
	resp, err := c.bDeleteEndpoint(ctx, BDeleteDeleteRequest{Id: id})
	if err != nil {
		return "", "", err
	}
	response := resp.(BDeleteDeleteResponse)
	return response.A, response.B, nil
}

func (c *clientB) Get(ctx context.Context, id int, name string, fname string, price float32, n int, b int, cc int) (User, error) {
	resp, err := c.bGetEndpoint(ctx, BGetGetRequest{Id: id, Name: name, Fname: fname, Price: price, N: n, B: b, Cc: cc})
	if err != nil {
		return User{}, err
	}
	response := resp.(User)
	return response, nil
}

func (c *clientB) GetAll(ctx context.Context, members Members) ([]*User, error) {
	resp, err := c.bGetAllEndpoint(ctx, BGetAllGetAllRequest{Members: members})
	if err != nil {
		return nil, err
	}
	response := resp.([]*User)
	return response, nil
}

func (c *clientB) TestMethod(data map[string]interface{}, ss interface{}) (map[string]map[int][]string, error) {
	resp, err := c.bTestMethodEndpoint(context.Background(), BTestMethodTestMethodRequest{Data: data, Ss: ss})
	if err != nil {
		return nil, err
	}
	response := resp.(map[string]map[int][]string)
	return response, nil
}

func (c *clientB) TestMethod2(ctx context.Context, ns string, utype string, user string, restype string, resource string, permission string) error {
	_, err := c.bTestMethod2Endpoint(ctx, BTestMethod2TestMethod2Request{Ns: ns, Utype: utype, User: user, Restype: restype, Resource: resource, Permission: permission})
	if err != nil {
		return err
	}
	return nil
}
