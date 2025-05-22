package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type PasswordResetVariables struct {
	FirstName string
	LastName  string
	ResetLink string
	AppName   string
}

type Channel interface {
	SendPasswordReset(ctx context.Context, recipient string, vars PasswordResetVariables) error
}

type PasswordResetHandler struct {
	orgName        string
	websiteBaseURL string
	emailChannel   Channel
}

func (r PasswordResetHandler) Handle(ctx context.Context, msg *transport.Message) {
	span := trace.SpanFromContext(ctx)

	err := r.handle(ctx, msg)
	if err != nil {
		slog.Error("unable to handle message", slog.String("id", msg.ID), slog.String("event", "PasswordResetEvent"), "err", err)
		span.SetStatus(codes.Error, "handle PasswordResetEvent failed")
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "ok")
	}
}

func NewPasswordResetHandler(
	orgName string,
	websiteBaseURL string,
	emailChannel Channel,
) PasswordResetHandler {
	return PasswordResetHandler{
		orgName:        orgName,
		websiteBaseURL: websiteBaseURL,
		emailChannel:   emailChannel,
	}
}

func (r PasswordResetHandler) handle(ctx context.Context, msg *transport.Message) error {
	var req transport.PasswordResetEvent
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return fmt.Errorf("unmarshalling %s request payload: %w", msg.ID, err)
	}

	vars := PasswordResetVariables{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		ResetLink: fmt.Sprintf("%s/reset-password?token=%s", r.websiteBaseURL, req.Token),
		AppName:   r.orgName,
	}

	switch req.Channel {
	case "email":
		return r.emailChannel.SendPasswordReset(ctx, req.Recipient, vars)
	}

	return fmt.Errorf("unsupported channel %s", req.Channel)
}
