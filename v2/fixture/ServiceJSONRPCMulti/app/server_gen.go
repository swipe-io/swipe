//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/l-vitaly/go-kit/transport/http/jsonrpc"
	"github.com/pquerna/ffjson/ffjson"
)

func MergeEndpointCodecMaps(ecms ...jsonrpc.EndpointCodecMap) jsonrpc.EndpointCodecMap {
	mergedECM := make(jsonrpc.EndpointCodecMap, 512)
	for _, ecm := range ecms {
		for key, codec := range ecm {
			mergedECM[key] = codec
		}
	}
	return mergedECM
}
func encodeResponseJSONRPC(_ context.Context, result interface{}) (json.RawMessage, error) {
	b, err := ffjson.Marshal(result)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func MakeAEndpointCodecMap(ep AEndpointSet, ns ...string) jsonrpc.EndpointCodecMap {
	var namespace string
	if len(ns) > 0 {
		namespace = strings.Join(ns, ".") + "."
	}
	ecm := jsonrpc.EndpointCodecMap{}
	if ep.TestMethodEndpoint != nil {
		ecm[namespace+"testMethod"] = jsonrpc.EndpointCodec{
			Endpoint: ep.TestMethodEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				return nil, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	return ecm
}

func MakeBEndpointCodecMap(ep BEndpointSet, ns ...string) jsonrpc.EndpointCodecMap {
	var namespace string
	if len(ns) > 0 {
		namespace = strings.Join(ns, ".") + "."
	}
	ecm := jsonrpc.EndpointCodecMap{}
	if ep.CreateEndpoint != nil {
		ecm[namespace+"create"] = jsonrpc.EndpointCodec{
			Endpoint: ep.CreateEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req BCreateCreateRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to BCreateCreateRequest: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	if ep.DeleteEndpoint != nil {
		ecm[namespace+"delete"] = jsonrpc.EndpointCodec{
			Endpoint: ep.DeleteEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req BDeleteDeleteRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to BDeleteDeleteRequest: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	if ep.GetEndpoint != nil {
		ecm[namespace+"get"] = jsonrpc.EndpointCodec{
			Endpoint: ep.GetEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req BGetGetRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to BGetGetRequest: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	if ep.GetAllEndpoint != nil {
		ecm[namespace+"getAll"] = jsonrpc.EndpointCodec{
			Endpoint: ep.GetAllEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req BGetAllGetAllRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to BGetAllGetAllRequest: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	if ep.TestMethodEndpoint != nil {
		ecm[namespace+"testMethod"] = jsonrpc.EndpointCodec{
			Endpoint: ep.TestMethodEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req BTestMethodTestMethodRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to BTestMethodTestMethodRequest: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	if ep.TestMethod2Endpoint != nil {
		ecm[namespace+"testMethod2"] = jsonrpc.EndpointCodec{
			Endpoint: ep.TestMethod2Endpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req BTestMethod2TestMethod2Request
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to BTestMethod2TestMethod2Request: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	return ecm
}

// HTTP JSONRPC Transport
func MakeHandlerJSONRPC(svcA InterfaceA, svcB InterfaceB, options ...ServerOption) (http.Handler, error) {
	opts := &serverOpts{}
	for _, o := range options {
		o(opts)
	}
	epSetA := MakeAEndpointSet(svcA)
	epSetA.TestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.aTestMethodEndpointMiddleware...))(epSetA.TestMethodEndpoint)
	epSetB := MakeBEndpointSet(svcB)
	epSetB.CreateEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bCreateEndpointMiddleware...))(epSetB.CreateEndpoint)
	epSetB.DeleteEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bDeleteEndpointMiddleware...))(epSetB.DeleteEndpoint)
	epSetB.GetEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bGetEndpointMiddleware...))(epSetB.GetEndpoint)
	epSetB.GetAllEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bGetAllEndpointMiddleware...))(epSetB.GetAllEndpoint)
	epSetB.TestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bTestMethodEndpointMiddleware...))(epSetB.TestMethodEndpoint)
	epSetB.TestMethod2Endpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.bTestMethod2EndpointMiddleware...))(epSetB.TestMethod2Endpoint)
	r := mux.NewRouter()
	handler := jsonrpc.NewServer(MergeEndpointCodecMaps(MakeAEndpointCodecMap(epSetA, "a"), MakeBEndpointCodecMap(epSetB, "b")), opts.genericServerOption...)
	r.Methods("POST").Path("").Handler(handler)
	return r, nil
}
