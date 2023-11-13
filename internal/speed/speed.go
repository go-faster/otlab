// Package speed implements speed counter.
package speed

import (
	"context"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// Start a new speed counter.
func Start(ctx context.Context, name string) *Speed {
	v := &Speed{name: name}
	v.watch(zctx.From(ctx))
	return v
}

// Speed counter.
type Speed struct {
	name    string
	counter atomic.Uint64
}

// Inc increments the speed counter by v.
func (s *Speed) Inc(v uint64) {
	s.counter.Add(v)
}

func (s *Speed) watch(lg *zap.Logger) {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			v := s.counter.Swap(0)
			lg.Info("Speed",
				zap.String("name", s.name),
				zap.String("speed", humanize.SI(float64(v), "")),
			)
		}
	}()
}
