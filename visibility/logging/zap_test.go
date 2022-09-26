package logging

import (
    "github.com/kami-zh/go-capturer"
    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"
    "strings"
    "testing"
)

func TestPrettyStacks(t *testing.T) {
    out := capturer.CaptureStderr(func() {
        devLogger := ConfigureDevLogger()
        stack := NewShortenedStackTrace(2, false, "")
        devLogger.Error("this is bad", stack.Field())
    })

    // Check that we got the stack back; the line number is the line of
    // NewShortenedStack, might change during refactoring
    assert.True(t, strings.Contains(out, "logging/zap_test.go:14"))
}

func TestPrettyStacksStr(t *testing.T) {
    out := capturer.CaptureStderr(func() {
        devLogger := ConfigureDevLogger()
        stack := NewShortenedStackTrace(2, false, "")
        devLogger.Error("this is bad",
            zap.String("stacktrace", stack.StringStack()), zap.Int64("haha", 123))
    })

    // Check that we got the stack back; the line number is the line of
    // NewShortenedStack, might change during refactoring
    assert.True(t, strings.Contains(out, "zap_test.go:26 TestPrettyStacksStr"))
}

func TestFieldOverride(t *testing.T) {
    out := capturer.CaptureStderr(func() {
        devLogger := ConfigureDevLogger()
        devLogger = devLogger.With(zap.String("field1", "hello"),
            zap.String("field2", "world"))
        devLogger = devLogger.With(zap.String("field1", "goodbye"))
        devLogger.Info("Everything is OK", zap.Int64("value", 42))
    })

    // The field1 initial value was overridden
    assert.True(t, strings.Contains(out,
        "Everything is OK\t{\"value\":42,\"field1\":\"goodbye\",\"field2\":\"world\"}"))
}

func TestLevels(t *testing.T) {
    out := capturer.CaptureStderr(func() {
        devLogger := ConfigureProdLogger()
        devLogger = devLogger.With(zap.String("field1", "goodbye"))
        devLogger.Debug("Will Get Eaten", zap.Int64("value", 42))
    })
    assert.True(t, out == "")
}
