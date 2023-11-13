package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"slices"

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
		flag.IntVar(&arg.ResourceCount, "resource-count", 100_000_000, "resource count")
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
		var (
			id    proto.ColUInt128
			value proto.ColStr
			buf   proto.Buffer
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
		write := func() error {
			// Write new block to io.Writer.
			//
			// You can write multiple blocks in sequence.
			b := proto.Block{
				Rows:    id.Rows(),
				Columns: len(input),
			}
			// Note that we are using version 54451, proto.Version will fail.
			if err := b.EncodeRawBlock(&buf, 54451, input); err != nil {
				return errors.Wrap(err, "encode")
			}

			// Write buffer to output io.Writer. In out case, it is os.Stdout.
			if _, err := os.Stdout.Write(buf.Buf); err != nil {
				return errors.Wrap(err, "write")
			}

			return nil
		}

		for {
			if total >= arg.ResourceCount {
				break
			}
			fillBatch()
			if err := write(); err != nil {
				return errors.Wrap(err, "write")
			}
		}

		return nil
	})
}
