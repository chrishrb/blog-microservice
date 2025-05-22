package source_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringSource(t *testing.T) {
	source := source.StringSourceProvider{
		Data: "hello world",
	}

	data, err := source.GetData()
	require.NoError(t, err)
	assert.Equal(t, []byte("hello world"), data)
}

func TestFileSource(t *testing.T) {
	source := source.FileSourceProvider{
		FileName: "testdata/jwt.key.pem",
	}

	data, err := source.GetData()
	require.NoError(t, err)
	assert.Contains(t, string(data), "-----BEGIN EC PRIVATE KEY-----")
}
