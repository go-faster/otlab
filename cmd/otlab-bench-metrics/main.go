package main

import (
	"context"
	"flag"
	"io"
	"math/rand"
	"slices"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/go-faster/city"
	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/app"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/otlab/internal/model"
	"github.com/go-faster/otlab/internal/speed"
)

func main() {
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
		var arg struct {
			Count   int
			Table   string
			Workers int
			Block   int
		}
		flag.IntVar(&arg.Count, "count", 1_000_000_000, "count")
		flag.StringVar(&arg.Table, "table", "resources", "table name")
		flag.IntVar(&arg.Workers, "j", 1, "workers")
		flag.IntVar(&arg.Block, "block", 100_000, "block size")
		flag.Parse()

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
		v := newResource()
		h := city.Hash128(v)
		now := time.Now()

		g, ctx := errgroup.WithContext(ctx)
		for j := 0; j < arg.Workers; j++ {
			g.Go(func() error {
				client, err := ch.Dial(ctx, ch.Options{})
				if err != nil {
					return err
				}
				var (
					id       proto.ColUInt128
					resource proto.ColUInt128
					attrs    proto.ColUInt128
					value    proto.ColFloat64
				)
				ts := new(proto.ColDateTime64).WithPrecision(proto.PrecisionNano)
				input := proto.Input{
					{Name: "id", Data: &id},
					{Name: "resource", Data: &resource},
					{Name: "attributes", Data: &attrs},
					{Name: "timestamp", Data: ts},
					{Name: "value", Data: &value},
				}
				spd := speed.Start(ctx, "inserts")

				var total int
				fillBatch := func() {
					input.Reset()
					for i := 0; i < arg.Block; i++ {
						id.Append(proto.UInt128(h))
						resource.Append(proto.UInt128(h))
						attrs.Append(proto.UInt128(h))
						ts.Append(now)
						value.Append(1.0)
					}
					spd.Inc(arg.Block)
					total += arg.Block
				}
				fillBatch()
				if err := client.Do(ctx, ch.Query{
					Input: input,
					Body:  input.Into(arg.Table),
					OnInput: func(ctx context.Context) error {
						if total >= arg.Count {
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

		return g.Wait()
	})
}
