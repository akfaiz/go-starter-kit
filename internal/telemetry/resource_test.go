package telemetry

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func TestNewResource_IncludesServiceAndSDKAttributes(t *testing.T) {
	res, err := newResource(context.Background(), config.Config{
		Telemetry: config.Telemetry{
			ServiceName: "test-service",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "test-service", attrValue(t, res, semconv.ServiceNameKey))
	assert.NotEmpty(t, attrValue(t, res, semconv.TelemetrySDKNameKey))
	assert.NotEmpty(t, attrValue(t, res, semconv.TelemetrySDKLanguageKey))
	assert.NotEmpty(t, attrValue(t, res, semconv.TelemetrySDKVersionKey))
}

func attrValue(t *testing.T, res *resource.Resource, key attribute.Key) string {
	t.Helper()

	for _, attr := range res.Attributes() {
		if attr.Key == key {
			return attr.Value.AsString()
		}
	}
	require.Failf(t, "attribute not found", "key=%s", key)
	return ""
}
