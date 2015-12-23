// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestTBinMarshalDecodeForceReflect(test *testing.T) {
	var line Polyline
	tdata, err := ioutil.ReadFile("../testdata/test.tbin")
	r := bytes.NewBuffer(tdata)
	dec := NewDecoder(r)
	if false {
		v := reflect.ValueOf(&line).Elem()
		dec.DecodeReflect(v)
	} else {
		dec.Decode(&line)
	}
	err = dec.Error()
	if err != nil {
		test.Errorf("Cannot decode: %v", err)
	}
}
