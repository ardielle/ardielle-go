// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package rdl

import (
	"testing"
)

func TestSchemaBuilder(test *testing.T) {
	schema, err := ParseRDLFile("../testdata/rdl.rdl", false, true, false)
	if err != nil {
		test.Errorf("Cannot load schema (rdl.rdl): %v", err)
		return
	}
	errmsg := CompareSchemas(schema, RdlSchema())
	if errmsg != "" {
		test.Errorf("TestSchemaBuilder: %s", errmsg)
		return
	}

	sb := NewSchemaBuilder("test")
	tb := NewStructTypeBuilder("Struct", "foo").Comment("description")
	tb.Field("field1", "Timestamp", false, nil, "The timestamp field")
	tb.Field("field2", "UUID", false, nil, "The uuid field")
	sb.AddType(tb.Build())
	schema = sb.Build()
	if schema == nil {
		test.Errorf("TestSchemaBuilder: Cannot build schema with certain base types")
	}
}
