# slog

This repository provides helper modules and packages for using [log/slog](https://pkg.go.dev/log/slog) effectively.

## github.com/goaux/slog/logger

Package logger is a placeholder package for creating and using a project-specific [slog.Logger][] across all modules in a program.

This package provides only two functions and will maintain this minimal API in the future:

- [New][]() (*[slog.Logger][], error)
- [NewName][](name string) (*[slog.Logger][], error)

Note that [New][] internally calls [NewName][], effectively providing a single core function.

[NewName][] returns a [slog.Logger][] created based on the value of an environment variable.
However, this functionality is specific to this package and may not be suitable for all programs.

To customize the logger for your project:

1. Create a new module with a custom `logger` package.
2. Implement your own `New` and `NewName` functions that return a project-specific [slog.Logger][].
3. Replace the logger module in the package containing main:

```
go mod edit -replace github.com/goaux/slog/logger=<project-specific-logger>
```

This replacement ensures that all modules using the github.com/goaux/slog/logger package
will use your project-specific logger instead of this placeholder package.

[slog.Logger]: https://pkg.go.dev/log/slog#Logger
[New]: https://pkg.go.dev/github.com/goaux/slog/logger#New
[NewName]: https://pkg.go.dev/github.com/goaux/slog/logger#NewName
