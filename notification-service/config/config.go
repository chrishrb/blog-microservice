package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"log/slog"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/internal/transport/kafka"
	"github.com/chrishrb/blog-microservice/notification-service/channels"
	"github.com/chrishrb/blog-microservice/notification-service/channels/email"
	"github.com/subnova/slog-exporter/slogtrace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Tracer               oteltrace.Tracer
	TracerProvider       *trace.TracerProvider
	MsgProducer          transport.Producer
	MsgConsumer          transport.Consumer
	PasswordResetHandler transport.MessageHandler
}

func Configure(ctx context.Context, cfg *BaseConfig) (c *Config, err error) {
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	c = &Config{}

	switch cfg.Observability.LogFormat {
	case "json":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	case "text":
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	default:
		return nil, fmt.Errorf("unknown log format: %s", cfg.Observability.LogFormat)
	}

	c.TracerProvider, err = getTracerProvider(ctx, cfg.Observability.OtelCollectorAddr)
	if err != nil {
		return nil, err
	}

	c.Tracer = c.TracerProvider.Tracer("notification-service")

	c.MsgProducer, err = getMsgProducer(&cfg.Transport, c.Tracer)
	if err != nil {
		return nil, err
	}

	c.MsgConsumer, err = getMsgConsumer(&cfg.Transport, c.Tracer)
	if err != nil {
		return nil, err
	}

	c.PasswordResetHandler, err = getPasswordResetHandler(cfg)
	if err != nil {
		return nil, err
	}

	return
}

func attributeFilter(_ attribute.KeyValue) bool {
	return true
}

func getTracerProvider(ctx context.Context, collectorAddr string) (*trace.TracerProvider, error) {
	var err error
	var res *resource.Resource
	var traceExporter trace.SpanExporter

	if collectorAddr != "" {
		res, err = resource.New(ctx,
			resource.WithDetectors(gcp.NewDetector()),
			resource.WithTelemetrySDK(),
			resource.WithAttributes(
				// the service name used to display traces in backends
				semconv.ServiceName("blog-notification-service"),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		conn, err := grpc.NewClient(collectorAddr,
			// Note the use of insecure transport here. TLS is recommended in production.
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		// Set up a trace exporter
		traceExporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	} else {
		res, err = resource.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource: %w", err)
		}

		traceExporter, err = slogtrace.New(attributeFilter)
		if err != nil {
			return nil, err
		}
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := trace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}

func getMsgProducer(cfg *TransportSettingsConfig, tracer oteltrace.Tracer) (transport.Producer, error) {
	switch cfg.Type {
	case "kafka":
		kafkaConnectTimeout, err := time.ParseDuration(cfg.Kafka.ConnectTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse mqtt connect timeout: %w", err)
		}

		kafkaProducer := kafka.NewProducer(
			kafka.WithKafkaBrokerUrls[kafka.Producer](cfg.Kafka.Urls),
			kafka.WithKafkaConnectSettings[kafka.Producer](kafkaConnectTimeout),
			kafka.WithOtelTracer[kafka.Producer](tracer))

		return kafkaProducer, nil
	default:
		return nil, fmt.Errorf("unknown transport type: %s", cfg.Type)
	}
}

func getMsgConsumer(cfg *TransportSettingsConfig, tracer oteltrace.Tracer) (transport.Consumer, error) {
	switch cfg.Type {
	case "kafka":
		kafkaConnectTimeout, err := time.ParseDuration(cfg.Kafka.ConnectTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse mqtt connect timeout: %w", err)
		}

		opts := []kafka.Opt[kafka.Consumer]{
			kafka.WithKafkaBrokerUrls[kafka.Consumer](cfg.Kafka.Urls),
			kafka.WithKafkaConnectSettings[kafka.Consumer](kafkaConnectTimeout),
			kafka.WithKafkaConsumerGroup(cfg.Kafka.Group),
			kafka.WithOtelTracer[kafka.Consumer](tracer),
		}

		return kafka.NewConsumer(opts...), nil
	default:
		return nil, fmt.Errorf("unknown transport type: %s", cfg.Type)
	}
}

func getPasswordResetHandler(cfg *BaseConfig) (transport.MessageHandler, error) {
	emailChannel, err := email.NewEmailChannel(
		cfg.Channels.Email.Host,
		cfg.Channels.Email.Port,
		cfg.Channels.Email.Username,
		cfg.Channels.Email.Password,
		cfg.Channels.Email.FromAddr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create email channel: %w", err)
	}

	return channels.NewPasswordResetHandler(
		cfg.General.OrgName,
		cfg.General.WebsiteBaseURL,
		emailChannel,
	), nil
}
