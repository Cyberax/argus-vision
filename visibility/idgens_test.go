package visibility

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIdGen(t *testing.T) {
	generator := NewCryptoSafeRandIdGenerator(true)

	traceId, _ := generator.NewIDs(context.Background())

	// XRay-compatible trace IDs are prefixed by the current time
	val := traceId.String()[0:8]
	var curTime int64
	_, _ = fmt.Sscanf(val, "%x", &curTime)
	assert.True(t, time.Now().Unix()-curTime < 10)
}

func TestPredictableIdGen(t *testing.T) {
	gen := NewPredictableIdGen(11)
	t1, s1 := gen.NewIDs(context.Background())
	s2 := gen.NewSpanID(context.Background(), t1)

	assert.Equal(t, t1.String(), "590c14409888b5b07d51a817ee07c3f2")
	assert.Equal(t, s1.String(), "145935bc7155e3c7")
	assert.Equal(t, s2.String(), "a76490c3e0aa0b6a")
}
