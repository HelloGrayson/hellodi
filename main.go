package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/yarpc"
	"go.uber.org/zap"

	"github.com/breerly/hellodi/fx2"
	"github.com/breerly/hellodi/hello"
	"github.com/breerly/hellodi/hello/helloclient"
	"github.com/breerly/hellodi/hello/helloserver"
)

func main() {
	service := fx2.New()

	service.RegisterType(newProcs)
	service.RegisterType(newHelloClient)
	service.RegisterType(newHelloHandler)

	service.Start()
	defer service.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}

func newProcs(helloHandler *helloHandler) *fx2.Procs {
	return &fx2.Procs{
		Value: helloserver.New(helloHandler),
	}
}

func newHelloClient(d *yarpc.Dispatcher) helloclient.Interface {
	return helloclient.New(d.ClientConfig("hello"))
}

func newHelloHandler(logger *zap.Logger, helloClient helloclient.Interface) *helloHandler {
	return &helloHandler{logger: logger, helloClient: helloClient}
}

type helloHandler struct {
	logger      *zap.Logger
	helloClient helloclient.Interface
}

func (h *helloHandler) Echo(ctx context.Context, req *hello.EchoRequest) (*hello.EchoResponse, error) {
	h.logger.Info("Echo", zap.Any("message", req.Message))
	return &hello.EchoResponse{Message: req.Message, Count: req.Count + 1}, nil
}

func (h *helloHandler) CallHome(ctx context.Context, req *hello.CallHomeRequest) (*hello.CallHomeResponse, error) {
	h.logger.Info("CallHome", zap.Any("echo", req.String()))
	return &hello.CallHomeResponse{Echo: &hello.EchoResponse{Message: "called home!"}}, nil
}
