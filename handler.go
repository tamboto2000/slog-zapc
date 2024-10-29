// Package slogzapc implements [log/slog.Handler] by using [go.uber.org/zap/zapcore.Core] implementation
// to handle log records.
package slogzapc

import (
	"context"
	"log/slog"
	"os"
	"runtime"

	"go.uber.org/zap/zapcore"
)

func upperCaseLvlEnc(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(level.CapitalString())
}

func defaultCore() zapcore.Core {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		EncodeLevel:  upperCaseLvlEnc,
		MessageKey:   slog.MessageKey,
		LevelKey:     slog.LevelKey,
		TimeKey:      slog.TimeKey,
		CallerKey:    slog.SourceKey,
		EncodeTime:   zapcore.RFC3339NanoTimeEncoder,
		EncodeCaller: zapcore.FullCallerEncoder,
	}), zapcore.AddSync(os.Stderr), zapcore.InfoLevel)

	return core
}

func logLevel(lvl slog.Level) zapcore.Level {
	switch lvl {
	case slog.LevelDebug:
		return zapcore.DebugLevel
	case slog.LevelInfo:
		return zapcore.InfoLevel
	case slog.LevelWarn:
		return zapcore.WarnLevel
	case slog.LevelError:
		return zapcore.ErrorLevel
	}

	return zapcore.InvalidLevel
}

func noopReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	return a
}

type HandlerOptions struct {
	// AddSource causes the handler to compute the source code position
	// of the log statement and add a SourceKey attribute to the output.
	AddSource bool

	// Level reports the minimum record level that will be logged.
	// The handler discards records with lower levels.
	// If Level is nil, the handler assumes LevelInfo.
	// The handler calls Level.Level for each record processed;
	// to adjust the minimum level dynamically, use a LevelVar.
	Level slog.Leveler

	// ReplaceAttr is called to rewrite each non-group attribute before it is logged.
	// The attribute's value has been resolved (see [Value.Resolve]).
	// If ReplaceAttr returns a zero Attr, the attribute is discarded.
	//
	// Normally, the built-in attributes with keys such as "time", "level", "source", and "msg"
	// are passed to this function and can be replaced, but since we are
	// using Zap's core implementation, the built-in keys can not be passed
	// through this function. If you want to change the built-in keys, you
	// must do it on handler creation by passing [go.uber.org/zap/zapcore.Core] implementation
	// with custom encoder config to New function
	ReplaceAttr ReplaceAttrFunc

	AttrFromCtx AttrsFromCtxFunc
}

// Handler implements [log/slog.Handler] interface.
type Handler struct {
	opts   HandlerOptions
	zcore  zapcore.Core
	attrs  []slog.Attr
	groups []string
}

// New creates [log/slog.Handler] implementation by using core
// as the logging backend.
// If core is nil, default zap core will be used
func New(opts HandlerOptions, core zapcore.Core) *Handler {
	if core == nil {
		core = defaultCore()
	}

	if opts.ReplaceAttr == nil {
		opts.ReplaceAttr = noopReplaceAttr
	}

	return &Handler{
		opts:  opts,
		zcore: core,
	}
}

func (h *Handler) Enabled(ctx context.Context, lvl slog.Level) bool {
	minLevel := slog.LevelDebug
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return lvl >= minLevel
}

func (h *Handler) Handle(ctx context.Context, rec slog.Record) error {
	ent := zapcore.Entry{
		Level:   logLevel(rec.Level),
		Time:    rec.Time,
		Message: rec.Message,
	}

	if h.opts.AddSource {
		frame, _ := runtime.CallersFrames([]uintptr{rec.PC}).Next()
		ent.Caller = zapcore.NewEntryCaller(0, frame.File, frame.Line, true)
	}

	ec := h.zcore.Check(ent, nil)
	if ec == nil {
		return nil
	}

	// TODO: might be a good addition
	// if we can have options or settings
	// on the sync behavior
	defer h.zcore.Sync()

	ec.Write(zapFields(ctx, h.groups, h.attrs, h.opts.ReplaceAttr, h.opts.AttrFromCtx, rec)...)

	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		opts:   h.opts,
		zcore:  h.zcore,
		attrs:  addToGroups(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	newH := &Handler{
		opts:   h.opts,
		zcore:  h.zcore,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}

	if len(h.groups) == 0 {
		newH.attrs = addToGroups(newH.groups, []slog.Attr{}, h.attrs...)
	}

	return newH
}
