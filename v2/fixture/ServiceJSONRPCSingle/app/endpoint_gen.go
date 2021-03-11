//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

func MakeServiceCreateEndpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateRequest)
		err := s.Create(ctx, req.NewData, req.Name, req.Data)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

}

func MakeServiceDeleteEndpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteRequest)
		a, b, err := s.Delete(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		return DeleteResponse{A: a, B: b}, nil
	}

}

func MakeServiceGetEndpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetRequest)
		result, err := s.Get(ctx, req.Id, req.Name, req.Fname, req.Price, req.N, req.B, req.Cc)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

}

func MakeServiceGetAllEndpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetAllRequest)
		result, err := s.GetAll(ctx, req.Members)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

}

func MakeServiceTestMethodEndpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(TestMethodRequest)
		result, err := s.TestMethod(req.Data, req.Ss)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

}

func MakeServiceTestMethod2Endpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(TestMethod2Request)
		err := s.TestMethod2(ctx, req.Ns, req.Utype, req.User, req.Restype, req.Resource, req.Permission)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

}

func MakeServiceTestMethodOptionalsEndpoint(s InterfaceB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(TestMethodOptionalsRequest)
		err := s.TestMethodOptionals(ctx, req.Ns, req.Options...)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

}

type ServiceEndpointSet struct {
	CreateEndpoint              endpoint.Endpoint
	DeleteEndpoint              endpoint.Endpoint
	GetEndpoint                 endpoint.Endpoint
	GetAllEndpoint              endpoint.Endpoint
	TestMethodEndpoint          endpoint.Endpoint
	TestMethod2Endpoint         endpoint.Endpoint
	TestMethodOptionalsEndpoint endpoint.Endpoint
}

func MakeServiceEndpointSet(svc InterfaceB) ServiceEndpointSet {
	return ServiceEndpointSet{
		CreateEndpoint:              MakeServiceCreateEndpoint(svc),
		DeleteEndpoint:              MakeServiceDeleteEndpoint(svc),
		GetEndpoint:                 MakeServiceGetEndpoint(svc),
		GetAllEndpoint:              MakeServiceGetAllEndpoint(svc),
		TestMethodEndpoint:          MakeServiceTestMethodEndpoint(svc),
		TestMethod2Endpoint:         MakeServiceTestMethod2Endpoint(svc),
		TestMethodOptionalsEndpoint: MakeServiceTestMethodOptionalsEndpoint(svc),
	}
}

type CreateRequest struct {
	NewData Data   `json:"newData"`
	Name    string `json:"name"`
	Data    []byte `json:"data"`
}
type DeleteRequest struct {
	Id uint `json:"id"`
}
type DeleteResponse struct {
	A string `json:"a"`
	B string `json:"b"`
}
type GetRequest struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Fname string    `json:"fname"`
	Price float32   `json:"price"`
	N     int       `json:"n"`
	B     int       `json:"b"`
	Cc    int       `json:"cc"`
}
type GetAllRequest struct {
	Members Members `json:"members"`
}
type TestMethodRequest struct {
	Data map[string]interface{} `json:"data"`
	Ss   interface{}            `json:"ss"`
}
type TestMethod2Request struct {
	Ns         string `json:"ns"`
	Utype      string `json:"utype"`
	User       string `json:"user"`
	Restype    string `json:"restype"`
	Resource   string `json:"resource"`
	Permission string `json:"permission"`
}
type TestMethodOptionalsRequest struct {
	Ns      string          `json:"ns"`
	Options []OptionService `json:"options"`
}
