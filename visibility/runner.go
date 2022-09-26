package visibility


//type Instrumenter struct {
//    obs               Observer
//    span              opentracing.Span
//    clientType        string
//    addSuccessMetrics bool
//    time              *TimeMeasurement
//    met               *MetricsContext
//    log               *zap.Logger
//}
//
//func StartInstrumenter(ctx context.Context, obs Observer, addSuccessMetrics bool,
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
////        span:              span,
//        clientType:        clientType,
//        addSuccessMetrics: addSuccessMetrics,
//        time:              bench,
//        met:               met,
//        log:               CL(ctx),
//    }, ctx
//}
//
//func (i *Instrumenter) Guard(err error) {
//    if p := recover(); p != nil {
//        i.panicHelper(p)
//        panic(p)
//    }
//    _ = i.Check(err)
//}
//
//func (i *Instrumenter) CheckAndIgnoreErr(err error) {
//    _ = i.Check(err)
//}
//
//func (i *Instrumenter) Check(err error) error {
//    if i.addSuccessMetrics {
//        // We have set Fault to 1 initially. If the code in between StartInstrumentation and Check panics,
//        // then we never reach this statement and the value of 1 propagates to the caller. However, if we
//        // do reach this, then it means that the fault (panic) hasn't happened, and we need to reset it.
//        i.met.AddCount("Fault", -1)
//
//        if err == nil {
//            i.met.AddCount("Success", 1)
//        } else {
//            i.met.AddCount("Error", 1)
//        }
//        i.time.Done()
//    }
//
//    //i.met.CopyToStatsd(i.obs.Stats, i.clientType)
//    //i.met.CopyToSpan(i.span)
//
//    if err != nil {
//        i.log.Info("Traced segment failed with error", zap.Error(err))
//        ext.Error.Set(i.span, true)
//        i.span.SetTag("error", err)
//        i.span.Finish()
//    } else {
//        i.span.Finish()
//    }
//
//    return err
//}
//
//func (i *Instrumenter) panicHelper(p interface{}) {
//    if i.addSuccessMetrics {
//        // We have set Fault to 1 initially. We simply leave it be.
//        i.time.Done()
//    }
//    // Process the stats. We set the metrics to Fault=1 in StartInstrumenter, so that's
//    // what we'll submit here.
//    //i.met.CopyToStatsd(i.obs.Stats, i.clientType)
//    //i.met.CopyToSpan(i.span)
//
//    // Create an error with a nice stack trace
//    stack := NewShortenedStackTrace(5, true,
//        fmt.Sprintf("%v", p))
//    i.span.SetTag(ErrorStack, stack.StringStack())
//    i.span.SetTag("panic", fmt.Sprintf("%v", p))
//
//    // Finalize the span with the error
//    errField := log.Error(fmt.Errorf("gopanic: %v", p))
//    fo := opentracing.FinishOptions{
//        LogRecords: []opentracing.LogRecord{{Fields: []log.Field{errField}}},
//    }
//    i.span.FinishWithOptions(fo)
//    i.log.Info("Task failed with panic", zap.Any("panic", p))
//}
//
//func (i *Instrumenter) BeOnGuardAgainstPanics() {
//    if p := recover(); p != nil {
//        i.panicHelper(p)
//        panic(p)
//    }
//}
//
//// RunInstrumented - traces the provided synchronous function by beginning and closing a new
//// subsegment around its execution. If the parent segment doesn't exist yet, then a new
//// top-level segment is created
//func RunInstrumented(ctx context.Context, obs Observer,
//    name string, fn func(context.Context) error) error {
//
//    i, ctx := StartInstrumenter(ctx, obs, false, name)
//    defer i.BeOnGuardAgainstPanics()
//
//    err := fn(ctx)
//
//    return i.Check(err)
//}
//
//func InstrumentWithMetrics(ctx context.Context, fn func(context.Context) error) error {
//    met := GetMetricsFromContext(ctx)
//    met.AddCount("Success", 0)
//    met.AddCount("Error", 0)
//    met.AddCount("Fault", 1) // Panic trick (see below)
//
//    bench := met.Benchmark("Time")
//    defer bench.Done()
//
//    err := fn(ctx)
//
//    // We have set Fault to 1 initially. If the function panics then we never reach
//    // this statement and the value of 1 propagates to the caller. However, if we
//    // do reach this, then it means that the fault (panic) hasn't happened, and we
//    // need to reset it. This is a small trick, to avoid the use of a defer/recover pair.
//    met.AddCount("Fault", -1)
//
//    if err == nil {
//        met.AddCount("Success", 1)
//    } else {
//        met.AddCount("Error", 1)
//    }
//
//    return err
//}
//
//func RunInstrumentedWithMetrics(ctx context.Context, obs Observer,
//    name string, fn func(context.Context) error) error {
//
//    i, ctx := StartInstrumenter(ctx, obs, true, name)
//    defer i.BeOnGuardAgainstPanics()
//
//    err := fn(ctx)
//
//    return i.Check(err)
//}
