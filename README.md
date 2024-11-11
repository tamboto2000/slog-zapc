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
	"os"

	slogzapc "github.com/tamboto2000/slog-zapc"
	"go.uber.org/zap/zapcore"
)

func main() {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:   slog.MessageKey,
		LevelKey:     slog.LevelKey,
		TimeKey:      slog.TimeKey,
		CallerKey:    slog.SourceKey,
		EncodeTime:   zapcore.RFC3339NanoTimeEncoder,
		EncodeCaller: zapcore.FullCallerEncoder,
	})

	// Custom zapcore.Core
	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(os.Stderr), zapcore.InfoLevel)

	// slog.Logger handler
	h := slogzapc.New(slogzapc.DefaultOptions, core)
	// slog.Logger
	logger := slog.New(h)

	logger.Info("Hello World!")
}
```

## License
MIT