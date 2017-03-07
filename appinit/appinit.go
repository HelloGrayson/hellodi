package appinit

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"

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

	container := dig.New()

	confData := parseConfData(config)
	// // delegate config keys to all participating components,
	// // allowing them to make their deps available on the container
	// for module, moduleConfig := range {
	//     module.Configure(container, moduleConfig)
	// }
	dispatcher := newDispatcher(confData["yarpc"])
	container.Register(dispatcher)

	// register framework types
	logger := newLogger()
	container.Register(logger)

	return &Service{config: config, container: container}
}

// Service is the service being bootstrapped
type Service struct {
	config    string
	container dig.Graph
}

// Procedures is a wrapper for []transport.Procedure
// since the container cant resolve lists
type Procedures struct {
	Register []transport.Procedure
}

// Provide adds a userland type to the container
func (s *Service) Provide(t interface{}) {
	s.container.Register(t)
}

// Start registers framework types, resolves
// the Procedures type from the container, registering
// it's contents with a YARPC dispatcher, and then
// starts the dispatcher
func (s *Service) Start() {
	logger := s.ResolveLogger()
	dispatcher := s.ResolveDispatcher()

	// resolve and register procs
	var procedures *Procedures
	s.container.ResolveAll(&procedures)
	if procedures != nil {
		logger.Info("Registering procedures.", zap.Any("procedures", procedures))
		dispatcher.Register(procedures.Register)
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

// ResolveLogger returns a configured zap.Logger
func (s *Service) ResolveLogger() *zap.Logger {
	var logger *zap.Logger
	s.container.ResolveAll(&logger)
	return logger
}

// ResolveDispatcher returns a configured dispatcher
func (s *Service) ResolveDispatcher() *yarpc.Dispatcher {
	var dispatcher *yarpc.Dispatcher
	s.container.ResolveAll(&dispatcher)
	return dispatcher
}

// parseConfData takes a path to a yaml and returns
// a map[string]interface{}, where the value is delegated
// to a participating configurator
func parseConfData(confPath string) map[string]interface{} {
	confFile, err := os.Open(confPath)
	if err != nil {
		log.Fatal(err)
	}
	defer confFile.Close()

	confData, err := ioutil.ReadAll(confFile)
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(confData, &data); err != nil {
		log.Fatal(err)
	}

	return data
}

// newLogger configures a zap.Logger,
// once configurators are in place, this can and should
// be delegated to the logger component
func newLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// newDispatcher takes the yarpc: key confData and
// configures a yarpc.Dispatcher
func newDispatcher(confData interface{}) *yarpc.Dispatcher {
	cfg := config.New()
	if err := http.RegisterTransport(cfg); err != nil {
		log.Fatal(err)
	}
	if err := tchannel.RegisterTransport(cfg); err != nil {
		log.Fatal(err)
	}

	builder, err := cfg.Load(confData)
	if err != nil {
		log.Fatal(err)
	}

	dispatcher, err := builder.BuildDispatcher()
	if err != nil {
		log.Fatal(err)
	}
	return dispatcher
}
