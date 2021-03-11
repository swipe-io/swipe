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

func MakeInterfaceBEndpointCodecMap(ep InterfaceBEndpointSet, ns ...string) jsonrpc.EndpointCodecMap {
	var namespace string
	if len(ns) > 0 {
		namespace = strings.Join(ns, ".") + "."
	}
	ecm := jsonrpc.EndpointCodecMap{}
	if ep.CreateEndpoint != nil {
		ecm[namespace+"create"] = jsonrpc.EndpointCodec{
			Endpoint: ep.CreateEndpoint,
			Decode: func(_ context.Context, msg json.RawMessage) (interface{}, error) {
				var req CreateRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to CreateRequest: %s", err)
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
				var req DeleteRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to DeleteRequest: %s", err)
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
				var req GetRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to GetRequest: %s", err)
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
				var req GetAllRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to GetAllRequest: %s", err)
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
				var req TestMethodRequest
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to TestMethodRequest: %s", err)
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
				var req TestMethod2Request
				err := ffjson.Unmarshal(msg, &req)
				if err != nil {
					return nil, fmt.Errorf("couldn't unmarshal body to TestMethod2Request: %s", err)
				}
				return req, nil
			},
			Encode: encodeResponseJSONRPC,
		}
	}
	return ecm
}

// HTTP JSONRPC Transport
func MakeHandlerJSONRPC(svcInterfaceB InterfaceB, options ...ServerOption) (http.Handler, error) {
	opts := &serverOpts{}
	for _, o := range options {
		o(opts)
	}
	epSet := MakeInterfaceBEndpointSet(svcInterfaceB)
	epSet.CreateEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBCreateEndpointMiddleware...))(epSet.CreateEndpoint)
	epSet.DeleteEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBDeleteEndpointMiddleware...))(epSet.DeleteEndpoint)
	epSet.GetEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBGetEndpointMiddleware...))(epSet.GetEndpoint)
	epSet.GetAllEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBGetAllEndpointMiddleware...))(epSet.GetAllEndpoint)
	epSet.TestMethodEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBTestMethodEndpointMiddleware...))(epSet.TestMethodEndpoint)
	epSet.TestMethod2Endpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.interfaceBTestMethod2EndpointMiddleware...))(epSet.TestMethod2Endpoint)
	r := mux.NewRouter()
	handler := jsonrpc.NewServer(MakeInterfaceBEndpointCodecMap(epSet), opts.genericServerOption...)
	r.Methods("POST").Path("").Handler(handler)
	return r, nil
}