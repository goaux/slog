// Package logger is a placeholder package for creating and using a project-specific
// [slog.Logger] across all modules in a program.
//
// This package provides only two functions and will maintain this minimal API in the future:
//
//   - [New]() (*[slog.Logger], error)
//   - [NewName](name string) (*[slog.Logger], error)
//
// Note that [New] internally calls [NewName], effectively providing a single core function.
//
// [NewName] returns a [slog.Logger] created based on the value of an environment variable.
// However, this functionality is specific to this package and may not be suitable for all programs.
//
// To customize the logger for your project:
//
//  1. Create a new module with a custom `logger` package.
//
//  2. Implement your own `New` and `NewName` functions that return a project-specific [slog.Logger].
//
//  3. Replace the logger module in the package containing main:
//
// Example:
//
//	go mod edit -replace github.com/goaux/slog/logger=<project-specific-logger>
//
// This replacement ensures that all modules using the [github.com/goaux/slog/logger] package
// will use your project-specific logger instead of this placeholder package.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/goaux/funcname"
	"github.com/goaux/slog/slogctx"
	"github.com/goaux/stacktrace/v2"
)

// New returns the result of calling [NewName](<package-name-of-caller>).
func New() (*slog.Logger, error) {
	pkgname, _ := funcname.SplitCaller(1)
	return newName(pkgname)
}

// NewName returns a [slog.Logger] configured with the environment variable.
// If name is not empty, NewName returns the result of calling [slog.Logger.With]([slog.String]("logger", name))
//
// The handler of the logger returned includes [slogctx.Handler] that automatically emits logging attribute attached to the context.
//
// # Environment Variable
//
// NewName uses the environment variable `SLOG_LOGGER`.
//
// If the environment variable is not defined or is an empty string,
// "json?output=stderr&level=error&addSource=true" is used as the default value.
//
// Format:
//
//	<type>?output=<output>&level=<level>&addSource=<bool>
//
// The value of the environment variable is parsed using [url.Parse].
// <type> is derived from [url.URL.Path], and parameters are obtained from [url.URL.Query] and [url.Values.Get].
//
// Type:
//
//   - "json": calls [slog.NewJSONHandler]
//   - "text": calls [slog.NewTextHandler]
//   - "discard": creates a [slog.Logger] that discards all logs
//
// Default is json.
//
// Output:
//
//   - "stderr" or "err": os.Stderr
//   - "stdout" or "out": os.Stdout
//   - number: os.NewFile(number)
//
// Default is stderr.
//
// Level:
//
// Level values are parsed using [slog.Level.UnmarshalText].
//
// It accepts any string produced by [slog.Level.MarshalText], ignoring case. It also
// accepts numeric offsets that would result in a different string on output.
// For example, "Error-8" would marshal as "INFO".
//
// Default is INFO.
//
// AddSource:
//
// A boolean value, parsed using [strconv.ParseBool].
//
// The level and addSource are used to create a [slog.HandlerOptions] which is
// passed to [slog.NewJSONHandler] and [slog.NewTextHandler].
//
// Default is false.
//
// # Customization
//
// The name of the environment variable can be changed at compile time.
//
// Example:
//
//	go build -ldflags "-X github.com/goaux/slog/logger.envKey=LOGGER" ...
//
// The label of slog.Attr for the `name` parameter can be changed at compile time.
//
// Example:
//
//	go build -ldflags "-X github.com/goaux/slog/logger.nameKey=LOGGER" ...
//
// The default logger used when the environment variable is missing can be changed at compile time.
//
// Example:
//
//	go build -ldflags "-X github.com/goaux/slog/logger.defaultLogger=text?level=debug" ...
func NewName(name string) (*slog.Logger, error) {
	return newName(name)
}

// envKey is the name of the environment variable used by [NewName].
//
// It is defined as a variable, allowing it to be modified at compile time using -ldflags.
var envKey = "SLOG_LOGGER"

// nameKey is the label for the 'name' parameter of [NewName].
//
// It is defined as a variable, allowing it to be modified at compile time using -ldflags.
var nameKey = "logger"

// defaultLogger is used when the environment variable is missing can be changed at compile time.
//
// It is defined as a variable, allowing it to be modified at compile time using -ldflags.
var defaultLogger = "json?output=stderr&level=error&addSource=true"

// newName is a function that provides the functionality of [NewName].
//
// It is created to ensure consistent stack trace depth between [New] and [NewName],
// as there may be cases where stacktrace is called.
// Functionally, [New] could directly execute [NewName], but by having both call newName,
// we ensure the same stack depth regardless of whether [New] or [NewName] is called.
func newName(name string) (*slog.Logger, error) {
	log, err := newRootOnce()
	if err != nil {
		return nil, stacktrace.NewError(err, stacktrace.Callers(2))
	}
	if name != "" {
		log = log.With(slog.String(nameKey, name))
	}
	return log, nil
}

// newRootOnce is a function that ensures the process of parsing the environment variable
// and creating a [slog.Logger] is executed at most once.
var newRootOnce = sync.OnceValues(newRoot)

// newRoot parses the environment variable and returns a [slog.Logger] created based on the parsed information.
func newRoot() (*slog.Logger, error) {
	s := os.Getenv(envKey)
	if s == "" {
		s = defaultLogger
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	name, values := u.Path, u.Query()

	switch name {
	case "":
		name = "json"
	case "json", "text":
		// ok. go ahead.
	case "discard":
		return slog.New(discardHandler{}), nil
	case "default":
		return slog.New(slogctx.NewHandler(slog.Default().Handler())), nil
	default:
		return nil, fmt.Errorf("unknown logger=`%s`", name)
	}

	output, err := getOutput(values)
	if err != nil {
		return nil, err
	}

	options, err := newHandlerOptions(values)
	if err != nil {
		return nil, err
	}

	switch name {
	case "json":
		return slog.New(slogctx.NewHandler(slog.NewJSONHandler(output, options))), nil
	case "text":
		return slog.New(slogctx.NewHandler(slog.NewTextHandler(output, options))), nil
	}
	return nil, fmt.Errorf("unknown logger=`%s`", name)
}

func getOutput(values url.Values) (io.Writer, error) {
	s := values.Get("output")
	switch s {
	case "stderr", "err", "":
		return os.Stderr, nil
	case "stdout", "out":
		return os.Stdout, nil
	case "discard":
		return io.Discard, nil
	}
	if fd, err := strconv.Atoi(s); err == nil {
		return os.NewFile(uintptr(fd), "&"+s), nil
	}
	return nil, fmt.Errorf("unknown output=`%s`, must be a file descriptor or one of `stdout`, `stderr` or `discard`", s)
}

func newHandlerOptions(values url.Values) (*slog.HandlerOptions, error) {
	var addSource bool
	var level slog.Level
	if s := values.Get("addSource"); s != "" {
		if v, err := strconv.ParseBool(s); err != nil {
			return nil, fmt.Errorf("invalid addSource=`%s`, must be parsed as a boolean", s)
		} else {
			addSource = v
		}
	}
	if s := values.Get("level"); s != "" {
		s = strings.ReplaceAll(s, " ", "+")
		if err := level.UnmarshalText([]byte(s)); err != nil {
			return nil, fmt.Errorf("invalid level=`%s`, e.g. `debug`, `warn`, `info` or `error`", s)
		}
	}
	options := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
	}
	return options, nil
}
