package slogctx_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/goaux/slog/slogctx"
)

func Example() {
	ctx := context.Background()

	// Adding logging attributes to the context.
	ctx = slogctx.With(ctx, "user", "alice", slog.Int("age", 42))
	// Adding another logging attributes to the context.
	ctx = slogctx.With(ctx, "state", "good")

	// To emit automatically logging attributes attached to the context, use slogctx.NewHandler.
	logger := slog.New(
		slogctx.NewHandler(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: removeTime}),
		),
	)
	// Log with context - attributes will be automatically included
	logger.InfoContext(ctx, "User logged in", "count", 7)
	// Output:
	// level=INFO msg="User logged in" count=7 state=good user=alice age=42
}

func ExampleWith() {
	ctx := context.Background()

	// Adding logging attributes to the context.
	ctx = slogctx.With(ctx, "a", "A", "b", "B")
	// Adding another logging attributes to the context.
	ctx = slogctx.With(ctx, "c", "C", slog.String("d", "D"))
	// Adding yet another logging attributes to the context.
	ctx = slogctx.With(ctx, slog.String("e", "E"), slog.Int("f", 7))

	// If you do not use slogctx.NewHandler for the logger,
	// Get attributes from context explicitly.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: removeTime}))
	logger.Info("User logged in", slogctx.Attrs(ctx, "g", "G")...)

	// Log with context - attributes will be automatically included
	logger2 := slog.New(slogctx.NewHandler(logger.Handler()))
	logger2.InfoContext(ctx, "User logged in", "g", "G")
	// Output:
	// level=INFO msg="User logged in" g=G e=E f=7 c=C d=D a=A b=B
	// level=INFO msg="User logged in" g=G e=E f=7 c=C d=D a=A b=B
}

func ExampleAttrs() {
	ctx := context.Background()

	// Adding logging attributes to the context.
	ctx = slogctx.With(ctx, "user", "alice", slog.Int("age", 42))
	// Adding another logging attributes to the context.
	ctx = slogctx.With(ctx, "state", "good")

	// If you do not use slogctx.NewHandler for the logger,
	// Get attributes from context explicitly.
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: removeTime}),
	)
	logger.InfoContext(ctx, "User logged in", slogctx.Attrs(ctx, "count", 7)...)
	// Output:
	// level=INFO msg="User logged in" count=7 state=good user=alice age=42
}

func ExampleReset() {
	ctx := context.Background()

	// Adding logging attributes to the context.
	ctx = slogctx.With(ctx, "user", "alice", slog.Int("age", 42))
	// Adding another logging attributes to the context.
	ctx = slogctx.With(ctx, "state", "good")

	logger := slog.New(
		slogctx.NewHandler(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: removeTime}),
		),
	)

	// When you use slog.Logger.WithGroup, the attributes you attach to the
	// context are output within a group.
	logger.
		WithGroup("GROUP").
		InfoContext(ctx, "User logged in", "count", 7)

	// If it's not what you expected, you can use Attrs and Reset to change the result.
	logger.
		With(slogctx.Attrs(ctx)...).
		WithGroup("GROUP").
		InfoContext(slogctx.Reset(ctx), "User logged in", "count", 7)
	// Output:
	// {"level":"INFO","msg":"User logged in","GROUP":{"count":7,"state":"good","user":"alice","age":42}}
	// {"level":"INFO","msg":"User logged in","state":"good","user":"alice","age":42,"GROUP":{"count":7}}
}

func removeTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

func TestAttrs(t *testing.T) {
	t.Run("", func(t *testing.T) {
		ctx := context.TODO()
		a0 := slogctx.Attrs(ctx)
		if len(a0) != 0 {
			t.Error()
		}
		a1 := slogctx.Attrs(ctx, "a", "A")
		if len(a1) != 1 {
			t.Error(len(a1))
		}
		ctx = slogctx.With(ctx, "b", "B")
	})

	t.Run("", func(t *testing.T) {
		ctx := context.TODO()
		a0 := slogctx.Attrs(ctx)
		if len(a0) != 0 {
			t.Error("must be 0")
		}
		ctx = slogctx.With(ctx, "a", "A")
		a1 := slogctx.Attrs(ctx)
		if len(a1) != 1 {
			t.Error("must be 1")
		}
		ctx = slogctx.With(ctx, "b", "B")
		a2 := slogctx.Attrs(ctx, "c", "C")
		if len(a2) != 3 {
			t.Error("must be 3")
		}
	})

	t.Run("", func(t *testing.T) {
		ctx := context.TODO()
		a0 := slogctx.Attrs(ctx, "a", "A")
		if len(a0) != 1 {
			t.Error("must be 1")
		}
		ctx = slogctx.With(ctx, "b", "B")
		a2 := slogctx.Attrs(ctx, "c", "C")
		if len(a2) != 2 {
			t.Error("must be 2")
		}
	})
}

func TestReset(t *testing.T) {
	ctx := context.TODO()
	ctx = slogctx.With(ctx, "a", "A")
	a0 := slogctx.Attrs(ctx)
	if len(a0) != 1 {
		t.Error()
	}
	ctx = slogctx.Reset(ctx)
	a1 := slogctx.Attrs(ctx)
	if len(a1) != 0 {
		t.Error()
	}
	ctx = slogctx.Reset(ctx, "a", "A")
	a2 := slogctx.Attrs(ctx)
	if len(a2) != 1 {
		t.Error()
	}
}

func TestWith(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		ctx := context.TODO()
		ctx0 := slogctx.With(ctx)
		a0 := slogctx.Attrs(ctx0)
		if len(a0) != 0 {
			t.Error()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			if s := recover(); s == nil {
				t.Error()
			} else if s != "cannot create context from nil parent" {
				t.Error()
			}
		}()
		slogctx.With(nil)
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			if s := recover(); s == nil {
				t.Error()
			} else if s != "cannot create context from nil parent" {
				t.Error()
			}
		}()
		slogctx.With(nil, "a", "A")
	})
}
