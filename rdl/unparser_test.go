// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package rdl

import (
	"bufio"
	"bytes"
	"testing"
)

func TestUnparse(test *testing.T) {
	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)
	schema := loadTestSchema(test, "basictypes.rdl")
	err := UnparseRDL(schema, writer)
	if err != nil {
		test.Errorf("Cannot unparse to RDL: %v", err)
	}
}
