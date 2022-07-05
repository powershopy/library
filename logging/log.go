package logging

import (
	"context"
	"github.com/powershopy/library/utils"
	"github.com/sirupsen/logrus"
	"os"
)

var log *logrus.Logger

type LoggerConfig struct {
	//LogType  string
	LogLevel string
	Pretty   bool
	//LogFile  string
}

func init() {
	log = logrus.StandardLogger()
}

func Config(config LoggerConfig) *logrus.Logger {
	//pretty := true
	//if setting.LogType == "file" {
	//	pretty = false
	//}
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   utils.TimeFormat,
		PrettyPrint:       config.Pretty,
		DisableHTMLEscape: true,
	})
	//if config.LogType == "file" {
	//	log.SetOutput(&lumberjack.Logger{
	//		Filename:  config.LogFile,
	//		MaxAge:    28,
	//		LocalTime: true,
	//		MaxSize:   500,
	//		Compress:  true,
	//	})
	//} else {
	log.SetOutput(os.Stdout)
	//}
	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)
	return log
}

func WithField(key string, value interface{}) *Entry {
	return &Entry{log.WithField(key, value)}
}

func WithFields(fields logrus.Fields) *Entry {
	return &Entry{log.WithFields(fields)}
}

// Debug logs a message at level Debug on the standard logger.
func Debug(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(ctx context.Context, args ...interface{}) {
	withStack(WithFields(utils.GetTraceLogEntryFromContext(ctx))).Error(ctx, args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(ctx context.Context, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(ctx context.Context, format string, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(ctx context.Context, format string, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(ctx context.Context, format string, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	withStack(WithFields(utils.GetTraceLogEntryFromContext(ctx))).Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(ctx context.Context, format string, args ...interface{}) {
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	traceInfo := utils.GetTraceInfoFromCtx(ctx)
	log.WithFields(map[string]interface{}{
		"trace_id": traceInfo.TraceID,
	})
	log.WithFields(utils.GetTraceLogEntryFromContext(ctx)).Fatalf(format, args...)
}
