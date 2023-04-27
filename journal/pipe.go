package journal

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

// Writer is an interface for writing to a journal. It is a combination of the io.Writer and io.Closer interfaces.
// This interface intended to be used by the dagger client to write to the journal.
type Writer interface {
	io.Writer
	io.Closer
}

// Reader is an interface for reading from a journal.
type Reader interface {
	ReadEntry() (*Entry, bool)
}

var (
	_ Writer = new(unboundedPipe)
	_ Reader = new(unboundedPipe)
)

// unboundedPipe is a simple implementation of the Writer and Reader interfaces.
type unboundedPipe struct {
	cond    *sync.Cond
	counter *atomic.Uint64
	buffer  []*Entry
	closed  bool
}

// Pipe creates a new Writer and Reader pair that can be used to write and read from a journal.
func Pipe() (Writer, Reader) {
	pipe := &unboundedPipe{
		cond:    sync.NewCond(new(sync.Mutex)),
		counter: &atomic.Uint64{},
	}
	return pipe, pipe
}

func (p *unboundedPipe) Write(data []byte) (n int, err error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	raw := string(data)

	scanner := bufio.NewScanner(strings.NewReader(raw))

	// Loop through each line and process it
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		p.buffer = append(p.buffer, parseEntry(int(p.counter.Add(1)), line))
		p.cond.Signal()
	}

	return len(data), nil
}

func (p *unboundedPipe) ReadEntry() (*Entry, bool) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	for len(p.buffer) == 0 && !p.closed {
		p.cond.Wait()
	}

	if len(p.buffer) == 0 && p.closed {
		return nil, false
	}

	value := p.buffer[0]
	p.buffer = p.buffer[1:]
	return value, true
}

func (p *unboundedPipe) Close() error {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.closed = true
	p.cond.Broadcast()
	return nil
}
