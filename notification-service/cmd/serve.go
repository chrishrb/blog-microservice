package cmd

import (
	"context"
	"log/slog"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/notification-service/config"
	"github.com/chrishrb/blog-microservice/notification-service/server"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the notification-service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig
		if configFile != "" {
			err := cfg.LoadFromFile(configFile)
			if err != nil {
				return err
			}
		}

		settings, err := config.Configure(context.Background(), &cfg)
		if err != nil {
			return err
		}
		defer func() {
			err := settings.TracerProvider.Shutdown(context.Background())
			if err != nil {
				slog.Warn("shutting down tracer provider", "error", err)
			}
		}()

		errCh := make(chan error, 1)

		// Start the server
		apiServer := server.New("api", cfg.Api.Addr, nil, server.NewApiHandler())
		apiServer.Start(errCh)

		// Start all consumers
		// TODO: refactor this somehow
		passwordResetConn, err := settings.MsgConsumer.Consume(context.Background(), transport.PasswordResetTopic, settings.PasswordResetHandler)
		if err != nil {
			errCh <- err
		}
		verifyAccountConn, err := settings.MsgConsumer.Consume(context.Background(), transport.VerifyAccountTopic, settings.VerifyAccountHandler)
		if err != nil {
			errCh <- err
		}

		err = <-errCh

		if passwordResetConn != nil {
			err := passwordResetConn.Disconnect(context.Background())
			if err != nil {
				slog.Warn("disconnecting from consumer", "err", err)
			}
		}
		if verifyAccountConn != nil {
			err := verifyAccountConn.Disconnect(context.Background())
			if err != nil {
				slog.Warn("disconnecting from consumer", "err", err)
			}
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&configFile, "config-file", "c", "/config/config.yaml",
		"The config file to use")
}
