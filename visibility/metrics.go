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

var MetricTagKey = utils.NewMutableContextKey[map[string]string]("metricTags")

type MetricHelper struct {
	lock sync.Mutex

	startingContext context.Context
	meter           metric.Meter
	metricPrefix    string

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

func ContextWithMetricHelper(ctx context.Context, m *MetricHelper) context.Context {
	return context.WithValue(ctx, metricContextKey, m)
}

func TryGetMetricHelperFromContext(ctx context.Context) *MetricHelper {
	value := ctx.Value(metricContextKey)
	if value == nil {
		return nil
	}
	return value.(*MetricHelper)
}

func GetMetricHelperFromContext(ctx context.Context) *MetricHelper {
	m := TryGetMetricHelperFromContext(ctx)
	utils.PanicIfF(m == nil, "No metric context was attached")
	return m
}

func NewMetricContext(ctx context.Context, meter metric.Meter) *MetricHelper {
	return NewMetricContextWithPrefix(ctx, meter, "")
}

func NewMetricContextWithPrefix(ctx context.Context, meter metric.Meter, metricPrefix string) *MetricHelper {
	res := &MetricHelper{
		meter:           meter,
		startingContext: ctx,
		metricPrefix:    metricPrefix,
		metricsToZero:   make(map[string]NamedMetric),
		metricsToSubmit: make(map[string]NamedMetric),
		metricValues:    make(map[string]float64),
	}
	return res
}

func (m *MetricHelper) Init(metrics ...NamedMetric) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, cur := range metrics {
		m.metricsToZero[cur.Name] = cur
	}
}

func (m *MetricHelper) InitCounts(metrics ...string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, cur := range metrics {
		m.metricsToZero[cur] = NamedMetric{Name: cur, Unit: Dimensionless}
	}
}

func (m *MetricHelper) AddCount(nm string, val float64) {
	m.Add(NamedMetric{Name: nm, Unit: Dimensionless}, val)
}

func (m *MetricHelper) Add(nm NamedMetric, val float64) {
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
func (m *MetricHelper) Close() {
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

func (m *MetricHelper) ExportToSpan(span trace.Span) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Set metrics as the span attributes
	for k, v := range m.metricValues {
		span.SetAttributes(attribute.Float64(m.metricPrefix+k, v))
	}
}

func (m *MetricHelper) getTags() []attribute.KeyValue {
	var attrs []attribute.KeyValue
	//for k, v := range m.tags {
	//	attrs = append(attrs, attribute.String(k, v))
	//}
	return attrs
}
