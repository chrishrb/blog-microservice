package email_test

import (
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/notification-service/channels"
	"github.com/chrishrb/blog-microservice/notification-service/channels/email"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPasswordReset(t *testing.T) {
	server, hostAddr, port := setupServer(t)

	c, err := email.NewEmailChannel(hostAddr, port, "", "", "user@example.com")
	require.NoError(t, err)

	err = c.SendPasswordReset(t.Context(), "john@example.com", channels.PasswordResetVariables{
		FirstName: "John",
		LastName:  "Doe",
		ResetLink: "https://example.com/reset-password?token=123",
		AppName:   "MyApp",
	})
	require.NoError(t, err)

	messages, err := server.WaitForMessages(1, 10*time.Millisecond)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "MAIL FROM:<user@example.com>", messages[0].MailfromRequest())
}

func TestSendVerifyAccount(t *testing.T) {
	server, hostAddr, port := setupServer(t)

	c, err := email.NewEmailChannel(hostAddr, port, "", "", "user@example.com")
	require.NoError(t, err)

	err = c.SendVerifyAccount(t.Context(), "john@example.com", channels.VerifyAccountVariables{
		FirstName:  "John",
		LastName:   "Doe",
		VerifyLink: "https://example.com/verify-account?token=123",
		AppName:    "MyApp",
	})
	require.NoError(t, err)

	messages, err := server.WaitForMessages(1, 10*time.Millisecond)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "MAIL FROM:<user@example.com>", messages[0].MailfromRequest())
}

func setupServer(t *testing.T) (*smtpmock.Server, string, int) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       true,
		LogServerActivity: true,
	})

	err := server.Start()
	require.NoError(t, err)

	return server, "127.0.0.1", server.PortNumber()
}
