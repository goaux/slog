package main

import (
	"context"
	"log/slog"

	"github.com/goaux/results"
	"github.com/goaux/slog/logger"
	"github.com/goaux/slog/slogctx"
)

var log = results.Must1(logger.New())

func main() {
	ctx := slogctx.With(context.TODO(), slog.String("hello", "world"))
	for i := -9; i <= 13; i++ {
		level := slog.Level(i)
		log.Log(ctx, level, level.String(), slog.Int("i", i))
	}
}
