package logging

import (
    "encoding/json"
    "fmt"
    "github.com/stretchr/testify/assert"
    "strings"
    "testing"
)

func TestStackTraceStringer(t *testing.T) {
    trace := NewShortenedStackTrace(1, false, fmt.Errorf("test error"))
    assert.Equal(t, "test error", trace.Error())

    trace2 := NewShortenedStackTrace(1,false, "test error")
    assert.Equal(t, "test error", trace2.Error())

    trace3 := NewShortenedStackTrace(1,false, 123)
    assert.Equal(t, "<int Value>", trace3.Error())

    trace4 := NewShortenedStackTrace(1, false, nil)
    assert.Equal(t, "recovered from panic", trace4.Error())
}

func TestStackTrace(t *testing.T) {
    st := NewShortenedStackTrace(2, false,"Hello")
    jsStack := st.Field().Interface

    jsStr, err := json.Marshal(jsStack)
    assert.NoError(t, err)

    var res []StackElement
    err = json.Unmarshal(jsStr, &res)
    assert.NoError(t, err)

    assert.Equal(t, "TestStackTrace", res[0].Fn)
    // This line must contain the line number of the NewShortenedStackTrace call,
    // might break during refactorings
    assert.True(t, strings.HasSuffix(res[0].Fl, "short_stacktrace_test.go:26"))

    // Now read the string-based version
    strStack := strings.Split(st.StringStack(), "\n")
    assert.True(t, strings.HasSuffix(strStack[0], "short_stacktrace_test.go:26 TestStackTrace"))
}

func TestPanicSearch(t *testing.T) {
    defer func() {
        recover()
        st := NewShortenedStackTrace(0, true,"Hello")
        strStack := strings.Split(st.StringStack(), "\n")
        // Must be the line number of the panic() call. Might fail after refactoring.
        if !strings.HasSuffix(strStack[0], "short_stacktrace_test.go:67 TestPanicSearch") {
            t.Fatal("Stack is bad")
        }
    }()

    defer func() {
        p := recover()
        panic(p)
    }()

    defer func() {
        p := recover()
        panic(p)
    }()

    panic("Hello")
}

