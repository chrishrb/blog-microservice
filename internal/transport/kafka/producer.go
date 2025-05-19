package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Producer struct {
	sync.Mutex
	connectionDetails
	conn   *kgo.Client
	tracer trace.Tracer
}

func NewProducer(opts ...Opt[Producer]) *Producer {
	e := new(Producer)
	for _, opt := range opts {
		opt(e)
	}
	ensureProducerDefaults(e)
	return e
}

func (p *Producer) Produce(ctx context.Context, topic string, message *transport.Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshaling message of type %s: %v", message.Type, err)
	}

	newCtx, span := p.tracer.Start(ctx,
		fmt.Sprintf("%s produce", topic),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystem("kafka"),
			semconv.MessagingMessagePayloadSizeBytes(len(payload)),
			semconv.MessagingOperationKey.String("produce"),
			semconv.MessagingMessageConversationID(message.Type),
			attribute.String("message_id", message.ID),
			attribute.String("message_type", string(message.Type)),
		))
	defer span.End()

	correlationMap := make(map[string]string)
	otel.GetTextMapPropagator().Inject(newCtx, propagation.MapCarrier(correlationMap))

	err = p.ensureConnection()
	if err != nil {
		return fmt.Errorf("connecting to kafka: %v", err)
	}

	record := &kgo.Record{
		Topic:   topic,
		Value:   payload,
		Headers: mapToRecordHeaders(correlationMap),
	}
	err = p.conn.ProduceSync(newCtx, record).FirstErr()
	if err != nil {
		return fmt.Errorf("publishing to %s: %v", topic, err)
	}
	return nil
}

func ensureProducerDefaults(p *Producer) {
	if len(p.kafkaBrokerUrls) == 0 {
		p.kafkaBrokerUrls = []string{"127.0.0.1:9092"}
	}
	if p.tracer == nil {
		p.tracer = noop.NewTracerProvider().Tracer("")
	}
	if p.kafkaConnectTimeout == 0 {
		p.kafkaConnectTimeout = 10 * time.Second
	}
}

func (p *Producer) ensureConnection() error {
	p.Lock()
	defer p.Unlock()
	if p.conn == nil {
		conn, err := kgo.NewClient(
			kgo.SeedBrokers(p.kafkaBrokerUrls...),
			kgo.DialTimeout(p.kafkaConnectTimeout),
			kgo.AllowAutoTopicCreation(),
			// TODO: add auth
		)
		if err != nil {
			return err
		}
		p.conn = conn
	}
	return nil
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
