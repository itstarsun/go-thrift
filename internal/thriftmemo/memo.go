package thriftmemo

import (
	"errors"
	"reflect"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

//go:generate go run memogen.go

// ErrBreak is returned when the step has reached its breakpoint.
var ErrBreak = errors.New("break")

// A Memo tracks [thriftwire.Reader] and [thriftwire.Writer].
type Memo struct {
	Breakpoint int // <= 0 means no breakpoint
	r, w       int

	steps  []string
	values [][]any
}

// Reset resets m.
func (m *Memo) Reset() {
	*m = Memo{
		steps:  m.steps[:0],
		values: m.values[:0],
	}
}

// Steps returns a copy of tracked steps.
func (m *Memo) Steps() []string {
	return append([]string(nil), m.steps...)
}

// Reader returns a [thriftwire.Reader] that is tracked by m.
//
// Any call to the methods of the returned [thriftwire.Reader] invalidates
// the [thriftwire.Writer] returned by [Memo.Writer].
func (m *Memo) Reader() thriftwire.Reader {
	return (*memoReader)(m)
}

// Writer returns a [thriftwire.Writer] that is tracked by m.
//
// The returned [thriftwire.Writer] is no longer valid if any method of
// the [thriftwire.Reader] returned by [Memo.Reader] is called.
func (m *Memo) Writer() thriftwire.Writer {
	return (*memoWriter)(m)
}

func (m *Memo) read(step string, values ...any) {
	m.steps = append(m.steps, step)
	r := m.r - 1
	if want := m.steps[r]; step != want {
		panic("invalid step: " + step + ", want " + want)
	}
	for i, v := range values {
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(m.values[r][i]))
	}
}

func (m *Memo) write(step string, values ...any) {
	m.steps = append(m.steps, step)
	m.values = append(m.values, values)
}

func (m *Memo) advance(p *int) error {
	if m.Breakpoint > 0 && *p+1 >= m.Breakpoint {
		return ErrBreak
	}
	*p++
	return nil
}

type memoReader Memo

func (m *memoReader) memo() *Memo {
	return (*Memo)(m)
}

func (m *memoReader) advance() error {
	return m.memo().advance(&m.r)
}

type memoWriter Memo

func (m *memoWriter) memo() *Memo {
	return (*Memo)(m)
}

func (m *memoWriter) advance() error {
	return m.memo().advance(&m.w)
}
