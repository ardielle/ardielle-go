// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

//
//Go implementation of the tbin encoding format
//

package tbin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"

	"github.com/ardielle/ardielle-go/rdl"
)

//
// Encode - this encoder method marshals an arbitrary object.
// It first checks of the object implements the TBinMarshallable interface,
// and if so calls the objects Marshal method. Otherwise, it uses reflection
// on the object to determine how to marshal it.
//
func (enc *Encoder) Encode(data interface{}) error {
	return enc.encode(data, true)
}

//
// EncodeReflect - encode the data *without* using any TbinMarshallable support, relying
// solely on reflection
//
func (enc *Encoder) EncodeReflect(data interface{}) error {
	return enc.encode(data, false) // ignore marshallable interface, just reflect always
}

func (enc *Encoder) encode(data interface{}, useMarshallable bool) error {
	if useMarshallable {
		m, ok := data.(TBinMarshallable)
		if ok {
			return m.MarshalTBin(enc) //burden on the app, but can be faster
		}
	}
	return enc.encodeData(data, useMarshallable)
}

//
// EncodeType - define the type of the data if it isn't already defined, and then
// emit the tag for it
//func (enc *Encoder) EncodeType(data interface{}, signature []byte) error {
//	v := reflect.ValueOf(data)
//	return enc.encodeReflectedStruct(v, false)
//}

//
// Flush - flush output, and reset buffer to encode again
//
func (enc *Encoder) Flush() error {
	if enc.err == nil && enc.out != nil {
		enc.out.Write(enc.buf.Bytes())
		enc.buf.Reset()
	}
	return enc.err
}

//
// Bytes - return the encoded data as a byte array
//
func (enc *Encoder) Bytes() []byte {
	return enc.buf.Bytes()
}

//
// Return the last encoding error, if any.
//
func (enc *Encoder) Error() error {
	return enc.err
}

func (enc *Encoder) encodeData(data interface{}, useMarshallable bool) error {
	if data == nil {
		return enc.EncodeNull()
	}
	switch d := data.(type) {
	case bool:
		return enc.EncodeBool(d)
	case int8:
		return enc.EncodeInt8(d)
	case int16:
		return enc.EncodeInt16(d)
	case int32:
		return enc.EncodeInt32(d)
	case int:
		return enc.EncodeInt32(int32(d))
	case int64:
		return enc.EncodeInt64(d)
	case uint8:
		return enc.EncodeInt8(int8(d))
	case uint16:
		return enc.EncodeInt16(int16(d))
	case uint32:
		return enc.EncodeInt32(int32(d))
	case uint:
		return enc.EncodeInt32(int32(d))
	case uint64:
		return enc.EncodeInt64(int64(d))
	case float32:
		return enc.EncodeFloat32(d)
	case float64:
		return enc.EncodeFloat64(d)
	case []byte:
		return enc.EncodeBytes(d)
	case string:
		return enc.EncodeString(d)
	case rdl.Timestamp:
		return enc.EncodeTimestamp(d)
	case rdl.UUID:
		return enc.EncodeUUID(d)
	case rdl.Symbol:
		return enc.EncodeSymbol(d)
	case rdl.Struct:
		return enc.EncodeStruct(d, useMarshallable)
	case []interface{}:
		return enc.EncodeArray(d, useMarshallable)
	}
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return fmt.Errorf("Cannot marshal value: %v", data)
	}
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return enc.EncodeNull()
		}
		vv := v.Elem()
		if t.Elem().Kind() == reflect.Array { //a pointer to an array can be sliced, which is faster
			vs := vv.Slice(0, vv.Len())
			return enc.encodeReflectedArray(vs, useMarshallable)
		}
		return enc.encode(vv.Interface(), useMarshallable)
	case reflect.Struct:
		return enc.encodeReflectedStruct(v, useMarshallable)
	case reflect.Array:
		return enc.encodeReflectedArray(v, useMarshallable)
	case reflect.Slice:
		if v.IsNil() {
			return enc.EncodeNull()
		}
		return enc.encodeReflectedArray(v, useMarshallable)
	case reflect.Map:
		return enc.encodeReflectedMap(v, useMarshallable)
	case reflect.Bool:
		return enc.EncodeBool(v.Bool())
	case reflect.Int8:
		return enc.EncodeInt8(int8(v.Int()))
	case reflect.Int16:
		return enc.EncodeInt16(int16(v.Int()))
	case reflect.Int:
		n := v.Int()
		s1 := fmt.Sprint(data)
		s2 := fmt.Sprint(n)
		if s1 != s2 {
			return enc.encodeReflectedEnum(v, useMarshallable)
		}
		fallthrough
	case reflect.Int32:
		return enc.EncodeInt32(int32(v.Int()))
	case reflect.Int64:
		return enc.EncodeInt64(v.Int())
	case reflect.Uint8:
		return enc.EncodeInt8(int8(v.Uint()))
	case reflect.Uint16:
		return enc.EncodeInt16(int16(v.Uint()))
	case reflect.Uint, reflect.Uint32:
		return enc.EncodeInt32(int32(v.Uint()))
	case reflect.Uint64:
		return enc.EncodeInt64(int64(v.Uint()))
	case reflect.Float32:
		return enc.EncodeFloat32(float32(v.Float()))
	case reflect.Float64:
		return enc.EncodeFloat64(v.Float())
	case reflect.String:
		return enc.EncodeString(v.String())
	default:
		enc.err = fmt.Errorf("(FIXME) Cannot marshal value (kind = %v): %v", t.Kind(), data)
		return enc.err
	}
}

func (enc *Encoder) EncodeNull() error {
	return enc.writeUnsigned(NullTag)
}

func (enc *Encoder) EncodeBool(val bool) error {
	enc.writeUnsigned(BoolTag)
	return enc.WriteBool(val)
}

func (enc *Encoder) EncodeInt8(val int8) error {
	enc.writeUnsigned(Int8Tag)
	return enc.WriteInt(int(val))
}

func (enc *Encoder) EncodeInt16(val int16) error {
	enc.writeUnsigned(Int16Tag)
	return enc.WriteInt(int(val))
}

func (enc *Encoder) EncodeInt32(val int32) error {
	enc.writeUnsigned(Int32Tag)
	return enc.WriteInt(int(val))
}

func (enc *Encoder) EncodeInt64(val int64) error {
	enc.writeUnsigned(Int64Tag)
	return enc.WriteInt64(val)
}

func (enc *Encoder) EncodeFloat32(val float32) error {
	enc.writeUnsigned(Float32Tag)
	return enc.WriteFloat32(val)
}

func (enc *Encoder) EncodeFloat64(val float64) error {
	enc.writeUnsigned(Float64Tag)
	return enc.WriteFloat64(val)
}

func (enc *Encoder) EncodeUUID(val rdl.UUID) error {
	if val == nil {
		return enc.EncodeNull()
	}
	enc.writeUnsigned(UUIDTag)
	return enc.WriteUUID(val)
}

func (enc *Encoder) EncodeTimestamp(val rdl.Timestamp) error {
	enc.writeUnsigned(TimestampTag)
	return enc.WriteTimestamp(val)
}

func (enc *Encoder) EncodeBytes(val []byte) error {
	if val == nil {
		return enc.EncodeNull()
	}
	enc.writeUnsigned(BytesTag)
	return enc.WriteBytes(val)
}

func (enc *Encoder) EncodeString(val string) error {
	n := len(val)
	if n <= TinyStrMaxlen {
		enc.writeUnsigned(TinyStrTag | n)
		if enc.err == nil {
			utf8 := []byte(val)
			_, enc.err = enc.buf.Write(utf8)
		}
		return enc.err
	}
	enc.writeUnsigned(StringTag)
	return enc.WriteString(val)
}

func (enc *Encoder) EncodeSymbol(val rdl.Symbol) error {
	enc.writeUnsigned(SymbolTag)
	return enc.WriteSymbol(string(val))
}

func (enc *Encoder) EncodeArray(val []interface{}, useMarshallable bool) error {
	n := len(val)
	enc.writeUnsigned(ArrayTag)
	enc.writeUnsigned(n)
	for _, item := range val {
		enc.encode(item, useMarshallable)
	}
	return enc.err
}

func (enc *Encoder) EncodeStruct(val rdl.Struct, useMarshallable bool) error {
	if val == nil {
		return enc.EncodeNull()
	}
	n := len(val)
	enc.writeUnsigned(StructTag)
	enc.writeUnsigned(n)
	for k, v := range val {
		enc.WriteSymbol(string(k))
		enc.encode(v, useMarshallable)
	}
	return enc.err
}

func (enc *Encoder) encodeReflectedStruct(v reflect.Value, useMarshallable bool) error {
	signature := TypeSignature(v.Interface())
	enc.WriteType(signature) //usually just writes the tag, but may write typedefs as a side-effect
	return enc.encodeValue(v, useMarshallable)
}

func (enc *Encoder) encodeReflectedEnum(v reflect.Value, useMarshallable bool) error {
	signature := TypeSignature(v.Interface())
	enc.WriteType(signature) //usually just writes the tag, but may write typedefs as a side-effect
	return enc.encodeValue(v, useMarshallable)
}

func (enc *Encoder) encodeValue(v reflect.Value, useMarshallable bool) error {
	k := v.Kind()
	if k == reflect.Invalid {
		enc.err = fmt.Errorf("cannot encode value %v", v)
		return enc.err
	}
	if useMarshallable {
		if v.CanInterface() {
			data := v.Interface()
			m, ok := data.(TBinMarshallable)
			if ok {
				return m.MarshalTBin(enc) //burden on the app, but can be faster
			}
		}
	}
	t := v.Type()
	typeName := t.String()
	switch typeName {
	case "rdl.UUID":
		return enc.WriteUUID(v.Interface().(rdl.UUID))
	case "rdl.Timestamp":
		return enc.WriteTimestamp(v.Interface().(rdl.Timestamp))
	case "rdl.Symbol":
		return enc.WriteSymbol(string(v.Interface().(rdl.Symbol)))
	case "rdl.Struct":
		//return enc.EncodeStruct(d, useMarshallable)
		panic("rdl.Struct -> fix me")
	}
	var err error
	switch k {
	case reflect.Struct:
		nfields := t.NumField()
		//first check if it is a union (`rdl:union`)
		if nfields > 0 {
			ft := t.Field(0)
			ftag := ft.Tag.Get("rdl")
			if ftag == "union" {
				nvar := int(v.Field(0).Int())
				enc.WriteUnsigned(nvar)
				//note: an uninitialized union has its tag set to zero. Emit nothing after the tag in that case.
				if nvar > 0 {
					enc.tagged = true //ensure the next WriteType doesn't actually do anything
					return enc.encodeValue(v.Field(nvar), useMarshallable)
				}
				enc.err = fmt.Errorf("Cannot marshal uninitialized union type %v in %v", t.Name, v)
				return enc.err
			}
		}
		for i := 0; i < nfields; i++ {
			f := v.Field(i)
			ft := t.Field(i)
			ftag := ft.Tag.Get("rdl")
			if ftag == "optional" {
				ddd := f.Interface()
				if IsZero(f) {
					err = enc.EncodeNull()
				} else {
					err = enc.encodeData(ddd, useMarshallable)
				}
			} else {
				fk := f.Kind()
				if fk == reflect.Ptr {
					if f.IsNil() {
						enc.err = fmt.Errorf("Cannot marshal null pointer for required field %v in %v", ft.Name, f)
						return enc.err
					}
					err = enc.encodeValue(f.Elem(), useMarshallable)
				} else {
					err = enc.encodeValue(f, useMarshallable)
				}
			}
		}
	case reflect.Map:
		n := v.Len()
		enc.WriteUnsigned(n)
		for _, k := range v.MapKeys() {
			enc.encodeValue(k, useMarshallable)
			enc.encodeValue(v.MapIndex(k), useMarshallable)
		}
		err = enc.err
	case reflect.Slice:
		n := v.Len()
		enc.WriteUnsigned(n)
		for i := 0; i < n; i++ {
			vv := v.Index(i)
			if vv.Kind() == reflect.Ptr {
				vv = vv.Elem()
			}
			enc.encodeValue(vv, useMarshallable)
		}
		err = enc.err
	case reflect.Ptr:
		//without context (of a field def, for example), we must assume "optional" for any pointer, because it can be nil
		if v.IsNil() {
			err = enc.EncodeNull()
		} else {
			err = enc.encode(v.Elem().Interface(), useMarshallable)
		}
	case reflect.Int8:
		return enc.WriteInt8(int8(v.Int()))
	case reflect.Int16:
		return enc.WriteInt16(int16(v.Int()))
	case reflect.Int:
		return enc.WriteInt32(int32(v.Int()))
	case reflect.Int32:
		return enc.WriteInt32(int32(v.Int()))
	case reflect.Int64:
		return enc.WriteInt64(v.Int())
	case reflect.Float32:
		return enc.WriteFloat32(float32(v.Float()))
	case reflect.Float64:
		return enc.WriteFloat64(v.Float())
	case reflect.String:
		return enc.WriteString(v.String())
	case reflect.Bool:
		return enc.WriteBool(v.Bool())
	case reflect.Interface: //any type
		return enc.encodeData(v.Interface(), useMarshallable)
	default:
		err = fmt.Errorf("Cannot determine type signature for reflect kind: '%v'", k)
		enc.err = err
	}
	return err
}

func pointerMeansOptional(v reflect.Type) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64, reflect.String:
		return true
	}
	return false
}

func (enc *Encoder) encodeReflectedArray(v reflect.Value, useMarshallable bool) error {
	t := v.Type()
	alen := v.Len()
	if t.Elem().Kind() == reflect.Uint8 {
		b, ok := v.Interface().([]byte)
		if !ok {
			//how to go more efficiently from a reflect.Value of an array to a slice? An array argument (by value) is not addressable
			b = make([]byte, alen)
			for i := 0; i < alen; i++ {
				b[i] = byte(v.Index(i).Uint())
			}
		}
		return enc.EncodeBytes(b)
	}
	//since we know the array type statically, use a typedef for it to make it more compact
	signature := buildTypeSignature(t)
	enc.WriteType(signature)
	return enc.encodeValue(v, useMarshallable)
}

func ValidMapKey(key interface{}) bool {
	return ValidMapKeyType(reflect.TypeOf(key))
}

func ValidMapKeyType(t reflect.Type) bool {
	return t.Kind() == reflect.String //this will include string-like types like rdl.Symbol
}

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
		//	case reflect.Bool: //this doesn't match the JSON choice
		//		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func (enc *Encoder) encodeReflectedMap(v reflect.Value, useMarshallable bool) error {
	t := v.Type()
	mlen := v.Len()
	if !ValidMapKeyType(t.Key()) {
		enc.err = fmt.Errorf("Unsupported type: %v", t)
		return enc.err
	}
	if v.IsNil() {
		return enc.EncodeNull()
	}
	switch t.Elem().Kind() {
	case reflect.Interface:
		enc.writeUnsigned(MapTag)
		enc.writeUnsigned(mlen)
		for _, k := range v.MapKeys() {
			enc.encode(k.Interface(), useMarshallable)
			enc.encode(v.MapIndex(k).Interface(), useMarshallable)
		}
		return enc.err
	}
	signature := buildTypeSignature(t)
	enc.WriteType(signature)
	return enc.encodeValue(v, useMarshallable)
}

//------------------------- low level encoding

type tagDef struct {
	tag       int
	signature string //this is the portablee identifier
	bytes     []byte //this is bound to the encoder stream state (
}

//given a signature, compile to a byte stream, binding typedefs to the encoder's tag set
// the definition, if needed, is emitted as a side effect.
// Int32 -> [Int32Tag]
// Struct -> [<newtag> DefStructTag <size> (<fname> compile(<ftype>))*]
func (enc *Encoder) compileSignature(sig *Signature, ref *bytes.Buffer) error {
	return enc.compileSignatureOptional(sig, ref, false)
}

func (enc *Encoder) compileSignatureOptional(sig *Signature, ref *bytes.Buffer, optional bool) error {
	if enc.err == nil {
		reftag := sig.Tag
		switch sig.Tag {
		case StructTag:
			if sig.Fields == nil {
				enc.err = EncodeUvarint(ref, sig.Tag) //no side effects, just add the tag to the ref
				return enc.err
			}
			key := sig.String()
			def, ok := enc.tags[key]
			if !ok {
				//not defined. Emit a defstruct before continuing
				nfields := len(sig.Fields)
				buf := &bytes.Buffer{}      //this struct's definition, gets built before defining its tag and emitting
				buf.WriteByte(DefStructTag) //note: bytes.Buffer never returns an error on write, so we'll ignore them
				EncodeUvarint(buf, nfields)
				for _, f := range sig.Fields {
					b := []byte(f.Name)
					EncodeUvarint(buf, len(b))
					buf.Write(b)
					if enc.compileSignatureOptional(f.Type, buf, f.optional) != nil {
						return enc.err
					}
				}
				//now bind to a tag and emit the typedef
				tag := enc.nextTag
				enc.nextTag++
				bsig := buf.Bytes()
				def = &tagDef{tag: tag}
				enc.tags[key] = def
				if enc.writeUnsigned(tag) == nil {
					_, enc.err = enc.buf.Write(bsig)
				}
			}
			reftag = def.tag
		case ArrayTag:
			if sig.Items == nil {
				//a generic array with no type (defaults to Any, meaning every element must be tagged)
				enc.err = EncodeUvarint(ref, sig.Tag)
				return enc.err
			}
			key := sig.String()
			def, ok := enc.tags[key]
			if !ok {
				buf := &bytes.Buffer{}
				buf.WriteByte(DefArrayTag)
				if enc.compileSignature(sig.Items, buf) != nil {
					return enc.err
				}
				tag := enc.nextTag
				enc.nextTag++
				bsig := buf.Bytes()
				def = &tagDef{tag: tag}
				enc.tags[key] = def
				if enc.writeUnsigned(tag) == nil {
					_, enc.err = enc.buf.Write(bsig)
				}
			}
			reftag = def.tag
		case MapTag:
			if sig.Keys == nil && sig.Items == nil {
				//a generic map with no type (defaults to String->Any, meaning every element must be tagged)
				enc.err = EncodeUvarint(ref, sig.Tag)
				return enc.err
			}
			key := sig.String()
			def, ok := enc.tags[key]
			if !ok {
				buf := &bytes.Buffer{}
				buf.WriteByte(DefMapTag)
				if enc.compileSignature(sig.Keys, buf) != nil {
					return enc.err
				}
				if enc.compileSignature(sig.Items, buf) != nil {
					return enc.err
				}
				tag := enc.nextTag
				enc.nextTag++
				bsig := buf.Bytes()
				def = &tagDef{tag: tag}
				enc.tags[key] = def
				if enc.writeUnsigned(tag) == nil {
					_, enc.err = enc.buf.Write(bsig)
				}
			}
			reftag = def.tag
		case EnumTag:
			key := sig.String()
			def, ok := enc.tags[key]
			if !ok {
				buf := &bytes.Buffer{}
				buf.WriteByte(DefEnumTag)
				nsyms := len(sig.Symbols)
				enc.err = EncodeUvarint(buf, nsyms)
				for _, sym := range sig.Symbols {
					utf8 := []byte(sym)
					enc.err = EncodeUvarint(buf, len(utf8))
					if enc.err == nil {
						_, enc.err = buf.Write(utf8)
					}
				}
				tag := enc.nextTag
				enc.nextTag++
				bsig := buf.Bytes()
				def = &tagDef{tag: tag}
				enc.tags[key] = def
				if enc.writeUnsigned(tag) == nil {
					_, enc.err = enc.buf.Write(bsig)
				}
			}
			reftag = def.tag
		case UnionTag:
			key := sig.String()
			def, ok := enc.tags[key]
			if !ok {
				buf := &bytes.Buffer{}
				buf.WriteByte(DefUnionTag)
				nvariants := len(sig.Variants)
				enc.err = EncodeUvarint(buf, nvariants) //the count is actually one greater than the array of variants (they start at one)
				for _, variant := range sig.Variants {
					if enc.compileSignature(variant, buf) != nil {
						return enc.err
					}
				}
				tag := enc.nextTag
				enc.nextTag++
				bsig := buf.Bytes()
				def = &tagDef{tag: tag}
				enc.tags[key] = def
				if enc.writeUnsigned(tag) == nil {
					_, enc.err = enc.buf.Write(bsig)
				}
			}
			reftag = def.tag
		default:
			//reftag is already sig.Tag, which we want
		}
		//now write the ref to it
		if optional {
			reftag = AnyTag
		}
		enc.err = EncodeUvarint(ref, reftag)
	}
	return enc.err
}

// WriteType takes a signature and writes the tag for it. If it is the first time
// the signature has been encountered, a new tag is allocated and written followed
// by its definition.
func (enc *Encoder) WriteType(sig *Signature) error {
	if enc.tagged {
		enc.tagged = false
		return nil
	}
	return enc.compileSignature(sig, &enc.buf) //as a side-effect, this defines new tags
}

func (enc *Encoder) writeByte(b byte) error {
	err := enc.buf.WriteByte(b)
	if err != nil {
		if enc.err == nil {
			enc.err = err
		}
	}
	return enc.err
}

func (enc *Encoder) writeUnsigned(n int) error {
	i := 0
	if (n & 0xffffff80) != 0 {
		enc.bytebuf[i] = byte((n | 0x80) & 0xff)
		i++
		n >>= 7
		if (n & 0xffffff80) != 0 {
			enc.bytebuf[i] = byte((n | 0x80) & 0xff)
			i++
			n >>= 7
			if (n & 0xffffff80) != 0 {
				enc.bytebuf[i] = byte((n | 0x80) & 0xff)
				i++
				n >>= 7
				if (n & 0xffffff80) != 0 {
					enc.bytebuf[i] = byte((n | 0x80) & 0xff)
					i++
					n >>= 7
				}
			}
		}
	}
	enc.bytebuf[i] = byte(n)
	i++
	_, err := enc.buf.Write(enc.bytebuf[:i])
	if err == nil {
		return nil
	}
	if enc.err == nil {
		enc.err = err
	}
	return enc.err
}

// WriteUnsigned - writes the int as a varuint. The argument is signed instead of
// unsigned because typical use cases (i.e. len(slice), or 0x40) themselves use int.
// An error is returned it the value is negative.
func (enc *Encoder) WriteUnsigned(val int) error {
	if val < 0 {
		enc.err = fmt.Errorf("negative value provided to WriteUnsigned")
		return enc.err
	}
	return enc.writeUnsigned(val)
}

// WriteInt - write the signed 32 bit int
func (enc *Encoder) WriteInt(n int) error {
	return enc.writeUnsigned((n << 1) ^ (n >> 31))
}

func (enc *Encoder) WriteSize(val int) error {
	return enc.writeUnsigned(val)
}

func (enc *Encoder) WriteTag(val int) error {
	return enc.writeUnsigned(val)
}

// WriteBool - writes the value of the bool
func (enc *Encoder) WriteBool(val bool) error {
	if val {
		return enc.writeUnsigned(1)
	}
	return enc.writeUnsigned(0)
}

// WriteInt8 - writes the signed 8 bit integer as a varint
func (enc *Encoder) WriteInt8(val int8) error {
	return enc.WriteInt(int(val))
}

// WriteInt16 - writes the signed 16 bit integer as a varint
func (enc *Encoder) WriteInt16(val int16) error {
	return enc.WriteInt(int(val))
}

// WriteInt32 - writes the signed 32 bit integer as a varint
func (enc *Encoder) WriteInt32(val int32) error {
	return enc.WriteInt(int(val))
}

// WriteInt64 - writes the signed 64 bit integer as a varint
func (enc *Encoder) WriteInt64(nn int64) error {
	n := binary.PutVarint(enc.bytebuf, nn)
	_, enc.err = enc.buf.Write(enc.bytebuf[:n])
	return enc.err
}

// WriteFloat32 - writes the signed 32 bit float as a varint
func (enc *Encoder) WriteFloat32(n float32) error {
	bits := math.Float32bits(n)
	enc.err = enc.buf.WriteByte(byte(bits >> 24))
	if enc.err == nil {
		enc.err = enc.buf.WriteByte(byte(bits >> 16))
		if enc.err == nil {
			enc.err = enc.buf.WriteByte(byte(bits >> 8))
			if enc.err == nil {
				enc.err = enc.buf.WriteByte(byte(bits))
			}
		}
	}
	return enc.err
}

// WriteFloat64 - writes the signed 64 bit float as a varint
func (enc *Encoder) WriteFloat64(n float64) error {
	bits := math.Float64bits(n)
	enc.err = enc.buf.WriteByte(byte(bits >> 56))
	if enc.err == nil {
		enc.err = enc.buf.WriteByte(byte(bits >> 48))
		if enc.err == nil {
			enc.err = enc.buf.WriteByte(byte(bits >> 40))
			if enc.err == nil {
				enc.err = enc.buf.WriteByte(byte(bits >> 32))
				if enc.err == nil {
					enc.err = enc.buf.WriteByte(byte(bits >> 24))
					if enc.err == nil {
						enc.err = enc.buf.WriteByte(byte(bits >> 16))
						if enc.err == nil {
							enc.err = enc.buf.WriteByte(byte(bits >> 8))
							if enc.err == nil {
								enc.err = enc.buf.WriteByte(byte(bits))
							}
						}
					}
				}
			}
		}
	}
	return enc.err
}

func (enc *Encoder) WriteUUID(u rdl.UUID) error {
	if enc.err == nil {
		_, enc.err = enc.buf.Write([]byte(u))
	}
	return enc.err
}

func (enc *Encoder) WriteTimestamp(val rdl.Timestamp) error {
	return enc.WriteFloat64(val.SecondsSinceEpoch())
}

func (enc *Encoder) WriteBytes(b []byte) error {
	if enc.err == nil {
		enc.writeUnsigned(len(b))
		if enc.err == nil {
			_, enc.err = enc.buf.Write(b)
		}
	}
	return enc.err
}

func (enc *Encoder) writeBytes(b []byte) error {
	if enc.err == nil {
		_, enc.err = enc.buf.Write(b)
	}
	return enc.err
}

func (enc *Encoder) WriteString(s string) error {
	if enc.err == nil {
		utf8 := []byte(s) //do I need to explicitly convert to UTF8?
		enc.writeUnsigned(len(utf8))
		if enc.err == nil {
			_, enc.err = enc.buf.Write(utf8)
		}
	}
	return enc.err
}

func (enc *Encoder) WriteSymbol(name string) error {
	if enc.err == nil {
		id, ok := enc.syms[name]
		if !ok {
			id = enc.nextSymId
			enc.nextSymId++
			enc.syms[name] = id
			enc.writeUnsigned(id)
			enc.WriteString(name)
		} else {
			enc.writeUnsigned(id)
		}
	}
	return enc.err
}

func (enc *Encoder) writeHeader() error {
	if enc.err == nil {
		enc.writeUnsigned(CurVersionTag)
	}
	return enc.err
}
