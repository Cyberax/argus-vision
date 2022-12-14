package logging

import (
    "fmt"
    "go.uber.org/zap"
    "reflect"
    "runtime"
    "strconv"
    "strings"
)

type ShortenedStackTrace struct {
    skipToFirstPanic bool
    stack            []uintptr
    msg              string
}

// NewShortenedStackTrace creates a new shortened stack trace, that can optionally skip all the frames
// after the first panic() call (typically deferred error handlers).
func NewShortenedStackTrace(skipFrames int, skipToFirstPanic bool,
    msg interface{}) *ShortenedStackTrace {
    // Register the stack trace inside the tracer segment
    s := make([]uintptr, 40)
    n := runtime.Callers(skipFrames, s)

    return &ShortenedStackTrace{skipToFirstPanic: skipToFirstPanic,
        stack: s[:n], msg: convertPanicMsg(msg)}
}

func convertPanicMsg(msg interface{}) string {
    if msg == nil {
        return "recovered from panic"
    }
    stringer, ok := msg.(fmt.Stringer)
    if ok {
        return stringer.String()
    }
    err, ok := msg.(error)
    if ok {
        return err.Error()
    }
    return reflect.ValueOf(msg).String()
}

func (s *ShortenedStackTrace) Error() string {
    return s.msg
}

func (s *ShortenedStackTrace) StackTrace() []uintptr {
    return s.stack
}

type StackElement struct {
    Fl string
    Fn string
}

// JSONStack creates a nice stack trace, skipping all the deferred frames after the first panic() call.
// This method returns the list of structures that can be nicely reflected into JSON.
func (s *ShortenedStackTrace) JSONStack() []StackElement {
    // Create the stack trace
    stackElements := make([]StackElement, 0, 20)
    panicsToSkip := 0
    if s.skipToFirstPanic {
        panicsToSkip = s.countPanics()
    }

    // Note: On the last iteration, frames.Next() returns false, with a valid
    // frame, but we ignore this frame. The last frame is the runtime frame which
    // adds noise, since it's only either runtime.main or runtime.goexit.
    frames := runtime.CallersFrames(s.stack)
    for frame, more := frames.Next(); more; frame, more = frames.Next() {
        path, line, label := s.parseFrame(frame)

        if panicsToSkip > 0 && strings.HasPrefix(path, "runtime/panic") && label == "gopanic" {
            panicsToSkip -= 1
            continue
        }
        if panicsToSkip > 0 {
            continue
        }

        stackElements = append(stackElements, StackElement{
            Fl: path + ":" + strconv.Itoa(line),
            Fn: label,
        })
    }
    return stackElements
}

// StringStack creates a nice stack trace, skipping all the deferred frames after the first panic() call.
// This method returns a human-readable multi-line string.
func (s *ShortenedStackTrace) StringStack() string {
    // Create the stack trace
    frames := runtime.CallersFrames(s.stack)

    panicsToSkip := 0
    if s.skipToFirstPanic {
        panicsToSkip = s.countPanics()
    }

    var res string
    // Note: On the last iteration, frames.Next() returns false, with a valid
    // frame, but we ignore this frame. The last frame is the runtime frame which
    // adds noise, since it's only either runtime.main or runtime.goexit.
    for frame, more := frames.Next(); more; frame, more = frames.Next() {
        path, line, label := s.parseFrame(frame)

        if panicsToSkip > 0 && strings.HasPrefix(path, "runtime/panic") && label == "gopanic" {
            panicsToSkip -= 1
            continue
        }
        if panicsToSkip > 0 {
            continue
        }

        res += path + ":" + strconv.Itoa(line) + " " + label + "\n"
    }

    return res
}

// The default stack trace contains the build environment full path as the first part of the file name.
// This adds no information to the stack trace and exposes the building environment,
// so process the stack trace to remove the building environment path.
func (s *ShortenedStackTrace) parseFrame(frame runtime.Frame) (string, int, string) {
    path, line, label := frame.File, frame.Line, frame.Function

    // Strip GOPATH from the path by counting the number of separators in label & path
    // For example:
    //   GOPATH = /home/user
    //   path   = /home/user/src/pkg/sub/file.go
    //   label  = pkg/sub.Type.Method
    // We want to set the path to:
    //    pkg/sub/file.go
    i := len(path)
    for n, g := 0, strings.Count(label, "/")+2; n < g; n++ {
        i = strings.LastIndex(path[:i], "/")
        if i == -1 {
            // Something went wrong and the path has fewer separators than we expected
            // Abort and leave i as -1 to counteract the +1 below
            break
        }
    }
    path = path[i+1:] // Trim the initial /

    // Strip the path from the function name as it's already in the path
    label = label[strings.LastIndex(label, "/")+1:]
    // Likewise, strip the package name
    label = label[strings.Index(label, ".")+1:]

    return path, line, label
}

func (s *ShortenedStackTrace) Field() zap.Field {
    return zap.Reflect("stacktrace", s.JSONStack())
}

// Add the number of go panic() calls in the stack trace
func (s *ShortenedStackTrace) countPanics() int {
    frames := runtime.CallersFrames(s.stack)
    panics := 0
    for frame, more := frames.Next(); more; frame, more = frames.Next() {
        path, _, label := s.parseFrame(frame)
        if strings.HasPrefix(path, "runtime/panic") && label == "gopanic" {
            panics += 1
        }
    }
    return panics
}
