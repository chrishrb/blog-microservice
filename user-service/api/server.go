package api

import (
	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/utils/clock"
)

type Server struct {
	clock   clock.PassiveClock
	openapi *openapi3.T
}

func NewServer(clock clock.PassiveClock) (*Server, error) {
	swagger, err := GetSwagger()
	if err != nil {
		return nil, err
	}

	return &Server{
		clock:   clock,
		openapi: swagger,
	}, nil
}
