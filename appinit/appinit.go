package appinit

import (
	"log"
	"os"

	"go.uber.org/fx/dig"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/http"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/yarpc/x/config"
	"go.uber.org/zap"
)

// New creates an app framework service
func New(config string) *Service {
	return &Service{config: config, container: dig.New()}
}

// Service is the service being bootstrapped
type Service struct {
	config    string
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
	dispatcher := newDispatcher(s.config)
	s.container.Register(dispatcher)

	// register framework types
	logger := newLogger()
	s.container.Register(logger)

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

func newDispatcher(yamlPath string) *yarpc.Dispatcher {
	cfg := config.New()
	if err := http.RegisterTransport(cfg); err != nil {
		log.Fatal(err)
	}
	if err := tchannel.RegisterTransport(cfg); err != nil {
		log.Fatal(err)
	}
	confFile, err := os.Open(yamlPath)
	if err != nil {
		log.Fatal(err)
	}
	defer confFile.Close()
	builder, err := cfg.LoadYAML(confFile)
	if err != nil {
		log.Fatal(err)
	}
	dispatcher, err := builder.BuildDispatcher()
	if err != nil {
		log.Fatal(err)
	}
	return dispatcher
}
