package model

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestEncodeMap(t *testing.T) {
	set := attribute.NewSet(
		attribute.String("service.name", "foo"),
		attribute.Int("service.id", 1),
		attribute.Float64("service.version", 1.0),
		attribute.Bool("service.enabled", true),
		attribute.Int64("service.instance.id", 1),
		attribute.Float64Slice("service.instance.version", []float64{1.0}),
	)
	e := &jx.Encoder{}
	require.NoError(t, EncodeAttrSet(set, e))
	t.Log(e.String())
}
