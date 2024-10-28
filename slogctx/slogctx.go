// Package slogctx provides context-aware attribute management for the standard [log/slog] package.
// It allows logging attributes to be attached to a [context.Context] and automatically included
// in log records when using the provided [Handler]. This enables hierarchical and contextual
// logging patterns where common attributes can be defined once and automatically included
// in all subsequent log entries within that context scope.
//
// The package is particularly useful in middleware and request processing scenarios where
// you want to attach common attributes (like request ID, user ID, etc.) at a higher level
// and have them automatically included in all logging calls further down the call stack.
package slogctx

import (
	"context"
	"log/slog"
	"time"
)

// With adds the given key-value pairs to the context as logging attributes.
// It returns a new context that includes the provided attributes in addition
// to any attributes from the parent context.
//
// With returns the new context that the result of calling [Attrs](parent, args...) is attached to.
// If len(args) is 0, then it returns parent as is.
//
// parent must not be nil, otherwise it panics.
//
// The attribute arguments follow the same rules as [slog.Logger.Log]:
//
//   - If an argument is an [slog.Attr], it is used as is.
//   - If an argument is a string and this is not the last argument, the following argument is treated as the value and the two are combined into an [slog.Attr].
//   - Otherwise, the argument is treated as a value with key "!BADKEY".
func With(parent context.Context, args ...any) context.Context {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	if len(args) == 0 {
		return parent
	}
	return context.WithValue(
		parent,
		withArgsKey{},
		&withArgs{attrs: appendAttrs(getAttrs(parent), args)},
	)
}

func getAttrs(ctx context.Context) [][]any {
	if v, ok := ctx.Value(withArgsKey{}).(*withArgs); ok {
		return v.attrs
	}
	return nil
}

func appendAttrs(parent [][]any, args []any) [][]any {
	n := len(parent)
	a := make([][]any, n+1)
	copy(a, parent)
	a[n] = args
	return a
}

type withArgsKey struct{}

type withArgs struct {
	attrs [][]any
}

// Attrs returns a slice containing the provided args followed by any attributes
// attached to the context. This is useful when you need to explicitly access
// the logging attributes stored in a context.
//
// The returned slice can be used directly with [slog.Logger] methods or as
// arguments to [slog.Logger.With] to create a new logger with combined attributes.
func Attrs(ctx context.Context, args ...any) []any {
	if v, ok := ctx.Value(withArgsKey{}).(*withArgs); ok {
		return argsToAttrs(v.attrs, args)
	} else if len(args) > 0 {
		return argsToAttrs(nil, args)
	}
	return nil
}

func argsToAttrs(list [][]any, args []any) []any {
	r := slog.NewRecord(time.Time{}, 0, "", 0)
	r.Add(args...)
	for i := len(list) - 1; i >= 0; i-- {
		r.Add(list[i]...)
	}
	var attrs []any
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	return attrs
}

// Reset returns a new context derived from parent that only includes the
// provided args as logging attributes, hiding any attributes from the parent
// context. This is useful when you want to start fresh with a new set of
// attributes while maintaining the parent context's other values.
func Reset(parent context.Context, args ...any) context.Context {
	return context.WithValue(parent, withArgsKey{}, &withArgs{attrs: toAttrs(args)})
}

func toAttrs(args []any) [][]any {
	if len(args) > 0 {
		return [][]any{args}
	}
	return nil
}
