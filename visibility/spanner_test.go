package visibility

import (
	"context"
	"github.com/Cyberax/argus-vision/visibility/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"runtime"
	"strings"
	"testing"
)

func TestSpanLeak(t *testing.T) {
	t.Parallel()
	obs, _ := NewRecordingObserver(zap.NewNop())

	panicMsg := atomic.NewString("")
	registerPanic := func(p any) {
		panicMsg.Store(p.(string))
	}

	// We leak this span
	sp, _ := BeginNewSpan(context.Background(), obs, "Leaked") // this line number --->
	sp.(*wrappedSpan).panic = registerPanic
	sp = nil
	runtime.GC() // Make sure finalizers fire

	assert.True(t, strings.HasPrefix(panicMsg.Load(), "A span has not been finalized"))
	assert.True(t, strings.HasSuffix(panicMsg.Load(), "spanner_test.go:24")) // <--- goes here
}

func TestSpannerHappyCase(t *testing.T) {
	ms, log := logging.NewMemorySinkLogger()
	obs, rec := NewRecordingObserver(log)

	func() {
		span, ctx := BeginNewSpan(context.Background(), obs, "TestSpan", WithMetrics())
		logging.L(ctx).Info("This is a test")
		defer CleanupSpan(span)
	}()

	spans := rec.Get()

	span1 := spans.Spans[0]
	assert.Equal(t, "TestSpan", span1.Name())
	assert.Equal(t, "f1405ced8b9968baf9109259515bf702", span1.SpanContext().TraceID().String())
	assert.Equal(t, "5a291b00ff7bfd6a", span1.SpanContext().SpanID().String())

	//assert.Equal(t, 1.0, spans.Metrics["TestSpanSuccess"])
	//assert.Equal(t, 0.0, spans.Metrics["TestSpanError"])
	//assert.Equal(t, 0.0, spans.Metrics["TestSpanFault"])

	logs := ms.String()
	exemplar := `{"level":"info","logger":"TestSpan","msg":"This is a test",` +
		`"dd.trace_id":"17947005427386152706","dd.span_id":"6496753628558916970"}`
	assert.Equal(t, exemplar, strings.TrimSpace(logs))
}

func TestError(t *testing.T) {
	//_, log := logging.NewMemorySinkLogger()
	//obs, rec := NewRecordingObserver(log)
	//
	//func() {
	//	span, _ := BeginNewSpan(context.Background(), obs, "TestSpan", WithMetrics())
	//	err := fmt.Errorf("bad error")
	//	defer CleanupWithErr(span, err)
	//}()

	//spans := rec.Get()

	//assert.Equal(t, 0.0, spans.Metrics["TestSpanSuccess"])
	//assert.Equal(t, 1.0, spans.Metrics["TestSpanError"])
	//assert.Equal(t, 0.0, spans.Metrics["TestSpanFault"])
}

func TestExternalError(t *testing.T) {
	//_, log := logging.NewMemorySinkLogger()
	//obs, rec := NewRecordingObserver(log)
	//
	//func() {
	//	span, _ := BeginNewSpan(context.Background(), obs, "TestSpan", WithMetrics())
	//	err := fmt.Errorf("bad error")
	//	span.RecordError(err)
	//	defer CleanupSpan(span)
	//}()

	//spans := rec.Get()

	//assert.Equal(t, 0.0, spans.Metrics["TestSpanSuccess"])
	//assert.Equal(t, 1.0, spans.Metrics["TestSpanError"])
	//assert.Equal(t, 0.0, spans.Metrics["TestSpanFault"])
}

func TestFault(t *testing.T) {
	_, log := logging.NewMemorySinkLogger()
	obs, rec := NewRecordingObserver(log)

	func() {
		defer func() {
			str := recover()
			assert.Equal(t, "run!", str)
		}()

		sp, _ := BeginNewSpan(context.Background(), obs, "TestSpan", WithMetrics())
		defer CleanupSpan(sp)
		panic("run!")
	}()

	spans := rec.Get()

	sp := spans.Spans[0]
	assert.Equal(t, "IsInPanic", string(sp.Attributes()[0].Key))
	assert.Equal(t, true, sp.Attributes()[0].Value.AsBool())

	//assert.Equal(t, 0.0, spans.Metrics["TestSpanSuccess"])
	//assert.Equal(t, 0.0, spans.Metrics["TestSpanError"])
	//assert.Equal(t, 1.0, spans.Metrics["TestSpanFault"])
}

func TestLinkedSpans(t *testing.T) {
	_, log := logging.NewMemorySinkLogger()
	obs, rec := NewRecordingObserver(log)

	sp, _ := BeginNewSpan(context.Background(), obs, "TestSpan", WithMetrics())
	spc := sp.SpanContext()
	CleanupSpan(sp)

	sp2, _ := BeginNewSpan(context.Background(), obs, "TestSpan2",
		WithLink(spc.TraceID(), spc.SpanID(), false), WithMetrics())
	CleanupSpan(sp2)

	sp3, _ := BeginNewSpan(context.Background(), obs, "TestSpan3",
		WithGraftedParent(spc.TraceID(), spc.SpanID(), false), WithMetrics())
	CleanupSpan(sp3)

	obs.Shutdown(context.Background())

	spans := rec.Get()

	span1 := spans.Spans[0]
	assert.Equal(t, "TestSpan", span1.Name())

	span2 := spans.Spans[1]
	assert.Equal(t, "TestSpan2", span2.Name())

	span3 := spans.Spans[2]
	assert.Equal(t, "TestSpan3", span3.Name())

	// Check that the spans are linked
	assert.Equal(t, span2.Links()[0].SpanContext.TraceID(), span1.SpanContext().TraceID())
	assert.Equal(t, span2.Links()[0].SpanContext.SpanID(), span1.SpanContext().SpanID())

	// Check that the grafted parent is correct
	assert.Equal(t, span1.SpanContext().TraceID(), span3.SpanContext().TraceID())
}

func TestCustomMetricTags(t *testing.T) {
	_, log := logging.NewMemorySinkLogger()
	obs, rec := NewRecordingObserver(log)

	sp, ctx := BeginNewSpan(context.Background(), obs, "TestSpan", WithMetrics())
	GetMetricHelperFromContext(ctx).AddCount("SomeCount", 5)
	CleanupSpan(sp)
	obs.Shutdown(context.Background())

	spans := rec.Get()

	span1 := spans.Spans[0]
	assert.Equal(t, "TestSpan", span1.Name())
	assert.Equal(t, "SomeCount", string(span1.Attributes()[0].Key))
	assert.Equal(t, true, span1.Attributes()[0].Value.AsBool())
}
