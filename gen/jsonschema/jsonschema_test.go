package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ardielle/ardielle-go/rdl"
)

func generate(filename string) (*rdl.Schema, JSONSchema, error) {
	schema, err := rdl.ParseRDLFile("../../testdata/"+filename, false, false, true)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("loaded", filename)
	js, err := Generate(schema)
	return schema, js, err
}

func pretty(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Sprintf("*** %v", err)
	}
	return string(b)
}

func schemaVersion(js JSONSchema) string {
	if v, ok := js["$schema"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func typeDefs(js JSONSchema) map[string]map[string]interface{} {
	if v, ok := js["definitions"]; ok {
		if defs, ok := v.(map[string]map[string]interface{}); ok {
			return defs
		} else {
			fmt.Println("what?!")
		}
	}
	return nil
}

func typeProperties(typedef map[string]interface{}) map[string]interface{} {
	if v, ok := typedef["properties"]; ok {
		if p, ok := v.(map[string]interface{}); ok {
			return p
		}
	}
	return nil
}

func TestMinimal(test *testing.T) {
	schema, js, err := generate("polyline.rdl")
	if err != nil {
		test.Errorf("TestMinimal: %v", err)
	}
	fmt.Println("schema:", schema)

	if schemaVersion(js) != "http://json-schema.org/draft-04/schema#" {
		test.Errorf("Result does not have a recognized JSON schema version ('$schema') value: %q", schemaVersion(js))
		return
	}
	defs := typeDefs(js)
	if len(defs) != 2 {
		test.Errorf("Schema should have 2 definitions, but only %d\n", len(defs))
		return
	}
	ptdef := defs["Point"]
	props := typeProperties(ptdef)
	if len(props) != 2 {
		test.Errorf("Point typedef should have 2 properties, but found %d\n", len(props))
		return
	}
	ptdef = defs["Polyline"]
	props = typeProperties(ptdef)
	if len(props) != 1 {
		test.Errorf("Polyline typedef should have 1 property, but found %d\n", len(props))
		return
	}

	fmt.Println("generated:")
	//	fmt.Println(pretty(js))
	fmt.Println(js)
}
