//+build !swipe

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/valyala/fasthttp"

	"github.com/go-kit/kit/log"

	"github.com/swipe-io/swipe/fixtures/config"
	"github.com/swipe-io/swipe/fixtures/service"
	"github.com/swipe-io/swipe/fixtures/transport/jsonrpc"
	"github.com/swipe-io/swipe/fixtures/transport/rest"
)

func main() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	logger := log.NewLogfmtLogger(os.Stdout)

	cfg, errs := config.LoadConfig()
	if len(errs) > 0 {
		for _, err := range errs {
			_ = logger.Log("err", err)
		}
		os.Exit(1)
	}

	fmt.Println(cfg)

	svc := service.Interface(&service.Service{})
	svc = jsonrpc.NewLoggingMiddlewareServiceInterface(svc, logger)

	httpJSONRPCHandler, err := jsonrpc.MakeHandlerJSONRPCServiceInterface(svc)
	if err != nil {
		_ = logger.Log("err", err)
		os.Exit(1)
	}

	httpRestHandler, err := rest.MakeHandlerRESTServiceInterface(svc)
	if err != nil {
		_ = logger.Log("err", err)
		os.Exit(1)
	}

	go func() {
		_ = http.ListenAndServe(":8080", httpJSONRPCHandler)
	}()
	go func() {
		_ = fasthttp.ListenAndServe(":8081", httpRestHandler)
	}()

	<-sigint
}
