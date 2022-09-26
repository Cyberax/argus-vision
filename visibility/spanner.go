package visibility

import (
	"context"
	"fmt"
	"github.com/Cyberax/argus-vision/utils"
	"github.com/Cyberax/argus-vision/visibility/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type wrappedSpan struct {
	trace.Span

	obs *Observer

	cfg       BeginSpanConfig
	startTime time.Time

	met *MetricHelper
	log *zap.Logger

	panic func(v any)

	mtx            sync.Mutex
	storedError    error
	endedWithError bool
	endOptions     []trace.SpanEndOption
}

func (s *wrappedSpan) End(options ...trace.SpanEndOption) {
	// We don't delegate to the original trace.Span.End implementation because we want
	// to be able to catch unwinding panics in the CleanupSpan function and override any
	// false successes.
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.endOptions = options
}

func (s *wrappedSpan) RecordError(err error, options ...trace.EventOption) {
	if err == nil {
		return
	}

	s.Span.RecordError(err, options...)

	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.storedError = err
}

func (s *wrappedSpan) SetStatus(code codes.Code, description string) {
	s.Span.SetStatus(code, description)

	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.endedWithError = code == codes.Error
}

func BeginNewSpan(ctx context.Context, obs *Observer, name string,
	options ...BeginSpanOption) (trace.Span, context.Context) {

	cfg := BeginSpanConfig{
		SpanName:       name,
		LibraryName:    obs.DefaultLibraryName,
		MetricNameBase: name,
	}

	for _, o := range options {
		o(&cfg)
	}

	return doBeginNewSpan(ctx, obs, cfg)
}

func BeginNewSpanWithConfig(ctx context.Context, obs *Observer,
	config BeginSpanConfig) (trace.Span, context.Context) {
	// Yes, we're just directly calling the `doBeginNewSpan` without doing anything, and
	// it seems like it can be replaced just by making `doBeginNewSpan` public.
	// But this is on purpose, to make sure that we skip the same number of frames in
	// armFinalizer on this code path and when called through BeginNewSpan.
	return doBeginNewSpan(ctx, obs, config)
}

func doBeginNewSpan(ctx context.Context, obs *Observer, config BeginSpanConfig) (trace.Span, context.Context) {
	startCtx := ctx
	if config.GraftedParent != nil {
		startCtx = trace.ContextWithSpanContext(startCtx, *config.GraftedParent)
	}

	_, span := obs.TraceProvider.Tracer(config.LibraryName).Start(
		startCtx, config.SpanName, config.StartSpanOptions...)

	canary := IsCanaryRequest(ctx)
	if canary {
		span.SetAttributes(attribute.Bool(CanaryAttributeName, true))
	}

	var mh *MetricHelper
	if config.AddMetrics {
		mh = obs.MakeMetricHelperWithPrefix(ctx, config.MetricPrefix)
		mh.InitCounts(config.MetricNameBase+"Success", config.MetricNameBase+"Error", config.MetricNameBase+"Fault")
		ctx = ContextWithMetricHelper(ctx, mh)
	}

	ctx = obs.ContextWithLogger(ctx, config.SpanName, obs.LogFieldsForSpan(span)...)

	// Wrap the span
	w := &wrappedSpan{
		Span:      span,
		obs:       obs,
		cfg:       config,
		startTime: time.Now(),
		met:       mh,
		log:       logging.L(ctx),

		panic: func(v any) { panic(v) },
	}

	if !config.WithoutLeakCheck {
		armFinalizer(w)
	}

	return w, trace.ContextWithSpan(ctx, w)
}

// armFinalizer arms the finalizer to detect unpaired calls to BeginNewSpan and CleanupSpan
func armFinalizer(span *wrappedSpan) {
	_, file, line, _ := runtime.Caller(3)
	finalizerArmedAt := file + ":" + strconv.Itoa(line)

	runtime.SetFinalizer(span, func(w *wrappedSpan) {
		w.panic("A span has not been finalized. Created at: " + finalizerArmedAt)
	})
}

func CleanupSpan(span trace.Span) {
	// We can not move the common code for panic recovery into doCleanupWithErr() because
	// recover() works only for the topmost deferred stack frame.
	thrownPanic := recover()
	doCleanupWithErr(span, nil, thrownPanic)
	if thrownPanic != nil {
		panic(thrownPanic)
	}
}

func CleanupWithErr(span trace.Span, err error) {
	// We can not move the common code for panic recovery into doCleanupWithErr() because
	// recover() works only for the topmost deferred stack frame.
	thrownPanic := recover()
	doCleanupWithErr(span, err, thrownPanic)
	if thrownPanic != nil {
		panic(thrownPanic)
	}
}

func doCleanupWithErr(span trace.Span, err error, thrownPanic any) {
	w, ok := span.(*wrappedSpan)
	utils.PanicIfF(!ok, "Trying to finalize a span not created by BeginNewSpan")

	if !w.cfg.WithoutLeakCheck {
		// Disarm the finalizer that we armed earlier in armFinalizer
		runtime.SetFinalizer(w, nil)
	}

	successMet := w.cfg.MetricPrefix + w.cfg.MetricNameBase + "Success"
	errorMet := w.cfg.MetricPrefix + w.cfg.MetricNameBase + "Error"
	failMet := w.cfg.MetricPrefix + w.cfg.MetricNameBase + "Fault"
	if w.met != nil {
		// Export metrics into the span as tags
		w.met.ExportToSpan(w.Span)

		w.met.InitCounts(successMet, errorMet, failMet)
	}

	w.mtx.Lock()
	defer w.mtx.Unlock()

	// The recover() call here will NOT stop the unwinding sequence, preserving the
	// full stacktrace for inspection/logging by callers up the stack.
	if thrownPanic != nil {
		// We're unwinding!

		if w.met != nil {
			w.met.AddCount(failMet, 1)
			w.met.Close()
		}

		err = fmt.Errorf("panic: %v", thrownPanic)
		w.endOptions = append(w.endOptions, trace.WithStackTrace(true))
		w.Span.RecordError(err, trace.WithStackTrace(true))
		w.Span.SetStatus(codes.Error, err.Error())
		w.Span.SetAttributes(attribute.Bool("IsInPanic", true))
		w.Span.End(w.endOptions...)
		return
	}

	// We have an error that we need to register
	if err != nil {
		if w.met != nil {
			w.met.AddCount(errorMet, 1)
			w.met.Close()
		}

		w.Span.RecordError(err)
		w.Span.SetStatus(codes.Error, err.Error())
		w.Span.End(w.endOptions...)
		return
	}

	// No error passed to this function
	if w.met != nil {
		if w.endedWithError || w.storedError != nil {
			w.met.AddCount(errorMet, 1)
		} else {
			w.met.AddCount(successMet, 1)
		}
		w.met.Close()
	}

	// Nothing special needs to be done
	w.Span.End(w.endOptions...)
}
