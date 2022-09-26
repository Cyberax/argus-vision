package visibility

//func TestSecureContext(t *testing.T) {
//	c1 := context.Background()
//	assert.Equal(t, "", TryGetSecureId(c1))
//	assert.Panics(t, func() {
//		GetSecureId(c1)
//	})
//
//	c2 := SetSecureReqId(c1, "Hello")
//	assert.Equal(t, "Hello", GetSecureId(c2))
//	assert.Equal(t, "Hello", TryGetSecureId(c2))
//}

//func TestMetricsContext(t *testing.T) {
//	ctx := MakeMetricHelper(context.Background(), "TestOp")
//	mctx := GetMetricHelperFromContext(ctx)
//
//	// Add metric can be called first, without a SetMetric
//	mctx.AddMetric("zonk", 10, cwtypes.StandardUnitCount)
//
//	mctx.SetCount("count1", 11)
//	mctx.SetCount("count1", 12) // Will override
//	mctx.AddCount("count1", 2)
//
//	mctx.SetMetric("speed", 123, cwtypes.StandardUnitGigabitsSecond)
//	mctx.AddMetric("speed", 2, cwtypes.StandardUnitGigabitsSecond)
//
//	mctx.SetDuration("duration", time.Millisecond*500)
//	mctx.AddDuration("duration", time.Second*2)
//
//	bench := mctx.Benchmark("delay")
//	time.Sleep(500 * time.Millisecond)
//	bench.Done()
//
//	fakeSink := NewRecordingSink()
//	mctx.CopyToStatsd(fakeSink, "ThisClientType")
//
//	assert.Equal(t, "client-type:ThisClientType", fakeSink.Tags["TestOp.duration"][0])
//	assert.Equal(t, "TestOp", mctx.OpName)
//
//	assert.Equal(t, float64(14), fakeSink.Distributions["TestOp.count1"])
//
//	assert.True(t, fakeSink.Distributions["TestOp.delay"] > 0.5*1000000)
//	assert.Equal(t, "unit:microseconds", fakeSink.Tags["TestOp.delay"][1])
//
//	assert.Equal(t, 2.5*1e6, fakeSink.Distributions["TestOp.duration"])
//	assert.Equal(t, "unit:microseconds", fakeSink.Tags["TestOp.duration"][1])
//
//	assert.Equal(t, 125.0*1024*1024*1024, fakeSink.Distributions["TestOp.speed"])
//	assert.Equal(t, "unit:bits_per_second", fakeSink.Tags["TestOp.speed"][1])
//
//	assert.Equal(t, float64(10), fakeSink.Distributions["TestOp.zonk"])
//
//	z1, zu := mctx.GetMetric("zonk")
//	assert.Equal(t, 10.0, z1)
//	assert.Equal(t, cwtypes.StandardUnitCount, zu)
//	assert.Equal(t, 10.0, mctx.GetMetricVal("zonk"))
//
//	// Non-existing metric
//	assert.Equal(t, 0.0, mctx.GetMetricVal("badbad"))
//
//	mctx.Reset()
//	mctx.Reset() // Idempotent
//
//	z1, zu = mctx.GetMetric("zonk")
//	assert.Equal(t, 0.0, z1)
//	assert.Equal(t, cwtypes.StandardUnitNone, zu)
//	assert.Equal(t, 0.0, mctx.GetMetricVal("zonk"))
//}
//
//func TestMetricsSubmission(t *testing.T) {
//	ctx := context.Background()
//	ctx = MakeMetricHelper(ctx, "TestCtxOriginal") // An original context
//	ctx = MakeMetricHelper(ctx, "TestCtx")         // Save metrics into the context
//
//	for i := 0; i < 17; i++ {
//		mctx := GetMetricHelperFromContext(ctx)
//		assert.Equal(t, "TestCtx", mctx.OpName)
//		mctx.AddCount(fmt.Sprintf("count%d", i), 2)
//		mctx.AddMetric(fmt.Sprintf("met%d", i), float64(i), cwtypes.StandardUnitBytes)
//	}
//
//	fc := &FakeSpan{tags: map[string]interface{}{}}
//	GetMetricHelperFromContext(ctx).CopyToSpan(fc)
//
//	for i := 0; i < 17; i++ {
//		assert.Equal(t, float64(2), fc.tags[fmt.Sprintf("count%d", i)])
//		assert.Nil(t, fc.tags[fmt.Sprintf("count%d_unit", i)])
//
//		assert.Equal(t, float64(i), fc.tags[fmt.Sprintf("met%d", i)])
//		assert.Equal(t, "bytes", fc.tags[fmt.Sprintf("met%d_unit", i)])
//	}
//}
//
//func TestTaggedMetrics(t *testing.T) {
//	ctx := context.Background()
//	ctx = MakeMetricHelper(ctx, "root") // An original context
//
//	mctx := GetMetricHelperFromContext(ctx)
//	tgd := mctx.Tagged("queue", "gpu", "env", "alpha")
//	assert.Same(t, tgd, mctx.Tagged("env", "alpha", "queue", "gpu"))
//	assert.NotSame(t, tgd, mctx.Tagged("queue", "gpu"))
//	assert.NotSame(t, tgd, mctx.Tagged("queue", "cpu", "env", "alpha"))
//
//	tgd.Tag("node", "n-1")
//
//	tgd.AddCount("ops_num", 22)
//
//	fc := &FakeSpan{tags: map[string]interface{}{}}
//	mctx.CopyToSpan(fc)
//	fakeSink := NewRecordingSink()
//	mctx.CopyToStatsd(fakeSink, "TestType")
//
//	assert.Equal(t, float64(22), fc.tags["queue_gpu_ops_num"])
//	assert.Equal(t, float64(22), fc.tags["node_n-1_ops_num"])
//	assert.Equal(t, "client-type:TestType", fakeSink.Tags["root.ops_num"][0])
//	assert.Equal(t, "env:alpha", fakeSink.Tags["root.ops_num"][1])
//	assert.Equal(t, "node:n-1", fakeSink.Tags["root.ops_num"][2])
//	assert.Equal(t, "queue:gpu", fakeSink.Tags["root.ops_num"][3])
//	assert.Equal(t, "unit:count", fakeSink.Tags["root.ops_num"][4])
//	assert.Equal(t, "n-1", tgd.Tags["node"])
//	assert.Equal(t, "gpu", tgd.Tags["queue"])
//}
//
//func TestTagInheritance(t *testing.T) {
//	ctx := context.Background()
//	ctx = MakeMetricHelper(ctx, "root") // An original context
//
//	mctx := GetMetricHelperFromContext(ctx)
//	mctx.Tag("root", "yes")
//
//	tgd := mctx.Tagged("queue", "gpu", "root", "no")
//	assert.Same(t, tgd, mctx.Tagged("root", "no", "queue", "gpu"))
//	assert.NotSame(t, tgd, mctx.Tagged("queue", "cpu"))
//
//	tgd.AddCount("ops_num", 22)
//
//	fc := &FakeSpan{tags: map[string]interface{}{}}
//	mctx.CopyToSpan(fc)
//	fakeSink := NewRecordingSink()
//	mctx.CopyToStatsd(fakeSink, "TestType")
//
//	assert.Equal(t, float64(22), fc.tags["queue_gpu_ops_num"])
//	assert.Equal(t, "queue:gpu", fakeSink.Tags["root.ops_num"][1])
//	assert.Equal(t, "root:no", fakeSink.Tags["root.ops_num"][2])
//}
//
//func TestMetricPath(t *testing.T) {
//	ctx := context.Background()
//	ctx = MakeMetricHelper(ctx, "root") // An original context
//
//	mctx := GetMetricHelperFromContext(ctx)
//	mctx.AddCount("/ops_num", 22)
//
//	fakeSink := NewRecordingSink()
//	mctx.CopyToStatsd(fakeSink, "TestType")
//
//	assert.Equal(t, float64(22), fakeSink.Distributions["ops_num"])
//}
