// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"io/ioutil"
	"testing"
)

func TestTBinMarshalEncodeForceReflect(test *testing.T) {
	data := polyline()

	enc := NewEncoder(nil)
	enc.EncodeReflect(data)
	err := enc.Error()
	if err != nil {
		test.Errorf("Cannot Encode: %v", err)
	}
	tdata := enc.Bytes()

	ioutil.WriteFile("../target/test_reflect.tbin", tdata, 0644)
}

func TestTBinMarshalEncodeForceReflect2(test *testing.T) {
	data := rect(1, 2, 3, 4)

	enc := NewEncoder(nil)
	enc.EncodeReflect(data)
	err := enc.Error()
	if err != nil {
		test.Errorf("Cannot Encode: %v", err)
	}
	tdata := enc.Bytes()
	if len(tdata) != 26 {
		test.Errorf("Bad encoding, expected 26 bytes, got %d", len(tdata))
	}
	ioutil.WriteFile("../target/test_rect.tbin", tdata, 0644)
}

func rectShape(r *Rect) *Shape {
	data := new(Shape)
	data.Variant = ShapeVariantRect
	data.Rect = r
	return data
}

func lineShape(l *Polyline) *Shape {
	data := new(Shape)
	data.Variant = ShapeVariantPolyline
	data.Polyline = l
	return data
}

func TestTBinMarshalEncodeForceReflect3(test *testing.T) {
	/*
		pointSig := Struct(Field("x", Int32), Field("y", Int32))
		lineSig := Struct(Field("points", pointSig))
		rectSig := Struct(Field("p1", pointSig), Field("p2", pointSig))
		sig := Union(lineSig, rectSig)
	*/
	//this is the generated signature I should use.
	// the String() method produces my human readable signature, but the object needs to parsing to traverse.
	// each type, as it gets traversed, can get its own tag assigned.
	// a Bytes() method would be nice, but needs the tag table for an encoder stream to bind them.
	//	panic("HERE")

	data := new(Drawing)
	data.Shapes = []*Shape{rectShape(rect(1, 2, 3, 4)), lineShape(polyline())}
	//data := rectShape(rect(1, 2, 3, 4))

	enc := NewEncoder(nil)
	enc.EncodeReflect(data)
	err := enc.Error()
	if err != nil {
		test.Errorf("Cannot Encode: %v", err)
	}
	tdata := enc.Bytes()
	ioutil.WriteFile("/tmp/test_drawing.tbin", tdata, 0644)
	if len(tdata) != 107 {
		test.Errorf("Encoding produced %d bytes, expected 107", len(tdata))
	}
}
