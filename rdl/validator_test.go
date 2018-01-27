package rdl

import (
	"fmt"
	"testing"
)

func TestValidatorBasicTypes(test *testing.T) {
	schema := RdlSchema()

	pdata := loadTestData(test, "basictypes_schema.json")
	if pdata == nil {
		return
	}
	expectedType := "Schema"
	data := *pdata
	validation := Validate(schema, expectedType, data)
	if validation.Error != "" {
		test.Errorf("Validation error: %v", validation)
	} else {
		if validation.Type != "" {
			if validation.Type == expectedType {
				fmt.Println("validated, determined the type to be", validation.Type)
			} else {
				test.Errorf("Validation error: chose the wrong type (should have been '%s': %v", expectedType, validation.Type)
			}
		} else {
			fmt.Println("Validation result:", validation)
		}
	}
}

type foo struct {
}

func (f *foo) Validate() error {
	return nil
}

func TestValidatorCustomTypes(test *testing.T) {

	sb := NewSchemaBuilder("test")
	tb := NewStructTypeBuilder("Struct", "foo").Comment("description")
	tb.Field("field1", "Timestamp", false, nil, "The timestamp field")
	tb.Field("field2", "UUID", false, nil, "The uuid field")
	sb.AddType(tb.Build())

	ta := NewAliasTypeBuilder("Timestamp", "mytimestamp")
	sb.AddType(ta.Build())

	ts := NewAliasTypeBuilder("string", "mystring")
	sb.AddType(ts.Build())

	tIdentifier := NewStringTypeBuilder("Identifier")
	tIdentifier.Comment("All names need to be of this restricted string type")
	tIdentifier.Pattern("[a-zA-Z_]+[a-zA-Z_0-9]*")
	sb.AddType(tIdentifier.Build())

	tTypeName := NewAliasTypeBuilder("Identifier", "TypeName")
	tTypeName.Comment("The identifier for an already-defined type")
	sb.AddType(tTypeName.Build())

	// Build the schema
	schema, err := sb.BuildParanoid()
	if err != nil {
		test.Fatal(err)
	}

	// Types that define their own Validate can do whatever they want regarding
	// whether a type validates or not (including not checking sub fields)
	positive := []struct {
		tname string
		v     interface{}
	}{
		{"string", "basic string"},
		{"mytimestamp", "2017-04-20T15:04:05.999Z"},
		{"foo", &foo{}},
	}

	for _, t := range positive {
		v := Validate(schema, t.tname, t.v)
		if v.Error != "" {
			test.Errorf("Validation error for type: %v", v)
		}
	}
	negative := []struct {
		tname string
		v     interface{}
	}{
		{"mytimestamp", "20170420T15:04:05.999Z"},
		{"foo", 4587},
		{"string", &foo{}},
	}

	for _, t := range negative {
		v := Validate(schema, t.tname, t.v)
		if v.Error == "" {
			test.Errorf("Validated incorrect type: %v", v)
		}
	}

}

func TestValidateStringSubtype(test *testing.T) {
	schema, err := parseRDLString(`
type Foo String (pattern="[a-z]*")
type Bar Foo (maxSize=4);
`)
	if err != nil {
		test.Errorf("cannot parse valid RDL: %v", err)
		return
	}
	validation := Validate(schema, "Bar", "1")
	if validation.Error == "" {
		test.Errorf("validation error did not occur, string has non-alpha characters: %v\nschema is: %v", validation, schema)
	}
	validation = Validate(schema, "Bar", "abcdef")
	if validation.Error == "" {
		test.Errorf("Validation error did not occur, string is too long: %v\nschema is: %v", validation, schema)
	}
}
