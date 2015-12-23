// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package rdl

import (
	"testing"
)

func check(test *testing.T, v1 interface{}, v2 interface{}, expected bool) {
	if Equal(v1, v2) != expected {
		test.Errorf("Equal(%v, %v) expected %v but got %v", v1, v2, expected, !expected)
		panic("here")
	}
}

func TestEqual(test *testing.T) {
	check(test, true, true, true)
	check(test, true, false, false)
	check(test, false, true, false)
	check(test, false, false, true)

	check(test, 23, 23, true)
	check(test, 23, 57, false)
	check(test, 57, 23, false)
	check(test, 23.0, 23.0, true)
	check(test, 23.0, 23, false)
	check(test, 23, 23.0, false)
	check(test, int8(23), int16(23), false)
	check(test, int(23), int32(23), false)
	check(test, float32(23.57), float64(23.57), false)
	check(test, float64(23.57), float64(23.57), true)

	check(test, "foo", "foo", true)
	check(test, "foo", "bar", false)

	check(test, []byte{1, 2}, []byte{1, 2}, true)
	check(test, []byte{1, 2}, []byte{1, 3}, false)
	check(test, []byte{1, 2}, []byte{1, 2, 3}, false)

	ts1 := TimestampNow()
	ts2, _ := TimestampParse("2015-05-17T01:37:09.534Z")
	ts3, _ := TimestampParse("2015-05-17T01:37:09.534Z")
	check(test, ts1, ts1, true)
	check(test, ts1, ts2, false)
	check(test, ts2, ts3, true)
	check(test, ts2, "2015-05-17T01:37:09.534Z", false)

	check(test, Symbol("foo"), Symbol("foo"), true)
	check(test, Symbol("foo"), Symbol("bar"), false)
	check(test, Symbol("foo"), "foo", false)
	check(test, "foo", Symbol("foo"), false)

	u1 := ParseUUID("88cfd476-fc35-11e4-acaa-14109fe4729f")
	u2 := ParseUUID("9f57aa86-fc35-11e4-9a55-14109fe4729f")
	u3 := ParseUUID("9f57aa86-fc35-11e4-9a55-14109fe4729f")
	check(test, u1, u1, true)
	check(test, u1, u2, false)
	check(test, u2, u3, true)
	check(test, u2, "9f57aa86-fc35-11e4-9a55-14109fe4729f", false)

	//array

	//map

	//struct

	//union

	//enum
}
