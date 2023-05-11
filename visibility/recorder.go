package visibility

import (
	"context"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"sync"
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
	pusher := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp)),
	)

	res.MeterController = pusher

	res.Shutdown = func(ctx context.Context) {
		_ = tp.Shutdown(ctx)
		_ = pusher.ForceFlush(ctx)
		_ = pusher.Shutdown(ctx)
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
	controller *metric.MeterProvider
	metrics    *recordingMetricExporter
	tracer     *recordingSpanExporter
}

func (r *Recorder) Get() Record {
	_ = r.controller.ForceFlush(context.Background())

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

var _ metric.Exporter = &recordingMetricExporter{}

func (e *recordingMetricExporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return metric.DefaultTemporalitySelector(kind)
}

func (e *recordingMetricExporter) Aggregation(kind metric.InstrumentKind) aggregation.Aggregation {
	return metric.DefaultAggregationSelector(kind)
}

func (e *recordingMetricExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.Sums == nil {
		e.Sums = make(map[string]float64)
	}

	for _, scope := range metrics.ScopeMetrics {
		for _, m := range scope.Metrics {
			agg := m.Data

			if sum, ok := agg.(*metricdata.Sum[float64]); ok {
				for _, p := range sum.DataPoints {
					e.Sums[m.Name] += p.Value
				}
			}
		}
	}

	return nil
}

func (e *recordingMetricExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func (e *recordingMetricExporter) Shutdown(ctx context.Context) error {
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
