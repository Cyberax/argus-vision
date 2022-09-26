package utils

import (
    "fmt"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestTime(t *testing.T) {
    clock := StaticClock(10000)
    assert.Equal(t, int64(10000), clock().Unix())
    assert.Equal(t, 0, clock().Nanosecond())

    at := FromTimeSec(clock())
    assert.Equal(t, int64(10000), at.ToUnix())
    assert.Equal(t, int64(10000), at.ToTime().Unix())
}

func TestPanicIfF(t *testing.T) {
    PanicIfF(false, "hello")

    assert.PanicsWithValue(t, "bad panic error", func() {
        PanicIfF(true, "bad panic %s", fmt.Errorf("error"))
    })

    assert.PanicsWithValue(t, "error", func() {
        PanicIfErr(fmt.Errorf("error"))
    })
}

func TestMakeRandomStr(t *testing.T) {
    assert.Equal(t, 20, len(MakeRandomStr(10)))
}

func TestCleaner(t *testing.T) {
    clean := false
    func() {
        cl := NewCleanup0(func() {
            clean = true
        })
        cl.Disarm()
    }()
    assert.False(t, clean)

    clean = false
    func() {
        cl := NewCleanup0(func() {
            clean = true
        })
        defer cl.Cleanup()
    }()
    assert.True(t, clean)

    clean = false
    func() {
        cl := NewCleanup1(func() error {
            clean = true
            return fmt.Errorf("ignored")
        })
        defer cl.Cleanup()
    }()
    assert.True(t, clean)
}
