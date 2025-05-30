// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

//the model is generated in polyline_model.go. Here we add methods to the Polyline type

// this is constant, once calculated, so could be produced from code generation
var polylineSignature = Struct(Field("points", Array(Struct(Field("x", Int32, false), Field("y", Int32, false))), false))

func (line Polyline) MarshalTBin(enc *Encoder) error {
	//the signature is calculated via reflection, and is pretty slow.
	//signature := TypeSignature(line)
	//But it is constant for the type, so the following code could be generated instead
	//signature := Struct(Field("points", Array(Struct(Field("x", Int32), Field("y", Int32)))))
	//even that has sizable cost, so have a global to hold it is quite a bit better:
	signature := polylineSignature
	enc.WriteType(signature)
	enc.WriteSize(len(line.Points)) //the array length
	for _, pt := range line.Points {
		enc.WriteInt32(pt.X) //and the data
		enc.WriteInt32(pt.Y) //for each item, packed
	}
	return enc.Error()
}

func TestTBinMarshallableEncode(test *testing.T) {
	line := polyline()

	jdata, err := json.Marshal(line)
	fmt.Println("json is", len(jdata), "bytes long")
	os.WriteFile("/tmp/test_marshal.json", jdata, 0644)

	tdata, err := Marshal(line)
	if err != nil {
		test.Errorf("Cannot marshal Polyline:, %v", err)
	} else {
		fmt.Println("tbin is", len(tdata), "bytes long")
		os.WriteFile("/tmp/test_marshal.tbin", tdata, 0644)
		if len(tdata) != testDataLengthTBinBest {
			test.Errorf("TBin data is not the right length. Expected %d, produced %d", testDataLengthTBinBest, len(tdata))
		}
	}
}
