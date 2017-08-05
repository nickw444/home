package services

import (
	"fmt"
)

type Registry struct {
	services map[string]interface{}
}

func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]interface{}),
	}
}

func (s *Registry) Register(id string, service interface{}) {
	s.services[id] = service
}

func (s *Registry) Get(id string) (interface{}, error) {
	if val, ok := s.services[id]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("Unknown service: %s", id)
}
