package main

import (
	"context"
	"log/slog"

	"github.com/goaux/results"
	"github.com/goaux/slog/logger"
)

var log = results.Must1(logger.New())

func main() {
	ctx := context.TODO()
	for i := -9; i <= 13; i++ {
		level := slog.Level(i)
		log.Log(ctx, level, level.String(), slog.Int("i", i))
	}
}
