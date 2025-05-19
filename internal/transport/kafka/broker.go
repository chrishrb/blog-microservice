package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kfake"
)

func NewBroker(t *testing.T) (*kfake.Cluster, []string) {
	cluster, err := kfake.NewCluster(kfake.AllowAutoTopicCreation())
	require.NoError(t, err)

	addrs := cluster.ListenAddrs()
	return cluster, addrs
}
