package fx2

import "go.uber.org/fx/dig"

// NewService creates an app framework service
func NewService() *Service {
	return &Service{container: dig.New()}
}

// Service is the service being bootstrapped
type Service struct {
	container dig.Graph

	types []interface{}
}

// RegisterType adds a userland type to the container
func (s *Service) RegisterType(t interface{}) {
	s.types = append(s.types, t)
}

// Start and starts the messaging framework
func (s *Service) Start() {
	// register framework types
	s.container.Register(newLogger)
	s.container.Register(newDispatcher)

	// register userland types and handlers
	for _, t := range s.types {
		s.container.Register(t)
	}

	// handlers should be registered with RegisterHandler
	// which we can resolve and register with the dispatcher
	// before calling dispatcher.Start()
	//var h *handler
	//s.container.ResolveAll(&h)
	//h.Hello()
}
