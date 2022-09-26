package visibility

//type SpanObserver struct {
//    obs  *Observer
//    span trace.Span
//
//    isCanary          bool
//
//    addSuccessMetrics bool
//
//    time              *TimeMeasurement
//    met               *MetricsContext
//    log               *zap.Logger
//}
//
//func StartSpanObserver(ctx context.Context, obs Observer, addSuccessMetrics bool,
//    name string) (*Instrumenter, context.Context) {
//
//    clientType := GetClientTypeFromContext(ctx)
//
//    ctx, span := obs.TraceProvider.Tracer("aa").Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))
//    //span.SetTag(ext.ResourceName, name)
//    //span.SetTag(ClientTypeTag, clientType)
//    //span.SetOperationName(name)
//
//    ctx = obs.ContextWithLogger(ctx, name, obs.LogFieldsForSpan(span)...)
//    ctx = MakeMetricContext(ctx, name) // Save metrics into the context
//
//    var bench *TimeMeasurement
//    met := GetMetricsFromContext(ctx)
//
//    if addSuccessMetrics {
//        met.AddCount("Success", 0)
//        met.AddCount("Error", 0)
//        met.AddCount("Fault", 1) // Panic trick (see below in Check method)
//        bench = met.Benchmark("Time")
//    }
//
//    return &Instrumenter{
//        obs:               obs,
//        //        span:              span,
//        clientType:        clientType,
//        addSuccessMetrics: addSuccessMetrics,
//        time:              bench,
//        met:               met,
//        log:               logging.CL(ctx),
//    }, ctx
//}
