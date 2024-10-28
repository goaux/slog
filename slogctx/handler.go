package slogctx

import (
	"context"
	"log/slog"
)

// Handler is a [slog.Handler] implementation that adds context attributes to log records.
// It wraps another [slog.Handler] and adds any attributes stored in the context
// to each log record before passing it to the wrapped handler.
//
// Handler automatically includes any attributes attached to the context via [With]()
// in the log records it processes. This allows for contextual logging where common
// attributes can be defined once at a higher level and automatically included in
// all subsequent logging calls.
type Handler struct {
	next slog.Handler
}

var _ slog.Handler = (*Handler)(nil)

// NewHandler creates a new [slog.Handler] that adds context attributes to log records.
// It wraps the provided next handler and enhances it with context awareness.
//
// The returned handler will automatically include any attributes attached to the
// context (via [With]) when processing log records. This allows for hierarchical
// logging where common attributes can be defined once and automatically included
// in all subsequent logging calls.
func NewHandler(next slog.Handler) *Handler {
	return &Handler{next: next}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	attrs := getAttrs(ctx)
	if len(attrs) == 0 {
		return h.next.Handle(ctx, r)
	}
	r2 := r.Clone()
	for i := len(attrs) - 1; i >= 0; i-- {
		r2.Add(attrs[i]...)
	}
	return h.next.Handle(Reset(ctx), r2)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{next: h.next.WithAttrs(attrs)}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{next: h.next.WithGroup(name)}
}
