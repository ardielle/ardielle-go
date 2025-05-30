// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestTBinTypes(test *testing.T) {
	var bt BigTest
	j, err := os.ReadFile("../testdata/bigtest.json")
	if err != nil {
		test.Errorf("Cannot read JSON file: %v", err)
	}
	err = json.Unmarshal(j, &bt)
	if err != nil {
		test.Errorf("Cannot parse JSON: %v", err)
	}
	jj, err := json.Marshal(bt)
	fmt.Printf("BigTest data marshaled into %d bytes of JSON\n", len(jj))
	os.WriteFile("/tmp/test_bigtest.json", jj, 0644)
	tdata, err := Marshal(bt)
	if err != nil {
		test.Errorf("Cannot marshal TBin: %v", err)
	}
	os.WriteFile("/tmp/test_bigtest.tbin", tdata, 0644)
	fmt.Printf("BigStruct data marshaled into %d bytes of tbin\n", len(tdata))
	var g interface{}
	err = Unmarshal(tdata, &g)
	if err != nil {
		test.Errorf("Cannot unmarshal TBin: %v", err)
	}
	var gref interface{}
	err = json.Unmarshal(j, &gref)
	if err != nil {
		test.Errorf("Cannot unmarshal JSON: %v", err)
	}
	enc := NewEncoder(nil)
	enc.EncodeReflect(bt)
	err = enc.Error()
	if err != nil {
		test.Errorf("Cannot Encode: %v", err)
	}
	tdata2 := enc.Bytes()

	var bt2 BigTest
	r := bytes.NewBuffer(tdata2)
	dec := NewDecoder(r)
	v := reflect.ValueOf(&bt2).Elem()
	err = dec.DecodeReflect(v)
	if err != nil {
		test.Errorf("Cannot decode: %v", err)
	}

	//the annotated function is a bit like JSON, but every type is annotated, and map keys are sorted. This allows
	//a comparison that reflect.DeepEqual would fail in a way that is visible in text
	s1 := annotated(bt)
	s2 := annotated(bt2)
	if s1 != s2 {
		os.WriteFile("/tmp/anno1.txt", []byte(s1), 0644)
		os.WriteFile("/tmp/anno2.txt", []byte(s2), 0644)
		fmt.Println("Different. See /tmp/anno{1,2}.txt for details")
		test.Errorf("Decode doesn't match original")
	}
}

func (bt BigTest) MarshalTBin(enc *Encoder) error {
	signature := TypeSignature(bt)
	enc.WriteType(signature)
	enc.WriteSize(len(bt.Stuff))
	for _, bs := range bt.Stuff {
		enc.WriteString(bs.MyName)
		enc.WriteString(bs.MyUtfname)
		enc.WriteBool(bs.MyBool)
		enc.WriteInt8(bs.MyByte)
		enc.WriteInt16(bs.MyShort)
		enc.WriteInt32(bs.MyInt)
		enc.WriteInt64(bs.MyLong)
		enc.WriteFloat32(bs.MyFloat)
		enc.WriteFloat64(bs.MyDouble)

		enc.WriteUnsigned(len(bs.MyIntArray))
		for _, v := range bs.MyIntArray {
			enc.WriteInt32(v)
		}

		enc.WriteUnsigned(len(bs.MyStringArray))
		for _, v := range bs.MyStringArray {
			enc.WriteString(v)
		}

		enc.WriteUnsigned(len(bs.MyMap))
		for k, v := range bs.MyMap {
			enc.WriteString(k)
			enc.WriteInt32(v)
		}

		enc.WriteUUID(bs.MyUuid)

		enc.WriteString(string(bs.MyStringSubtype))
		enc.WriteInt32(int32(bs.MyInt32Subtype))
		enc.WriteFloat64(float64(bs.MyFloat64Subtype))
		enc.WriteTimestamp(bs.MyTime)
	}
	return nil
}
