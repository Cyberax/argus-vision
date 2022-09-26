package logging

import (
    "bytes"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "sync"
)

var initMutex sync.Mutex
var initialized = false

func ConfigureZapGlobals() {
    initMutex.Lock()
    defer initMutex.Unlock()
    if initialized {
        return
    }

    err := zap.RegisterEncoder("prettyconsole",
        func(config zapcore.EncoderConfig) (zapcore.Encoder, error) {
            ce := NewPrettyConsoleEncoder(config)
            return ce, nil
        })

    if err != nil {
        panic(err.Error())
    }

    initialized = true
}


func ConfigureDevLogger() *zap.Logger {
    ConfigureZapGlobals()

    config := zap.NewDevelopmentConfig()
    config.Encoding = "prettyconsole"
    config.DisableStacktrace = true
    logger, err := config.Build(MakeFieldsUnique(true))
    if err != nil {
        panic(err.Error())
    }
    return logger
}

func ConfigureProdLogger() *zap.Logger {
    ConfigureZapGlobals()

    config := zap.NewProductionConfig()
    logger, err := config.Build(MakeFieldsUnique(true))
    if err != nil {
        panic(err.Error())
    }
    return logger
}

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
    bytes.Buffer
}
func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

func NewMemorySinkLogger() (*MemorySink, *zap.Logger) {
    sink := &MemorySink{}
    config := zap.NewProductionEncoderConfig()
    config.TimeKey = ""
    core := zapcore.NewCore(zapcore.NewJSONEncoder(config), sink, zap.DebugLevel)
    logger := zap.New(core)
    return sink, logger
}
