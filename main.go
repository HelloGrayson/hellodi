package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/http"
	"go.uber.org/zap"

	"github.com/breerly/hellodi/hello"
	"github.com/breerly/hellodi/hello/helloclient"
	"github.com/breerly/hellodi/hello/helloserver"
)

func main() {
	logger, _ := zap.NewDevelopment()

	dispatcher := newDispatcher()

	client := newHelloClient(dispatcher)
	handler := newHelloHandler(logger, client)

	dispatcher.Register(helloserver.New(handler))

	if err := dispatcher.Start(); err != nil {
		logger.Fatal("Unable to start service", zap.Any("err", err))
	}
	defer dispatcher.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}

func newDispatcher() *yarpc.Dispatcher {
	http := http.NewTransport()
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: "hello",
		Inbounds: yarpc.Inbounds{
			http.NewInbound(":8086"),
		},
		Outbounds: yarpc.Outbounds{
			"hello": {
				Unary: http.NewSingleOutbound("http://127.0.0.1:8086"),
			},
		},
	})
	return dispatcher
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

	resp, err := h.helloClient.Echo(ctx, req.Echo)
	if err != nil {
		h.logger.Fatal("Failed to call home", zap.Any("request", req.Echo))
	}

	return &hello.CallHomeResponse{Echo: resp}, nil
}
