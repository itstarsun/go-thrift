package thriftwire

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func testReadString(t *testing.T, readString func(*bufio.Reader, int) (string, error)) {
	test := func(t *testing.T, wantErr error, wantSize, readSize int) {
		b := bufio.NewReaderSize(nil, 0)
		want := strings.Repeat("a", b.Size()+wantSize)
		b.Reset(strings.NewReader(want))
		got, err := readString(b, len(want)+readSize)
		if err != wantErr {
			t.Errorf("got %v, want %v", err, wantErr)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	t.Run("Small", func(t *testing.T) {
		test(t, nil, -1, 0)
	})
	t.Run("SmallUnexpectedEOF", func(t *testing.T) {
		test(t, io.ErrUnexpectedEOF, -1, +1)
	})

	t.Run("Large", func(t *testing.T) {
		test(t, nil, +1, 0)
	})
	t.Run("LargeUnexpectedEOF", func(t *testing.T) {
		test(t, io.ErrUnexpectedEOF, +0, +1)
	})
}

func TestReadString(t *testing.T) {
	testReadString(t, ReadString)
}

func TestReadBytes(t *testing.T) {
	testReadString(t, func(r *bufio.Reader, n int) (string, error) {
		prefix := []byte("prefix")
		b, err := ReadBytes(r, n, prefix)
		b, hasPrefix := bytes.CutPrefix(b, prefix)
		if err == nil && !hasPrefix {
			err = fmt.Errorf("ReadBytes does not append to the original buffer")
		}
		return string(b), err
	})
}
