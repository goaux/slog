# slogctx

slogctx is a Go package that extends the standard `log/slog` package with context-aware attribute management. It allows you to attach logging attributes to a `context.Context` and have them automatically included in log records.

## Features

- Attach logging attributes to context.Context
- Automatically include context attributes in log records
- Compatible with standard log/slog handlers
- Zero external dependencies
- Clean and simple API

Compared to the other similar slogctx packages, this package has the following differences:

- Attributes added later are considered more specific and are output first
- Attributes in a context is resettable and has affinity with slog.Logger.WithGroup
- Append-like attribute getter

## Usage

### Basic Usage

```go
// Create a context with logging attributes
ctx := context.Background()
ctx = slogctx.With(ctx, "user", "alice", slog.Int("age", 42))
ctx = slogctx.With(ctx, "requestID", "req-123")

// Create a logger with the context-aware handler
logger := slog.New(slogctx.NewHandler(slog.Default().Handler()))

// Log with context - attributes will be automatically included
logger.InfoContext(ctx, "User action performed", "action", "login")
// Output: level=INFO msg="User action performed" action=login user=alice age=42 requestID=req-123
```

### Manual Attribute Access

```go
// Get attributes from context explicitly
logger.Info("User action performed", slogctx.Attrs(ctx, "action", "login")...)
```

### Reset Context Attributes

```go
// Create new context without inherited attributes
newCtx := slogctx.Reset(ctx, "fresh", "start")
```

## Best Practices

1. Use `With` to attach common attributes at middleware or high-level handlers
2. Pass the context through your application's call chain
3. Use `NewHandler` to automatically include context attributes in logs
4. Use `Reset` when you need to start fresh with new attributes
5. Use `Attrs` when you need manual control over attribute inclusion