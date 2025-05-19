package kafka_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/internal/transport/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestProducerSendsMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// start the broker
	broker, clientUrl := kafka.NewBroker(t)
	defer func() {
		broker.Close()
	}()

	producer := kafka.NewProducer(
		kafka.WithKafkaBrokerUrls[kafka.Producer](clientUrl),
	)

	// subscribe to the output channel
	rcvdCh := make(chan struct{})

	handler := func(ctx context.Context, msg *transport.Message, headers []kgo.RecordHeader) {
		assert.Equal(t, "my-message-id", msg.ID)
		assert.Equal(t, transport.MessageTypeUserRegistered, msg.Type)
		assert.Equal(t, json.RawMessage(`{"someKey":"someValue"}`), msg.Data)
		rcvdCh <- struct{}{}
	}
	client := listenForMessageSent(t, ctx, clientUrl, "topic123", handler)

	defer func() {
		client.Close()
	}()

	// publish a message to the input channel
	msg := transport.Message{
		ID:   "my-message-id",
		Type: "UserRegistered",
		Data:     json.RawMessage(`{"someKey":"someValue"}`),
	}
	err := producer.Produce(context.Background(), "topic123", &msg)
	require.NoError(t, err)

	// wait for success
	select {
	case <-rcvdCh:
		// success
	case <-ctx.Done():
		assert.Fail(t, "timeout waiting for test")
	}
}

func TestProducerAddsCorrelationData(t *testing.T) {
	traceExporter := tracetest.NewInMemoryExporter()
	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithSyncer(traceExporter),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tracer := tracerProvider.Tracer("test")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// start the broker
	broker, clientUrl := kafka.NewBroker(t)
	defer func() {
		broker.Close()
	}()

	producer := kafka.NewProducer(
		kafka.WithKafkaBrokerUrls[kafka.Producer](clientUrl),
		kafka.WithOtelTracer[kafka.Producer](tracer))

	// subscribe to the output channel
	rcvdCh := make(chan struct{})

	handler := func(ctx context.Context, msg *transport.Message, headers []kgo.RecordHeader) {
		require.NotEmpty(t, headers)
		correlationMap := kafka.RecordHeadersToMap(headers)
		assert.NotEmpty(t, correlationMap["traceparent"])

		rcvdCh <- struct{}{}
	}
	client := listenForMessageSent(t, ctx, clientUrl, "topic123", handler)

	defer func() {
		client.Close()
	}()

	// publish a message to the input channel
	msg := transport.Message{
		ID:   "my-message-id",
		Type: "UserRegistered",
		Data:     json.RawMessage(`{"someKey":"someValue"}`),
	}
	err := producer.Produce(context.Background(), "topic123", &msg)
	require.NoError(t, err)

	// wait for success
	select {
	case <-rcvdCh:
		// success
	case <-ctx.Done():
		assert.Fail(t, "timeout waiting for test")
	}
}

func listenForMessageSent(t *testing.T, ctx context.Context, addrs []string, topic string, handler func(ctx context.Context, msg *transport.Message, headers []kgo.RecordHeader)) *kgo.Client {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(addrs...),
		kgo.ConsumeTopics(topic),
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.DisableIdempotentWrite(),
		kgo.MaxProduceRequestsInflightPerBroker(2),
	)
	require.NoError(t, err)

	kadmClient := kadm.NewClient(client)
	t.Cleanup(kadmClient.Close)

	_, err = kadmClient.CreateTopics(ctx, 1, 1, nil, topic)
	require.NoError(t, err)

	go func() {
		for {
			fetches := client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}
			fetches.EachRecord(func(r *kgo.Record) {
				var msg transport.Message
				err := json.Unmarshal(r.Value, &msg)
				require.NoError(t, err)

				handler(ctx, &msg, r.Headers)
			})
		}
	}()

	return client
}
