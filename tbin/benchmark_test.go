// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

/*

Polyline test data, with 13 points.

Size:
 JSON:                      250 bytes
 TBIN generic:              160 bytes
 Protobuf                   96 bytes (+ schema = 308)
 TBIN optimized:            70 bytes (- schema = 46 bytes, i.e. in a session where types are reused)
 Avro                       46 bytes (+ schema = 230 bytes)

TBin encodes the schema inline with the data (for its first occurence), unlike PB and Avro. This eliminates
the need to manage it separetely, although the default case looks bigger than avro as a result. In this test
the combined schema + data is still smaller than protobuf data only

Marshal ops/ns:
 JSON generic:              25547
 JSON marshallable:         12511   // used with RDL only for unions
 JSON reflect:              5991    // common with RDL models

 TBIN generic:              22878
 TBIN reflect:              11938   // easiest to use, handles unions, etc. Not optimized.
 TBIN marshallable:         4638    // common with RDL models (using codegen)
 TBIN marshallable inlined: 4544    // smarter codegen could achieve this, or you can hand code this

Unmarshal ops/ns:
 JSON unmarshallable:       59754   // common with RDL models
 JSON generic:              24573   // decode to map
 JSON reflect:              22277   // no validation of object to type (i.e. required fields, etc)

 TBIN reflect:              23029
 TBIN generic:              20364
 TBIN unmarshallable:       10535   // common with RDL models

*/

var _ = fmt.Println

func BenchmarkJSONMarshalReflect(b *testing.B) {
	line := polyline()
	var jdata []byte
	for n := 0; n < b.N; n++ {
		jdata, _ = json.Marshal(line)
	}
	if len(jdata) != testDataLengthJSON {
		b.Errorf("Expected %d bytes of JSON encoded data, got %d", testDataLengthJSON, len(jdata))
	}
}

//this forces reflection, bypassing any TBinMarshallable functionality. If you write no special code
//for your types (i.e. no MarshalTBin method), this is what happens.
func BenchmarkTBinMarshalReflect(b *testing.B) {
	line := polyline()
	var tdata []byte
	//signature := TypeSignature(line) //factored out of the benchmark, it is constant for the type
	for n := 0; n < b.N; n++ {
		enc := NewEncoder(nil)
		//enc.WriteType(signature)
		enc.EncodeReflect(line)
		//enc.EncodeType(line, signature)
		tdata = enc.Bytes()
	}
	if len(tdata) != testDataLengthTBinBest {
		b.Errorf("Expected %d bytes of TBIN encoded data, got %d", testDataLengthTBinBest, len(tdata))
	}
}

//The normal default, invoke the TBinMarshallable method when present (as it is in this test)
//Note that the TBinMarshallable code in the example factors out the signature calculation, since
//it is constant. So this ends up being not much slower than the CodeGen benchmark below.
func BenchmarkTBinMarshalUser(b *testing.B) {
	line := polyline()
	var tdata []byte
	for n := 0; n < b.N; n++ {
		tdata, _ = Marshal(line)
	}
	if len(tdata) != testDataLengthTBinBest {
		b.Errorf("Expected %d bytes of TBIN encoded data, got %d", testDataLengthTBinBest, len(tdata))
	}
}

//inlined encoding. Generated code could produce this sort of thing, with the signature as a generated constant
func BenchmarkTBinMarshalCodeGen(b *testing.B) {
	line := polyline()

	signature := TypeSignature(line) //factored out of the benchmark, it is constant for the type
	var tdata []byte
	for n := 0; n < b.N; n++ {
		encoder := NewEncoder(nil)
		if encoder.Error() == nil {
			encoder.WriteType(signature)
			encoder.WriteSize(len(line.Points))
			for _, pt := range line.Points {
				encoder.WriteInt32(pt.X)
				encoder.WriteInt32(pt.Y)
			}
		}
		if encoder.Error() != nil {
			b.Errorf("*** %v", encoder.Error())
		} else {
			tdata = encoder.Bytes()
		}
	}
	if len(tdata) != testDataLengthTBinBest {
		b.Errorf("Expected %d bytes of encoding, got %d", testDataLengthTBinBest, len(tdata))
	}

}

//The normal default, invoke the MarshalJSON method when present (as it is in this test)
func BenchmarkJsonMarshalGeneric(b *testing.B) {
	line := polyline()
	td, _ := Marshal(line)
	var generic interface{}
	Unmarshal(td, &generic)
	for n := 0; n < b.N; n++ {
		json.Marshal(generic)
	}
}

//The normal default, invoke the TBinMarshallable method when present (as it is in this test)
func BenchmarkTBinMarshalGeneric(b *testing.B) {
	line := polyline()
	td, _ := Marshal(line)
	var generic interface{}
	Unmarshal(td, &generic)
	var tdata []byte
	for n := 0; n < b.N; n++ {
		tdata, _ = Marshal(generic)
	}
	if len(tdata) != testDataLengthTBinGeneric {
		b.Errorf("Expected %d bytes of TBIN encoded data, got %d", testDataLengthTBinGeneric, len(tdata))
	}
}

func BenchmarkJsonUnmarshalGeneric(b *testing.B) {
	line := polyline()
	jd, _ := json.Marshal(line)
	for n := 0; n < b.N; n++ {
		var generic interface{}
		json.Unmarshal(jd, &generic)
	}
}

func BenchmarkTBinUnmarshalGeneric(b *testing.B) {
	line := polyline()
	jd, _ := Marshal(line)
	for n := 0; n < b.N; n++ {
		var generic interface{}
		Unmarshal(jd, &generic)
	}
}

func BenchmarkJSONUnmarshalUser(b *testing.B) {
	line := polyline()
	jd, _ := json.Marshal(line)
	for n := 0; n < b.N; n++ {
		var line2 Polyline
		json.Unmarshal(jd, &line2)
	}
}

func BenchmarkTBinUnmarshalUser(b *testing.B) {
	line := polyline()
	jd, _ := Marshal(line)
	for n := 0; n < b.N; n++ {
		var line2 Polyline
		Unmarshal(jd, &line2)
	}
}

func BenchmarkTBinUnmarshalReflect(b *testing.B) {
	line := polyline()
	tdata, _ := Marshal(line)
	for n := 0; n < b.N; n++ {
		var line2 Polyline
		r := bytes.NewBuffer(tdata)
		dec := NewDecoder(r)
		v := reflect.ValueOf(&line2).Elem()
		dec.DecodeReflect(v)
	}
}
