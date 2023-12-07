package thriftframed

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

const (
	MessageBegin = "MessageBegin"
	MessageEnd   = "MessageEnd"
)

func TestWriter(t *testing.T) {
	var bf bufferFlusher
	var flushes int

	p := New(testProtocol{
		readMessageBegin: func(r io.Reader) error {
			return readString(r, MessageBegin)
		},
		readMessageEnd: func(r io.Reader) error {
			return readString(r, MessageEnd)
		},
		writeMessageBegin: func(w io.Writer) error {
			_, err := io.WriteString(w, MessageBegin)
			return err
		},
		writeMessageEnd: func(w io.Writer) error {
			_, err := io.WriteString(w, MessageEnd)
			return err
		},
		flush: func(w io.Writer) error {
			flushes++
			return nil
		},
	})

	w := p.NewWriter(&bf)

	if err := w.WriteMessageBegin(thriftwire.MessageHeader{}); err != nil {
		t.Fatal(err)
	}
	if flushes != 0 || bf.flushes != 0 {
		t.Errorf("got (%d, %d), want (%d, %d)", flushes, bf.flushes, 0, 0)
	}

	if err := w.WriteMessageEnd(); err != nil {
		t.Fatal(err)
	}
	if flushes != 1 || bf.flushes != 0 {
		t.Errorf("got (%d, %d), want (%d, %d)", flushes, bf.flushes, 1, 0)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if flushes != 1 || bf.flushes != 1 {
		t.Errorf("got (%d, %d), want (%d, %d)", flushes, bf.flushes, 1, 1)
	}

	r := p.NewReader(&bf)
	if _, err := r.ReadMessageBegin(); err != nil {
		t.Fatal(err)
	}
	if err := r.ReadMessageEnd(); err != nil {
		t.Fatal(err)
	}
}

func readString(r io.Reader, want string) error {
	buf := make([]byte, len(want))
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	if got := string(buf); got != want {
		return fmt.Errorf("got %q, want %q", got, want)
	}
	return nil
}

func TestErrors(t *testing.T) {
	t.Run("Reader", func(t *testing.T) {
		p := New(testProtocol{})
		r := p.NewReader(iotest.ErrReader(io.EOF)).(*reader)

		// Mocking that the previous message is not fully consumed
		// or that the underlying reader has not fully consumed the message
		r.msg.WriteString(MessageBegin)

		if _, err := r.ReadMessageBegin(); err == nil {
			t.Errorf("got %v, want error", err)
		}
		if err := r.ReadMessageEnd(); err == nil {
			t.Errorf("got %v, want error", err)
		}
	})

	t.Run("Writer", func(t *testing.T) {
		p := New(testProtocol{})
		w := p.NewWriter(io.Discard).(*writer)

		// Mocking that the previous message was not fully sent.
		w.msg.WriteString(MessageBegin)

		if err := w.WriteMessageBegin(thriftwire.MessageHeader{}); err == nil {
			t.Errorf("got %v, want error", err)
		}
	})
}

type bufferFlusher struct {
	bytes.Buffer
	flushes int
}

func (bf *bufferFlusher) Flush() error {
	bf.flushes++
	return nil
}

type testProtocol struct {
	readMessageBegin, readMessageEnd   func(io.Reader) error
	writeMessageBegin, writeMessageEnd func(io.Writer) error
	flush                              func(io.Writer) error
}

func (p testProtocol) NewReader(r io.Reader) thriftwire.Reader {
	return &testReader{testProtocol: p, r: r}
}

func (p testProtocol) NewWriter(w io.Writer) thriftwire.Writer {
	return &testWriter{testProtocol: p, w: w}
}

type testReader struct {
	testProtocol
	thriftwire.Reader
	r io.Reader
}

func (x *testReader) ReadMessageBegin() (h thriftwire.MessageHeader, _ error) {
	if f := x.readMessageBegin; f != nil {
		return h, f(x.r)
	}
	return h, nil
}

func (w *testReader) ReadMessageEnd() error {
	if f := w.readMessageEnd; f != nil {
		return f(w.r)
	}
	return nil
}

func (x *testReader) Reset(r io.Reader) {
	x.r = r
}

type testWriter struct {
	testProtocol
	thriftwire.Writer
	w io.Writer
}

func (x *testWriter) WriteMessageBegin(h thriftwire.MessageHeader) error {
	if f := x.writeMessageBegin; f != nil {
		return f(x.w)
	}
	return nil
}

func (x *testWriter) WriteMessageEnd() error {
	if f := x.writeMessageEnd; f != nil {
		return f(x.w)
	}
	return nil
}

func (x *testWriter) Flush() error {
	if f := x.flush; f != nil {
		return f(x.w)
	}
	return nil
}

func (x *testWriter) Reset(w io.Writer) {
	x.w = w
}
