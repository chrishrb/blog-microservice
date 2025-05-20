package cmd

import (
	"context"
	"log/slog"

	"github.com/chrishrb/blog-microservice/user-service/config"
	"github.com/chrishrb/blog-microservice/user-service/server"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
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

		apiServer := server.New("api", cfg.Api.Addr, nil,
			server.NewApiHandler(settings.Api, settings.Storage, settings.JWSVerifier, settings.JWSSigner))
		errCh := make(chan error, 1)
		apiServer.Start(errCh)
		err = <-errCh

		return err
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&configFile, "config-file", "c", "/config/config.yaml",
		"The config file to use")
}
