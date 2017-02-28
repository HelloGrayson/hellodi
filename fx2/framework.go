package fx2

import (
	"log"

	"go.uber.org/fx/dig"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/http"
	"go.uber.org/zap"
)

// New creates an app framework service
func New() *Service {
	return &Service{container: dig.New()}
}

// Service is the service being bootstrapped
type Service struct {
	container dig.Graph
}

// Procs is a wrapper for []transport.Procedure
// since the container cant resolve lists
type Procs struct {
	Value []transport.Procedure
}

// RegisterType adds a userland type to the container
func (s *Service) RegisterType(t interface{}) {
	s.container.Register(t)
}

// Start and starts the messaging framework
func (s *Service) Start() {
	// register framework types
	logger := newLogger()
	s.container.Register(logger)
	dispatcher := newDispatcher()
	s.container.Register(dispatcher)

	// resolve and register procs
	// note we have to use an internal type here,
	// which we wouldnt have to if there was named deps support
	var procs *Procs
	s.container.ResolveAll(&procs)
	if procs != nil {
		logger.Info("Found procs, registering.", zap.Any("procs", procs))
		dispatcher.Register(procs.Value)
	} else {
		logger.Fatal("found no procs, exiting.")
	}

	if err := dispatcher.Start(); err != nil {
		log.Fatal(err)
	}
}

// Stop stops the service
func (s *Service) Stop() {
	var d *yarpc.Dispatcher
	s.container.ResolveAll(&d)
	d.Stop()
}

func newLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
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
