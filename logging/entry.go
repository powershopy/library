package logging

import (
	"context"
	"github.com/powershopy/library/utils"
	"github.com/sirupsen/logrus"
)

type Entry struct {
	*logrus.Entry
}

func (e *Entry) withStack() *Entry {
	return withStack(e)
}

func (e *Entry) WithFields(fields logrus.Fields) *Entry {
	return &Entry{e.Entry.WithFields(fields)}
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	return &Entry{e.Entry.WithField(key, value)}
}

func (e *Entry) Debug(ctx context.Context, args ...interface{}) {
	e.Entry.WithFields(utils.GetCommonMetaFromCtx(ctx)).Debug(args)
}

func (e *Entry) Info(ctx context.Context, args ...interface{}) {
	e.Entry.WithFields(utils.GetCommonMetaFromCtx(ctx)).Info(args...)
}

func (e *Entry) Warn(ctx context.Context, args ...interface{}) {
	e.Entry.WithFields(utils.GetCommonMetaFromCtx(ctx)).Warn(args...)
}

func (e *Entry) Warning(ctx context.Context, args ...interface{}) {
	e.Entry.WithFields(utils.GetCommonMetaFromCtx(ctx)).Warning(args...)
}

func (e *Entry) Error(ctx context.Context, args ...interface{}) {
	e.withStack().Entry.WithFields(utils.GetCommonMetaFromCtx(ctx)).Error(args...)
}
