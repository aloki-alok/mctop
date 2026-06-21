package mcp

import (
	"bytes"
	"strings"
	"sync"
	"time"
)

// A Frame is one JSON-RPC message that crossed the connection, tagged with its
// direction and the time it was seen. The trace view shows these so a user can
// watch the protocol the way the Inspector's network pane does, in the terminal.
type Frame struct {
	Dir  Direction
	Data string // the JSON-RPC message as one compact line
	At   time.Time
}

// Direction is which way a frame travelled.
type Direction int

const (
	Sent     Direction = iota // mctop to server
	Received                  // server to mctop
	Failed                    // a read or write error
)

// recorder is the io.Writer handed to the SDK's LoggingTransport. The transport
// writes one "write: {json}" or "read: {json}" line per message; the recorder
// parses those into frames. It keeps only the most recent maxFrames so a long
// session cannot grow without bound.
type recorder struct {
	mu     sync.Mutex
	frames []Frame
	buf    []byte
}

const maxFrames = 500

func (r *recorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf = append(r.buf, p...)
	for {
		i := bytes.IndexByte(r.buf, '\n')
		if i < 0 {
			break
		}
		r.record(string(r.buf[:i]))
		r.buf = r.buf[i+1:]
	}
	return len(p), nil
}

func (r *recorder) record(line string) {
	var dir Direction
	var data string
	switch {
	case strings.HasPrefix(line, "write: "):
		dir, data = Sent, line[len("write: "):]
	case strings.HasPrefix(line, "read: "):
		dir, data = Received, line[len("read: "):]
	case strings.HasPrefix(line, "write error: "), strings.HasPrefix(line, "read error: "):
		dir, data = Failed, line
	default:
		return
	}
	r.frames = append(r.frames, Frame{Dir: dir, Data: data, At: time.Now()})
	if len(r.frames) > maxFrames {
		r.frames = r.frames[len(r.frames)-maxFrames:]
	}
}

func (r *recorder) snapshot() []Frame {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Frame, len(r.frames))
	copy(out, r.frames)
	return out
}
