package visibility

import (
    "context"
    "github.com/Cyberax/argus-vision/utils"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/metric/instrument"
    "go.opentelemetry.io/otel/metric/unit"
    "go.opentelemetry.io/otel/trace"
    "sync"
)

const metricContextKey = "MetricContext"

type MetricsContext struct {
    lock sync.Mutex

    meter        metric.Meter
    metricPrefix string
    tags         map[string]string

    metricsToZero   map[string]NamedMetric
    metricsToSubmit map[string]NamedMetric
    metricValues    map[string]float64
}

type NamedMetric struct {
    Name string
    Unit unit.Unit
}

func Named(nm string, un unit.Unit) NamedMetric {
    return NamedMetric{
        Name: nm,
        Unit: un,
    }
}

func ContextWithMetrics(ctx context.Context, m *MetricsContext) context.Context {
    return context.WithValue(ctx, metricContextKey, m)
}

func TryGetMetricsFromContext(ctx context.Context) *MetricsContext {
    value := ctx.Value(metricContextKey)
    if value == nil {
        return nil
    }
    return value.(*MetricsContext)
}

func GetMetricsFromContext(ctx context.Context) *MetricsContext {
    m := TryGetMetricsFromContext(ctx)
    utils.PanicIfF(m == nil, "No metric context was attached")
    return m
}

func NewMetricContext(meter metric.Meter) *MetricsContext {
    return NewMetricContextWithPrefix(meter, "")
}

func NewMetricContextWithPrefix(meter metric.Meter, metricPrefix string) *MetricsContext {
    return &MetricsContext{
        meter:           meter,
        metricPrefix:    metricPrefix,
        tags:            make(map[string]string),
        metricsToZero:   make(map[string]NamedMetric),
        metricsToSubmit: make(map[string]NamedMetric),
        metricValues:    make(map[string]float64),
    }
}

func (m *MetricsContext) AddTag(k, v string) {
    m.lock.Lock()
    defer m.lock.Unlock()

    m.tags[k] = v
}

func (m *MetricsContext) Init(metrics ...NamedMetric) {
    m.lock.Lock()
    defer m.lock.Unlock()

    for _, cur := range metrics {
        m.metricsToZero[cur.Name] = cur
    }
}

func (m *MetricsContext) InitCounts(metrics ...string) {
    m.lock.Lock()
    defer m.lock.Unlock()

    for _, cur := range metrics {
        m.metricsToZero[cur] = NamedMetric{Name: cur, Unit: Dimensionless}
    }
}

func (m *MetricsContext) Add(nm NamedMetric, val float64) {
    m.lock.Lock()
    defer m.lock.Unlock()

    // Make sure we don't submit the zero metric at the end of the call
    delete(m.metricsToZero, nm.Name)

    // Register the metrics in our segment
    curMet, ok := m.metricsToSubmit[nm.Name]
    if ok && curMet.Unit != nm.Unit {
        panic("Inconsistent units for metric " + nm.Name)
    }
    m.metricsToSubmit[nm.Name] = nm
    m.metricValues[nm.Name] = m.metricValues[nm.Name] + val

    // Record the counter
    hg, err := m.meter.SyncFloat64().UpDownCounter(
        m.metricPrefix+nm.Name, instrument.WithUnit(nm.Unit))
    utils.PanicIfErr(err)
    hg.Add(context.Background(), val, m.getTags()...)

    // Counters lose details since they are not submitted immediately, so make sure we submit
    // the number of samples taken to be able to calculate the average value.
    hgCnt, err := m.meter.SyncFloat64().Counter(
        m.metricPrefix+nm.Name+"_num", instrument.WithUnit(Dimensionless))
    utils.PanicIfErr(err)
    hgCnt.Add(context.Background(), 1, m.getTags()...)
}

// Close submits all the remaining zero-valued metrics
func (m *MetricsContext) Close() {
    m.lock.Lock()
    defer m.lock.Unlock()

    attrs := m.getTags()

    for _, val := range m.metricsToZero {
        counter, err := m.meter.SyncFloat64().UpDownCounter(
            m.metricPrefix+val.Name, instrument.WithUnit(val.Unit))
        utils.PanicIfErr(err)
        counter.Add(context.Background(), 0, attrs...)
    }
}

func (m *MetricsContext) ExportToSpan(span trace.Span) {
    m.lock.Lock()
    defer m.lock.Unlock()

    // Set metrics as the span attributes
    for k, v := range m.metricValues {
        span.SetAttributes(attribute.Float64(m.metricPrefix+k, v))
    }
}

func (m *MetricsContext) getTags() []attribute.KeyValue {
    var attrs []attribute.KeyValue
    for k, v := range m.tags {
        attrs = append(attrs, attribute.String(k, v))
    }
    return attrs
}
