package channels

import (
	"context"
)

type PasswordResetVariables struct {
	FirstName string
	LastName  string
	ResetLink string
	AppName   string
}

type VerifyAccountVariables struct {
	FirstName  string
	LastName   string
	VerifyLink string
	AppName    string
}

type Channel interface {
	SendPasswordReset(ctx context.Context, recipient string, vars PasswordResetVariables) error
	SendVerifyAccount(ctx context.Context, recipient string, vars VerifyAccountVariables) error
}
