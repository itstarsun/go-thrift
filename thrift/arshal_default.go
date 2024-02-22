package thrift

import (
	"reflect"
	"sync"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

var (
	setType   = reflect.TypeOf((*interface{ set() })(nil)).Elem()
	listType  = reflect.TypeOf((*interface{ list() })(nil)).Elem()
	bytesType = reflect.TypeOf((*[]byte)(nil)).Elem()
	uuidType  = reflect.TypeOf((*[16]byte)(nil)).Elem()
)

func makeDefaultArshaler(t reflect.Type) *arshaler {
	switch t.Kind() {
	case reflect.Bool:
		return makeBoolArshaler(t)
	case reflect.Float32, reflect.Float64:
		return makeDoubleArshaler(t)
	case reflect.Int8:
		return makeIntArshaler(t, thriftwire.Byte, "WriteByte", thriftwire.Writer.WriteByte, "ReadByte", thriftwire.Reader.ReadByte)
	case reflect.Int16:
		return makeIntArshaler(t, thriftwire.I16, "WriteI16", thriftwire.Writer.WriteI16, "ReadI16", thriftwire.Reader.ReadI16)
	case reflect.Int32:
		return makeIntArshaler(t, thriftwire.I32, "WriteI32", thriftwire.Writer.WriteI32, "ReadI32", thriftwire.Reader.ReadI32)
	case reflect.Int64:
		return makeIntArshaler(t, thriftwire.I64, "WriteI64", thriftwire.Writer.WriteI64, "ReadI64", thriftwire.Reader.ReadI64)
	case reflect.Uint8:
		return makeUintArshaler(t, thriftwire.Byte, "WriteByte", thriftwire.Writer.WriteByte, "ReadByte", thriftwire.Reader.ReadByte)
	case reflect.Uint16:
		return makeUintArshaler(t, thriftwire.I16, "WriteI16", thriftwire.Writer.WriteI16, "ReadI16", thriftwire.Reader.ReadI16)
	case reflect.Uint32:
		return makeUintArshaler(t, thriftwire.I32, "WriteI32", thriftwire.Writer.WriteI32, "ReadI32", thriftwire.Reader.ReadI32)
	case reflect.Uint64:
		return makeUintArshaler(t, thriftwire.I64, "WriteI64", thriftwire.Writer.WriteI64, "ReadI64", thriftwire.Reader.ReadI64)
	case reflect.String:
		return makeStringArshaler(t)
	case reflect.Struct:
		return makeStructArshaler(t)
	case reflect.Map:
		return makeMapArshaler(t)
	case reflect.Slice:
		switch {
		case t.Implements(setType):
			return makeSetArshaler(t)
		case t.Implements(listType):
			return makeListArshaler(t)
		case t.AssignableTo(bytesType):
			return makeBytesArshaler(t)
		default:
			return makeListArshaler(t)
		}
	case reflect.Array:
		if t == uuidType {
			return makeUUIDArshaler(t)
		}
		return makeInvalidArshaler(t)
	case reflect.Pointer:
		return makePointerArshaler(t)
	default:
		return makeInvalidArshaler(t)
	}
}

func makeBoolArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	fncs.wireType = thriftwire.Bool
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		err := w.WriteBool(va.Bool())
		if err != nil {
			err := &wireError{action: "WriteBool", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Bool, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != thriftwire.Bool {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		v, err := r.ReadBool()
		if err != nil {
			err := &wireError{action: "ReadBool", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Bool, GoType: t, Err: err}
		}
		va.SetBool(v)
		return nil
	}
	return &fncs
}

func makeDoubleArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	fncs.wireType = thriftwire.Double
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		err := w.WriteDouble(va.Float())
		if err != nil {
			err := &wireError{action: "WriteDouble", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Double, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != thriftwire.Double {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		v, err := r.ReadDouble()
		if err != nil {
			err := &wireError{action: "ReadDouble", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Double, GoType: t, Err: err}
		}
		va.SetFloat(v)
		return nil
	}
	return &fncs
}

type ints interface {
	byte | int16 | int32 | int64
}

func makeIntArshaler[T ints](
	t reflect.Type,
	wt thriftwire.Type,
	writeAction string,
	write func(thriftwire.Writer, T) error,
	readAction string,
	read func(thriftwire.Reader) (T, error),
) *arshaler {
	return makeIntArshaler2(t, wt, writeAction, write, readAction, read, addressableValue.Int, addressableValue.SetInt)
}

func makeUintArshaler[T ints](
	t reflect.Type,
	wt thriftwire.Type,
	writeAction string,
	write func(thriftwire.Writer, T) error,
	readAction string,
	read func(thriftwire.Reader) (T, error),
) *arshaler {
	return makeIntArshaler2(t, wt, writeAction, write, readAction, read, addressableValue.Uint, addressableValue.SetUint)
}

func makeIntArshaler2[T ints, U int64 | uint64](
	t reflect.Type,
	wireType thriftwire.Type,
	writeAction string,
	write func(thriftwire.Writer, T) error,
	readAction string,
	read func(thriftwire.Reader) (T, error),
	get func(addressableValue) U,
	set func(addressableValue, U),
) *arshaler {
	var fncs arshaler
	fncs.wireType = wireType
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		err := write(w, T(get(va)))
		if err != nil {
			err := &wireError{action: writeAction, err: err}
			return &SemanticError{action: "marshal", ThriftType: wireType, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != wireType {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		v, err := read(r)
		if err != nil {
			err := &wireError{action: readAction, err: err}
			return &SemanticError{action: "unmarshal", ThriftType: wireType, GoType: t, Err: err}
		}
		set(va, U(v))
		return nil
	}
	return &fncs
}

func makeStringArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	fncs.wireType = thriftwire.String
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		err := w.WriteString(va.String())
		if err != nil {
			err := &wireError{action: "WriteString", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.String, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != thriftwire.String {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		v, err := r.ReadString()
		if err != nil {
			err := &wireError{action: "ReadString", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.String, GoType: t, Err: err}
		}
		va.SetString(v)
		return nil
	}
	return &fncs
}

func makeBytesArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	fncs.wireType = thriftwire.String
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		err := w.WriteBytes(va.Bytes())
		if err != nil {
			err := &wireError{action: "WriteBytes", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.String, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != thriftwire.String {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		v, err := r.ReadBytes(va.Bytes()[:0])
		if err != nil {
			err := &wireError{action: "ReadBytes", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.String, GoType: t, Err: err}
		}
		va.SetBytes(v)
		return nil
	}
	return &fncs
}

func makeUUIDArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	fncs.wireType = thriftwire.UUID
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		err := w.WriteUUID((*[16]byte)(va.Bytes()))
		if err != nil {
			err := &wireError{action: "WriteUUID", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.UUID, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != thriftwire.UUID {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		err := r.ReadUUID((*[16]byte)(va.Bytes()))
		if err != nil {
			err := &wireError{action: "ReadUUID", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.UUID, GoType: t, Err: err}
		}
		return nil
	}
	return &fncs
}

func makeStructArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	var (
		once    sync.Once
		fields  structFields
		errInit *SemanticError
	)
	init := func() {
		fields, errInit = makeStructFields(t)
	}
	fncs.wireType = thriftwire.Struct
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		once.Do(init)
		if errInit != nil {
			err := *errInit // shallow copy SemanticError
			err.action = "marshal"
			return &err
		}
		err := w.WriteStructBegin(thriftwire.StructHeader{
			Name: t.Name(),
		})
		if err != nil {
			err := &wireError{action: "WriteStructBegin", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
		}
		for i := range fields.sorted {
			f := &fields.sorted[i]
			if f.fncs.wireType == thriftwire.Stop {
				return &SemanticError{action: "marshal", GoType: f.typ}
			}
			v := addressableValue{va.Field(f.index[0])} // addressable if struct value is addressable
			if len(f.index) > 1 {
				v = v.fieldByIndex(f.index[1:], false)
				if !v.IsValid() {
					continue // implies a nil inlined field
				}
			}
			if !f.required && ((f.isZero == nil && v.IsZero()) || (f.isZero != nil && f.isZero(v))) {
				continue
			}
			if err := w.WriteFieldBegin(thriftwire.FieldHeader{
				Name: f.name,
				Type: f.fncs.wireType,
				ID:   f.id,
			}); err != nil {
				err := &wireError{action: "WriteFieldBegin", err: err}
				return &SemanticError{action: "marshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
			}
			if err := f.fncs.marshal(w, v, mo); err != nil {
				return err
			}
			if err := w.WriteFieldEnd(); err != nil {
				err := &wireError{action: "WriteFieldEnd", err: err}
				return &SemanticError{action: "marshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
			}
		}
		err = w.WriteStructEnd()
		if err != nil {
			err := &wireError{action: "WriteStructEnd", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		once.Do(init)
		if errInit != nil {
			err := *errInit // shallow copy SemanticError
			err.action = "unmarshal"
			return &err
		}
		if wt != thriftwire.Struct {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		_, err := r.ReadStructBegin()
		if err != nil {
			err := &wireError{action: "ReadStructBegin", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
		}
		for {
			h, err := r.ReadFieldBegin()
			if err != nil {
				err := &wireError{action: "ReadFieldBegin", err: err}
				return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
			}
			if h.Type == thriftwire.Stop {
				break
			}
			f, ok := fields.byID[h.ID]
			if !ok {
				if err := thriftwire.Skip(r, h.Type); err != nil {
					return err
				}
			} else {
				v := addressableValue{va.Field(f.index[0])} // addressable if struct value is addressable
				if len(f.index) > 1 {
					v = v.fieldByIndex(f.index[1:], true)
				}
				if err := f.fncs.unmarshal(r, v, uo, h.Type); err != nil {
					return err
				}
			}
			if err := r.ReadFieldEnd(); err != nil {
				err := &wireError{action: "ReadFieldEnd", err: err}
				return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
			}
		}
		err = r.ReadStructEnd()
		if err != nil {
			err := &wireError{action: "ReadStructEnd", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Struct, GoType: t, Err: err}
		}
		return nil
	}
	return &fncs
}

func (va addressableValue) fieldByIndex(index []int, mayAlloc bool) addressableValue {
	for _, i := range index {
		va = va.indirect(mayAlloc)
		if !va.IsValid() {
			return va
		}
		va = addressableValue{va.Field(i)} // addressable if struct value is addressable
	}
	return va
}

func (va addressableValue) indirect(mayAlloc bool) addressableValue {
	if va.Kind() == reflect.Pointer {
		if va.IsNil() {
			if !mayAlloc {
				return addressableValue{}
			}
			va.Set(reflect.New(va.Type().Elem()))
		}
		va = addressableValue{va.Elem()} // dereferenced pointer is always addressable
	}
	return va
}

func makeMapArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	keyFncs := lookupArshaler(t.Key())
	valFncs := lookupArshaler(t.Elem())
	fncs.wireType = thriftwire.Map
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		if keyFncs.wireType == thriftwire.Stop || valFncs.wireType == thriftwire.Stop {
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Map, GoType: t}
		}
		n := va.Len()
		err := w.WriteMapBegin(thriftwire.MapHeader{
			Key:   keyFncs.wireType,
			Value: valFncs.wireType,
			Size:  n,
		})
		if err != nil {
			err := &wireError{action: "WriteMapBegin", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Map, GoType: t, Err: err}
		}
		if n > 0 {
			k := newAddressableValue(t.Key())
			v := newAddressableValue(t.Elem())
			for iter := va.MapRange(); iter.Next(); {
				k.SetIterKey(iter)
				if err = keyFncs.marshal(w, k, mo); err != nil {
					return err
				}
				v.SetIterValue(iter)
				if err = valFncs.marshal(w, v, mo); err != nil {
					return err
				}
			}
		}
		err = w.WriteMapEnd()
		if err != nil {
			err := &wireError{action: "WriteMapEnd", err: err}
			return &SemanticError{action: "marshal", ThriftType: thriftwire.Map, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != thriftwire.Map {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		h, err := r.ReadMapBegin()
		if err != nil {
			err := &wireError{action: "ReadMapBegin", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Map, GoType: t, Err: err}
		}
		if h.Size > 0 {
			if va.IsNil() {
				va.Set(reflect.MakeMap(t))
			}
			k := newAddressableValue(t.Key())
			v := newAddressableValue(t.Elem())
			for i := 0; i < h.Size; i++ {
				k.SetZero()
				if err = keyFncs.unmarshal(r, k, uo, h.Key); err != nil {
					return err
				}
				if v2 := va.MapIndex(k.Value); v2.IsValid() {
					v.Set(v2)
				} else {
					v.SetZero()
				}
				err = valFncs.unmarshal(r, v, uo, h.Value)
				va.SetMapIndex(k.Value, v.Value)
				if err != nil {
					return err
				}
			}
		}
		err = r.ReadMapEnd()
		if err != nil {
			err := &wireError{action: "ReadMapEnd", err: err}
			return &SemanticError{action: "unmarshal", ThriftType: thriftwire.Map, GoType: t, Err: err}
		}
		return nil
	}
	return &fncs
}

func makeSetListArshaler[H thriftwire.SetHeader | thriftwire.ListHeader](
	t reflect.Type,
	wireType thriftwire.Type,
	writeBeginAction string,
	writeBegin func(thriftwire.Writer, H) error,
	writeEndAction string,
	writeEnd func(thriftwire.Writer) error,
	readBeginAction string,
	readBegin func(thriftwire.Reader) (H, error),
	readEndAction string,
	readEnd func(thriftwire.Reader) error,
) *arshaler {
	var fncs arshaler
	valFncs := lookupArshaler(t.Elem())
	fncs.wireType = wireType
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		if valFncs.wireType == thriftwire.Stop {
			return &SemanticError{action: "marshal", ThriftType: wireType, GoType: t}
		}
		n := va.Len()
		err := writeBegin(w, H{
			Element: valFncs.wireType,
			Size:    n,
		})
		if err != nil {
			err := &wireError{action: writeBeginAction, err: err}
			return &SemanticError{action: "marshal", ThriftType: wireType, GoType: t, Err: err}
		}
		for i := 0; i < n; i++ {
			v := addressableValue{va.Index(i)} // indexed slice element is always addressable
			if err := valFncs.marshal(w, v, mo); err != nil {
				return err
			}
		}
		err = writeEnd(w)
		if err != nil {
			err := &wireError{action: writeEndAction, err: err}
			return &SemanticError{action: "marshal", ThriftType: wireType, GoType: t, Err: err}
		}
		return nil
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if wt != wireType {
			return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
		}
		h, err := readBegin(r)
		if err != nil {
			err := &wireError{action: readBeginAction, err: err}
			return &SemanticError{action: "unmarshal", ThriftType: wireType, GoType: t, Err: err}
		}
		sh := thriftwire.SetHeader(h)
		if sh.Size > 0 {
			mustZero := true // we do not know the cleanliness of unused capacity
			cap := va.Cap()
			if cap > 0 {
				va.SetLen(cap)
			}
			var i int
			for i < sh.Size {
				if i == cap {
					va.Grow(1)
					cap = va.Cap()
					va.SetLen(cap)
					mustZero = false // reflect.Value.Grow ensures new capacity is zero-initialized
				}
				v := addressableValue{va.Index(i)} // indexed slice element is always addressable
				i++
				if mustZero {
					v.SetZero()
				}
				if err = valFncs.unmarshal(r, v, uo, sh.Element); err != nil {
					va.SetLen(i)
					return err
				}
			}
			va.SetLen(i)
		} else {
			va.SetLen(0)
		}
		err = readEnd(r)
		if err != nil {
			err := &wireError{action: readEndAction, err: err}
			return &SemanticError{action: "unmarshal", ThriftType: wireType, GoType: t, Err: err}
		}
		return nil
	}
	return &fncs
}

func makeSetArshaler(t reflect.Type) *arshaler {
	return makeSetListArshaler(t,
		thriftwire.Set,
		"WriteSetBegin",
		thriftwire.Writer.WriteSetBegin,
		"WriteSetEnd",
		thriftwire.Writer.WriteSetEnd,
		"ReadSetBegin",
		thriftwire.Reader.ReadSetBegin,
		"ReadSetEnd",
		thriftwire.Reader.ReadSetEnd,
	)
}

func makeListArshaler(t reflect.Type) *arshaler {
	return makeSetListArshaler(t,
		thriftwire.List,
		"WriteListBegin",
		thriftwire.Writer.WriteListBegin,
		"WriteListEnd",
		thriftwire.Writer.WriteListEnd,
		"ReadListBegin",
		thriftwire.Reader.ReadListBegin,
		"ReadListEnd",
		thriftwire.Reader.ReadListEnd,
	)
}

func makePointerArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	valFncs := lookupArshaler(t.Elem())
	fncs.wireType = valFncs.wireType
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		if va.IsNil() {
			v := newAddressableValue(t.Elem())
			return valFncs.marshal(w, v, mo)
		}
		v := addressableValue{va.Elem()} // dereferenced pointer is always addressable
		return valFncs.marshal(w, v, mo)
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		if va.IsNil() {
			va.Set(reflect.New(t.Elem()))
		}
		v := addressableValue{va.Elem()} // dereferenced pointer is always addressable
		return valFncs.unmarshal(r, v, uo, wt)
	}
	return &fncs
}

func makeInvalidArshaler(t reflect.Type) *arshaler {
	var fncs arshaler
	fncs.marshal = func(w thriftwire.Writer, va addressableValue, mo marshalOptions) error {
		return &SemanticError{action: "marshal", GoType: t}
	}
	fncs.unmarshal = func(r thriftwire.Reader, va addressableValue, uo unmarshalOptions, wt thriftwire.Type) error {
		return &SemanticError{action: "unmarshal", ThriftType: wt, GoType: t}
	}
	return &fncs
}
