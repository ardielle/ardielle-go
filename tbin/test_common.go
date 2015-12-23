// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package tbin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

var _ = testing.Verbose

const testDataJSON = `{"points":[{"x":1,"y":11},{"x":2,"y":22},{"x":3,"y":33},{"x":10,"y":100},{"x":-23,"y":100},{"x":-23,"y":-33},{"x":10,"y":-33},{"x":103,"y":333},{"x":300,"y":1000},{"x":1234,"y":1234},{"x":12345678,"y":12321312},{"x":321321321,"y":33},{"x":1,"y":11}]}`

const testDataLengthJSON = 250
const testDataLengthTBinGeneric = 160
const testDataLengthTBinBest = 70

func polyline() *Polyline {
	var line Polyline
	err := json.Unmarshal([]byte(testDataJSON), &line)
	if err != nil {
		return nil
	}
	return &line
}

func (line Polyline) String() string {
	data, err := json.Marshal(line)
	if err == nil {
		return string(data)
	}
	return "<Polyline>"
}

func rect(x, y, w, h int32) *Rect {
	return &Rect{&Point{x, y}, &Point{x + w, y + h}}
}

func Pretty(o interface{}) string {
	data, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		return fmt.Sprint(o)
	}
	return string(data)
}

func Equal(o1 interface{}, o2 interface{}) bool {
	//doesn't work when the hash order is different:
	//return reflect.DeepEqual(o1, o2)
	return annotated(o1) == annotated(o2)
}

func annotated(o interface{}) string {
	return rannotated(reflect.TypeOf(o), reflect.ValueOf(o))
}

func rannotated(t reflect.Type, v reflect.Value) string {
	typeName := t.String()
	switch typeName {
	case "rdl.UUID":
		return fmt.Sprintf("UUID>~%v", v.Interface())
	case "rdl.Timestamp":
		return fmt.Sprintf("Timestamp~%v", v.Interface())
	case "rdl.Struct":
		return fmt.Sprintf("Struct~%v", v.Interface())
	case "rdl.Symbol":
		return fmt.Sprintf("Symbol~%v", v.Interface())
	}
	switch v.Kind() {
	case reflect.Struct:
		size := v.NumField()
		s := fmt.Sprintf("%s~{", v.Type().Name())
		for i := 0; i < size; i++ {
			f := v.Field(i)
			ft := t.Field(i)
			if i > 0 {
				s = s + ","
			}
			s = s + ft.Name + ":" + rannotated(ft.Type, f) + "\n"
		}
		return s + "}"
	case reflect.Slice:
		size := v.Len()
		s := fmt.Sprintf("array<%s>#%d~[", t.Elem().Name(), size)
		for i := 0; i < size; i++ {
			if i > 0 {
				s = s + ","
			}
			s = s + rannotated(t.Elem(), v.Index(i))
		}
		return s + "]"
	case reflect.Map:
		vkeys := v.MapKeys()
		var keys []string
		km := make(map[string]reflect.Value)
		for _, vk := range vkeys {
			k := vk.String()
			km[k] = vk
			keys = append(keys, k)
		}
		sort.Strings(keys)
		size := v.Len()
		s := fmt.Sprintf("<%v>#%d>{", t, size)
		for i, k := range keys {
			if i > 0 {
				s = s + ","
			}
			kv := km[k]
			vv := v.MapIndex(kv)
			s = s + fmt.Sprintf("%s:%v", k, rannotated(vv.Type(), vv))
		}
		return s + "}"
	case reflect.Ptr:
		if v.IsNil() {
			return "nil"
		} else {
			return "*" + annotated(v.Elem().Interface())
		}
	case reflect.String:
		return fmt.Sprintf("string~%q", v.String())
	case reflect.Bool:
		return fmt.Sprintf("bool~%v", v.Interface())
	case reflect.Int8:
		return fmt.Sprintf("int8~%d", v.Int())
	case reflect.Int16:
		return fmt.Sprintf("int16~%d", v.Int())
	case reflect.Int:
		return fmt.Sprintf("int~%d", v.Int())
	case reflect.Int32:
		return fmt.Sprintf("int32~%d", v.Int())
	case reflect.Int64:
		return fmt.Sprintf("int64~%d", v.Int())
	case reflect.Float32:
		return fmt.Sprintf("float32~%g", v.Float())
	case reflect.Float64:
		return fmt.Sprintf("float64~%g", v.Float())
	case reflect.Uint8:
		return fmt.Sprintf("<uint8>")
	case reflect.Interface:
		return fmt.Sprintf("<any>")
	default:
		panic(fmt.Sprintf("Handle this kind: %v", v.Kind()))
	}
}
