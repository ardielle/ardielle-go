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

func TestMapDefinition(t *testing.T) {
	schema := loadTestSchema(t, "maptest.rdl")
	if schema == nil {
		return
	}
	reg := NewTypeRegistry(schema)
	ty := reg.FindType("AttachedVolume")
	switch ty.Variant {
	case TypeVariantMapTypeDef:
	default:
		t.Errorf("parsed Map type was inlined when it should not have been")
	}
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
	parseGoodRDL(test, `type A String (pattern="[a-z]"); type B A; type C B; type D string (pattern="{C}-{A}");`)
	parseGoodRDL(test, `type foo struct { String foo; }`)
	parseGoodRDL(test, `type Bar enum { ONE TWO }`)
	parseGoodRDL(test, `
type MultiLine Enum {
	ONE
	TWO
}`)

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

func TestConsumes(test *testing.T) {
	schema, err := parseRDLString(`
resource Any GET "/foo" {
  consumes application/json, application/xml    ,   text/plain   // some comment
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with consumes: %v", err)
	} else {
		if len(schema.Resources) != 1 {
			test.Errorf("Did not parse expected number of resources: %v", schema)
		}
		r := schema.Resources[0]
		if len(r.Consumes) != 3 {
			test.Errorf("Did not parse expected number of consumes: %v", len(r.Consumes))
		}
		if r.Consumes[0] != "application/json" {
			test.Errorf("Did not parse consumes value correctly: %v (expected: application/json)", r.Consumes[0])
		}
		if r.Consumes[1] != "application/xml" {
			test.Errorf("Did not parse consumes value correctly: %v (expected: application/xml)", r.Consumes[1])
		}
		if r.Consumes[2] != "text/plain" {
			test.Errorf("Did not parse consumes value correctly: %v (expected: text/plain)", r.Consumes[2])
		}
		if r.Comment != "some comment" {
			test.Errorf("Did not parse trailing comment correctly: %v", r.Comment)
		}
	}
}

func TestProduces(test *testing.T) {
	schema, err := parseRDLString(`
resource Any GET "/foo" {
  produces application/json, application/xml    ,   text/plain   // some comment
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with produces: %v", err)
	} else {
		if len(schema.Resources) != 1 {
			test.Errorf("Did not parse expected number of resources: %v", schema)
		}
		r := schema.Resources[0]
		if len(r.Produces) != 3 {
			test.Errorf("Did not parse expected number of produces: %v", len(r.Produces))
		}
		if r.Produces[0] != "application/json" {
			test.Errorf("Did not parse produces value correctly: %v (expected: application/json)", r.Produces[0])
		}
		if r.Produces[1] != "application/xml" {
			test.Errorf("Did not parse produces value correctly: %v (expected: application/xml)", r.Produces[1])
		}
		if r.Produces[2] != "text/plain" {
			test.Errorf("Did not parse produces value correctly: %v (expected: text/plain)", r.Produces[2])
		}
		if r.Comment != "some comment" {
			test.Errorf("Did not parse trailing comment correctly: %v", r.Comment)
		}
	}
}

func TestResourceName(test *testing.T) {
	schema, err := parseRDLString(`
resource Any GET "/nil" {
  expected OK;
}

resource Any GET "/foo" (name=myFoo) {
  expected OK;
}

resource Any GET "/bar" (async, name    =  myBar, x_something   ) {
  expected OK;
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with resource name: %v", err)
	} else {
		if len(schema.Resources) != 3 {
			test.Errorf("Did not parse expected number of resources: %v", schema)
		}
		if schema.Resources[0].Name != "" {
			test.Errorf("Did not parse resource name correctly (expected nil)")
		}
		r := schema.Resources[1]
		if r.Name != "myFoo" {
			test.Errorf("Did not parse resource name correctly: %v (expected myFoo)", r.Name)
		}
		r = schema.Resources[2]
		if r.Name != "myBar" {
			test.Errorf("Did not parse resource name correctly: %v (expected myBar)", r.Name)
		}
	}
}

func TestStructFieldRestrictions(test *testing.T) {
	schema, err := parseRDLString(`
type Foo Struct {
    String (pattern="y_*") bar2 (optional); //normal syntax, the options are onthe type
    String bar (optional, pattern="y_*"); //alternate syntax: the options for the field are applied to type
    String blah (maxsize=20, minsize=5, x_foo="hey");
    String hmm (values=["one","two","three"])
    UUID id (values=["901dfb52-39b5-11e7-adba-6c4008a30aa6"], optional)
    Timestamp ts (values=["2017-05-15T21:30:10.742Z"], optional)
    Symbol sym (values=["one","two"])
    Int32 num (max=100,min=50)
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with resource name: %v", err)
	}
	if len(schema.Types) != 9 {
		test.Errorf("expected 5 types in schema, found %d", len(schema.Types))
	}
}

func TestNestedTypes(test *testing.T) {
	var err error
	_, err = parseRDLString(`
type Foo Struct {
    Struct {
        String name
    } bar;
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL with resource name: %v", err)
	}
	_, err = parseRDLString(`
type Foo Struct {
    Struct {
        Struct {
            String name
        } foo;
    } bar;
}
`)
}

func TestEnumComments(test *testing.T) {
	var err error
	egood, err := parseRDLString(`
// Comment for TestEnum
type TestEnum enum {
    ONE, // Comment for ONE
    TWO // Comment for TWO
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
	}
	ebad, err := parseRDLString(`
// Comment for TestEnum
type TestEnum enum {

    // Comment for ONE
    ONE,

    // Comment for TWO
    TWO
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
	}
	type1 := egood.Types[0] //.EnumTypeDef.Elements[0]
	type2 := ebad.Types[0]  //.EnumTypeDef.Elements[0]
	if !EquivalentTypes(type1, type2) {
		test.Errorf("Types don't match: %v, %v", type1, type2)
	}
}

func TestFieldComments(test *testing.T) {
	var err error
	t1, err := parseRDLString(`
//type comment
type TestStruct Struct {
    String one; //comment for field 1
    String two; //comment for field 2
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
	}
	t2, err := parseRDLString(`
//type comment
type TestStruct Struct {

    //comment for field 1
    String one;

    //comment for field 2
    String two;
}
`)
	type1 := t1.Types[0]
	type2 := t2.Types[0]
	if !EquivalentTypes(type1, type2) {
		test.Errorf("Types don't match: %v, %v", type1, type2)
	}
}

func EquivalentTypes(t1, t2 *Type) bool {
	//cheesy
	b1, err := json.Marshal(t1)
	if err != nil {
		return false
	}
	b2, err := json.Marshal(t2)
	if err != nil {
		return false
	}
	return string(b1) == string(b2)
}

func TestEnumElementComments(test *testing.T) {
	var err error
	s1, err := parseRDLString(`
//type comment
type TestEnum Enum {
    ONE (x_index="1"),
    TWO (x_index="2"), //comment for TWO
    THREE, //comment for THREE
    FOUR (x_index="4") //comment for FOUR
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
		return
	}
	if s1.Types[0].EnumTypeDef.Name != "TestEnum" || len(s1.Types[0].EnumTypeDef.Elements) != 4 {
		test.Errorf("Enum type parsed incorrectly: %v", s1)
	}
	e := s1.Types[0].EnumTypeDef.Elements
	if e[0].Annotations["x_index"] != "1" || e[1].Annotations["x_index"] != "2" || e[3].Annotations["x_index"] != "4" {
		test.Errorf("Enum type parsed incorrectly: %v", s1)
	}
}

func TestSchemaAnnotations(test *testing.T) {
	_, err := parseRDLString(`
//this is a schema annotation test
name foo;
version 1
x_something="23"
x_blah = "blah"

type Foo Struct {
    String text
}
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
		return
	}
}

func TestAliasAnnotation(test *testing.T) {
	_, err := parseRDLString(`
	   type DateTime string (x_date_time)
	   type Response struct {
	     DateTime dateTime (optional);
	   }`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
		return
	}
}

func TestAliasAnnotation2(test *testing.T) {
	v, err := parseRDLString(`
	   type MyBase Struct { String message; }
	   type MySubtype1 MyBase;
	   type MySubtype2 MyBase (x_y="z");`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
		return
	}
	//
	for _, td := range v.Types {
		if td.AliasTypeDef != nil {
			if td.AliasTypeDef.Type == "___forward_reference___" {
				test.Errorf("Improperly parsed alias with annotations: %v", td)
			}
			if td.AliasTypeDef.Name == "MySubtype2" {
				if td.AliasTypeDef.Annotations == nil || len(td.AliasTypeDef.Annotations) != 1 {
					test.Errorf("Annotations did not survive the parse: %v", td)
				}
				for k, v := range td.AliasTypeDef.Annotations {
					if k != "x_y" || v != "z" {
						test.Errorf("Annotations did not survive the parse: %v", td)
					}
				}
			}
		}
	}
}
