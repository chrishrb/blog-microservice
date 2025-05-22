package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Consumer struct {
	connectionDetails
	tracer trace.Tracer
}

func NewConsumer(opts ...Opt[Consumer]) *Consumer {
	c := new(Consumer)
	for _, opt := range opts {
		opt(c)
	}
	ensureConsumerDefaults(c)
	// TODO: check if connection can be established
	return c
}

func ensureConsumerDefaults(c *Consumer) {
	if len(c.kafkaBrokerUrls) == 0 {
		c.kafkaBrokerUrls = []string{"127.0.0.1:9092"}
	}
	if c.tracer == nil {
		c.tracer = noop.NewTracerProvider().Tracer("")
	}
	if c.kafkaConnectTimeout == 0 {
		c.kafkaConnectTimeout = 10 * time.Second
	}
	if c.kafkaConsumerGroup == "" {
		c.kafkaConsumerGroup = "analytics-service"
	}
}

func (c *Consumer) Consume(ctx context.Context, topic string, handler transport.MessageHandler) (transport.Connection, error) {
	var err error

	clientId := fmt.Sprintf("%s-%s", c.kafkaConsumerGroup, randSeq(5))

	conn := new(connection)
	conn.kafkaClient, err = kgo.NewClient(
		kgo.SeedBrokers(c.kafkaBrokerUrls...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumerGroup(c.kafkaConsumerGroup),
		kgo.ClientID(clientId),
		kgo.DialTimeout(c.kafkaConnectTimeout),
	)

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			fetches := conn.kafkaClient.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}

			if errs := fetches.Errors(); len(errs) > 0 {
				slog.Error("error fetching message from broker", "topic", topic, "errs", errs)
			}

			fetches.EachRecord(func(r *kgo.Record) {
				ctx := context.Background()

				if len(r.Headers) != 0 {
					correlationMap := RecordHeadersToMap(r.Headers)
					ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(correlationMap))
				}

				// create span
				newCtx, span := c.tracer.Start(ctx,
					fmt.Sprintf("%s receive", topic),
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingSystem("kafka"),
						semconv.MessagingConsumerID(clientId),
						semconv.MessagingMessagePayloadSizeBytes(len(r.Value)),
						semconv.MessagingOperationKey.String("receive"),
					))
				defer span.End()

				// unmarshal the message
				var msg transport.Message
				err := json.Unmarshal(r.Value, &msg)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "unable to unmarshal message")
					slog.Warn("unable to unmarshal message", "err", err)
					return
				}

				// add additional span attributes
				span.SetAttributes(
					attribute.String("message_id", msg.ID),
					attribute.String("message_topic", topic),
					semconv.MessagingMessageConversationID(msg.ID),
				)

				// execute the handler
				handler.Handle(newCtx, &msg)
			})
		}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("timeout waiting for kafka setup")
	default:
		return conn, nil
	}
}

type connection struct {
	kafkaClient *kgo.Client
}

func (c *connection) Disconnect(ctx context.Context) error {
	if c.kafkaClient != nil {
		c.kafkaClient.Close()
		c.kafkaClient = nil
	}
	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		//#nosec G404 - client suffix does not require secure random number generator
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RecordHeadersToMap(headers []kgo.RecordHeader) map[string]string {
	headersMap := make(map[string]string)
	for _, header := range headers {
		headersMap[header.Key] = string(header.Value)
	}
	return headersMap
}
