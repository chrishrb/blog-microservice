package kafka_test

import (
	"context"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/transport/kafka"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestNewBroker(t *testing.T) {
	topic := "test"
	recordValue := "testRecordValue"

	broker, addrs := kafka.NewBroker(t)
	defer func() {
		broker.Close()
	}()

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = kadmClient.CreateTopics(ctx, 1, 1, nil, topic)
	require.NoError(t, err)

	pr := client.ProduceSync(ctx,
		&kgo.Record{
			Value: []byte(recordValue),
			Topic: topic,
		})
	require.NoError(t, pr.FirstErr())

	fetches := client.PollFetches(ctx)
	require.True(t, len(fetches.Records()) == 1)

	fetches.EachRecord(func(record *kgo.Record) {
		require.Equal(t, string(record.Value), recordValue)
	})
}
