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

type VerifyAccountHandler struct {
	orgName        string
	websiteBaseURL string
	emailChannel   Channel
}

func NewVerifyAccountHandler(
	orgName string,
	websiteBaseURL string,
	emailChannel Channel,
) VerifyAccountHandler {
	return VerifyAccountHandler{
		orgName:        orgName,
		websiteBaseURL: websiteBaseURL,
		emailChannel:   emailChannel,
	}
}

func (r VerifyAccountHandler) Handle(ctx context.Context, msg *transport.Message) {
	span := trace.SpanFromContext(ctx)

	err := r.handle(ctx, msg)
	if err != nil {
		slog.Error("unable to handle message", slog.String("id", msg.ID), slog.String("event", "VerifyAccountEvent"), "err", err)
		span.SetStatus(codes.Error, "handle VerifyAccountEvent failed")
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "ok")
	}
}

func (r VerifyAccountHandler) handle(ctx context.Context, msg *transport.Message) error {
	var req transport.VerifyAccountEvent
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return fmt.Errorf("unmarshalling %s request payload: %w", msg.ID, err)
	}

	vars := VerifyAccountVariables{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		VerifyLink: fmt.Sprintf("%s/verify-account?token=%s", r.websiteBaseURL, req.Token),
		AppName:    r.orgName,
	}

	switch req.Channel {
	case "email":
		return r.emailChannel.SendVerifyAccount(ctx, req.Recipient, vars)
	}

	return fmt.Errorf("unsupported channel %s", req.Channel)
}
