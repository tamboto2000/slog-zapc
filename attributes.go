package slogzapc

import (
	"context"
	"log/slog"

	"go.uber.org/zap/zapcore"
)

type ReplaceAttrFunc func(groups []string, a slog.Attr) slog.Attr
type AttrsFromCtxFunc func(ctx context.Context) []slog.Attr

func addToGroups(groups []string, oldAttrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		return append(oldAttrs, newAttrs...)
	}

	for i, attr := range oldAttrs {
		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			gval := slog.GroupValue(addToGroups(groups[1:], attr.Value.Group(), newAttrs...)...)
			attr.Value = gval
			oldAttrs[i] = attr
			return oldAttrs
		}
	}

	gval := slog.GroupValue(addToGroups(groups[1:], []slog.Attr{}, newAttrs...)...)
	gattr := slog.Group(groups[0])
	gattr.Value = gval

	return append(oldAttrs, gattr)
}

func zapFields(ctx context.Context, groups []string, withAttrs []slog.Attr, replace ReplaceAttrFunc, fromCtx AttrsFromCtxFunc, rec slog.Record) []zapcore.Field {
	// TODO: if groups len is 0, no grouping logic
	var fields []zapcore.Field

	// var traverseGroups []string	
	var gattr slog.Attr
	var objEnc zapcore.ObjectEncoder

	if len(withAttrs) != 0 {
		gattr = withAttrs[0]
	}

	for _, g := range groups {
		if gattr.Key == "" {

		}
	}
	

	// TODO: replace and convert attrs from record
	// TODO: replace and convert attrs from

	return fields
}
