//
// This file generated by rdl v0.7.7 2015-05-14T19:53:06Z
//

package tbin

import (
	"encoding/json"
	"fmt"

	"github.com/ardielle/ardielle-go/rdl"
)

// var _ = testing.Verbose()
var _ = json.Marshal
var _ = fmt.Printf

//
// SimpleName -
//
type SimpleName string

//
// CompoundName -
//
type CompoundName string

//
// Options - options comment
//
type Options int

//
// Options constants
//
const (
	_ Options = iota
	ONE
	TWO
	THREE
)

var namesOptions = []string{
	ONE:   "ONE",
	TWO:   "TWO",
	THREE: "THREE",
}

//
// NewOptions - return a string representation of the enum
//
func NewOptions(init ...interface{}) Options {
	if len(init) == 1 {
		switch v := init[0].(type) {
		case Options:
			return v
		case int:
			return Options(v)
		case int32:
			return Options(v)
		case string:
			for i, s := range namesOptions {
				if s == v {
					return Options(i)
				}
			}
		default:
			panic("Bad init value for Options enum")
		}
	}
	return Options(0) //default to the first enum value
}

//
// String - return a string representation of the enum
//
func (e Options) String() string {
	return namesOptions[e]
}

//
// MarshalJSON is defined for proper JSON encoding of a Options
//
func (e Options) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

//
// UnmarshalJSON is defined for proper JSON decoding of a Options
//
func (e *Options) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err == nil {
		s := string(j)
		for v, s2 := range namesOptions {
			if s == s2 {
				*e = Options(v)
				return nil
			}
		}
		err = fmt.Errorf("Bad enum symbol for type Options: %s", s)
	}
	return err
}

//
// ComplicatedOptions -
//
type ComplicatedOptions string

//
// AlphaName - AlphaName def one or more alpha characters
//
type AlphaName string

//
// YEncoded -
//
type YEncoded string

//
// StringTest -
//
type StringTest struct {
	Name   SimpleName   `json:"name"`
	Parent CompoundName `json:"parent"`
	Names  []SimpleName `json:"names,omitempty" rdl:"optional"`
	Enc    YEncoded     `json:"enc,omitempty" rdl:"optional"`
}

//
// NewStringTest - creates an initialized StringTest instance, returns a pointer to it
//
func NewStringTest(init ...*StringTest) *StringTest {
	var o *StringTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(StringTest)
	}
	return o
}

type rawStringTest StringTest

//
// UnmarshalJSON is defined for proper JSON decoding of a StringTest
//
func (pTypeDef *StringTest) UnmarshalJSON(b []byte) error {
	var r rawStringTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := StringTest(r)
		*pTypeDef = o
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *StringTest) Validate() error {
	if pTypeDef.Name == "" {
		return fmt.Errorf("StringTest: Missing required field: name")
	}
	if pTypeDef.Parent == "" {
		return fmt.Errorf("StringTest: Missing required field: parent")
	}
	return nil
}

//
// azAZ -
//
type azAZ string

//
// TinyInt -
//
type TinyInt int8

//
// SmallInt -
//
type SmallInt int16

//
// RegularInt -
//
type RegularInt int32

//
// LargeInt -
//
type LargeInt int64

//
// Year -
//
type Year int32

//
// Latitude -
//
type Latitude float64

//
// Pi -
//
type Pi float64

//
// LongNumber -
//
type LongNumber int64

//
// MapTest -
//
type MapTest struct {
	Locations map[string]int32 `json:"locations"`
}

//
// NewMapTest - creates an initialized MapTest instance, returns a pointer to it
//
func NewMapTest(init ...*MapTest) *MapTest {
	var o *MapTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(MapTest)
	}
	return o.Init()
}

//
// Init - sets up the instance according to its default field values, if any
//
func (pTypeDef *MapTest) Init() *MapTest {
	if pTypeDef.Locations == nil {
		pTypeDef.Locations = make(map[string]int32)
	}
	return pTypeDef
}

type rawMapTest MapTest

//
// UnmarshalJSON is defined for proper JSON decoding of a MapTest
//
func (pTypeDef *MapTest) UnmarshalJSON(b []byte) error {
	var r rawMapTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := MapTest(r)
		*pTypeDef = *((&o).Init())
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *MapTest) Validate() error {
	if pTypeDef.Locations == nil {
		return fmt.Errorf("MapTest: Missing required field: locations")
	}
	return nil
}

//
// ArrayOfInt -
//
type ArrayOfInt []int32

//
// MapArrayTest -
//
type MapArrayTest struct {
	Locations map[string]ArrayOfInt `json:"locations"`
}

//
// NewMapArrayTest - creates an initialized MapArrayTest instance, returns a pointer to it
//
func NewMapArrayTest(init ...*MapArrayTest) *MapArrayTest {
	var o *MapArrayTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(MapArrayTest)
	}
	return o.Init()
}

//
// Init - sets up the instance according to its default field values, if any
//
func (pTypeDef *MapArrayTest) Init() *MapArrayTest {
	if pTypeDef.Locations == nil {
		pTypeDef.Locations = make(map[string]ArrayOfInt)
	}
	return pTypeDef
}

type rawMapArrayTest MapArrayTest

//
// UnmarshalJSON is defined for proper JSON decoding of a MapArrayTest
//
func (pTypeDef *MapArrayTest) UnmarshalJSON(b []byte) error {
	var r rawMapArrayTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := MapArrayTest(r)
		*pTypeDef = *((&o).Init())
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *MapArrayTest) Validate() error {
	if pTypeDef.Locations == nil {
		return fmt.Errorf("MapArrayTest: Missing required field: locations")
	}
	return nil
}

//
// IntOOBTest -
//
type IntOOBTest struct {
	Theyear Year `json:"theyear"`
}

//
// NewIntOOBTest - creates an initialized IntOOBTest instance, returns a pointer to it
//
func NewIntOOBTest(init ...*IntOOBTest) *IntOOBTest {
	var o *IntOOBTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(IntOOBTest)
	}
	return o
}

type rawIntOOBTest IntOOBTest

//
// UnmarshalJSON is defined for proper JSON decoding of a IntOOBTest
//
func (pTypeDef *IntOOBTest) UnmarshalJSON(b []byte) error {
	var r rawIntOOBTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := IntOOBTest(r)
		*pTypeDef = o
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *IntOOBTest) Validate() error {
	return nil
}

//
// NegativeNumberTest -
//
type NegativeNumberTest struct {
	Mylatitude Latitude `json:"mylatitude"`
}

//
// NewNegativeNumberTest - creates an initialized NegativeNumberTest instance, returns a pointer to it
//
func NewNegativeNumberTest(init ...*NegativeNumberTest) *NegativeNumberTest {
	var o *NegativeNumberTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(NegativeNumberTest)
	}
	return o
}

type rawNegativeNumberTest NegativeNumberTest

//
// UnmarshalJSON is defined for proper JSON decoding of a NegativeNumberTest
//
func (pTypeDef *NegativeNumberTest) UnmarshalJSON(b []byte) error {
	var r rawNegativeNumberTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := NegativeNumberTest(r)
		*pTypeDef = o
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *NegativeNumberTest) Validate() error {
	return nil
}

//
// UUIDTest -
//
type UUIDTest struct {
	Myid rdl.UUID `json:"myid"`
}

//
// NewUUIDTest - creates an initialized UUIDTest instance, returns a pointer to it
//
func NewUUIDTest(init ...*UUIDTest) *UUIDTest {
	var o *UUIDTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(UUIDTest)
	}
	return o
}

type rawUUIDTest UUIDTest

//
// UnmarshalJSON is defined for proper JSON decoding of a UUIDTest
//
func (pTypeDef *UUIDTest) UnmarshalJSON(b []byte) error {
	var r rawUUIDTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := UUIDTest(r)
		*pTypeDef = o
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *UUIDTest) Validate() error {
	if pTypeDef.Myid == nil {
		return fmt.Errorf("UUIDTest: Missing required field: myid")
	}
	return nil
}

//
// TimestampTest -
//
type TimestampTest struct {
	Mytime rdl.Timestamp `json:"mytime"`
}

//
// NewTimestampTest - creates an initialized TimestampTest instance, returns a pointer to it
//
func NewTimestampTest(init ...*TimestampTest) *TimestampTest {
	var o *TimestampTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(TimestampTest)
	}
	return o
}

type rawTimestampTest TimestampTest

//
// UnmarshalJSON is defined for proper JSON decoding of a TimestampTest
//
func (pTypeDef *TimestampTest) UnmarshalJSON(b []byte) error {
	var r rawTimestampTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := TimestampTest(r)
		*pTypeDef = o
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *TimestampTest) Validate() error {
	if pTypeDef.Mytime.IsZero() {
		return fmt.Errorf("TimestampTest: Missing required field: mytime")
	}
	return nil
}

//
// BigStruct -
//
type BigStruct struct {
	MyName           string           `json:"myName"`
	MyUtfname        string           `json:"myUtfname"`
	MyBool           bool             `json:"myBool"`
	MyByte           int8             `json:"myByte"`
	MyShort          int16            `json:"myShort"`
	MyInt            int32            `json:"myInt"`
	MyLong           int64            `json:"myLong"`
	MyFloat          float32          `json:"myFloat"`
	MyDouble         float64          `json:"myDouble"`
	MyIntArray       []int32          `json:"myIntArray"`
	MyStringArray    []string         `json:"myStringArray"`
	MyMap            map[string]int32 `json:"myMap"`
	MyUuid           rdl.UUID         `json:"myUuid"`
	MyStringSubtype  azAZ             `json:"myStringSubtype"`
	MyInt32Subtype   Year             `json:"myInt32Subtype"`
	MyFloat64Subtype Pi               `json:"myFloat64Subtype"`
	MyTime           rdl.Timestamp    `json:"myTime"`
}

//
// NewBigStruct - creates an initialized BigStruct instance, returns a pointer to it
//
func NewBigStruct(init ...*BigStruct) *BigStruct {
	var o *BigStruct
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(BigStruct)
	}
	return o.Init()
}

//
// Init - sets up the instance according to its default field values, if any
//
func (pTypeDef *BigStruct) Init() *BigStruct {
	if pTypeDef.MyIntArray == nil {
		pTypeDef.MyIntArray = make([]int32, 0)
	}
	if pTypeDef.MyStringArray == nil {
		pTypeDef.MyStringArray = make([]string, 0)
	}
	if pTypeDef.MyMap == nil {
		pTypeDef.MyMap = make(map[string]int32)
	}
	return pTypeDef
}

type rawBigStruct BigStruct

//
// UnmarshalJSON is defined for proper JSON decoding of a BigStruct
//
func (pTypeDef *BigStruct) UnmarshalJSON(b []byte) error {
	var r rawBigStruct
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := BigStruct(r)
		*pTypeDef = *((&o).Init())
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *BigStruct) Validate() error {
	if pTypeDef.MyName == "" {
		return fmt.Errorf("BigStruct: Missing required field: myName")
	}
	if pTypeDef.MyUtfname == "" {
		return fmt.Errorf("BigStruct: Missing required field: myUtfname")
	}
	if pTypeDef.MyIntArray == nil {
		return fmt.Errorf("BigStruct: Missing required field: myIntArray")
	}
	if pTypeDef.MyStringArray == nil {
		return fmt.Errorf("BigStruct: Missing required field: myStringArray")
	}
	if pTypeDef.MyMap == nil {
		return fmt.Errorf("BigStruct: Missing required field: myMap")
	}
	if pTypeDef.MyUuid == nil {
		return fmt.Errorf("BigStruct: Missing required field: myUuid")
	}
	if pTypeDef.MyStringSubtype == "" {
		return fmt.Errorf("BigStruct: Missing required field: myStringSubtype")
	}
	if pTypeDef.MyTime.IsZero() {
		return fmt.Errorf("BigStruct: Missing required field: myTime")
	}
	return nil
}

//
// BigTest -
//
type BigTest struct {
	Stuff []*BigStruct `json:"stuff"`
}

//
// NewBigTest - creates an initialized BigTest instance, returns a pointer to it
//
func NewBigTest(init ...*BigTest) *BigTest {
	var o *BigTest
	if len(init) == 1 {
		o = init[0]
	} else {
		o = new(BigTest)
	}
	return o.Init()
}

//
// Init - sets up the instance according to its default field values, if any
//
func (pTypeDef *BigTest) Init() *BigTest {
	if pTypeDef.Stuff == nil {
		pTypeDef.Stuff = make([]*BigStruct, 0)
	}
	return pTypeDef
}

type rawBigTest BigTest

//
// UnmarshalJSON is defined for proper JSON decoding of a BigTest
//
func (pTypeDef *BigTest) UnmarshalJSON(b []byte) error {
	var r rawBigTest
	err := json.Unmarshal(b, &r)
	if err == nil {
		o := BigTest(r)
		*pTypeDef = *((&o).Init())
		err = pTypeDef.Validate()
	}
	return err
}

//
// Validate - checks for missing required fields, etc
//
func (pTypeDef *BigTest) Validate() error {
	if pTypeDef.Stuff == nil {
		return fmt.Errorf("BigTest: Missing required field: stuff")
	}
	return nil
}
