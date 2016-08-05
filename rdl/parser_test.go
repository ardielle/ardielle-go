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

func parseGoodRDL(test *testing.T, s string) {
	_, err := parseRDLString(s)
	if err != nil {
		test.Errorf("Fail to parse (%s): %v", s, err)
	}
}

func parseBadRDL(test *testing.T, s string) {
	_, err := parseRDLString(s)
	if err == nil {
		test.Errorf("Expected failure parsing (%s), but it didn't fail", s)
	}
}

func TestParse(test *testing.T) {
	parseBadRDL(test, `type Contact Struct { String foo; String foo; }`)
	parseBadRDL(test, `type Foo Struct { String foo; } type Bar Foo { String foo; }`)
	parseBadRDL(test, `type Foo Struct { String bar; } resource Foo GET "/foo?d={debug}" {String debug (optinal); }`)
	parseGoodRDL(test, `type Foo Any; type X Struct { Any y; } type Y Struct { Foo y;}`)

	schema, err := parseRDLString(`type Base Struct { String bar; } type Foo Base;`)
	if err != nil {
		test.Errorf("Cannot parse: %v\n", err)
	} else {
		reg := NewTypeRegistry(schema)
		t1 := reg.FindType("Foo")
		if t1 == nil {
			test.Errorf("Expected type, found nothing")
		} else {
			_, tSuper, _ := TypeInfo(t1)
			assertStringEquals(test, "supertype", "Base", string(tSuper))
		}
	}
}

func hasAnnotation(annotations map[ExtendedAnnotation]string, name string, value string) bool {
	if annotations == nil {
		return false
	}
	v, ok := annotations[ExtendedAnnotation(name)]
	if !ok {
		return false
	}
	if v != value {
		return false
	}
	return true
}

func TestAnnotations(test *testing.T) {
	schema, err := parseRDLString(`
type MyType Struct (x_one="two", x_three="four") {
  String myId (x_hasRemoteId="yes");
  String myField;
  String badName (x_name="goodName");
  Int32 hasCustomRange (x_range="1;127"); // in practice would use rdl's built-in support for ranges
}
resource MyType GET "/foo/{bar}" (x_r_one="two", x_r_three="four") {
  String bar (x_five="bletch")
  String glorp (out, header="X_GLORP", x_whatever="xxx", x_oh_yeah)
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with annotations: %v", err)
	} else {
		reg := NewTypeRegistry(schema)
		t1 := reg.FindType("MyType")
		if t1 == nil || t1.Variant != TypeVariantStructTypeDef || t1.StructTypeDef == nil {
			test.Errorf("Bad 'MyType' definition")
		} else {
			if !hasAnnotation(t1.StructTypeDef.Annotations, "x_one", "two") {
				test.Errorf("Bad 'x_one' annotation")
			}
			if !hasAnnotation(t1.StructTypeDef.Annotations, "x_three", "four") {
				test.Errorf("Bad 'x_three' annotation")
			}
			for _, f := range t1.StructTypeDef.Fields {
				if f.Name == "myId" {
					if !hasAnnotation(f.Annotations, "x_hasRemoteId", "yes") {
						test.Errorf("Bad annotation on struct field 'myId'")
					}
				}
			}
			if len(schema.Resources) != 1 {
				test.Errorf("Did not parse expected number of resources: %v", schema)
			}
			for _, r := range schema.Resources {
				if !hasAnnotation(r.Annotations, "x_r_one", "two") {
					test.Errorf("Bad annotation on resource: 'x_r_one'")
				}
				if !hasAnnotation(r.Annotations, "x_r_three", "four") {
					test.Errorf("Bad annotation on resource: 'x_r_three'")
				}
				for _, i := range r.Inputs {
					if i.Name == "bar" && !hasAnnotation(i.Annotations, "x_five", "bletch") {
						test.Errorf("Bad annotation on resource input parameter 'bar'")
					}
				}
				for _, o := range r.Outputs {
					if o.Name == "glorp" && !hasAnnotation(o.Annotations, "x_whatever", "xxx") {
						test.Errorf("Bad annotation on resource output parameter 'glorp'")
					}
				}
			}
		}
	}
}

func TestRecursive(test *testing.T) {
	_, err := parseRDLString(`
type Node Struct {
  Node left;
  String value;
  Node right (optional);
}
`)
	if err == nil {
		test.Errorf("recursive field must be optional")
	}

	_, err = parseRDLString(`
type Node Struct {
  Node left (optional);
  String value;
  Node right (optional);
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
	}
}

func TestIncludeTypeLookup(test *testing.T) {
	//this tests that type lookup is correct across multiple included files
	loadTestSchema(test, "k1_a.rdl")
}

func TestAccept(test *testing.T) {
	schema, err := parseRDLString(`
resource Any GET "/foo" {
  accept application/json, application/xml    ,   text/plain   // some comment
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with accept: %v", err)
	} else {
		if len(schema.Resources) != 1 {
			test.Errorf("Did not parse expected number of resources: %v", schema)
		}
		r :=  schema.Resources[0]
		if len(r.Accept) != 3 {
			test.Errorf("Did not parse expected number of accept: %v", len(r.Accept))
		}
		if r.Accept[0] != "application/json" {
			test.Errorf("Did not parse accept value correctly: %v (expected: application/json)", r.Accept[0])
		}
		if r.Accept[1] != "application/xml" {
			test.Errorf("Did not parse accept value correctly: %v (expected: application/xml)", r.Accept[1])
		}
		if r.Accept[2] != "text/plain" {
			test.Errorf("Did not parse accept value correctly: %v (expected: text/plain)", r.Accept[2])
		}
		if r.Comment != "some comment" {
			test.Errorf("Did not parse trailing comment correctly: %v", r.Comment)
		}
	}
}

func TestContentType(test *testing.T) {
	schema, err := parseRDLString(`
resource Any GET "/foo" {
  content-type application/json, application/xml    ,   text/plain   // some comment
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with accept: %v", err)
	} else {
		if len(schema.Resources) != 1 {
			test.Errorf("Did not parse expected number of resources: %v", schema)
		}
		r :=  schema.Resources[0]
		if len(r.ContentType) != 3 {
			test.Errorf("Did not parse expected number of accept: %v", len(r.ContentType))
		}
		if r.ContentType[0] != "application/json" {
			test.Errorf("Did not parse accept value correctly: %v (expected: application/json)", r.ContentType[0])
		}
		if r.ContentType[1] != "application/xml" {
			test.Errorf("Did not parse accept value correctly: %v (expected: application/xml)", r.ContentType[1])
		}
		if r.ContentType[2] != "text/plain" {
			test.Errorf("Did not parse accept value correctly: %v (expected: text/plain)", r.ContentType[2])
		}
		if r.Comment != "some comment" {
			test.Errorf("Did not parse trailing comment correctly: %v", r.Comment)
		}
	}
}
