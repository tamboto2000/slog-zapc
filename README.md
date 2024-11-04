# Go's `slog` handler implementation using Uber's `zap/zapcore`
Minimal `log/slog.Handler` implementation by using `go.uber.org/zap/zapcore.Core`

## Examples
### 1. Using defaults
```go
package main

import (
    "log/slog"

    slogzapc "github.com/tamboto2000/slog-zapc"
)

func main() {
    h := slogzapc.New(slogzapc.DefaultOptions, nil)
    logger := slog.New(h)
    logger.Info("Hello World!", slog.String("note", "I like apple"))
}
```

### 2. Using custom `zapcore.Core`
```go
package main

import (
    "log/slog"

    slogzapc "github.com/tamboto2000/slog-zapc"
)

func main() {
    

    h := slogzapc.New(slogzapc.DefaultOptions, nil)
    logger := slog.New(h)
    logger.Info("Hello World!", slog.String("note", "I like apple"))
}
```