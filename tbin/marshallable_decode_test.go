// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"fmt"
	"os"
	"testing"
)

func (line *Polyline) UnmarshalTBin(dec *Decoder) error {
	//the signature is calculated via reflection, and is pretty slow.
	//signature := TypeSignature(line)
	//but can be generated with other code:
	//signature := Struct(Field("points", Array(Struct(Field("x", Int32), Field("y", Int32)))))
	signature := polylineSignature
	refSig, err := dec.ReadType()
	if err != nil {
		return err
	}
	if signature.String() != refSig.String() {
		return fmt.Errorf("Signature mismatch on decode: %v vs %v", signature, refSig)
	}
	size := dec.ReadSize()
	if dec.Error() == nil {
		proto := Polyline{}
		proto.Points = make([]*Point, size)
		for i := 0; i < size; i++ {
			x := dec.ReadInt32()
			y := dec.ReadInt32()
			proto.Points[i] = &Point{x, y}
		}
		*line = proto
		return nil
	}
	return dec.Error()
}

func TestTBinMarshallableDecode(test *testing.T) {
	var line Polyline
	tdata, err := os.ReadFile("../testdata/test.tbin")
	err = Unmarshal(tdata, &line)
	if err != nil {
		test.Errorf("Cannot unmarshal test.tbin: %v", err)
	}
	line2 := polyline()
	if !Equal(&line, line2) {
		test.Errorf("unmarshalled value not equal to reference value: %v\n%v", line, *line2)
	}
}
