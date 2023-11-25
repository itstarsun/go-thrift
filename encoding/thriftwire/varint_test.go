package thriftwire

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
	"testing"
)

func testReadVarint[T ~int64 | uint64](
	t *testing.T,
	append func([]byte, T) []byte,
	read func(*bufio.Reader) (T, error),
) {
	for n := 1; n <= 64; n <<= 1 {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			want := T(1) << (n - 1)
			buf := append(nil, want)
			r := bufio.NewReader(bytes.NewReader(buf))
			got, err := read(r)
			if err != nil {
				t.Errorf("got %v, want %v", err, nil)
			}
			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}

	t.Run("Overflow", func(t *testing.T) {
		r := bufio.NewReader(bytes.NewBuffer(bytes.Repeat([]byte{0x80}, binary.MaxVarintLen64+1)))
		_, err := read(r)
		if err == nil || !strings.Contains(err.Error(), "overflow") {
			t.Errorf("got %v, want overflow error", err)
		}
	})
}

func TestReadVarint(t *testing.T) {
	testReadVarint(t, binary.AppendVarint, ReadVarint)
}

func TestReadUvarint(t *testing.T) {
	testReadVarint(t, binary.AppendUvarint, ReadUvarint)
}
