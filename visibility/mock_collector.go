package visibility

import (
    "context"
    metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
    collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
    v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
    tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
    "google.golang.org/grpc"
    "net"
    "sync"
    "testing"
    "time"
)

type mockMetricService struct {
    metricspb.UnimplementedMetricsServiceServer

    mtx             sync.Mutex
    resourceMetrics []*v1.ResourceMetrics
}

func (m *mockMetricService) getResourceMetrics() []*v1.ResourceMetrics {
    m.mtx.Lock()
    defer m.mtx.Unlock()
    res := make([]*v1.ResourceMetrics, len(m.resourceMetrics))
    copy(res, m.resourceMetrics)
    return res
}

func (m *mockMetricService) Export(_ context.Context,
    request *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {

    m.mtx.Lock()
    defer m.mtx.Unlock()
    m.resourceMetrics = append(m.resourceMetrics, request.GetResourceMetrics()...)

    return &metricspb.ExportMetricsServiceResponse{}, nil
}

type mockTraceService struct {
    collectortracepb.UnimplementedTraceServiceServer

    mtx           sync.Mutex
    resourceSpans []*tracepb.ResourceSpans
}

func (m *mockTraceService) getResourceSpans() []*tracepb.ResourceSpans {
    m.mtx.Lock()
    defer m.mtx.Unlock()
    res := make([]*tracepb.ResourceSpans, len(m.resourceSpans))
    copy(res, m.resourceSpans)
    return res
}

func (m *mockTraceService) Export(_ context.Context,
    exp *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {

    m.mtx.Lock()
    defer m.mtx.Unlock()

    reply := &collectortracepb.ExportTraceServiceResponse{}
    m.resourceSpans = append(m.resourceSpans, exp.GetResourceSpans()...)

    return reply, nil
}

type mockCollector struct {
    t *testing.T

    server    *grpc.Server
    traceSvc  *mockTraceService
    metricSvc *mockMetricService

    endpoint string
}

var _ collectortracepb.TraceServiceServer = &mockTraceService{}
var _ metricspb.MetricsServiceServer = &mockMetricService{}

func (mc *mockCollector) Stop() {
    mc.server.Stop()
}

func (mc *mockCollector) Get() ([]*tracepb.ResourceSpans, []*v1.ResourceMetrics) {
    return mc.traceSvc.getResourceSpans(), mc.metricSvc.getResourceMetrics()
}

func runMockCollector(t *testing.T) *mockCollector {
    ln, err := net.Listen("tcp", "localhost:0")
    if err != nil {
        t.Fatalf("Failed to get an endpoint: %v", err)
    }

    srv := grpc.NewServer()
    mc := &mockCollector{
        t:         t,
        traceSvc:  &mockTraceService{},
        metricSvc: &mockMetricService{},
    }

    collectortracepb.RegisterTraceServiceServer(srv, mc.traceSvc)
    metricspb.RegisterMetricsServiceServer(srv, mc.metricSvc)

    go func() {
        _ = srv.Serve(ln)
    }()

    addr := ln.Addr()

    // Wait for the server to start
    for {
        conn, err := net.Dial(addr.Network(), addr.String())
        if err == nil {
            _ = conn.Close()
            break
        }
        time.Sleep(10 * time.Millisecond)
    }

    mc.server = srv
    mc.endpoint = addr.String()

    return mc
}
