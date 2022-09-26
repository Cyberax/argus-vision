package visibility

import (
    "context"
    "github.com/Cyberax/argus-vision/visibility/logging"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
    "strconv"
    "time"

    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
    controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
    processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
    selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

type Observer struct {
    Logger *zap.Logger

    // The name of the application or library that is being traced.
    // E.g. if you are instrumenting YourCoolApp then set this to "YourCoolApp"
    DefaultLibraryName string
    Resource           *resource.Resource

    TraceProvider   trace.TracerProvider
    MeterClient     otlpmetric.Client
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

    return ObserverOptions{
        MetricsEndpoint: "localhost:4317",
        TracingEndpoint: "localhost:4317",
        LibraryName:     libraryName,
        Resource:        envInfo,
        IdGenerator:     &CryptoSafeRandIDGenerator{},
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
    var pusher *controller.Controller

    // Metrics
    if opts.MetricsEndpoint != "" {
        options := []otlpmetricgrpc.Option{
            otlpmetricgrpc.WithInsecure(),
            otlpmetricgrpc.WithEndpoint(opts.MetricsEndpoint),
        }

        client := otlpmetricgrpc.NewClient(options...)
        exp, err := otlpmetric.New(context.Background(), client)
        if err != nil {
            return nil, err
        }

        pusher = controller.New(
            processor.NewFactory(
                selector.NewWithHistogramDistribution(),
                exp,
            ),
            controller.WithResource(opts.Resource),
            controller.WithExporter(exp),
            controller.WithCollectPeriod(2*time.Second),
            controller.WithCollectTimeout(4*time.Second),
        )

        res.MeterClient = client
        res.MeterController = pusher

        err = pusher.Start(context.Background())
        if err != nil {
            return nil, err
        }
    } else {
        res.MeterController = metric.NewNoopMeterProvider()
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
        if pusher != nil {
            _ = pusher.Collect(ctx)
            _ = pusher.Stop(ctx)
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
