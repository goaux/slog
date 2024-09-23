package logger_test

import (
	"log/slog"

	"github.com/goaux/results"
	"github.com/goaux/slog/logger"
)

var log = results.Must1(logger.New())

func Example() {
	log.Info("guide", slog.Int("the meaning of life", 42))
	// Output:
}

func ExampleNewName() {
	log := results.Must1(logger.NewName("example"))
	log.Info("guide", slog.Int("the meaning of life", 42))
	// Output:
}
