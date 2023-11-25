package thriftwire

import (
	"bufio"
	"io"
	"slices"
	"strings"
)

// Next is like [bufio.Reader.Peek], but advance the reader.
// If an [io.EOF] happens after reading some but not all the bytes,
// Next returns [io.ErrUnexpectedEOF].
func Next(b *bufio.Reader, n int) ([]byte, error) {
	buf, err := b.Peek(n)
	_, _ = b.Discard(len(buf))
	if len(buf) == n {
		err = nil
	} else if len(buf) > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return buf, err
}

// ReadString reads the next n bytes and returns them as a string.
func ReadString(b *bufio.Reader, n int) (string, error) {
	if b.Size() >= n {
		c, err := Next(b, n)
		return string(c), err
	}
	var sb strings.Builder
	sb.Grow(b.Size())
	left := n
	for left > 0 {
		c, err := Next(b, min(left, b.Size()))
		sb.Write(c)
		left -= len(c)
		if err != nil {
			if left != n && err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return sb.String(), err
		}
	}
	return sb.String(), nil
}

// ReadBytes reads the next n bytes, appends them to buf,
// and returns the extended buffer.
func ReadBytes(b *bufio.Reader, n int, buf []byte) ([]byte, error) {
	if b.Size() >= n {
		c, err := Next(b, n)
		return append(buf, c...), err
	}
	buf = slices.Grow(buf, b.Size())
	left := n
	for left > 0 {
		c, err := Next(b, min(left, b.Size()))
		buf = append(buf, c...)
		left -= len(c)
		if err != nil {
			if left != n && err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return buf, err
		}
	}
	return buf, nil
}
