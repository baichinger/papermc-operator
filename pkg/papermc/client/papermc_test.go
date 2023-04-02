package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const version = "1.19.2"

func TestNewClient(t *testing.T) {
	client := NewPapermcClient(context.TODO())

	assert.NotNil(t, client)
}

func TestGetCurrentBuildForVersion(t *testing.T) {
	client := NewPapermcClient(context.TODO())
	build, err := client.GetBuildForVersion(version)

	assert.NoError(t, err)
	assert.Greater(t, build, 0)

	t.Logf("build for version=%s: %d", version, build)
}

func TestGetUrlForVersionBuildDownload(t *testing.T) {
	client := NewPapermcClient(context.TODO())
	build, err := client.GetBuildForVersion(version)

	require.NoError(t, err)
	require.Greater(t, build, 0)

	url, err := client.GetUrlForVersionBuildDownload(version, build)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)

	t.Logf("download url for version=%s build=%d: %s", version, build, url)
}
