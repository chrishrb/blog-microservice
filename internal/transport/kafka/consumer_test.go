package kafka_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/testutil"
	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/internal/transport/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func TestListenerProcessesMessagesReceivedFromTheBroker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// start the broker
	broker, clientUrl := kafka.NewBroker(t)
	defer func() {
		broker.Close()
	}()

	// setup the handler
	receivedMsgCh := make(chan struct{})
	handler := func(ctx context.Context, msg *transport.Message) {
		assert.Equal(t, "my-message-id", msg.ID)
		assert.Equal(t, json.RawMessage(`{"someKey":"someValue"}`), msg.Data)
		receivedMsgCh <- struct{}{}
	}

	// connect the consumer to the broker
	consumer := kafka.NewConsumer(
		kafka.WithKafkaBrokerUrls[kafka.Consumer](clientUrl),
	)
	conn, err := consumer.Consume(ctx, "topic123", transport.MessageHandlerFunc(handler))
	require.NoError(t, err)
	defer func() {
		if conn != nil {
			err := conn.Disconnect(ctx)
			require.NoError(t, err)
		}
	}()

	// publish message
	publishMessage(t, ctx, broker.ListenAddrs(), "topic123", transport.Message{
		ID:   "my-message-id",
		Data: json.RawMessage(`{"someKey":"someValue"}`),
	})

	// wait for message to be received / timeout
	select {
	case <-ctx.Done():
		assert.Fail(t, "timeout waiting for test to complete")
	case <-receivedMsgCh:
		// do nothing
	}
}

func TestListenerAddsTraceInformation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tracer, exporter := testutil.GetTracer()

	// start the broker
	broker, clientUrl := kafka.NewBroker(t)
	defer func() {
		broker.Close()
	}()

	// setup the handler
	receivedMsgCh := make(chan struct{})
	handler := func(ctx context.Context, msg *transport.Message) {
		receivedMsgCh <- struct{}{}
	}

	// connect the listener to the broker
	consumer := kafka.NewConsumer(
		kafka.WithKafkaBrokerUrls[kafka.Consumer](clientUrl),
		kafka.WithOtelTracer[kafka.Consumer](tracer),
	)
	conn, err := consumer.Consume(ctx, "topic123", transport.MessageHandlerFunc(handler))
	require.NoError(t, err)
	defer func() {
		if conn != nil {
			err := conn.Disconnect(ctx)
			require.NoError(t, err)
		}
	}()

	// publish message
	newCtx, span := tracer.Start(ctx, "test span")
	defer span.End()
	publishMessage(t, newCtx, broker.ListenAddrs(), "topic123", transport.Message{
		ID:   "my-message-id",
		Data: json.RawMessage(`{"someKey":"someValue"}`),
	})

	// wait for message to be received / timeout
	select {
	case <-ctx.Done():
		assert.Fail(t, "timeout waiting for test to complete")
	case <-receivedMsgCh:
		require.Greater(t, len(exporter.GetSpans()), 0)
		assert.True(t, exporter.GetSpans()[0].Parent.HasTraceID())
		testutil.AssertSpan(t, &exporter.GetSpans()[0], "topic123 receive", map[string]any{
			"messaging.system":                     "kafka",
			"messaging.operation":                  "receive",
			"messaging.message.payload_size_bytes": 53,
			"message_id":                           "my-message-id",
			"message_topic":                         "topic123",
			"messaging.consumer.id": func(val attribute.Value) bool {
				return strings.HasPrefix(val.AsString(), "analytics-service-")
			},
			"messaging.message.conversation_id": func(val attribute.Value) bool {
				return val.AsString() != ""
			},
		})
	}
}

func publishMessage(t *testing.T, ctx context.Context, addrs []string, topic string, msg transport.Message) {
	msgBytes, err := json.Marshal(msg)
	require.NoError(t, err)

	correlationMap := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(correlationMap))

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

	pr := client.ProduceSync(ctx,
		&kgo.Record{
			Value:   msgBytes,
			Topic:   topic,
			Headers: mapToRecordHeaders(correlationMap),
		})
	require.NoError(t, pr.FirstErr())
}

func mapToRecordHeaders(headersMap map[string]string) []kgo.RecordHeader {
	var headers []kgo.RecordHeader
	for key, value := range headersMap {
		headers = append(headers, kgo.RecordHeader{
			Key:   key,
			Value: []byte(value),
		})
	}
	return headers
}
