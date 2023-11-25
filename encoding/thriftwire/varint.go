package thriftwire

import (
	"bufio"
	"encoding/binary"
)

// ReadUvarint is like [binary.ReadUvarint] but optimized for [bufio.Reader].
func ReadUvarint(b *bufio.Reader) (uint64, error) {
	if buf, _ := b.Peek(binary.MaxVarintLen64); len(buf) > 0 {
		if x, n := binary.Uvarint(buf); n > 0 {
			_, err := b.Discard(n)
			return x, err
		}
	}
	return binary.ReadUvarint(b)
}

// ReadVarint is like [binary.ReadVarint] but optimized for [bufio.Reader].
func ReadVarint(b *bufio.Reader) (int64, error) {
	if buf, _ := b.Peek(binary.MaxVarintLen64); len(buf) > 0 {
		if x, n := binary.Varint(buf); n > 0 {
			_, err := b.Discard(n)
			return x, err
		}
	}
	return binary.ReadVarint(b)
}
