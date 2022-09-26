package utils

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
)

// AbsoluteTimeSec is time in milliseconds since the Unix epoch
type AbsoluteTimeSec int64

func (at AbsoluteTimeSec) ToTime() time.Time {
    return time.Unix(int64(at), 0)
}

func (at AbsoluteTimeSec) ToUnix() int64 {
    return int64(at)
}

func FromTimeSec(tm time.Time) AbsoluteTimeSec {
    return AbsoluteTimeSec(tm.Unix())
}

func StaticClock(sec int64) func() time.Time {
    return func() time.Time {
        return time.Unix(sec, 0)
    }
}

func MakeRandomStr(numBytes int) string {
    bytesSlice := make([]byte, numBytes)
    _, err := rand.Read(bytesSlice)
    PanicIfF(err != nil, "failed to read random numbers: %s", err)
    return hex.EncodeToString(bytesSlice)
}

func PanicIfF(cond bool, msg string, args ...interface{}) {
    if cond {
        panic(fmt.Sprintf(msg, args...))
    }
}

func PanicIfErr(err error) {
    if err != nil {
        panic(fmt.Sprintf("%v", err))
    }
}

type Cleaner[T any] struct {
    cleanup func() T
}

func NewCleanup1[T any](cl func() T) *Cleaner[T] {
    return &Cleaner[T]{cleanup: cl}
}

func NewCleanup0(cl func()) *Cleaner[bool] {
    thunk := func() bool {
        cl()
        return true
    }
    return &Cleaner[bool]{cleanup: thunk}
}

func (c *Cleaner[T]) Disarm() {
    c.cleanup = nil
}

func (c *Cleaner[T]) Cleanup() {
    if c.cleanup != nil {
        c.cleanup()
    }
}
