package visibility

import "go.opentelemetry.io/otel/trace"

type BeginSpanConfig struct {
	SpanName         string
	LibraryName      string
	StartSpanOptions []trace.SpanStartOption

	AddMetrics     bool
	MetricPrefix   string
	MetricNameBase string

	WithoutLeakCheck bool

	GraftedParent *trace.SpanContext
}

type BeginSpanOption func(cfg *BeginSpanConfig)

func WithGraftedParent(traceId trace.TraceID, spanId trace.SpanID, remote bool) BeginSpanOption {
	spc := trace.SpanContext{}.WithTraceID(traceId).WithSpanID(spanId).WithRemote(remote)
	return func(cfg *BeginSpanConfig) {
		cfg.GraftedParent = &spc
	}
}

func WithLinkedContext(spc trace.SpanContext) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		link := trace.Link{
			SpanContext: spc,
		}
		cfg.StartSpanOptions = append(cfg.StartSpanOptions, trace.WithLinks(link))
	}
}

func WithLink(traceId trace.TraceID, spanId trace.SpanID, remote bool) BeginSpanOption {
	spc := trace.SpanContext{}.WithTraceID(traceId).WithSpanID(spanId).WithRemote(remote)
	return WithLinkedContext(spc)
}

func WithSpanStartOptions(options ...trace.SpanStartOption) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.StartSpanOptions = append(cfg.StartSpanOptions, options...)
	}
}

func WithCustomLibraryName(name string) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.LibraryName = name
	}
}

func WithMetrics() BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.AddMetrics = true
	}
}

func WithCustomMetricPrefix(prefix string) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.MetricPrefix = prefix
	}
}

func WithCustomMetricNameBase(base string) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.MetricNameBase = base
	}
}

func WithoutLeakCheck() BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.WithoutLeakCheck = true
	}
}
