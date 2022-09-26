package logging

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestMemorySinkLogger(t *testing.T) {
	sink, logger := NewMemorySinkLogger()
	logger.Info("hello, world")
	assert.Equal(t, "{\"level\":\"info\",\"msg\":\"hello, world\"}\n", sink.String())
	_ = sink.Close()
}

func TestContextLogging(t *testing.T) {
	ctx := context.Background()

	sink, logger := NewMemorySinkLogger()

	imbued := ImbueContext(ctx, logger)
	L(imbued).Info("Hello this is a test", zap.Int64("test", 123))
	SL(imbued).Infof("Hello this is a test %d", 123)

	res := sink.String()
	splits := strings.Split(res, "\n")
	assert.True(t, strings.HasSuffix(splits[0],
		`"msg":"Hello this is a test","test":123}`))
	assert.True(t, strings.HasSuffix(splits[1],
		`"msg":"Hello this is a test 123"}`))
}

func TestNoLog(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() {
		L(ctx)
	})
	assert.Panics(t, func() {
		SL(ctx)
	})
}
