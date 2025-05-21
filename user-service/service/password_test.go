package service_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/stretchr/testify/require"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "my_secure_password"

	hashedPassword, err := service.HashPassword(password)
	require.NoError(t, err)

	if !service.VerifyPassword(password, hashedPassword) {
		t.Errorf("Expected password verification to succeed")
	}

	if service.VerifyPassword("wrong_password", hashedPassword) {
		t.Errorf("Expected password verification to fail")
	}
}
