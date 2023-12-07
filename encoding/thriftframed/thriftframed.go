// Package thriftframed implements the Thrift Framed protocol encoding.
package thriftframed

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

// New returns a new framed [thriftwire.Protocol] that uses p
// as the underlying [thriftwire.Protocol].
func New(p thriftwire.Protocol) thriftwire.Protocol {
	return protocol{p}
}

type protocol struct {
	thriftwire.Protocol
}

func (p protocol) NewReader(r io.Reader) thriftwire.Reader {
	var x reader
	x.Reader = p.Protocol.NewReader(&x.msg)
	x.lr.R = r
	return &x
}

func (p protocol) NewWriter(w io.Writer) thriftwire.Writer {
	var x writer
	x.Writer = p.Protocol.NewWriter(&x.msg)
	x.w = w
	return &x
}

var _ thriftwire.Protocol = (*protocol)(nil)

type reader struct {
	thriftwire.Reader
	lr  io.LimitedReader
	msg bytes.Buffer
	buf [4]byte
}

func (x *reader) ReadMessageBegin() (h thriftwire.MessageHeader, err error) {
	if x.msg.Len() > 0 {
		// The previous message is not fully consumed.
		return h, fmt.Errorf("thriftframed: ReadMessageBegin called without matching ReadMessageEnd")
	}
	if err := x.readMessage(); err != nil {
		return h, err
	}
	return x.Reader.ReadMessageBegin()
}

// ReadMessageEnd guarantees that the message is fully consumed (x.msg.Len() == 0).
func (x *reader) ReadMessageEnd() error {
	if err := x.Reader.ReadMessageEnd(); err != nil {
		return err
	}
	if x.msg.Len() > 0 {
		return fmt.Errorf("thriftframed: the underlying Reader does not fully consumed the message")
	}
	return nil
}

func (x *reader) readMessage() error {
	n, err := x.readUint32()
	if err != nil {
		return err
	}
	x.lr.N = int64(n)
	if _, err := io.Copy(&x.msg, &x.lr); err != nil {
		return err
	}
	return nil
}

func (x *reader) readUint32() (uint32, error) {
	buf := x.buf[:4]
	_, err := io.ReadFull(x.lr.R, buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func (x *reader) Reset(r io.Reader) {
	x.Reader.Reset(&x.msg)
	x.msg.Reset()
	x.lr.R = r
}

type writer struct {
	thriftwire.Writer
	w   io.Writer
	msg bytes.Buffer
	buf [4]byte
}

func (x *writer) WriteMessageBegin(h thriftwire.MessageHeader) error {
	if x.msg.Len() > 0 {
		// The previous message is not fully sent.
		return fmt.Errorf("thriftframed: WriteMessageBegin called twice without matching WriteMessageEnd")
	}
	return x.Writer.WriteMessageBegin(h)
}

// WriteMessageEnd guarantees that the message is fully sent (x.msg.Len() == 0).
func (x *writer) WriteMessageEnd() error {
	if err := x.Writer.WriteMessageEnd(); err != nil {
		return err
	}
	return x.writeMessage()
}

func (x *writer) writeMessage() error {
	if err := x.writeUint32(uint32(x.msg.Len())); err != nil {
		return err
	}
	if _, err := x.msg.WriteTo(&x.msg); err != nil {
		return err
	}
	return nil
}

func (x *writer) writeUint32(v uint32) error {
	buf := x.buf[:4]
	binary.BigEndian.PutUint32(buf, v)
	_, err := x.w.Write(buf)
	return err
}

func (x *writer) Flush() error {
	return thriftwire.Flush(x.w)
}

func (x *writer) Reset(w io.Writer) {
	x.Writer.Reset(&x.msg)
	x.msg.Reset()
	x.w = w
}
