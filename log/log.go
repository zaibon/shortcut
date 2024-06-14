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

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	singletonLogger.Log(ctx, level, msg, args...)
}
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	singletonLogger.LogAttrs(ctx, level, msg, attrs...)
}
func Debug(msg string, args ...any) {
	singletonLogger.Debug(msg, args...)
}
func DebugContext(ctx context.Context, msg string, args ...any) {
	singletonLogger.DebugContext(ctx, msg, args...)
}
func Info(msg string, args ...any) {
	singletonLogger.Info(msg, args...)
}
func InfoContext(ctx context.Context, msg string, args ...any) {
	singletonLogger.InfoContext(ctx, msg, args...)
}
func Warn(msg string, args ...any) {
	singletonLogger.Warn(msg, args...)
}
func WarnContext(ctx context.Context, msg string, args ...any) {
	singletonLogger.WarnContext(ctx, msg, args...)
}
func Error(msg string, args ...any) {
	singletonLogger.Error(msg, args...)
}
func ErrorContext(ctx context.Context, msg string, args ...any) {
	singletonLogger.ErrorContext(ctx, msg, args...)
}
func Fatal(v ...any) {
	singletonLogger.Error(fmt.Sprint(v...))
	os.Exit(1)
}
func Fatalf(format string, args ...any) {
	singletonLogger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}
