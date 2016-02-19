// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ardielle/ardielle-go/rdl"
)

var _ = testing.Verbose
var _ = rdl.BaseTypeAny
var _ = ioutil.WriteFile

func bytesString(bytes []byte) string {
	var buf []byte
	buf = append(buf, '{')
	for i, b := range bytes {
		if i > 0 {
			buf = append(buf, ',')
			buf = append(buf, ' ')
		}
		for _, cb := range []byte(fmt.Sprintf("0x%02x", b)) {
			buf = append(buf, cb)
		}
	}
	buf = append(buf, '}')
	return string(buf)
}

func checkExpectedError(test *testing.T, msg string, err error) {
	if err == nil {
		fmt.Printf("*** %s\n", msg)
		test.Errorf("%s", msg)
		os.Exit(1)
	}
}

func checkError(test *testing.T, msg string, tdata []byte, err error, expectedLen int, expected []byte) {
	if err != nil {
		fmt.Printf("*** %s - %v\n", msg, err)
		test.Errorf("%s - %v", msg, err)
		os.Exit(1)
	}
	if expectedLen >= 0 && expectedLen != len(tdata) {
		fmt.Printf("*** %s - expected %d bytes, got %d\n", msg, expectedLen, len(tdata))
		test.Errorf("%s - %v", msg, err)
		os.Exit(1)
	}
	if expected != nil && !bytes.Equal(tdata, expected) {
		fmt.Printf("*** %s - expected %v, got %v\n", msg, bytesString(expected), bytesString(tdata))
		test.Errorf("%s - %v", msg, err)
		os.Exit(1)
	}
	fmt.Printf("%s - %v\n", msg, bytesString(tdata))
}

type ptrtest struct {
	I32  int32
	Pi32 *int32
	Oi32 *int32 `rdl:"optional"`
}

func TestMarshalSimpleTypes(test *testing.T) {
	//null
	var null interface{}
	tdata, err := Marshal(null)
	checkError(test, "nil", tdata, err, -1, []byte{24, 0})

	//bool
	var b bool
	tdata, err = Marshal(b)
	checkError(test, "bool false", tdata, err, -1, []byte{24, 1, 0})
	b = true
	tdata, err = Marshal(b)
	checkError(test, "bool true", tdata, err, -1, []byte{24, 1, 1})

	//int8
	var i8 int8
	tdata, err = Marshal(i8)
	checkError(test, "i8", tdata, err, -1, []byte{24, 2, 0})
	i8 = 23
	tdata, err = Marshal(i8)
	checkError(test, "i8 23", tdata, err, -1, []byte{24, 2, 46})
	i8 = -23
	tdata, err = Marshal(i8)
	checkError(test, "i8 -23", tdata, err, -1, []byte{24, 2, 45})
	i8 = 127
	tdata, err = Marshal(i8)
	checkError(test, "i8 127", tdata, err, -1, []byte{24, 2, 254, 1})
	i8 = -128
	tdata, err = Marshal(i8)
	checkError(test, "i8 -128", tdata, err, -1, []byte{24, 2, 255, 1})
	var u8 uint8 = 0xff
	tdata, err = Marshal(int8(u8))
	checkError(test, "u8 255", tdata, err, -1, []byte{24, 2, 1}) //uint8(0xff) == int8(-1)

	//int16
	var i16 int16 = 23
	tdata, err = Marshal(i16)
	checkError(test, "i16 23", tdata, err, -1, []byte{24, 3, 46})
	i16 = -23
	tdata, err = Marshal(i16)
	checkError(test, "i16 -23", tdata, err, -1, []byte{24, 3, 45})
	i16 = 32767
	tdata, err = Marshal(i16)
	checkError(test, "i16 32767", tdata, err, -1, []byte{24, 3, 254, 255, 3})
	i16 = -32768
	tdata, err = Marshal(i16)
	checkError(test, "i16 -32768", tdata, err, -1, []byte{24, 3, 255, 255, 3})
	var u16 uint16 = 0xffff
	tdata, err = Marshal(int16(u16))
	checkError(test, "u16 65535", tdata, err, -1, []byte{24, 3, 1}) //uint16(0xffff) == int16(-1)

	//int32
	var i32 int32 = 23
	tdata, err = Marshal(i32)
	checkError(test, "i32 23", tdata, err, -1, []byte{24, 4, 46})
	i32 = -23
	tdata, err = Marshal(i32)
	checkError(test, "i32 -23", tdata, err, -1, []byte{24, 4, 45})
	i32 = 0x7fffffff
	tdata, err = Marshal(i32)
	checkError(test, fmt.Sprintf("i32 %d", i32), tdata, err, -1, []byte{24, 4, 254, 255, 255, 255, 15})
	i32 = -(0x7fffffff) - 1
	tdata, err = Marshal(i32)
	checkError(test, fmt.Sprintf("i32 %d", i32), tdata, err, -1, []byte{24, 4, 255, 255, 255, 255, 15})
	var i = 23
	tdata, err = Marshal(i)
	checkError(test, "i 23", tdata, err, -1, []byte{24, 4, 46})
	var u32 uint32 = 0xffffffff
	tdata, err = Marshal(int32(u32))
	checkError(test, "u32 4294967295", tdata, err, -1, []byte{24, 4, 1}) //uint32(0xffffffff) == int32(-1)

	i32 = 23
	pt := ptrtest{1, &i32, nil}
	tdata, err = Marshal(pt)
	checkError(test, "pt.Pi32 &23", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x03, 0x03, 0x49, 0x33, 0x32, 0x04, 0x04, 0x50, 0x69, 0x33, 0x32, 0x04, 0x04, 0x4f, 0x69, 0x33, 0x32, 0x10, 0x40, 0x02, 0x2e, 0x00})
	pt.Oi32 = &i32
	tdata, err = Marshal(pt)
	checkError(test, "pt.Pi32 &23 pt.Oi32 = &23", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x03, 0x03, 0x49, 0x33, 0x32, 0x04, 0x04, 0x50, 0x69, 0x33, 0x32, 0x04, 0x04, 0x4f, 0x69, 0x33, 0x32, 0x10, 0x40, 0x02, 0x2e, 0x04, 0x2e})

	//int64
	var i64 int64 = 23
	tdata, err = Marshal(i64)
	checkError(test, "i64 23", tdata, err, -1, []byte{24, 5, 46})
	i64 = -23
	tdata, err = Marshal(i64)
	checkError(test, "i64 -23", tdata, err, -1, []byte{24, 5, 45})
	i64 = 0x7fffffffffffffff
	tdata, err = Marshal(i64)
	checkError(test, fmt.Sprintf("i64 %d", i64), tdata, err, -1, []byte{24, 5, 254, 255, 255, 255, 255, 255, 255, 255, 255, 1})
	i64 = -i64 - 1
	tdata, err = Marshal(i64)
	checkError(test, fmt.Sprintf("i64 %d", i64), tdata, err, -1, []byte{24, 5, 255, 255, 255, 255, 255, 255, 255, 255, 255, 1})
	var u64 uint64 = 0xffffffffffffffff
	tdata, err = Marshal(int64(u64))
	checkError(test, "u64 18446744073709551615", tdata, err, -1, []byte{24, 5, 1}) //uint64(0xffffffffffffffff) == int64(-1)

	//float32
	var f32 float32
	tdata, err = Marshal(f32)
	checkError(test, "f32", tdata, err, -1, []byte{24, 6, 0, 0, 0, 0})

	f32 = 23.57
	tdata, err = Marshal(f32)
	checkError(test, "f32 23.57", tdata, err, -1, []byte{24, 6, 65, 188, 143, 92})

	//float64
	var f64 float64
	tdata, err = Marshal(f64)
	checkError(test, "f64", tdata, err, -1, []byte{24, 7, 0, 0, 0, 0, 0, 0, 0, 0})

	f64 = 23.57
	tdata, err = Marshal(f64)
	checkError(test, "f64 23.57", tdata, err, -1, []byte{24, 7, 64, 55, 145, 235, 133, 30, 184, 82})

	//bytes
	var bs []byte
	tdata, err = Marshal(bs)
	checkError(test, "bytes nil slice", tdata, err, -1, []byte{24, 0})
	bs = make([]byte, 0) //empty byte slice
	tdata, err = Marshal(bs)
	checkError(test, "bytes empty slice", tdata, err, -1, []byte{24, 8, 0})
	bs = []byte{1, 2, 3, 4, 5} //initialize byte slice
	tdata, err = Marshal(bs)
	checkError(test, "bytes slice with 5 elements", tdata, err, -1, []byte{24, 8, 5, 1, 2, 3, 4, 5})
	var ba [2]byte
	tdata, err = Marshal(&ba) //
	checkError(test, "bytes pointer to array of length 2", tdata, err, -1, []byte{24, 8, 2, 0, 0})
	tdata, err = Marshal(ba) //this is slower, reflect has a hard time with array values instead of slices
	checkError(test, "bytes array of length 2", tdata, err, -1, []byte{24, 8, 2, 0, 0})

	//string
	var s string
	tdata, err = Marshal(s)
	checkError(test, "s empty string", tdata, err, -1, []byte{24, 32})
	s = "foo"
	tdata, err = Marshal(s)
	checkError(test, "s tiny", tdata, err, -1, []byte{24, 35, 102, 111, 111})
	tdata, err = Marshal(&s)
	checkError(test, "s pointer to tiny", tdata, err, -1, []byte{24, 35, 102, 111, 111})
	s = "*can* fit into the tiny format"
	tdata, err = Marshal(s)
	checkError(test, "s largest tiny", tdata, err, -1, []byte{24, 62, 42, 99, 97, 110, 42, 32, 102, 105, 116, 32, 105, 110, 116, 111, 32, 116, 104, 101, 32, 116, 105, 110, 121, 32, 102, 111, 114, 109, 97, 116})
	s = "*can't* fit into the tiny format"
	tdata, err = Marshal(s)
	checkError(test, "s not tiny", tdata, err, -1, []byte{24, 9, 32, 42, 99, 97, 110, 39, 116, 42, 32, 102, 105, 116, 32, 105, 110, 116, 111, 32, 116, 104, 101, 32, 116, 105, 110, 121, 32, 102, 111, 114, 109, 97, 116})

	s = "tiny 姚冀清"
	tdata, err = Marshal(s)
	checkError(test, "s tiny multibyte", tdata, err, -1, []byte{24, 46, 116, 105, 110, 121, 32, 229, 167, 154, 229, 134, 128, 230, 184, 133})

	s = "looks tiny but is bigger姚冀清"
	tdata, err = Marshal(s)
	checkError(test, "s not quite tiny multibyte", tdata, err, -1, []byte{24, 9, 33, 108, 111, 111, 107, 115, 32, 116, 105, 110, 121, 32, 98, 117, 116, 32, 105, 115, 32, 98, 105, 103, 103, 101, 114, 229, 167, 154, 229, 134, 128, 230, 184, 133})

	//timestamp
	ts, _ := rdl.TimestampParse("2015-05-16T19:50:21.002Z")
	tdata, err = Marshal(ts)
	checkError(test, "timestamp", tdata, err, -1, []byte{24, 10, 65, 213, 85, 231, 223, 64, 32, 197})

	//uuid
	u := rdl.ParseUUID("373ab4c4-fc05-11e4-a198-14109fe4729f")
	tdata, err = Marshal(u)
	checkError(test, "uuid", tdata, err, -1, []byte{24, 12, 55, 58, 180, 196, 252, 5, 17, 228, 161, 152, 20, 16, 159, 228, 114, 159})

	//symbol
	sym := rdl.Symbol("foo")
	tdata, err = Marshal(sym)
	checkError(test, "sym", tdata, err, -1, []byte{24, 11, 0, 3, 102, 111, 111})
	sym2 := rdl.Symbol("bar")
	enc := NewEncoder(nil)
	enc.Encode(sym)
	enc.Encode(sym2)
	enc.Encode(sym)
	enc.Encode(sym2)
	tdata = enc.Bytes()
	checkError(test, "sym reuse", tdata, err, -1, []byte{24, 11, 0, 3, 102, 111, 111, 11, 1, 3, 98, 97, 114, 11, 0, 11, 1})
}

func TestMarshalArrays(test *testing.T) {
	//generic array
	var a []interface{}
	tdata, err := Marshal(a)
	checkError(test, "empty generic array", tdata, err, -1, []byte{24, 13, 0})
	a = []interface{}{int32(23)}
	tdata, err = Marshal(a)
	checkError(test, "generic array of one int32", tdata, err, -1, []byte{24, 13, 1, 4, 46})
	a = []interface{}{23}
	tdata, err = Marshal(a)
	checkError(test, "generic array of one int", tdata, err, -1, []byte{24, 13, 1, 4, 46})
	a = []interface{}{int8(23)}
	tdata, err = Marshal(a)
	checkError(test, "generic array of one int8", tdata, err, -1, []byte{24, 13, 1, 2, 46})
	a = []interface{}{byte(23)}
	tdata, err = Marshal(a)
	checkError(test, "generic array of one byte", tdata, err, -1, []byte{24, 13, 1, 2, 46})

	//typed array
	var aint32 []int32
	tdata, err = Marshal(aint32)
	checkError(test, "empty array", tdata, err, -1, []byte{0x18, 0x00})
	aint32 = []int32{1, 2, 3, 4, 5}
	tdata, err = Marshal(aint32)
	checkError(test, "array of five int32 values", tdata, err, -1, []byte{24, 64, 17, 4, 64, 5, 2, 4, 6, 8, 10})

	astr := []string{"one", "two", "three"}
	tdata, err = Marshal(astr)
	checkError(test, "array of 3 string values", tdata, err, -1, []byte{24, 64, 17, 9, 64, 3, 3, 111, 110, 101, 3, 116, 119, 111, 5, 116, 104, 114, 101, 101})
}

func TestMarshalMaps(test *testing.T) {
	var tdata []byte
	var err error
	var expected []byte //maps with more than one entry may fail due to randomized hash order. So just check length.

	mbad := make(map[int]interface{})
	tdata, err = Marshal(mbad)
	if err == nil {
		test.Errorf("Map with non-string key did not get rejected as expected")
	}

	//generic map
	var m map[string]interface{}
	tdata, err = Marshal(m)
	checkError(test, "nil generic map", tdata, err, -1, []byte{24, 0})

	m = make(map[string]interface{})
	tdata, err = Marshal(m)
	checkError(test, "nil generic map", tdata, err, -1, []byte{24, 14, 0})

	//m["foo"] = int32(23)
	m["foo"] = 23 //int and int32 get encoded the same
	tdata, err = Marshal(m)
	checkError(test, "generic map of one string-to-int32", tdata, err, -1, []byte{24, 14, 1, 35, 102, 111, 111, 4, 46})
	m["bar"] = "blah"
	tdata, err = Marshal(m)
	expected = []byte{24, 14, 2, 35, 102, 111, 111, 4, 46, 35, 98, 97, 114, 36, 98, 108, 97, 104}
	checkError(test, "generic map of two items", tdata, err, len(expected), nil)

	//other key types
	m2 := make(map[rdl.Symbol]interface{})
	tdata, err = Marshal(m2)
	checkError(test, "empty symbol-to-any map", tdata, err, -1, []byte{24, 14, 0})

	m2 = make(map[rdl.Symbol]interface{})
	//m2[rdl.Symbol("foo")] = "bar"
	m2["foo"] = "bar" //a symbol *is* a string, so you can just pass it in
	tdata, err = Marshal(m2)
	//checkError(test, "symbol-to-any map, 1 entry", tdata, err, -1, []byte{24, 14, 1, 11, 0, 3, 102, 111, 111, 35, 98, 97, 114})

	m2["bar"] = rdl.Symbol("foo")
	tdata, err = Marshal(m2)
	expected = []byte{24, 14, 2, 11, 0, 3, 102, 111, 111, 35, 98, 97, 114, 11, 1, 3, 98, 97, 114, 11, 0}
	//checkError(test, "symbol-to-any map, 2 entries", tdata, err, len(expected), nil)
	var mmm map[rdl.Symbol]interface{}
	err = Unmarshal(tdata, &mmm)
	if err != nil {
		test.Errorf("Cannot unmarshal map[Symbol]interface{}")
	}

	//typed map
	var mii map[int]int
	tdata, err = Marshal(mii)
	if err == nil {
		//the type gets rejected, even though its nil value could have been encoded, to avoid programming errors
		test.Errorf("Map with non-string key did not get rejected as expected")
	}

	msi := make(map[string]int)
	tdata, err = Marshal(msi)
	checkError(test, "map<string,int> empty", tdata, err, -1, []byte{24, 64, 18, 9, 4, 64, 0}) //defines a new type tag (64), uses it

	msi = make(map[string]int)
	msi["foo"] = 23
	tdata, err = Marshal(msi)
	checkError(test, "map<string,int> 1 entry", tdata, err, -1, []byte{24, 64, 18, 9, 4, 64, 1, 3, 102, 111, 111, 46})

	var mmmm interface{}
	err = Unmarshal(tdata, &mmmm)
	if err != nil {
		test.Errorf("Cannot generic unmarshal map[string]int")
	}

	msymi := make(map[rdl.Symbol]int)
	msymi["foo"] = 23
	tdata, err = Marshal(msymi)
	//checkError(test, "map<rdl.Symbol,int> 1 entry", tdata, err, -1, []byte{24, 64, 18, 11, 4, 64, 1, 3, 102, 111, 111, 46})

	msi["bar"] = 57
	tdata, err = Marshal(msi)
	expected = []byte{24, 64, 18, 9, 4, 64, 2, 3, 102, 111, 111, 46, 3, 98, 97, 114, 114}
	checkError(test, "map<string,int> 2 entries", tdata, err, len(expected), nil)
}

func TestMarshalStructs(test *testing.T) {
	var err error
	var tdata []byte

	//generic struct
	var gs rdl.Struct
	tdata, err = Marshal(gs)
	checkError(test, "nil generic struct", tdata, err, -1, []byte{24, 0})
	gs = make(rdl.Struct)
	gs["foo"] = 23
	tdata, err = Marshal(gs)
	checkError(test, "generic struct with one field", tdata, err, -1, []byte{0x18, 0x0f, 0x01, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x04, 0x2e})
	gs["bar"] = "Hello"

	tdata, err = Marshal(gs)
	expected := []byte{0x18, 0x0f, 0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x04, 0x2e, 0x01, 0x03, 0x62, 0x61, 0x72, 0x25, 0x48, 0x65, 0x6c, 0x6c, 0x6f}
	checkError(test, "generic struct with two fields", tdata, err, len(expected), nil)

	//typed struct
	var pt Point
	tdata, err = Marshal(pt)
	checkError(test, "Point, uninitialized", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x02, 0x01, 0x78, 0x04, 0x01, 0x79, 0x04, 0x40, 0x00, 0x00})

	pt = Point{23, 57}
	tdata, err = Marshal(pt)
	checkError(test, "Point{23,57}", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x02, 0x01, 0x78, 0x04, 0x01, 0x79, 0x04, 0x40, 0x2e, 0x72})
	tdata, err = Marshal(&pt)
	checkError(test, "pointer to Point{23,57}", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x02, 0x01, 0x78, 0x04, 0x01, 0x79, 0x04, 0x40, 0x2e, 0x72})

	tdata, err = Marshal(rect(1, 2, 3, 4))
	checkError(test, "Rect", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x02, 0x01, 0x78, 0x04, 0x01, 0x79, 0x04, 0x41, 0x13, 0x02, 0x02, 0x70, 0x31, 0x40, 0x02, 0x70, 0x32, 0x40, 0x41, 0x02, 0x04, 0x08, 0x0c})
	ioutil.WriteFile("/tmp/test_rect.tbin", tdata, 0644)
}

func TestMarshalUnions(test *testing.T) {
	var err error
	var tdata []byte

	var shp Shape
	tdata, err = Marshal(shp)
	checkExpectedError(test, "Shape, uninitialized", err)

	shp = Shape{ShapeVariantRect, nil, rect(1, 2, 3, 4)}
	tdata, err = Marshal(shp)
	checkError(test, "Shape with rect", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x02, 0x01, 0x78, 0x04, 0x01, 0x79, 0x04, 0x41, 0x11, 0x40, 0x42, 0x13, 0x01, 0x06, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x41, 0x43, 0x13, 0x02, 0x02, 0x70, 0x31, 0x40, 0x02, 0x70, 0x32, 0x40, 0x44, 0x14, 0x02, 0x42, 0x43, 0x44, 0x02, 0x02, 0x04, 0x08, 0x0c})

	ioutil.WriteFile("/tmp/test_rect_shape.tbin", tdata, 0644)

	shp = Shape{ShapeVariantPolyline, polyline(), nil}
	tdata, err = Marshal(shp)
	checkError(test, "Shape with line", tdata, err, -1, []byte{0x18, 0x40, 0x13, 0x02, 0x01, 0x78, 0x04, 0x01, 0x79, 0x04, 0x41, 0x11, 0x40, 0x42, 0x13, 0x01, 0x06, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x41, 0x43, 0x13, 0x02, 0x02, 0x70, 0x31, 0x40, 0x02, 0x70, 0x32, 0x40, 0x44, 0x14, 0x02, 0x42, 0x43, 0x44, 0x01, 0x0d, 0x02, 0x16, 0x04, 0x2c, 0x06, 0x42, 0x14, 0xc8, 0x01, 0x2d, 0xc8, 0x01, 0x2d, 0x41, 0x14, 0x41, 0xce, 0x01, 0x9a, 0x05, 0xd8, 0x04, 0xd0, 0x0f, 0xa4, 0x13, 0xa4, 0x13, 0x9c, 0x85, 0xe3, 0x0b, 0xc0, 0x88, 0xe0, 0x0b, 0xd2, 0xe5, 0xb7, 0xb2, 0x02, 0x42, 0x02, 0x16})
	ioutil.WriteFile("/tmp/test_line_shape.tbin", tdata, 0644)

}

type FooB bool
type Foo8 int8
type Foo16 int16
type Foo int
type Foo32 int32
type Foo64 int64
type FooU8 uint8
type FooU16 uint16
type FooU uint
type FooU32 uint32
type FooU64 uint64
type FooF32 float32
type FooF64 float64

func TestMarshalMisc(test *testing.T) {
	tdata, err := Marshal(FooB(true))
	checkError(test, "FooB", tdata, err, -1, []byte{0x18, 0x01, 0x01})

	tdata, err = Marshal(Foo8(23))
	checkError(test, "Foo8", tdata, err, -1, []byte{0x18, 0x02, 0x2e})
	tdata, err = Marshal(Foo16(23))
	checkError(test, "Foo16", tdata, err, -1, []byte{0x18, 0x03, 0x2e})
	tdata, err = Marshal(Foo32(23))
	checkError(test, "Foo32", tdata, err, -1, []byte{0x18, 0x04, 0x2e})
	tdata, err = Marshal(Foo(23))
	checkError(test, "Foo", tdata, err, -1, []byte{0x18, 0x04, 0x2e})
	tdata, err = Marshal(Foo64(23))
	checkError(test, "Foo64", tdata, err, -1, []byte{0x18, 0x05, 0x2e})

	tdata, err = Marshal(FooU8(23))
	checkError(test, "FooU8", tdata, err, -1, []byte{0x18, 0x02, 0x2e})
	tdata, err = Marshal(FooU16(23))
	checkError(test, "FooU16", tdata, err, -1, []byte{0x18, 0x03, 0x2e})
	tdata, err = Marshal(FooU32(23))
	checkError(test, "FooU32", tdata, err, -1, []byte{0x18, 0x04, 0x2e})
	tdata, err = Marshal(FooU(23))
	checkError(test, "FooU", tdata, err, -1, []byte{0x18, 0x04, 0x2e})
	tdata, err = Marshal(FooU64(23))
	checkError(test, "FooU64", tdata, err, -1, []byte{0x18, 0x05, 0x2e})

}

func TestMarshalEnums(test *testing.T) {
	var err error
	var tdata []byte

	var opt Options
	tdata, err = Marshal(opt)
	checkError(test, "empty Options", tdata, err, -1, []byte{0x18, 0x04, 0x00})

	opt = ONE
	tdata, err = Marshal(opt)
	checkError(test, "Options value of ONE", tdata, err, -1, []byte{0x18, 0x04, 0x02})

}
