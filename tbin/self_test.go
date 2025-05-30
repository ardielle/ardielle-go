// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ardielle/ardielle-go/rdl"
)

func TestRdlSchema(test *testing.T) {
	bytes, err := os.ReadFile("../testdata/rdl_schema.json")
	if err != nil {
		fmt.Printf("Cannot read file (rdl_schema.json): %v\n", err)
	}
	var schema rdl.Schema
	if err = json.Unmarshal(bytes, &schema); err != nil {
		fmt.Printf("Cannot parse schema: %v\n", err)
	}
	//fmt.Println(Pretty(schema))
	tbinData, err := Marshal(schema)
	if err != nil {
		fmt.Printf("Cannot encode tbin: %v\n", err)
	}
	os.WriteFile("../target/rdl_schema.tbin", tbinData, 0644)
	var schema2 rdl.Schema
	if err = Unmarshal(tbinData, &schema2); err != nil {
		fmt.Printf("Cannot decode schema2: %v\n", err)
	}
	if !Equal(schema, schema2) {
		fmt.Println("Hmm, not equal. See ../target/anno*.txt")
		anno1 := annotated(schema)
		os.WriteFile("../target/anno1.txt", []byte(anno1), 0644)
		anno2 := annotated(schema2)
		os.WriteFile("../target/anno2.txt", []byte(anno2), 0644)
	} else {
		fmt.Println("Schemas serialize correctly with tbin")
	}
}
