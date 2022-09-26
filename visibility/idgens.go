package visibility

import (
    "context"
    "crypto/rand"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
    unsaferand "math/rand"
    "sync"
)

type CryptoSafeRandIDGenerator struct {
}

var _ sdktrace.IDGenerator = &CryptoSafeRandIDGenerator{}

// NewSpanID returns a new crypto-safe non-zero span ID
func (gen *CryptoSafeRandIDGenerator) NewSpanID(ctx context.Context, _ trace.TraceID) trace.SpanID {
    sid := trace.SpanID{}
    _, _ = rand.Read(sid[:])
    return sid
}

// NewIDs returns new crypto-safe random non-zero trace ID and span ID
func (gen *CryptoSafeRandIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
    tid := trace.TraceID{}
    _, _ = rand.Read(tid[:])
    sid := trace.SpanID{}
    _, _ = rand.Read(sid[:])
    return tid, sid
}


type PredictableIdGen struct {
    mtx sync.Mutex
    randSource *unsaferand.Rand
}

var _ sdktrace.IDGenerator = &PredictableIdGen{}

func NewPredictableIdGen(seed int64) *PredictableIdGen {
    source := unsaferand.NewSource(seed)
    return &PredictableIdGen{randSource: unsaferand.New(source)}
}

// NewSpanID returns a non-zero span ID from the pre-seeded pseudorandom source
func (g *PredictableIdGen) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
    g.mtx.Lock()
    defer g.mtx.Unlock()
    sid := trace.SpanID{}
    _, _ = g.randSource.Read(sid[:])
    return sid
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from the pre-seeded pseudorandom source
func (g *PredictableIdGen) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
    g.mtx.Lock()
    defer g.mtx.Unlock()
    tid := trace.TraceID{}
    _, _ = g.randSource.Read(tid[:])
    sid := trace.SpanID{}
    _, _ = g.randSource.Read(sid[:])
    return tid, sid
}
