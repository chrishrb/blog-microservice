package api

import (
	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/utils/clock"
)

type Server struct {
	engine      store.Engine
	clock       clock.PassiveClock
	openapi     *openapi3.T
	jwsVerifier auth.JWSVerifier
	jwsSigner   auth.JWSSigner
	producer    transport.Producer
}

func NewServer(
	engine store.Engine,
	clock clock.PassiveClock,
	jwsVerifier auth.JWSVerifier,
	jwsSigner auth.JWSSigner,
	producer transport.Producer,
) (*Server, error) {
	swagger, err := GetSwagger()
	if err != nil {
		return nil, err
	}

	return &Server{
		engine:      engine,
		clock:       clock,
		openapi:     swagger,
		jwsVerifier: jwsVerifier,
		jwsSigner:   jwsSigner,
		producer:    producer,
	}, nil
}
