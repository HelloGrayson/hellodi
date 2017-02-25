package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/breerly/hellodi/hello"
	"github.com/breerly/hellodi/hello/helloclient"
	"github.com/breerly/hellodi/hello/helloserver"

	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/http"
)

func main() {
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

	dispatcher.Register(helloserver.New(&helloHandler{}))

	if err := dispatcher.Start(); err != nil {
		log.Fatal(err)
	}
	defer dispatcher.Stop()

	client := helloclient.New(dispatcher.ClientConfig("hello"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Echo(ctx, &hello.EchoRequest{Message: "Hello world", Count: 1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}

type helloHandler struct{}

func (h helloHandler) Echo(ctx context.Context, e *hello.EchoRequest) (*hello.EchoResponse, error) {
	return &hello.EchoResponse{Message: e.Message, Count: e.Count + 1}, nil
}
