// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

package rdl

import (
	"fmt"
	"math/rand"
	"runtime/debug"
	"strings"
	"testing"
	"time"
	"unicode"
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
	schema, err = sb.BuildParanoid()
	if err != nil {
		test.Fatal(err)
	}
	if schema == nil {
		test.Errorf("TestSchemaBuilder: Cannot build schema with certain base types")
	}
}

// This can be replaced by a direct call to test.Run(name, testFunc) in go-1.7
func runTest(test *testing.T, name string, testFunc func(*testing.T)) {
	testFunc(test)
}

func TestSBBaseTypeCaseSensitivity(test *testing.T) {

	type testType struct {
		supertype string
		kind      TypeVariantTag
	}

	rand.Seed(time.Now().UnixNano())

	var makeTypes = func(sb *SchemaBuilder, superType, ref string, builder func(s, r string) *Type) (names []string) {
		const (
			title int = iota
			uc
			lc
			mixed
			_end
		)
		for i := 0; i < _end; i++ {
			var name = fmt.Sprintf("%s_%d", ref, i)
			var super string = superType
			switch i {
			case title:
				super = strings.Title(super)
			case uc:
				super = strings.ToUpper(super)
			case lc:
				super = strings.ToLower(super)
			case mixed:
				var b []rune
				for _, r := range super {
					if rand.Float32() >= 0.5 {
						b = append(b, unicode.ToUpper(r))
					} else {
						b = append(b, unicode.ToLower(r))
					}

				}
				super = string(b)
			default:
			}
			names = append(names, name)
			sb.AddType(builder(super, name))
		}
		return names
	}

	for idx, b := range namesBaseType {
		baseType := BaseType(idx)
		switch baseType {
		case 0, BaseTypeAny, BaseTypeBytes, BaseTypeSymbol, BaseTypeUUID, BaseTypeTimestamp:
			continue
		default:
		}
		baseTypeName := b
		runTest(test, baseTypeName, func(test *testing.T) {
			// test.Parallel() - use after migrating to go-1.7
			defer func() {
				if r := recover(); r != nil {
					test.Errorf("TestSBBaseTypeCaseSensitivity (%s) %s: %s", baseTypeName, r, debug.Stack())
				}
			}()

			var cases []string
			sb := NewSchemaBuilder(baseTypeName)

			switch baseTypeName {
			case "":
				test.Skip("No type specified")
			case "Bool":
				cases = makeTypes(sb, baseTypeName, "TestAlias", func(s, r string) *Type {
					return NewAliasTypeBuilder(s, r).Build()
				})
			case "Int8", "Int16", "Int32", "Int64", "Float32", "Float64":
				cases = makeTypes(sb, baseTypeName, "TestNumber", func(s, r string) *Type {
					return NewNumberTypeBuilder(s, r).Build()
				})
			case "String":
				cases = makeTypes(sb, baseTypeName, "TestString", func(s, r string) *Type {
					return NewStringTypeBuilder(r).Build()
				})
			case "Struct":
				cases = makeTypes(sb, baseTypeName, "TestStruct", func(s, r string) *Type {
					sb := NewStructTypeBuilder(s, r)
					sb.Field("field", "Any", false, nil, "")
					return sb.Build()
				})
			case "Array":
				cases = makeTypes(sb, baseTypeName, "TestArray", func(s, r string) *Type {
					ab := NewArrayTypeBuilder(s, r)
					ab.Items("Any")
					return ab.Build()
				})
			case "Map":
				cases = makeTypes(sb, baseTypeName, "TestMap", func(s, r string) *Type {
					mb := NewMapTypeBuilder(s, r)
					mb.Keys("Any").Items("Any")
					return mb.Build()
				})
			case "Union":
				cases = makeTypes(sb, baseTypeName, "TestUnion", func(s, r string) *Type {
					ub := NewUnionTypeBuilder(s, r)
					ub.Variant("Any").Variant("Int8")
					return ub.Build()
				})
			case "Enum":
				cases = makeTypes(sb, baseTypeName, "TestEnum", func(s, r string) *Type {
					eb := NewEnumTypeBuilder(s, r)
					eb.Element("Foo1", "").Element("Foo2", "")
					return eb.Build()
				})
			default:
				test.Skipf("Basetype %s not tested", baseTypeName)
			}
			schema, err := sb.BuildParanoid()
			if err != nil {
				test.Fatal(err)
			}
			if schema == nil {
				test.Errorf("TestSBBaseTypeCaseSensitivity: Cannot build schema")
			}
			types := NewTypeRegistry(schema)
			for _, c := range cases {
				if types.FindType(TypeRef(c)) == nil {
					test.Errorf("TestSBBaseTypeCaseSensitivity: Could not find typeName %s", c)
				}
				if types.FindBaseType(TypeRef(c)) != baseType {
					test.Errorf("TestSBBaseTypeCaseSensitivity: typeName %s doesn't have correct BaseType %s", c, namesBaseType[baseType])
				}
			}
		})
	}
}
