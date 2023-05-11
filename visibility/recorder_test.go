package visibility

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	namedBytes := Named("hello_world_test2", UnitBytes)

	obs, rec := NewRecordingObserver(zap.NewNop())
	defer obs.Shutdown(context.Background())

	// Make sure that the recorder can be used multiple times
	for i := 0; i < 3; i++ {
		tr := obs.MakeTracer()
		_, span := tr.Start(context.Background(), "SpanTest",
			trace.WithAttributes(semconv.ServiceNameKey.String("DynamoDB")))

		span.SetStatus(codes.Ok, "Everything's fine")
		time.Sleep(10 * time.Millisecond)

		mc := NewMetricContext(context.Background(), obs.MeterController.Meter("ArgusApp"))
		mc.Add(namedBytes, 123)
		mc.Add(namedBytes, 321)
		mc.ExportToSpan(span)
		mc.Close()
		span.End()

		values := rec.Get()
		//TODO: fix
		//assert.Equal(t, 444., values.Metrics[namedBytes.Name])

		sp := values.Spans[0]
		assert.True(t, sp.EndTime().Sub(sp.StartTime()) >= 10*time.Millisecond)
		assert.Equal(t, "SpanTest", sp.Name())
		assert.Equal(t, "service.name", string(sp.Attributes()[0].Key))
		assert.Equal(t, "DynamoDB", sp.Attributes()[0].Value.AsString())
		assert.Equal(t, namedBytes.Name, string(sp.Attributes()[1].Key))
		assert.Equal(t, 444., sp.Attributes()[1].Value.AsFloat64())

		empty := rec.Get()
		assert.True(t, len(empty.Metrics) == 0)
	}

	// Test different metric aggregators
	histogram, err := obs.MeterController.Meter("TestHist").Float64Histogram("Hist1")
	assert.NoError(t, err)
	histogram.Record(context.Background(), 123)
	histogram.Record(context.Background(), 1)

	//histo := rec.Get()
	//TODO: fix
	//assert.Equal(t, 124., histo.Metrics["Hist1"])
}
