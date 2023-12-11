package thrift

import (
	"errors"
	"reflect"
	"sync"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

type marshalOptions struct{}

// Marshal serializes a Go value into a [thriftwire.Writer].
func Marshal(out thriftwire.Writer, in any) error {
	return marshalOptions{}.Marshal(out, in)
}

func (mo marshalOptions) Marshal(out thriftwire.Writer, in any) error {
	v := reflect.ValueOf(in)
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		var t reflect.Type
		if v.Kind() == reflect.Pointer {
			t = v.Type().Elem()
		}
		err := errors.New("value must be passed as a non-nil pointer reference")
		return &SemanticError{action: "marshal", GoType: t, Err: err}
	}
	// Shallow copy non-pointer values to obtain an addressable value.
	// It is beneficial to performance to always pass pointers to avoid this.
	if v.Kind() != reflect.Pointer {
		v2 := reflect.New(v.Type())
		v2.Elem().Set(v)
		v = v2
	}
	va := addressableValue{v.Elem()}
	fncs := lookupArshaler(va.Type())
	return fncs.marshal(out, va, mo)
}

type unmarshalOptions struct{}

// Unmarshal deserializes a Go value from a [thriftwire.Reader].
func Unmarshal(in thriftwire.Reader, out any) error {
	return unmarshalOptions{}.Unmarshal(in, out)
}

func (uo unmarshalOptions) Unmarshal(in thriftwire.Reader, out any) error {
	v := reflect.ValueOf(out)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		var t reflect.Type
		if v.IsValid() {
			t = v.Type()
			if t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
		}
		err := errors.New("value must be passed as a non-nil pointer reference")
		return &SemanticError{action: "unmarshal", GoType: t, Err: err}
	}
	va := addressableValue{v.Elem()}
	fncs := lookupArshaler(va.Type())
	return fncs.unmarshal(in, va, uo, fncs.wireType)
}

// addressableValue is a reflect.Value that is guaranteed to be addressable
// such that calling the Addr and Set methods do not panic.
//
// There is no compile magic that enforces this property,
// but rather the need to construct this type makes it easier to examine each
// construction site to ensure that this property is upheld.
type addressableValue struct{ reflect.Value }

// newAddressableValue constructs a new addressable value of type t.
func newAddressableValue(t reflect.Type) addressableValue {
	return addressableValue{reflect.New(t).Elem()}
}

type (
	marshaler   = func(thriftwire.Writer, addressableValue, marshalOptions) error
	unmarshaler = func(thriftwire.Reader, addressableValue, unmarshalOptions, thriftwire.Type) error
)

type arshaler struct {
	wireType  thriftwire.Type
	marshal   marshaler
	unmarshal unmarshaler
}

var lookupArshalerCache sync.Map // map[reflect.Type]*arshaler

func lookupArshaler(t reflect.Type) *arshaler {
	if v, ok := lookupArshalerCache.Load(t); ok {
		return v.(*arshaler)
	}

	fncs := makeDefaultArshaler(t)

	// Use the last stored so that duplicate arshalers can be garbage collected.
	v, _ := lookupArshalerCache.LoadOrStore(t, fncs)
	return v.(*arshaler)
}
