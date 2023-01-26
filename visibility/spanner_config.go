package visibility

import "go.opentelemetry.io/otel/trace"

// BeginSpanConfig is used for the span configuration
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

// BeginSpanOption is used to customize the span options
type BeginSpanOption func(cfg *BeginSpanConfig)

// WithGraftedParent sets the parent of the span, overriding any existing parents that might
// be passed from the context. This is useful for spans that are created asynchronously, even
// after their parent request may be completed.
func WithGraftedParent(traceId trace.TraceID, spanId trace.SpanID, remote bool) BeginSpanOption {
	spc := trace.SpanContext{}.WithTraceID(traceId).WithSpanID(spanId).WithRemote(remote)
	return func(cfg *BeginSpanConfig) {
		cfg.GraftedParent = &spc
	}
}

// WithLinkedContext adds the specified trace.SpanContext as a span link. It can be used to
// link untrusted remote spans.
func WithLinkedContext(spc trace.SpanContext) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		link := trace.Link{
			SpanContext: spc,
		}
		cfg.StartSpanOptions = append(cfg.StartSpanOptions, trace.WithLinks(link))
	}
}

// WithLink see WithLinkedContext
func WithLink(traceId trace.TraceID, spanId trace.SpanID, remote bool) BeginSpanOption {
	spc := trace.SpanContext{}.WithTraceID(traceId).WithSpanID(spanId).WithRemote(remote)
	return WithLinkedContext(spc)
}

// WithSpanStartOptions allows customizing the span start options
func WithSpanStartOptions(options ...trace.SpanStartOption) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.StartSpanOptions = append(cfg.StartSpanOptions, options...)
	}
}

// WithCustomLibraryName overrides the "instrumentation library name".
// Instrumented Library and Instrumentation Library may be the same library if it has built-in
// OpenTelemetry instrumentation.
//
// The inspiration of the OpenTelemetry project is to make every library and application observable
// out-of-the-box by having them call OpenTelemetry API directly. However, many libraries will not
// have such integration, and as such there is a need for a separate library which would inject
// such calls, using mechanisms such as wrapping interfaces, subscribing to library-specific callbacks,
// or translating existing telemetry into the OpenTelemetry model.
//
// A library that enables OpenTelemetry observability for another library is called an Instrumentation Library.
//
// An instrumentation library should be named to follow any naming conventions of the instrumented
// library (e.g. middleware for a web framework).
func WithCustomLibraryName(name string) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.LibraryName = name
	}
}

// WithMetrics turns on the automatic metrics for the span. It will create a new metric context
// and submit the following metrics:
// <CustomMetricsPrefix><SpanName>Success=1 in case the span succeeds
// <CustomMetricsPrefix><SpanName>Error=1 in case the span fails
// <CustomMetricsPrefix><SpanName>Fault=1 in case the span panics
func WithMetrics() BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.AddMetrics = true
	}
}

// WithCustomMetricPrefix customizes the span metrics prefix
func WithCustomMetricPrefix(prefix string) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.MetricPrefix = prefix
	}
}

// WithCustomMetricNameBase overrides the span name for the success/error/fault metrics
func WithCustomMetricNameBase(base string) BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.MetricNameBase = base
	}
}

// WithoutLeakCheck disables the span leak checker. Leak checker imposes a slight overhead
// that might be inappropriate for very tight inner loops (but then, why do you want
// to run them as separate spans?)
func WithoutLeakCheck() BeginSpanOption {
	return func(cfg *BeginSpanConfig) {
		cfg.WithoutLeakCheck = true
	}
}
