// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package rdl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func loadTestSchema(test *testing.T, filename string) *Schema {
	schema, err := ParseRDLFile("../testdata/"+filename, false, false, true)
	if err != nil {
		test.Errorf("Cannot load schema (%s): %v", filename, err)
		return nil
	}
	fmt.Println("loaded", filename)
	return schema
}

func loadTestData(test *testing.T, filename string) *map[string]interface{} {
	var data map[string]interface{}
	bytes, err := ioutil.ReadFile("../testdata/" + filename)
	if err != nil {
		fmt.Printf("Cannot read data(%s): %v", filename, err)
		test.Errorf("Cannot read data(%s): %v", filename, err)
		return nil
	} else if err = json.Unmarshal(bytes, &data); err != nil {
		fmt.Printf("Cannot unmarshal data (%s): %v", filename, err)
		test.Errorf("Cannot unmarshal data (%s): %v", filename, err)
		return nil
	} else {
		fmt.Println("loaded", filename)
		return &data
	}
}

func assertStringEquals(test *testing.T, msg string, expected string, val string) bool {
	if val != expected {
		test.Errorf("Expected %s to be '%s', but it was '%s'", msg, expected, val)
		return false
	}
	return true
}

func TestBasicTypes(test *testing.T) {
	schema := loadTestSchema(test, "basictypes.rdl")
	if schema == nil {
		return
	}
	if !assertStringEquals(test, "namespace", "test", string(schema.Namespace)) {
		return
	}

	pdata := loadTestData(test, "basictypes.json")
	if pdata == nil {
		return
	}
	data := *pdata
	validation := Validate(schema, "Test", data)
	if validation.Error != "" {
		test.Errorf("Validation error: %v", validation)
	} else {
		if validation.Type != "" {
			if validation.Type == "Test" {
				fmt.Println("validated, determined the type to be", validation.Type)
			} else {
				test.Errorf("Validation error: chose the wrong type (should have been 'Test': %v", validation.Type)
			}
		} else {
			fmt.Println("Validation result:", validation)
		}
	}
}

func parseRDLString(s string) (*Schema, error) {
	r := strings.NewReader(s)
	return parseRDL(nil, "", r, true, true, false)
}

func TestParse(test *testing.T) {
	badRdl := `type Contact Struct { String foo; String foo; }`
	_, err := parseRDLString(badRdl)
	if err == nil {
		test.Errorf("Expected parse fail, it didn't")
	}
	badRdl = `type Foo Struct { String foo; } type Bar Foo { String foo; }`
	_, err = parseRDLString(badRdl)
	if err == nil {
		test.Errorf("Expected parse fail, it didn't")
	}
}
