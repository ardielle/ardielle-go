// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

//
// Go implementation of the tbin encoding format
//

package tbin

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

const CurrentVersion = 1 // the first versioned version

const NullTag = 0x00      // "nil" or "null"
const BoolTag = 0x01      // "uvarint(b? 1 : 0)"
const Int8Tag = 0x02      // "INT8TAG varint(n)"
const Int16Tag = 0x03     // "INT16TAG varint(n)"
const Int32Tag = 0x04     // "INT32TAG varint(n)"
const Int64Tag = 0x05     // "INT64TAG varint(n)"
const Float32Tag = 0x06   // "FLOAT32TAG 4_bytes"
const Float64Tag = 0x07   // "FLOAT64TAG 8_bytes"
const BytesTag = 0x08     // "BYTESTAG uvarint(len) byte*"
const StringTag = 0x09    // "STRINGTAG uvarint(utflen) utf8bytes*"
const TimestampTag = 0x0a // "TIMESTAMPTAG double" - represented as seconds since epoch (1970)
const SymbolTag = 0x0b    // "SYMBOLTAG uvarint(id) [string(name)]" the name is only included the first occurrence
const UUIDTag = 0x0c      // "UUIDTAG byte[16]" = written as 16 bytes, no count
const ArrayTag = 0x0d     // "ARRAYTAG uvarint(size) value*"
const MapTag = 0x0e       // "MAPTAG uvarint(size) (<value> <value>)*"
const StructTag = 0x0f    // "STRUCTTAG uvarint(size) (<symbol> <value>)*" -- for generic structs (native Go structs are their ownb type/tag)
const AnyTag = 0x10       // used only as a value for array items, map items and keys, and in defstruct.
const DefArrayTag = 0x11  // "<newtag> DefArrayTag itemsTag"
const DefMapTag = 0x12    // "<newtag> DefMapTag keysTag itemsTag"
const DefStructTag = 0x13 // "<newtag> DefStructTag <fieldcount> (uvarint(fnamelen) fnamebytes* ftypeTag)*"
const DefUnionTag = 0x14  // "<newtag> DefUnionTag <count> variantTag*"
const DefEnumTag = 0x15   // "<newtag> DefEnumTag <count> (uvarint(symlen) symbytes*)*"

const UnionTag = 0x16

// unused, reserved tags:
const EnumTag = 0x17

// A version tag should be the first tag in the stream. v1..v8 are thus supported, after which additional byte(s)
// will be required. Bits encode (version - 1), i.e. currently encoded "0001 0000".
const VersionTag = 0x18 // "uvarint(0001 1xxx)"
const VersionTagMask = 0xf8
const VersionDataMask = 0x07
const MinVersionTag = VersionTag
const CurVersionTag = VersionTag + (CurrentVersion - 1) // "0001 1000
const MaxVersionTag = VersionTag + VersionDataMask
const MaxVersion = VersionDataMask + 1

// Tiny strings have a length up to 31 utf8 bytes
// again, this optimization has no effect on packed structs, just the generic encoding (187->160 for my test data)
// if symbols are used instead of strings, the savings are even bigger. But maps have strings, not symbols, as keys
const TinyStrTag = 0x20 // "001x xxxx" <utf8byte>*
const TinyStrTagMask = 0xe0
const TinyStrDataMask = 0x1f
const TinyStrMaxlen = TinyStrDataMask

const FirstUserTag = 0x40 //0x40..0x7f all fit in a single byte tag. Subsequent tags take more. The tag is an unsigned varint.

func TagName(tag int) string {
	if (tag & TinyStrTagMask) == TinyStrTag {
		return "String"
	}
	switch tag {
	case NullTag:
		return "Null"
	case BoolTag:
		return "Bool"
	case Int8Tag:
		return "Int8"
	case Int16Tag:
		return "Int16"
	case Int32Tag:
		return "Int32"
	case Int64Tag:
		return "Int64"
	case Float32Tag:
		return "Float32"
	case Float64Tag:
		return "Float64"
	case BytesTag:
		return "Bytes"
	case StringTag:
		return "String"
	case TimestampTag:
		return "Timestamp"
	case SymbolTag:
		return "Symbol"
	case UUIDTag:
		return "UUID"
	case ArrayTag:
		return "Array"
	case MapTag:
		return "Map"
	case StructTag:
		return "Struct"
	case AnyTag:
		return "Any"
	case EnumTag:
		return "Enum"
	default:
		return fmt.Sprintf("0x%02x", tag)
	}
}

type Symbolic interface {
	Symbol() string
}

//
// Signature - a minimal description of a type
//
type Signature struct {
	Tag      int               `json:"tag"`
	Fields   []*FieldSignature `json:"fields,omitempty"`
	Items    *Signature        `json:"items,omitempty"`
	Keys     *Signature        `json:"keys,omitempty"`
	Variants []*Signature      `json:"variants,omitempty"`
	Symbols  []string          `json:"symbols,omitempty"`
	key      string
}

//
// FieldSignature - the description of a single field in a Struct
//
type FieldSignature struct {
	Name     string     `json:"name"`
	Type     *Signature `json:"type"`
	optional bool
}

//
// String - Signature's String() method produces a compact flat representation of
// the signature.
//
func (sig *Signature) String() string {
	if sig.key == "" {
		sig.key = sig.cacheKey()
	}
	return sig.key
}

func (sig *Signature) cacheKey() string {
	switch sig.Tag {
	case StructTag:
		if sig.Fields == nil {
			return TagName(sig.Tag) //naked struct, implies map[rdl.Symbol]rdl.Any
		} else {
			s := TagName(sig.Tag) + "{"
			for i, f := range sig.Fields {
				if i > 0 {
					s += ","
				}
				s += f.Name
				s += ":"
				s += f.Type.String()
			}
			return s + "}"
		}
	case ArrayTag:
		return TagName(sig.Tag) + "<" + sig.Items.String() + ">"
	case MapTag:
		return TagName(sig.Tag) + "<" + sig.Keys.String() + "," + sig.Items.String() + ">"
	case AnyTag:
		return "Any"
	case UnionTag:
		s := "Union<"
		for i, t := range sig.Variants {
			if i > 0 {
				s += ","
			}
			s += t.String()
		}
		return s + ">"
	case EnumTag:
		s := "Enum<"
		for i, t := range sig.Symbols {
			if i > 0 {
				if i > 1 {
					s += ","
				}
				s += t
			}
		}
		return s + ">"
	default:
		return TagName(sig.Tag)
	}
}

func Field(n string, t *Signature, opt bool) *FieldSignature {
	return &FieldSignature{Name: n, Type: t, optional: opt}
}

func Struct(fields ...*FieldSignature) *Signature {
	return &Signature{Tag: StructTag, Fields: fields}
}

func Array(items *Signature) *Signature {
	return &Signature{Tag: ArrayTag, Items: items}
}

func Map(keys *Signature, items *Signature) *Signature {
	return &Signature{Tag: MapTag, Keys: keys, Items: items}
}

func Union(variants ...*Signature) *Signature {
	return &Signature{Tag: UnionTag, Variants: variants}
}

func Enum(symbols ...string) *Signature {
	return &Signature{Tag: EnumTag, Symbols: symbols}
}

var Null = &Signature{Tag: NullTag}
var Bool = &Signature{Tag: BoolTag}
var Int8 = &Signature{Tag: Int8Tag}
var Int16 = &Signature{Tag: Int16Tag}
var Int32 = &Signature{Tag: Int32Tag}
var Int64 = &Signature{Tag: Int64Tag}
var Float32 = &Signature{Tag: Float32Tag}
var Float64 = &Signature{Tag: Float64Tag}
var Bytes = &Signature{Tag: BytesTag}
var String = &Signature{Tag: StringTag}
var Timestamp = &Signature{Tag: TimestampTag}
var Symbol = &Signature{Tag: SymbolTag}
var UUID = &Signature{Tag: UUIDTag}
var Any = &Signature{Tag: AnyTag}

// EncodeUvarint - encode the uvarint to the Writer
func EncodeUvarint(out io.Writer, n int) error {
	var buf [16]byte
	var err error
	i := 0
	if (n & 0xffffff80) != 0 {
		buf[i] = byte((n | 0x80) & 0xff)
		i++
		n >>= 7
		if (n & 0xffffff80) != 0 {
			buf[i] = byte((n | 0x80) & 0xff)
			i++
			n >>= 7
			if (n & 0xffffff80) != 0 {
				buf[i] = byte((n | 0x80) & 0xff)
				i++
				n >>= 7
				if (n & 0xffffff80) != 0 {
					buf[i] = byte((n | 0x80) & 0xff)
					i++
					n >>= 7
				}
			}
		}
	}
	buf[i] = byte(n)
	_, err = out.Write(buf[:i+1])
	return err
}

// EncodeVarint - encode the varint to the Writer
func EncodeVarint(out io.Writer, n int) error {
	return EncodeUvarint(out, (n<<1)^(n>>31))
}

// DecodeUvarint - decode the uvarint from the Reader
func DecodeUvarint(in io.Reader) (uint, error) {
	var abuf [1]byte
	buf := abuf[:]
	var n uint = 0
	var shift uint = 0
	for shift < 32 {
		c, err := in.Read(buf)
		if c == 1 {
			b := buf[0]
			n |= (uint(b) & 0x7f) << shift
			if (b & 0x80) == 0 {
				return n, nil
			}
		}
		if err != nil {
			return 0, err
		}
		shift += 7
	}
	return 0, fmt.Errorf("Bad varint encoding")
}

// DecodeVarint - decode the varint from the Reader
func DecodeVarint(in io.Reader) (int, error) {
	u, err := DecodeUvarint(in)
	n := int(u)
	return (n >> 1) ^ -(n & 1), err
}

//
// Return a Signature for the type of the given data. Reflection is used.
//
func TypeSignature(val interface{}) *Signature {
	t := reflect.TypeOf(val)
	return buildTypeSignature(t)
}

func buildTypeSignature(t reflect.Type) *Signature {
	//	return buildTypeSignature2(t, false)
	//}

	//func buildTypeSignature2(t reflect.Type, optional bool) *Signature {
	typeName := t.String()
	switch typeName {
	case "rdl.UUID":
		return UUID
	case "rdl.Timestamp":
		return Timestamp
	case "rdl.Symbol":
		return Symbol
	case "rdl.Struct":
		//a naked struct with no schema, essentially a map[rdl.Symbol]interface{}
		return &Signature{Tag: StructTag}
	}
	var err error
	k := t.Kind()
	switch k {
	case reflect.Struct:
		nfields := t.NumField()
		if nfields > 0 {
			ft := t.Field(0)
			ftag := ft.Tag.Get("rdl")
			if ftag == "union" {
				var variants []*Signature
				for i := 1; i < nfields; i++ {
					f := t.Field(i)
					variant := buildTypeSignature(f.Type)
					variants = append(variants, variant)
				}
				return Union(variants...)
			}
		}
		//a regular struct with fields
		var fields []*FieldSignature
		for i := 0; i < nfields; i++ {
			f := t.Field(i)
			fn := f.Name
			ftag := f.Tag.Get("json")
			if ftag != "" {
				if ftag == "-" {
					//should never happen with rdl models.
					continue
				}
				ftags := strings.Split(ftag, ",")
				n := ftags[0]
				if n != "" {
					fn = n
				}
			}
			ftag = f.Tag.Get("rdl")
			opt := (ftag == "optional")
			ft := buildTypeSignature(f.Type)
			if ft != nil {
				fields = append(fields, Field(fn, ft, opt))
			}
		}
		return Struct(fields...)
	case reflect.Slice:
		items := buildTypeSignature(t.Elem())
		return Array(items)
	case reflect.Ptr:
		return buildTypeSignature(t.Elem())
	case reflect.String:
		return String
	case reflect.Bool:
		return Bool
	case reflect.Int8, reflect.Uint8:
		return Int8
	case reflect.Int16, reflect.Uint16:
		return Int16
	case reflect.Int:
		syms := enumSymbols(t)
		if syms != nil {
			return Enum(syms...)
		}
		return Int32
	case reflect.Int32, reflect.Uint, reflect.Uint32:
		return Int32
	case reflect.Int64, reflect.Uint64:
		return Int64
	case reflect.Float32:
		return Float32
	case reflect.Float64:
		return Float64
	case reflect.Map:
		keys := buildTypeSignature(t.Key())
		items := buildTypeSignature(t.Elem())
		return Map(keys, items)
	case reflect.Interface:
		return Any
	default:
		err = fmt.Errorf("Cannot determine type signature for reflect kind of %v", k)
		fmt.Print("***", err)
		//panic(err.Error())
	}
	return Any
}

func enumSymbols(t reflect.Type) []string {
	if _, ok := t.MethodByName("SymbolSet"); ok {
		//an enum type that RDL generated. Create a dummy instance and call to get the symbol set
		tmp := reflect.New(t)
		vv := tmp.Elem()
		vv.SetInt(1)
		symbols := tmp.MethodByName("SymbolSet").Call([]reflect.Value{})[0].Interface()
		return symbols.([]string)[1:]
	}
	return nil
}
