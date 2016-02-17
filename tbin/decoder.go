// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

//
// Go implementation of the tbin encoding format
//

package tbin

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/ardielle/ardielle-go/rdl"
)

func (d *Decoder) readHeader() error {
	tag := int(d.ParseUnsigned())
	if (tag & VersionTagMask) == VersionTag {
		d.dataVersion = (tag & VersionDataMask) + 1
		if d.dataVersion != CurrentVersion {
			d.err = fmt.Errorf("TBin version not supported: %d", d.dataVersion)
			return d.err
		}
		d.types = make([]*Signature, 0)
		return nil
	}
	d.err = fmt.Errorf("not a valid tbin file")
	return d.err
}

func (d *Decoder) Decode(data interface{}) error {
	rv := reflect.ValueOf(data)
	if !rv.IsNil() && rv.Kind() == reflect.Ptr {
		v := rv.Elem()
		if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
			//Empty interface, just do the generic decode to a map
			d, err := d.decode()
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(d))
			return nil
		}
		if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
			v = v.Addr()
		}
		if v.Kind() == reflect.Ptr {
			m, ok := v.Interface().(TBinUnmarshallable)
			if ok {
				err := m.UnmarshalTBin(d)
				return err
			}
			v = v.Elem()
		}
		return d.DecodeReflect(v)
	}
	return fmt.Errorf("Cannot decode into this: %v", data)
}

func (d *Decoder) Error() error {
	return d.err
}

func (d *Decoder) CurrentCount() int {
	return d.currentCount
}

func (d *Decoder) parseType() *Signature {
	tag := d.ParseUnsigned()
	if tag >= FirstUserTag {
		idx := int(tag - FirstUserTag)
		if idx >= len(d.types) {
			d.err = fmt.Errorf("ref to a undefined tag: 0x%02x", tag)
			return nil
		}
		return d.types[idx]
	}
	switch tag {
	case ArrayTag:
		//all arrays get typedef'd now, this code may be dead
		itemsType := d.parseType()
		return Array(itemsType)
	case MapTag:
		//all maps get typedef'd now, this code may be dead
		keysType := d.parseType()
		itemsType := d.parseType()
		return Map(keysType, itemsType)
	case DefArrayTag:
		itemType := d.parseType()
		return Array(itemType)
	case DefMapTag:
		keysType := d.parseType()
		itemsType := d.parseType()
		return Map(keysType, itemsType)
	case DefStructTag:
		size := int(d.ParseUnsigned())
		fields := make([]*FieldSignature, size)
		for i := 0; i < size; i++ {
			fname, _ := d.ParseString()
			ftype := d.parseType()
			fields[i] = Field(fname, ftype, (ftype == Any))
		}
		return Struct(fields...)
	case DefUnionTag:
		size := int(d.ParseUnsigned())
		var variants []*Signature
		for i := 0; i < size; i++ {
			variantType := d.parseType()
			variants = append(variants, variantType)
		}
		return Union(variants...)
	case DefEnumTag:
		size := int(d.ParseUnsigned())
		syms := []string{""}
		for i := 0; i < size; i++ {
			sym, err := d.ParseString()
			if err != nil {
				d.err = err
				return nil
			}
			syms = append(syms, sym)
		}
		return Enum(syms...)
	case BoolTag:
		return Bool
	case Int8Tag:
		return Int8
	case Int16Tag:
		return Int16
	case Int32Tag:
		return Int32
	case Int64Tag:
		return Int64
	case Float32Tag:
		return Float32
	case Float64Tag:
		return Float64
	case StringTag:
		return String
	case UUIDTag:
		return UUID
	case TimestampTag:
		return Timestamp
	case AnyTag:
		return Any
		//	case StructTag:
		//		//fix: it means symbol/any, the any needs a tag to parse
		//		panic("Struct!")
	default:
		d.err = fmt.Errorf("Unexpected tag definition type in TBin stream: 0x%2x", tag)
		return nil
	}
}

func (d *Decoder) decode() (interface{}, error) {
again:
	tag := d.ParseUnsigned()
	if d.err != nil {
		return nil, d.err
	}
	if tag >= FirstUserTag {
		idx := int(tag - FirstUserTag)
		if idx < len(d.types) {
			ttype := d.types[idx]
			return d.decodeType(ttype)
		}
		ttype := d.parseType()
		if ttype == nil {
			d.err = fmt.Errorf("First use of a user tag must be followed by a typedef.")
			return nil, d.err
		}
		d.types = append(d.types, ttype)
		goto again
	} else {
		if (tag & TinyStrTagMask) == TinyStrTag {
			n := tag & TinyStrDataMask
			tinybuf := make([]byte, n)
			d.err = d.readBytes(tinybuf)
			return string(tinybuf), d.err
		}
		switch tag {
		case NullTag:
			return nil, nil
		case BoolTag:
			n, err := d.in.ReadByte()
			if err != nil {
				return nil, err
			}
			if n != 0 {
				return true, nil
			}
			return false, nil
		case Int8Tag:
			n := int8(d.ParseInt())
			return n, d.err
		case Int16Tag:
			n := int16(d.ParseInt())
			return n, d.err
		case Int32Tag:
			n := int32(d.ParseInt())
			return n, d.err
		case Int64Tag:
			n := d.ParseInt64()
			return n, d.err
		case Float32Tag:
			return d.ParseFloat32()
		case Float64Tag:
			return d.ParseFloat64()
		case BytesTag:
			return d.ParseBytes()
		case StringTag:
			return d.ParseString()
		case SymbolTag:
			return d.ParseSymbol()
		case TimestampTag:
			return d.ParseTimestamp()
		case UUIDTag:
			return d.ParseUUID()
		case StructTag:
			return d.DecodeStruct()
		case ArrayTag:
			return d.DecodeArray()
		case MapTag:
			return d.DecodeMap()
		}
	}
	return nil, fmt.Errorf("Unexpected tag value: 0x%02x", tag)
}

func (d *Decoder) DecodeStruct() (map[string]interface{}, error) {
	nfields := int(d.ParseUnsigned())
	result := make(map[string]interface{})
	for i := 0; i < nfields; i++ {
		name, _ := d.ParseSymbol()
		val, _ := d.decode()
		result[name] = val
	}
	return result, d.err
}

func (d *Decoder) DecodeArray() ([]interface{}, error) {
	count := int(d.ParseUnsigned())
	var result []interface{}
	for i := 0; i < count; i++ {
		val, _ := d.decode()
		result = append(result, val)
	}
	return result, d.err
}

func (d *Decoder) DecodeMap() (map[string]interface{}, error) {
	count := int(d.ParseUnsigned())
	result := make(map[string]interface{})
	for i := 0; i < count; i++ {
		key, _ := d.decode()
		val, _ := d.decode()
		skey, ok := key.(string)
		if !ok {
			d.err = fmt.Errorf("Map keys must derive from strings")
			break
		}
		result[skey] = val
	}
	return result, d.err
}

func (d *Decoder) decodeType(tt *Signature) (interface{}, error) {
	switch tt.Tag {
	case StructTag:
		result := make(map[string]interface{}, 0)
		for _, f := range tt.Fields {
			tmp, err := d.decodeType(f.Type)
			if err != nil {
				return nil, err
			}
			result[f.Name] = tmp
		}
		return result, nil
	case MapTag:
		mlen := int(d.ParseUnsigned())
		if d.err != nil {
			return nil, d.err
		}
		keys := tt.Keys
		items := tt.Items
		result := make(map[string]interface{}) //! I could do better
		for i := 0; i < mlen; i++ {
			ke, err := d.decodeType(keys)
			if err != nil {
				return nil, err
			}
			it, err := d.decodeType(items)
			if err != nil {
				return nil, err
			}
			if skey, ok := ke.(string); ok {
				result[skey] = it
			} else {
				return nil, fmt.Errorf("cannot decode Map from tbin: key is non-string derived: %v", ke)
			}
		}
		return result, nil
	case ArrayTag:
		alen := int(d.ParseUnsigned())
		if d.err != nil {
			return nil, d.err
		}
		items := tt.Items
		switch items.Tag {
		default:
			result := make([]interface{}, alen)
			for i := 0; i < alen; i++ {
				tmp, err := d.decodeType(items)
				if err != nil {
					return nil, err
				}
				result[i] = tmp
			}
			return result, nil
		}
	case AnyTag:
		return d.decode()
	case EnumTag:
		nsym := d.ParseInt()
		return tt.Symbols[nsym], d.err
	case UnionTag:
		nvariant := d.ParseUnsigned() - 1
		return d.decodeType(tt.Variants[nvariant])
	case Int8Tag:
		n := int8(d.ParseInt())
		return n, d.err
	case Int16Tag:
		n := int16(d.ParseInt())
		return n, d.err
	case Int32Tag:
		n := int32(d.ParseInt())
		return n, d.err
	case Int64Tag:
		n := d.ParseInt64()
		return n, d.err
	case Float32Tag:
		return d.ParseFloat32()
	case Float64Tag:
		return d.ParseFloat64()
	case StringTag:
		return d.ParseString()
	case TimestampTag:
		return d.ParseTimestamp()
	case UUIDTag:
		return d.ParseUUID()
	case BoolTag:
		n := d.ParseUnsigned()
		if d.err == nil {
			if n != 0 {
				return true, nil
			}
		}
		return false, d.err
	}
	d.err = fmt.Errorf("decode unhandled type (0x%02x): %v", tt.Tag, tt)
	return nil, d.err
}

func (d *Decoder) ParseUnsigned() uint {
	if d.err == nil {
		n := uint(0)
		var shift uint
		for shift < 32 {
			b, err := d.in.ReadByte()
			if err != nil {
				d.err = err
				return 0
			}
			n |= (uint(b) & 127) << shift
			if (b & 0x80) == 0 {
				return n
			}
			shift += 7
		}
		d.err = fmt.Errorf("Bad int encoding")
	}
	return 0
}

func (d *Decoder) ParseUnsigned64() uint64 {
	if d.err == nil {
		n := uint64(0)
		var shift uint
		for shift < 64 {
			b, err := d.in.ReadByte()
			if err != nil {
				d.err = err
				return 0
			}
			n |= (uint64(b) & 127) << shift
			if (b & 0x80) == 0 {
				return n
			}
			shift += 7
		}
		d.err = fmt.Errorf("Bad int encoding")
	}
	return 0
}

func (d *Decoder) ParseSymbol() (string, error) {
	if d.err == nil {
		id := d.ParseUnsigned()
		if d.err == nil {
			if int(id) == len(d.syms) {
				name, err := d.ParseString()
				if err == nil {
					d.syms = append(d.syms, name)
					return name, nil
				}
			} else {
				name := d.syms[id]
				return name, nil
			}
		}
	}
	return "", d.err
}

func (d *Decoder) ParseBool() bool {
	b := false
	n, err := d.in.ReadByte()
	if err != nil {
		d.err = err
	} else if n != 0 {
		b = true
	}
	return b
}

func (d *Decoder) ParseInt() int {
	n := int(d.ParseUnsigned())
	return (n >> 1) ^ -(n & 1) // back to two's-complement
}

func (d *Decoder) ParseInt64() int64 {
	n := int(d.ParseUnsigned64())
	return int64((n >> 1) ^ -(n & 1)) // back to two's-complement
}

func (d *Decoder) BoolValue() bool {
	switch d.currentTag {
	case BoolTag:
		if d.ParseUnsigned() != 0 {
			return true
		}
		return false
	default:
		d.err = fmt.Errorf("Cannot get bool value for this tag: %v", d.currentTag)
		return false
	}
}

func (d *Decoder) Int8Value() int8 {
	return int8(d.ParseInt())
}
func (d *Decoder) Int16Value() int16 {
	return int16(d.ParseInt())
}
func (d *Decoder) ReadInt32() int32 {
	return int32(d.ParseInt())
}

func (d *Decoder) Int32Value() int32 {
	return int32(d.ParseInt())
}
func (d *Decoder) Int64Value() int64 {
	return d.ParseInt64()
}

func (d *Decoder) ParseFloat32() (float32, error) {
	var bit24, bit16, bit8, bit0 byte
	if d.err == nil {
		bit24, d.err = d.in.ReadByte()
		if d.err == nil {
			bit16, d.err = d.in.ReadByte()
			if d.err == nil {
				bit8, d.err = d.in.ReadByte()
				if d.err == nil {
					bit0, d.err = d.in.ReadByte()
					if d.err == nil {
						bits := (uint32(bit24) << 24) |
							(uint32(bit16) << 16) |
							(uint32(bit8) << 8) |
							uint32(bit0)
						return math.Float32frombits(bits), nil
					}
				}
			}
		}
	}
	return 0, d.err
}

func (d *Decoder) ParseFloat64() (float64, error) {
	var bit56, bit48, bit40, bit32, bit24, bit16, bit8, bit0 byte
	if d.err == nil {
		bit56, d.err = d.in.ReadByte()
		if d.err == nil {
			bit48, d.err = d.in.ReadByte()
			if d.err == nil {
				bit40, d.err = d.in.ReadByte()
				if d.err == nil {
					bit32, d.err = d.in.ReadByte()
					if d.err == nil {
						bit24, d.err = d.in.ReadByte()
						if d.err == nil {
							bit16, d.err = d.in.ReadByte()
							if d.err == nil {
								bit8, d.err = d.in.ReadByte()
								if d.err == nil {
									bit0, d.err = d.in.ReadByte()
									if d.err == nil {
										bits := (uint64(bit56) << 56) |
											(uint64(bit48) << 48) |
											(uint64(bit40) << 40) |
											(uint64(bit32) << 32) |
											(uint64(bit24) << 24) |
											(uint64(bit16) << 16) |
											(uint64(bit8) << 8) |
											uint64(bit0)
										return math.Float64frombits(bits), nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return 0, d.err
}

func (d *Decoder) readBytes(buf []byte) error {
	if d.err == nil {
		offset := 0
		end := len(buf)
		remaining := len(buf)
		for remaining > 0 {
			n, err := d.in.Read(buf[offset:end])
			if err != nil {
				d.err = err
				break
			}
			remaining -= n
			offset += n
		}
	}
	return d.err
}

func (d *Decoder) ParseBytes() ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}
	n := d.ParseUnsigned()
	buf := make([]byte, n)
	err := d.readBytes(buf)
	return buf, err
}

func (d *Decoder) ParseString() (string, error) {
	buf, err := d.ParseBytes()
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (d *Decoder) ParseTimestamp() (rdl.Timestamp, error) {
	secs, err := d.ParseFloat64()
	if err != nil {
		var ts rdl.Timestamp
		return ts, err
	}
	return rdl.TimestampFromEpoch(secs), nil
}

func (d *Decoder) ParseUUID() (rdl.UUID, error) {
	u := make([]byte, 16)
	n, err := d.in.Read(u)
	if err != nil || n != 16 {
		d.err = fmt.Errorf("Bad UUID value in tbin stream")
		return nil, d.err
	}
	return rdl.UUID(u), nil
}

func (d *Decoder) ReadInt() int {
	return d.ParseInt()
}

func (d *Decoder) ReadUnsigned() int {
	return int(d.ParseUnsigned())
}
func (d *Decoder) ReadSize() int {
	return d.ReadUnsigned()
}

func (d *Decoder) ReadType() (*Signature, error) {
again:
	tag := d.ParseUnsigned()
	if d.err != nil {
		return nil, d.err
	}
	if tag < FirstUserTag {
		return nil, fmt.Errorf("Expected custom type, found 0x%02x", tag)
	}
	idx := int(tag - FirstUserTag)
	var ttype *Signature
	if idx < len(d.types) {
		ttype = d.types[idx]
	} else {
		ttype = d.parseType()
		if ttype == nil {
			d.err = fmt.Errorf("First use of a user tag must be followed by a typedef.")
			return nil, d.err
		}
		d.types = append(d.types, ttype)
		goto again
	}
	return ttype, nil
}

func (d *Decoder) buildTypeSignature(t *Signature, out *bytes.Buffer) error {
	var err error
	tag := t.Tag
	switch tag {
	case StructTag:
		out.WriteByte(DefStructTag)
		nfields := len(t.Fields)
		err = EncodeUvarint(out, nfields)
		if err == nil {
			for _, f := range t.Fields {
				fn := f.Name
				b := []byte(fn)
				err = EncodeUvarint(out, len(b))
				if err == nil {
					_, err = out.Write(b)
					if err == nil {
						err = d.buildTypeSignature(f.Type, out)
					}
				}
			}
		}
	case ArrayTag:
		err = out.WriteByte(ArrayTag)
		if err == nil {
			err = d.buildTypeSignature(t.Items, out)
		}
	default:
		err = EncodeUvarint(out, int(tag))
	}
	return err
}

func (d *Decoder) DecodeReflect(v reflect.Value) error {
again:
	var tag int
	if d.pendingTag >= 0 {
		tag = d.pendingTag
		d.pendingTag = -1
	} else {
		tag = int(d.ParseUnsigned())
	}
	if d.err != nil {
		return d.err
	}
	if tag >= FirstUserTag {
		idx := int(tag - FirstUserTag)
		if idx < len(d.types) {
			ttype := d.types[idx]
			return d.decodeTypeReflect(ttype, v)
		}
		ttype := d.parseType()
		if ttype == nil {
			d.err = fmt.Errorf("First use of a user tag must be followed by a typedef.")
			return d.err
		}
		d.types = append(d.types, ttype)
		goto again
	} else {
		if (tag & TinyStrTagMask) == TinyStrTag {
			n := tag & TinyStrDataMask
			tinybuf := make([]byte, n)
			d.err = d.readBytes(tinybuf)
			if d.err == nil {
				s := string(tinybuf)
				if v.Kind() == reflect.Interface {
					vv := reflect.New(reflect.TypeOf(s))
					v.Set(vv)
				} else {
					v.SetString(s)
				}
			}
			return d.err
		}
		switch tag {
		case NullTag:
			//already zeroed
			return nil
		case BoolTag:
			b := d.ParseBool()
			if v.Kind() == reflect.Interface {
				v.Set(reflect.New(reflect.TypeOf(b)))
			} else {
				v.SetBool(b)
			}
			return d.err
		case Int8Tag:
			n := int8(d.ParseInt())
			if d.err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(n)))
				} else {
					v.SetInt(int64(n))
				}
			}
			return d.err
		case Int16Tag:
			n := int16(d.ParseInt())
			if d.err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(n)))
				} else {
					v.SetInt(int64(n))
				}
			}
			return d.err
		case Int32Tag:
			n := d.ParseInt()
			if d.err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(n)))
				} else {
					v.SetInt(int64(n))
				}
			}
			return d.err
		case Int64Tag:
			n := d.ParseInt64()
			if d.err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(n)))
				} else {
					v.SetInt(n)
				}
			}
			return d.err
		case Float32Tag:
			n, err := d.ParseFloat32()
			if err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(n)))
				} else {
					v.SetFloat(float64(n))
				}
			}
			return err
		case Float64Tag:
			n, err := d.ParseFloat64()
			if err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(n)))
				} else {
					v.SetFloat(n)
				}
			}
			return err
		case BytesTag:
			b, err := d.ParseBytes()
			if err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(b)))
				} else {
					v.SetBytes(b)
				}
			}
			return err
		case StringTag:
			s, err := d.ParseString()
			if err == nil {
				if v.Kind() == reflect.Interface {
					v.Set(reflect.New(reflect.TypeOf(s)))
				} else {
					v.SetString(s)
				}
			}
			return err
		case SymbolTag:
			s, err := d.ParseSymbol()
			if err == nil {
				v.Set(reflect.ValueOf(rdl.Symbol(s)))
			}
			return err
		case TimestampTag:
			ts, err := d.ParseTimestamp()
			if err == nil {
				v.Set(reflect.ValueOf(ts))
			}
			return err
		case UUIDTag:
			u, err := d.ParseUUID()
			if err == nil {
				v.Set(reflect.ValueOf(u))
			}
			return err
		case StructTag:
			return d.DecodeStructReflect(v)
		case ArrayTag:
			return d.DecodeArrayReflect(v)
		case MapTag:
			return d.DecodeMapReflect(v)
		}
	}
	return fmt.Errorf("Unexpected tag value: 0x%02x", tag)
}

func (d *Decoder) reflectFieldByIndex(v reflect.Value, idx int) reflect.Value {
	t := v.Type()
	if v.Kind() == reflect.Ptr {
		vv := v.Elem()
		if !vv.IsValid() {
			vv = reflect.New(t.Elem())
			v.Set(vv)
		}
		v = v.Elem()
		t = t.Elem()
	}
	return v.Field(idx)
}

func (d *Decoder) reflectField(v reflect.Value, fname string) (reflect.Value, error) {
	t := v.Type()
	if v.Kind() == reflect.Ptr {
		vv := v.Elem()
		if !vv.IsValid() {
			vv = reflect.New(t.Elem())
			v.Set(vv)
		}
		v = v.Elem()
		t = t.Elem()
	}
	size := t.NumField()
	for i := 0; i < size; i++ {
		field := t.Field(i)
		fn := field.Name
		ftag := field.Tag.Get("json")
		if ftag != "" {
			//if ftag == "-" { continue } //should never happen with rdl models. If it does, our count was wrong!
			ftags := strings.Split(ftag, ",")
			n := ftags[0]
			if n != "" {
				fn = n
			}
		}
		if strings.EqualFold(fname, fn) {
			return v.Field(i), nil
		}
	}
	var junk reflect.Value
	return junk, fmt.Errorf("No such field: %v", fname)
}

func (d *Decoder) decodeTypeReflect(tt *Signature, v reflect.Value) error {
	switch tt.Tag {
	case StructTag:
		if !v.CanSet() {
			d.err = fmt.Errorf("Cannot set struct")
			return d.err
		}
		t := v.Type()
		vv := v
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
			t := t.Elem()
			if !v.IsValid() {
				v = reflect.New(t)
			}
		}
		for _, f := range tt.Fields {
			fn := f.Name
			field, err := d.reflectField(v, fn)
			if err != nil {
				continue
			}
			if !field.CanSet() {
				d.err = fmt.Errorf("Cannot set struct field")
				return d.err
			}
			err = d.decodeTypeReflect(f.Type, field)
			if err != nil {
				return err
			}
		}
		vv.Set(v)
		return nil
	case MapTag:
		mlen := int(d.ParseUnsigned())
		if d.err != nil {
			return d.err
		}
		keys := tt.Keys
		items := tt.Items
		t := v.Type()
		itemType := t.Elem()
		keyType := t.Key()
		if !v.CanSet() {
			d.err = fmt.Errorf("Cannot set array element")
			return d.err
		}
		v.Set(reflect.MakeMap(reflect.MapOf(keyType, itemType)))
		for i := 0; i < mlen; i++ {
			keyV := reflect.New(keyType).Elem()    //we don't want a pointer
			err := d.decodeTypeReflect(keys, keyV) //? need a reflect.Value, not a reflect.Type
			if err != nil {
				return err
			}
			itemV := reflect.New(itemType)
			if itemV.Kind() != reflect.Struct {
				itemV = itemV.Elem()
			}
			err = d.decodeTypeReflect(items, itemV)
			if err != nil {
				return err
			}
			v.SetMapIndex(keyV, itemV)
		}
		return nil
	case ArrayTag:
		alen := int(d.ParseUnsigned())
		if d.err != nil {
			return d.err
		}
		items := tt.Items
		switch items.Tag {
		default:
			t := v.Type()
			itemType := t.Elem()
			if !v.CanSet() {
				d.err = fmt.Errorf("Cannot set array element")
				return d.err
			}
			v.Set(reflect.MakeSlice(reflect.SliceOf(itemType), alen, alen))
			for i := 0; i < alen; i++ {
				item := v.Index(i)
				itemType := item.Type()
				if item.Kind() == reflect.Ptr {
					item = item.Elem()
					itemType := itemType.Elem()
					if !item.IsValid() {
						item = reflect.New(itemType)
					}
				}
				if itemType.Kind() == reflect.Ptr {
					d.decodeTypeReflect(items, item.Elem()) //BUG: this bypasses the UnmarshalTBin method of the object!
				} else {
					d.decodeTypeReflect(items, item)
				}
				v.Index(i).Set(item)
				if d.err != nil {
					return d.err
				}
			}
			return nil
		}
	case BoolTag:
		n := d.ParseUnsigned()
		if d.err == nil {
			b := false
			if n != 0 {
				b = true
			}
			v.SetBool(b)
		}
		return d.err
	case Int32Tag, Int8Tag, Int16Tag:
		n := d.ParseInt()
		if d.err == nil {
			v.SetInt(int64(n))
		}
		return d.err
	case Int64Tag:
		n := d.ParseInt64()
		if d.err == nil {
			v.SetInt(n)
		}
		return d.err
	case Float32Tag:
		n, err := d.ParseFloat32()
		if err == nil {
			v.SetFloat(float64(n))
		}
		return err
	case Float64Tag:
		n, err := d.ParseFloat64()
		if err == nil {
			v.SetFloat(n)
		}
		return err
	case StringTag:
		s, err := d.ParseString()
		if err == nil {
			v.SetString(s)
		}
		return err
	case UUIDTag:
		u, err := d.ParseUUID()
		if err == nil {
			v.Set(reflect.ValueOf(u))
		}
		return err
	case TimestampTag:
		ts, err := d.ParseTimestamp()
		if err == nil {
			v.Set(reflect.ValueOf(ts))
		}
		return err
	case EnumTag:
		n := d.ParseInt()
		if d.err == nil {
			v.SetInt(int64(n))
		}
		return d.err
	case UnionTag:
		n := int(d.ParseUnsigned())
		if n > len(tt.Variants) || len(tt.Variants) != v.NumField()-1 {
			d.err = fmt.Errorf("Variant id out of range for target union type: %v -- %v == %v", v, tt, v.NumField())
			return d.err
		}
		if !v.IsValid() {
			un := reflect.New(v.Type())
			vu := un.Elem()
			v.Set(vu)
		}
		field := v.Field(0)
		if !field.CanSet() {
			d.err = fmt.Errorf("Cannot set struct field")
			return d.err
		}
		field.SetInt(int64(n))
		field = v.Field(n)
		if !field.CanSet() {
			d.err = fmt.Errorf("Cannot set struct field")
			return d.err
		}
		return d.decodeTypeReflect(tt.Variants[n-1], field)
	case AnyTag:
		peekTag := d.ParseUnsigned()
		if d.err == nil && peekTag != NullTag {
			d.pendingTag = int(peekTag)
			if v.Kind() == reflect.Ptr {
				vv := reflect.New(v.Type().Elem())
				v.Set(vv)
				v = v.Elem()
			}
			return d.DecodeReflect(v)
		}
		return d.err
	}
	d.err = fmt.Errorf("decode unhandled type (0x%02x): %v", tt.Tag, tt)
	return d.err
}

func (d *Decoder) DecodeStructReflect(v reflect.Value) error {
	count := int(d.ParseUnsigned())
	t := v.Type()
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t := t.Elem()
		if !v.IsValid() {
			v = reflect.New(t)
		}
	}
	for i := 0; i < count; i++ {
		fname, _ := d.ParseSymbol()
		if d.err != nil {
			break
		}
		field, err := d.reflectField(v, fname)
		if err != nil {
			d.decode() //and toss it
		} else {
			if !field.CanSet() {
				d.err = fmt.Errorf("Cannot set struct field")
				return d.err
			}
			d.DecodeReflect(field)
		}
	}
	return d.err
}

func (d *Decoder) DecodeArrayReflect(v reflect.Value) error {
	count := int(d.ParseUnsigned())
	t := v.Type()
	itemType := t.Elem()
	if !v.CanSet() {
		d.err = fmt.Errorf("Cannot set array element")
		return d.err
	}
	v.Set(reflect.MakeSlice(reflect.SliceOf(itemType), count, count))
	for i := 0; i < count; i++ {
		item := v.Index(i)
		itemType := item.Type()
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
			itemType := itemType.Elem()
			if !item.IsValid() {
				item = reflect.New(itemType)
			}
		}
		if itemType.Kind() == reflect.Ptr {
			d.DecodeReflect(item.Elem())
		} else {
			d.DecodeReflect(item)
		}
		v.Index(i).Set(item)
		if d.err != nil {
			return d.err
		}
	}
	return d.err
}

func (d *Decoder) DecodeMapReflect(v reflect.Value) error {
	count := int(d.ParseUnsigned())
	t := v.Type()
	keyType := t.Key()
	itemType := t.Elem()
	if !v.CanSet() {
		d.err = fmt.Errorf("Cannot set mapentry")
		return d.err
	}
	v.Set(reflect.MakeMap(reflect.MapOf(keyType, itemType)))
	for i := 0; i < count; i++ {
		key := reflect.New(keyType)
		d.DecodeReflect(key.Elem())
		key = key.Elem()
		var item reflect.Value
		if itemType.Kind() == reflect.Interface && itemType.NumMethod() == 0 {
			tmp, err := d.decode()
			if err != nil {
				return err
			}
			item = reflect.ValueOf(tmp)
		} else {
			//need a test case for this path
			item = reflect.New(itemType)
			itemType := item.Type()
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
				itemType := itemType.Elem()
				if !item.IsValid() {
					item = reflect.New(itemType)
				}
			}
			if itemType.Kind() == reflect.Ptr {
				d.DecodeReflect(item.Elem())
			} else {
				d.DecodeReflect(item)
			}
		}
		if d.err != nil {
			return d.err
		}
		v.SetMapIndex(key, item)
	}
	return d.err
}
