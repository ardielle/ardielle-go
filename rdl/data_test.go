// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package rdl

import (
	"testing"
)

func TestTimestampEpoch(test *testing.T) {
	ts := TimestampNow()
	secs := ts.SecondsSinceEpoch()
	ts2 := TimestampFromEpoch(secs)
	if !ts.Equal(ts2) {
		test.Errorf("%v should be equal to %v", ts, ts2)
	}
}

func TestSymbolMap(test *testing.T) {
	s1 := Symbol("one")
	s1b := Symbol("one")
	s2 := Symbol("two")
	m := make(map[Symbol]string)
	m[s1] = "This is one"
	if len(m) != 1 {
		test.Errorf("map length failure: %v", m)
	}
	if m[s1] != "This is one" {
		test.Errorf("map[s1] failure")
	}
	if m[s1b] != "This is one" {
		test.Errorf("map[s1b] failure")
	}
	if _, ok := m["one"]; !ok {
		test.Errorf("map[\"one\"] failure, should be present")
	}
	if v, ok := m[s2]; ok {
		test.Errorf("map[s2] failure, should be absent: %v", v)
	}
}
