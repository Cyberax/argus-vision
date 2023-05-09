package visibility

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	unsaferand "math/rand"
	"sync"
	"time"
)

type cryptoSafeRandIDGenerator struct {
	xrayTimePrefixed bool
}

var _ sdktrace.IDGenerator = &cryptoSafeRandIDGenerator{}

func NewCryptoSafeRandIdGenerator(xrayTimePrefixed bool) sdktrace.IDGenerator {
	return &cryptoSafeRandIDGenerator{xrayTimePrefixed: xrayTimePrefixed}
}

// NewSpanID returns a new crypto-safe non-zero span ID
func (gen *cryptoSafeRandIDGenerator) NewSpanID(ctx context.Context, _ trace.TraceID) trace.SpanID {
	sid := trace.SpanID{}
	_, _ = rand.Read(sid[:])
	return sid
}

// NewIDs returns new crypto-safe random non-zero trace ID and span ID,
// optionally replace the first 4 bytes of the resulting trace ID with
// the current time.
func (gen *cryptoSafeRandIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	tid := trace.TraceID{}
	_, _ = rand.Read(tid[:])

	// Replace the first 4 bytes of the trace ID with the Unix time, as specified
	// in https://docs.aws.amazon.com/xray/latest/devguide/xray-api-sendingdata.html#xray-api-traceids
	// This is OK until around the year 2106.
	if gen.xrayTimePrefixed {
		nowTime := uint32(time.Now().Unix())
		binary.BigEndian.PutUint32(tid[:], nowTime)
	}

	sid := trace.SpanID{}
	_, _ = rand.Read(sid[:])
	return tid, sid
}

type PredictableIdGen struct {
	mtx        sync.Mutex
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
