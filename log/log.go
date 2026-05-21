package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

var singletonLogger *slog.Logger

func SetupLogger(logger *slog.Logger) {
	if singletonLogger == nil {
		singletonLogger = logger
	}
}

func getLogger() *slog.Logger {
	if singletonLogger == nil {
		singletonLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return singletonLogger
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	getLogger().Log(ctx, level, msg, args...)
}
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	getLogger().LogAttrs(ctx, level, msg, attrs...)
}
func Debug(msg string, args ...any) {
	getLogger().Debug(msg, args...)
}
func DebugContext(ctx context.Context, msg string, args ...any) {
	getLogger().DebugContext(ctx, msg, args...)
}
func Info(msg string, args ...any) {
	getLogger().Info(msg, args...)
}
func InfoContext(ctx context.Context, msg string, args ...any) {
	getLogger().InfoContext(ctx, msg, args...)
}
func Warn(msg string, args ...any) {
	getLogger().Warn(msg, args...)
}
func WarnContext(ctx context.Context, msg string, args ...any) {
	getLogger().WarnContext(ctx, msg, args...)
}
func Error(msg string, args ...any) {
	getLogger().Error(msg, args...)
}
func ErrorContext(ctx context.Context, msg string, args ...any) {
	getLogger().ErrorContext(ctx, msg, args...)
}
func Fatal(v ...any) {
	getLogger().Error(fmt.Sprint(v...))
	os.Exit(1)
}
func Fatalf(format string, args ...any) {
	getLogger().Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}
