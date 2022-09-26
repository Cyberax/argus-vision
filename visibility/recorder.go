package visibility

import (
    "context"
    "go.opentelemetry.io/otel/sdk/instrumentation"
    controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
    "go.opentelemetry.io/otel/sdk/metric/export"
    "go.opentelemetry.io/otel/sdk/metric/export/aggregation"
    processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
    "go.opentelemetry.io/otel/sdk/metric/sdkapi"
    selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.uber.org/zap"
    "sync"
    "time"
)

func NewRecordingObserver(rootLogger *zap.Logger) (*Observer, *Recorder) {
    res := &Observer{
        Logger:           rootLogger,
        LogFieldsForSpan: DatadogLogDerivation,
    }

    tracerExp := &recordingSpanExporter{}

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
        sdktrace.WithSyncer(tracerExp),
        sdktrace.WithIDGenerator(NewPredictableIdGen(123)),
    )
    res.TraceProvider = tp

    exp := &recordingMetricExporter{}
    pusher := controller.New(
        processor.NewFactory(
            selector.NewWithInexpensiveDistribution(),
            exp,
        ),
        controller.WithExporter(exp),
        controller.WithCollectPeriod(2*time.Second),
    )
    res.MeterController = pusher
    _ = pusher.Start(context.Background())

    res.Shutdown = func(ctx context.Context) {
        _ = tp.Shutdown(ctx)
        _ = pusher.Collect(ctx)
        _ = pusher.Stop(ctx)
    }

    return res, &Recorder{
        controller: pusher,
        metrics:    exp,
        tracer:     tracerExp,
    }
}

type Record struct {
    Metrics map[string]float64
    Spans   []trace.ReadOnlySpan
}

type Recorder struct {
    controller *controller.Controller
    metrics    *recordingMetricExporter
    tracer     *recordingSpanExporter
}

func (r *Recorder) Get() Record {
    // Force the metric collection (yes, it's the only way)
    _ = r.controller.Stop(context.Background())
    _ = r.controller.Start(context.Background())

    var res Record

    r.metrics.mtx.Lock()
    defer r.metrics.mtx.Unlock()
    res.Metrics = r.metrics.Sums
    r.metrics.Sums = nil

    if res.Metrics == nil {
        res.Metrics = make(map[string]float64)
    }

    r.tracer.mtx.Lock()
    defer r.tracer.mtx.Unlock()
    res.Spans = r.tracer.spans
    r.tracer.spans = nil

    return res
}

type recordingMetricExporter struct {
    mtx sync.Mutex

    Sums map[string]float64
}

var _ export.Exporter = &recordingMetricExporter{}

func (e *recordingMetricExporter) TemporalityFor(desc *sdkapi.Descriptor, kind aggregation.Kind) aggregation.Temporality {
    return aggregation.StatelessTemporalitySelector().TemporalityFor(desc, kind)
}

func (e *recordingMetricExporter) Export(_ context.Context, res *resource.Resource,
    reader export.InstrumentationLibraryReader) error {

    e.mtx.Lock()
    defer e.mtx.Unlock()

    if e.Sums == nil {
        e.Sums = make(map[string]float64)
    }

    err := reader.ForEach(func(lib instrumentation.Scope, mr export.Reader) error {
        return mr.ForEach(e, func(record export.Record) error {
            agg := record.Aggregation()

            if sum, ok := agg.(aggregation.Sum); ok {
                value, err := sum.Sum()
                if err != nil {
                    return err
                }
                e.Sums[record.Descriptor().Name()] += value.AsFloat64()
            } else if lv, ok := agg.(aggregation.LastValue); ok {
                value, _, err := lv.LastValue()
                if err != nil {
                    return err
                }
                e.Sums[record.Descriptor().Name()] += value.AsFloat64()
            }

            return nil
        })
    })
    if err != nil {
        return err
    }

    return nil
}

type recordingSpanExporter struct {
    mtx   sync.Mutex
    spans []trace.ReadOnlySpan
}

var _ trace.SpanExporter = &recordingSpanExporter{}

// ExportSpans writes spans in json format to stdout.
func (e *recordingSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
    e.mtx.Lock()
    defer e.mtx.Unlock()

    e.spans = append(e.spans, spans...)
    return nil
}

// Shutdown is called to stop the exporter, it preforms no action.
func (e *recordingSpanExporter) Shutdown(_ context.Context) error {
    return nil
}
