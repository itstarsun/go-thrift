package thriftwire

// Skip skips over the next value.
func Skip(r Reader, t Type) (err error) {
	switch t {
	default:
		return InvalidTypeError(t)
	case Bool:
		_, err = r.ReadBool()
	case Byte:
		_, err = r.ReadByte()
	case Double:
		_, err = r.ReadDouble()
	case I16:
		_, err = r.ReadI16()
	case I32:
		_, err = r.ReadI32()
	case I64:
		_, err = r.ReadI64()
	case String:
		return r.SkipString()
	case Struct:
		if _, err = r.ReadStructBegin(); err != nil {
			return err
		}
		for {
			h, err := r.ReadFieldBegin()
			if err != nil {
				return err
			}
			if h.Type == Stop {
				break
			}
			if err := Skip(r, h.Type); err != nil {
				return err
			}
			if err := r.ReadFieldEnd(); err != nil {
				return err
			}
		}
		return r.ReadStructEnd()
	case Map:
		h, err := r.ReadMapBegin()
		if err != nil {
			return err
		}
		for i := 0; i < h.Size; i++ {
			if err := Skip(r, h.Key); err != nil {
				return err
			}
			if err := Skip(r, h.Value); err != nil {
				return err
			}
		}
		return r.ReadMapEnd()
	case Set:
		h, err := r.ReadSetBegin()
		if err != nil {
			return err
		}
		for i := 0; i < h.Size; i++ {
			if err := Skip(r, h.Element); err != nil {
				return err
			}
		}
		return r.ReadSetEnd()
	case List:
		h, err := r.ReadListBegin()
		if err != nil {
			return err
		}
		for i := 0; i < h.Size; i++ {
			if err := Skip(r, h.Element); err != nil {
				return err
			}
		}
		return r.ReadListEnd()
	case UUID:
		return r.SkipUUID()
	}
	return err
}
