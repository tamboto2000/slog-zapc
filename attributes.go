package slogzapc

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ReplaceAttrFunc func(groups []string, a slog.Attr) slog.Attr
type AttrsFromCtxFunc func(ctx context.Context) []slog.Attr

type attrObjEncoder struct {
	attr slog.Attr
}

func (ae attrObjEncoder) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range ae.attr.Value.Group() {
	ATTR_VAL_SWITCH:
		switch attr.Value.Kind() {
		case slog.KindAny:
			zany := zap.Any(attr.Key, attr.Value.Any())
			zany.AddTo(enc)
		case slog.KindBool:
			enc.AddBool(attr.Key, attr.Value.Bool())
		case slog.KindDuration:
			enc.AddDuration(attr.Key, attr.Value.Duration())
		case slog.KindFloat64:
			enc.AddFloat64(attr.Key, attr.Value.Float64())
		case slog.KindInt64:
			enc.AddInt64(attr.Key, attr.Value.Int64())
		case slog.KindString:
			enc.AddString(attr.Key, attr.Value.String())
		case slog.KindTime:
			enc.AddTime(attr.Key, attr.Value.Time())
		case slog.KindUint64:
			enc.AddUint64(attr.Key, attr.Value.Uint64())
		case slog.KindGroup:
			enc.AddObject(attr.Key, attrObjEncoder{attr: attr})
		case slog.KindLogValuer:
			av := attr.Value.LogValuer().LogValue()
			attr = slog.Attr{
				Key:   attr.Key,
				Value: av,
			}

			goto ATTR_VAL_SWITCH
		}
	}

	return nil
}

func attrToZapField(attr slog.Attr) zapcore.Field {
	switch attr.Value.Kind() {
	case slog.KindAny:
		return zap.Any(attr.Key, attr.Value.Any())
	case slog.KindBool:
		return zap.Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return zap.Duration(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return zap.Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return zap.Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return zap.String(attr.Key, attr.Value.String())
	case slog.KindTime:
		return zap.Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return zap.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		return zap.Object(attr.Key, attrObjEncoder{attr: attr})
	case slog.KindLogValuer:
		newAttr := slog.Attr{
			Key:   attr.Key,
			Value: attr.Value.LogValuer().LogValue(),
		}

		return attrToZapField(newAttr)
	default:
		return zap.Any(attr.Key, attr.Value.Any())
	}
}

func addToGroups(groups []string, withAttrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		return append(withAttrs, newAttrs...)
	}

	for i, attr := range withAttrs {
		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			gval := slog.GroupValue(addToGroups(groups[1:], attr.Value.Group(), newAttrs...)...)
			attr.Value = gval
			withAttrs[i] = attr
			return withAttrs
		}
	}

	gval := slog.GroupValue(addToGroups(groups[1:], []slog.Attr{}, newAttrs...)...)
	gattr := slog.Group(groups[0])
	gattr.Value = gval

	return append(withAttrs, gattr)
}

func zapFields(ctx context.Context, groups []string, withAttrs []slog.Attr, replace ReplaceAttrFunc, fromCtx AttrsFromCtxFunc, rec slog.Record) []zapcore.Field {
	var fields []zapcore.Field

	var traverseGroups []string
	groupCount := len(groups)
	if groupCount != 0 {
		for i, g := range groups {
			traverseGroups = append(traverseGroups, g)
			fields = append(fields, zap.Namespace(g))

			var gattr slog.Attr
			for _, wa := range withAttrs {
				if wa.Key == g && wa.Value.Kind() == slog.KindGroup {
					gattr = wa
					break
				}
			}

			if gattr.Key == "" {
				continue
			}

			var nextGroup string
			if i+1 < groupCount {
				nextGroup = groups[i+1]
			}

			for _, attr := range gattr.Value.Group() {
				if ignoreAttr(attr) {
					continue
				}

				if nextGroup != "" && attr.Key == nextGroup && attr.Value.Kind() == slog.KindGroup {
					withAttrs[0] = attr
					continue
				}

				attr = replace(traverseGroups, attr)
				if ignoreAttr(attr) {
					continue
				}

				fields = append(fields, attrToZapField(attr))
			}
		}

		withAttrs = []slog.Attr{}
	}

	for _, attr := range withAttrs {
		if ignoreAttr(attr) {
			continue
		}

		attr = replace(traverseGroups, attr)

		if ignoreAttr(attr) {
			continue
		}

		fields = append(fields, attrToZapField(attr))
	}

	rec.Attrs(func(a slog.Attr) bool {
		if ignoreAttr(a) {
			return true
		}

		a = replace(traverseGroups, a)

		if ignoreAttr(a) {
			return true
		}

		fields = append(fields, attrToZapField(a))

		return true
	})

	if fromCtx != nil {
		for _, attr := range fromCtx(ctx) {
			if ignoreAttr(attr) {
				continue
			}

			attr = replace(traverseGroups, attr)

			if ignoreAttr(attr) {
				continue
			}

			fields = append(fields, attrToZapField(attr))
		}
	}

	return fields
}

func ignoreAttr(a slog.Attr) bool {
	if a.Key == "" || a.Equal(slog.Attr{}) {
		return true
	}

	return false
}
