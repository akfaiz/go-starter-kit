package telemetry

import (
	"context"
	"fmt"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func newResource(ctx context.Context, cfg config.Config) (*resource.Resource, error) {
	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			attribute.String("service.name", cfg.Telemetry.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("init telemetry resource: %w", err)
	}
	return res, nil
}
