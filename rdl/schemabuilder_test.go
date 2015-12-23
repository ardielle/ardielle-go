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
}
