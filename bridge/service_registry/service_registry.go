package service_registry

import (
	"fmt"
)

type ServiceRegistry struct {
	services map[string]interface{}
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]interface{}),
	}
}

func (s *ServiceRegistry) Register(id string, service interface{}) {
	s.services[id] = service
}

func (s *ServiceRegistry) Get(id string) (interface{}, error) {
	if val, ok := s.services[id]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("Unknown service: %s", id)
}
