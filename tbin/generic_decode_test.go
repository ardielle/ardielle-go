// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"encoding/json"
	"os"
	"testing"
)

func TestTBinMarshalGenericDecode(test *testing.T) {
	var line interface{}
	tdata, err := os.ReadFile("../testdata/test.tbin")
	err = Unmarshal(tdata, &line)
	if err != nil {
		test.Errorf("Cannot unmarshal test.tbin: %v", err)
	}
	fromJSON, err := os.ReadFile("../testdata/test.json")
	if err != nil {
		test.Errorf("Cannot read test.json")
	}

	fromTbin, _ := json.Marshal(line)
	if !Equal(fromTbin, fromJSON) {
		test.Errorf("Original data serialized and generically read is different than the original")
	}
}
