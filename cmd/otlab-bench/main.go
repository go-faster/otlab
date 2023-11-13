package main

import (
	"context"
	"flag"
	"io"
	"math/rand"
	"slices"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/go-faster/city"
	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/app"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/go-faster/otlab/internal/model"
)

func main() {
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
		// Fill resources table.
		var arg struct {
			ResourceCount int
		}
		flag.IntVar(&arg.ResourceCount, "resource-count", 10_000_000, "resource count")
		flag.Parse()
		client, err := ch.Dial(ctx, ch.Options{})
		if err != nil {
			return err
		}
		res, err := app.Resource(ctx)
		if err != nil {
			return err
		}
		rnd := rand.New(rand.NewSource(0)) // #nosec G404
		newResource := func() []byte {
			attrs := slices.Clone(res.Attributes())
			attrs = append(attrs, attribute.Int("service.id", rnd.Int()))
			data, err := model.MarshalAttrSet(attribute.NewSet(attrs...))
			if err != nil {
				panic(err)
			}
			return data
		}
		var (
			id    proto.ColUInt128
			value proto.ColStr
		)
		input := proto.Input{
			{Name: "id", Data: &id},
			{Name: "value", Data: &value},
		}
		var total int
		fillBatch := func() {
			input.Reset()
			const size = 200_000
			for i := 0; i < size; i++ {
				v := newResource()
				h := city.Hash128(v)
				id.Append(proto.UInt128(h))
				value.AppendBytes(v)
			}
			total += size
		}
		fillBatch()
		if err := client.Do(ctx, ch.Query{
			Input: input,
			Body:  input.Into("resources"),
			OnInput: func(ctx context.Context) error {
				if total >= arg.ResourceCount {
					return io.EOF
				}
				fillBatch()
				return nil
			},
		}); err != nil {
			return errors.Wrap(err, "resources")
		}

		return nil
	})
}
