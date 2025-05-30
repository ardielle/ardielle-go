// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"fmt"
	"os"
	"testing"
)

func TestTBinMarshalGenericEncode(test *testing.T) {
	var line interface{}
	tdata, err := os.ReadFile("../testdata/test.tbin")
	err = Unmarshal(tdata, &line)
	if err != nil {
		test.Errorf("Cannot unmarshal test.tbin: %v", err)
	}
	err = Unmarshal(tdata, &line)
	if err != nil {
		test.Errorf("Cannot unmarshal test.tbin: %v", err)
	}
	tdata2, err := Marshal(line)
	if err != nil {
		test.Errorf("Cannot marshal generic data:, %v", err)
	} else {
		fmt.Println("tbin (generic) is", len(tdata2), "bytes long")
		os.WriteFile("../target/test_generic.tbin", tdata2, 0644)
	}
}
