package source_test

import (
	"context"
	"testing"

	"github.com/chrishrb/blog-microservice/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringSource(t *testing.T) {
	source := source.StringSourceProvider{
		Data: "hello world",
	}

	data, err := source.GetData(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "hello world", data)
}

func TestFileSource(t *testing.T) {
	source := source.FileSourceProvider{
		FileName: "testdata/jwt.key.pem",
	}

	data, err := source.GetData(context.TODO())
	require.NoError(t, err)
	assert.Contains(t, data, "-----BEGIN EC PRIVATE KEY-----")
}
