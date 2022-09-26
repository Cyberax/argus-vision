package visibility

import (
    "context"
    "github.com/Cyberax/argus-vision/visibility/logging"
    "github.com/stretchr/testify/assert"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/metric/unit"
    semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
    "go.opentelemetry.io/otel/trace"
    v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
    "go.uber.org/zap"
    "strings"
    "testing"
)

func TestObserverWithMockCollector(t *testing.T) {
    namedBytes := Named("hello_world_test2", unit.Bytes)

    mc := runMockCollector(t)
    t.Cleanup(mc.Stop)

    opts, err := NewDefaultObserverOptions("ArgusApp", "TracedDB", "alpha")
    assert.NoError(t, err)
    opts.TracingEndpoint = mc.endpoint
    opts.MetricsEndpoint = mc.endpoint

    development, err := zap.NewDevelopment()
    assert.NoError(t, err)
    obs, err := NewObserver(development, opts)
    assert.NoError(t, err)

    tr := obs.MakeTracer()
    _, span := tr.Start(context.Background(), "SpanTest",
        trace.WithAttributes(semconv.ServiceNameKey.String("SomeLib")))

    metricCtx := obs.MakeMetricHelperWithPrefix(context.Background(), "prod.")
    metricCtx.Add(namedBytes, 123)
    metricCtx.Add(namedBytes, 321)
    metricCtx.ExportToSpan(span)
    metricCtx.Close()

    span.SetStatus(codes.Ok, "Everything's fine")
    span.End()

    obs.Shutdown(context.Background())

    spans, metrics := mc.Get()
    sp1 := spans[0].GetScopeSpans()[0].GetSpans()[0]
    assert.Equal(t, "SpanTest", sp1.Name)
    assert.Equal(t, "ArgusApp", spans[0].ScopeSpans[0].Scope.Name)

    resAttrs := make(map[string]string)
    for _, v := range spans[0].Resource.Attributes {
        resAttrs[v.Key] = v.Value.GetStringValue()
    }
    assert.Equal(t, "TracedDB", resAttrs["service.name"])
    assert.Equal(t, "alpha", resAttrs["deployment.environment"])

    assert.Equal(t, "prod.hello_world_test2", sp1.Attributes[1].Key)
    assert.Equal(t, 123.+321., sp1.Attributes[1].Value.GetDoubleValue())

    metricsMap := make(map[string]*v1.Metric)
    for _, m := range metrics[0].ScopeMetrics[0].Metrics {
        metricsMap[m.Name] = m
    }

    m0 := metricsMap["prod.hello_world_test2_num"]
    assert.Equal(t, 2., m0.GetSum().GetDataPoints()[0].GetAsDouble())

    m1 := metricsMap["prod.hello_world_test2"]
    assert.Equal(t, 123.+321., m1.GetSum().GetDataPoints()[0].GetAsDouble())
}

func TestObserverLogging(t *testing.T) {
    sink, logger := logging.NewMemorySinkLogger()

    observer, _ := NewRecordingObserver(logger)
    defer observer.Shutdown(context.Background())

    tr := observer.MakeTracer()

    _, span := tr.Start(context.Background(), "SpanTest")
    ctxWithLog := observer.ContextWithLogger(context.Background(), "",
        observer.LogFieldsForSpan(span)...)
    logging.L(ctxWithLog).Info("This is a test", zap.Int32("key", 123))
    span.End()

    messages := sink.Bytes()
    assert.True(t, strings.Contains(string(messages),
        `"dd.trace_id":"17947005427386152706","dd.span_id":"6496753628558916970"`))
}

func TestNoTracing(t *testing.T) {
    sink, logger := logging.NewMemorySinkLogger()

    obs, err := NewObserver(logger, NewBlindObserverOptions())
    assert.NoError(t, err)
    defer obs.Shutdown(context.Background())

    tr := obs.MakeTracer()

    _, span := tr.Start(context.Background(), "SpanTest")
    assert.False(t, span.SpanContext().HasSpanID())
    assert.False(t, span.SpanContext().HasTraceID())

    ctxWithLog := obs.ContextWithLogger(context.Background(), "", obs.LogFieldsForSpan(span)...)
    logging.L(ctxWithLog).Info("This is a test", zap.Int32("key", 123))
    span.End()

    messages := sink.Bytes()
    assert.True(t, !strings.Contains(string(messages), "dd.trace_id"))

    metricCtx := obs.MakeMetricHelper(context.Background())
    namedBytes := Named("hello_world_test2", unit.Bytes)
    metricCtx.Add(namedBytes, 123)
    metricCtx.Close()
}
