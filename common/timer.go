package common

import (
	"context"
	"time"

	"github.com/nyttikord/gokord/bot"
)

func NewTimer(ctx context.Context, d time.Duration, fn func(context.Context, context.CancelFunc)) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	ctx = bot.SetLogger(ctx, bot.Logger(ctx).With("module", "timer"))
	go func(ctx context.Context, d time.Duration) {
		ticker := time.NewTicker(d)
		fn(ctx, cancel)
		for {
			select {
			case <-ticker.C:
				fn(ctx, cancel)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}(ctx, d)
	return cancel
}
