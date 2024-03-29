package visibility

import (
	"context"
	"github.com/Cyberax/argus-vision/visibility/logging"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
)

type Observer struct {
	Logger *zap.Logger

	// The name of the application or library that is being traced.
	// E.g. if you are instrumenting YourCoolApp then set this to "YourCoolApp"
	DefaultLibraryName string
	Resource           *resource.Resource

	TraceProvider   trace.TracerProvider
	MeterController metric.MeterProvider

	LogFieldsForSpan func(span trace.Span) []zap.Field

	Shutdown func(ctx context.Context)
}

type ObserverOptions struct {
	MetricsEndpoint string // localhost:4317 is the default
	TracingEndpoint string // localhost:4317 is the default

	// The name of the application or library that is being traced.
	// E.g. if you are instrumenting YourCoolApp then set this to "YourCoolApp"
	LibraryName string
	Resource    *resource.Resource

	// The ID generator for the spans, can be customized to produce predictable IDs
	IdGenerator sdktrace.IDGenerator
}

func NewDefaultObserverOptions(libraryName, serviceName, envName string) (ObserverOptions, error) {
	envInfo, err := resource.New(context.Background(),
		// pull attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables
		resource.WithFromEnv(),
		// This option configures a set of Detectors that discovers the process information
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
		resource.WithAttributes(semconv.DeploymentEnvironmentKey.String(envName)),
	)
	if err != nil {
		return ObserverOptions{}, err
	}

	// ENV vars as specified in:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md
	both := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if both == "" {
		both = "localhost:4317"
	}
	metrics := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")
	if metrics == "" {
		metrics = both
	}
	tracing := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if tracing == "" {
		tracing = both
	}

	return ObserverOptions{
		MetricsEndpoint: metrics,
		TracingEndpoint: tracing,
		LibraryName:     libraryName,
		Resource:        envInfo,
		IdGenerator:     NewCryptoSafeRandIdGenerator(true),
	}, nil
}

func NewBlindObserverOptions() ObserverOptions {
	return ObserverOptions{
		MetricsEndpoint: "",
		TracingEndpoint: "",
	}
}

func NewObserver(rootLogger *zap.Logger, opts ObserverOptions) (*Observer, error) {
	res := &Observer{
		Logger:             rootLogger,
		DefaultLibraryName: opts.LibraryName,
		Resource:           opts.Resource,

		LogFieldsForSpan: DatadogLogDerivation,
	}

	var tp *sdktrace.TracerProvider
	var pusher metric.MeterProvider

	// Metrics
	if opts.MetricsEndpoint != "" {
		options := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithEndpoint(opts.MetricsEndpoint),
		}

		client, err := otlpmetricgrpc.New(context.Background(), options...)
		if err != nil {
			return nil, err
		}

		pusher = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(client,
				sdkmetric.WithInterval(2*time.Second),
				sdkmetric.WithTimeout(2*time.Second),
			)),
			sdkmetric.WithResource(opts.Resource),
		)

		res.MeterController = pusher

		if err != nil {
			return nil, err
		}
	} else {
		res.MeterController = noop.NewMeterProvider()
	}

	// Traces
	if opts.TracingEndpoint != "" {
		tracerOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(opts.TracingEndpoint),
		}
		traceClient := otlptracegrpc.NewClient(tracerOpts...)
		traceExporter, err := otlptrace.New(context.Background(), traceClient)
		if err != nil {
			return nil, err
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithResource(opts.Resource),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithIDGenerator(opts.IdGenerator),
			sdktrace.WithBatcher(
				traceExporter,
				sdktrace.WithBatchTimeout(5*time.Second),
				sdktrace.WithMaxExportBatchSize(10),
			),
		)

		res.TraceProvider = tp
	} else {
		res.TraceProvider = trace.NewNoopTracerProvider()
	}

	res.Shutdown = func(ctx context.Context) {
		if tp != nil {
			_ = tp.Shutdown(ctx)
		}
	}

	return res, nil
}

func DatadogLogDerivation(span trace.Span) []zap.Field {
	spanCtx := span.SpanContext()

	if !spanCtx.HasSpanID() || !spanCtx.HasTraceID() {
		return []zap.Field{}
	}

	return []zap.Field{
		zap.String("dd.trace_id", convertDatadogTraceIDToLogId(spanCtx.TraceID().String())),
		zap.String("dd.span_id", convertDatadogTraceIDToLogId(spanCtx.SpanID().String())),
	}
}

// See: https://docs.datadoghq.com/tracing/other_telemetry/connect_logs_and_traces/opentelemetry/?tab=go
func convertDatadogTraceIDToLogId(id string) string {
	if len(id) < 16 {
		return ""
	}
	if len(id) > 16 {
		id = id[16:]
	}
	intValue, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(intValue, 10)
}

func (o *Observer) ContextWithLogger(parent context.Context, name string, fields ...zap.Field) context.Context {
	parentLogger := logging.TryGetLoggerFromContext(parent)
	if parentLogger == nil {
		parentLogger = o.Logger
	}
	logger := parentLogger
	if name != "" {
		logger = parentLogger.Named(name)
	}
	logger = logger.With(fields...)
	return logging.ImbueContext(parent, logger)
}

func (o *Observer) MakeMetricHelper(ctx context.Context) *MetricHelper {
	return NewMetricContext(ctx, o.MeterController.Meter(o.DefaultLibraryName))
}

func (o *Observer) MakeMetricHelperWithPrefix(ctx context.Context, prefix string) *MetricHelper {
	return NewMetricContextWithPrefix(ctx, o.MeterController.Meter(o.DefaultLibraryName), prefix)
}

func (o *Observer) MakeTracer() trace.Tracer {
	return o.TraceProvider.Tracer(o.DefaultLibraryName)
}
