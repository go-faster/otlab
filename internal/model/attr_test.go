package model

import (
	"context"
	"slices"
	"testing"

	"github.com/go-faster/sdk/app"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestAttributes(t *testing.T) {
	res, err := app.Resource(context.Background())
	require.NoError(t, err)

	attrs := slices.Clone(res.Attributes())
	attrs = append(attrs, attribute.Int("service.id", 1))
	set := attribute.NewSet(attrs...)
	data, err := set.MarshalJSON()
	require.NoError(t, err)

	t.Log(string(data))
}
