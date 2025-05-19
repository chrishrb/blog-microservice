package kafka

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type connectionDetails struct {
	kafkaBrokerUrls     []string
	kafkaConsumerGroup  string
	kafkaConnectTimeout time.Duration
}

type Opt[T any] func(h *T)

func WithKafkaBrokerUrl[T Producer | Consumer](brokerUrl string) Opt[T] {
	return func(h *T) {
		switch x := any(h).(type) {
		case *Producer:
			x.kafkaBrokerUrls = append(x.kafkaBrokerUrls, brokerUrl)
		case *Consumer:
			x.kafkaBrokerUrls = append(x.kafkaBrokerUrls, brokerUrl)
		}
	}
}

func WithKafkaBrokerUrls[T Producer | Consumer](brokerUrls []string) Opt[T] {
	return func(h *T) {
		switch x := any(h).(type) {
		case *Producer:
			x.kafkaBrokerUrls = brokerUrls
		case *Consumer:
			x.kafkaBrokerUrls = brokerUrls
		}
	}
}

func WithKafkaConnectSettings[T Producer | Consumer](kafkaConnectTimeout time.Duration) Opt[T] {
	return func(h *T) {
		switch x := any(h).(type) {
		case *Producer:
			x.kafkaConnectTimeout = kafkaConnectTimeout
		case *Consumer:
			x.kafkaConnectTimeout = kafkaConnectTimeout
		}
	}
}

func WithOtelTracer[T Producer | Consumer](tracer trace.Tracer) Opt[T] {
	return func(h *T) {
		switch x := any(h).(type) {
		case *Producer:
			x.tracer = tracer
		case *Consumer:
			x.tracer = tracer
		}
	}
}

func WithKafkaConsumerGroup[T Consumer](kafkaConsumerGroup string) Opt[T] {
	return func(h *T) {
		switch x := any(h).(type) {
		case *Consumer:
			x.kafkaConsumerGroup = kafkaConsumerGroup
		}
	}
}
