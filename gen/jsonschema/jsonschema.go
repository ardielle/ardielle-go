// Copyright 2018 Lee Boynton (lee@boynton.com, github.com/boynton)
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.
//
package jsonschema

//
// export RDL types to JSON Schema
//
import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ardielle/ardielle-go/rdl"
)

type JSONSchema map[string]interface{}

func (js JSONSchema) String() string {
	b, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return fmt.Sprintf("*** %v", err)
	}
	return string(b)
}

//always generates schemas of the form {"definitions": { ... }}, unless no types are defined at all, then just {}
func Generate(schema *rdl.Schema) (JSONSchema, error) {
	reg := rdl.NewTypeRegistry(schema)
	js := make(JSONSchema) //map[string]interface{})
	js["$schema"] = "http://json-schema.org/draft-04/schema#"
	if len(schema.Types) > 0 {
		defs := make(map[string]map[string]interface{})
		js["definitions"] = defs
		for _, t := range schema.Types {
			ref := jsTypeDef(reg, t)
			if ref != nil {
				tName, _, _ := rdl.TypeInfo(t)
				defs[string(tName)] = ref
			}
		}
	}
	return js, nil
}

func jsTypeRef(reg rdl.TypeRegistry, itemTypeName rdl.TypeRef) (string, string, interface{}) {
	itype := string(itemTypeName)
	switch reg.FindBaseType(itemTypeName) {
	case rdl.BaseTypeInt8:
		return "string", "byte", nil //?
	case rdl.BaseTypeInt16, rdl.BaseTypeInt32, rdl.BaseTypeInt64:
		return "integer", strings.ToLower(itype), nil
	case rdl.BaseTypeFloat32:
		return "number", "float", nil
	case rdl.BaseTypeFloat64:
		return "number", "double", nil
	case rdl.BaseTypeString:
		return "string", "", nil
	case rdl.BaseTypeTimestamp:
		return "string", "date-time", nil
	case rdl.BaseTypeUUID, rdl.BaseTypeSymbol:
		return "string", strings.ToLower(itype), nil
	default:
		s := make(map[string]interface{})
		s["$ref"] = "#/definitions/" + itype
		return "", "", s
	}
}

func jsTypeDef(reg rdl.TypeRegistry, t *rdl.Type) map[string]interface{} {
	st := make(map[string]interface{})
	bt := reg.BaseType(t)
	switch t.Variant {
	case rdl.TypeVariantStructTypeDef:
		typedef := t.StructTypeDef
		if typedef.Comment != "" {
			st["description"] = typedef.Comment
		}
		props := make(map[string]interface{})
		var required []string
		if len(typedef.Fields) > 0 {
			for _, f := range typedef.Fields {
				if !f.Optional {
					required = append(required, string(f.Name))
				}
				ft := reg.FindType(f.Type)
				fbt := reg.BaseType(ft)
				prop := make(map[string]interface{})
				if f.Comment != "" {
					prop["description"] = f.Comment
				}
				switch fbt {
				case rdl.BaseTypeArray:
					if ft.Variant != rdl.TypeVariantBaseType {
						name, _, _ := rdl.TypeInfo(ft)
						prop["$ref"] = "#/definitions/" + name
					} else {
						prop["type"] = "array"
						if ft.Variant == rdl.TypeVariantArrayTypeDef && f.Items == "" {
							f.Items = ft.ArrayTypeDef.Items
						}
						if f.Items != "" {
							fitems := string(f.Items)
							items := make(map[string]interface{})
							switch fitems {
							case "String":
								items["type"] = strings.ToLower(fitems)
							case "Int32", "Int64", "Int16":
								items["type"] = "integer"
								//not supported by all validators: items["format"] = strings.ToLower(fitems)
							default:
								items["$ref"] = "#/definitions/" + fitems
							}
							prop["items"] = items
						}
					}
				case rdl.BaseTypeString:
					if ft.Variant != rdl.TypeVariantBaseType {
						name, _, _ := rdl.TypeInfo(ft)
						prop["$ref"] = "#/definitions/" + name
					} else {
						prop["type"] = "string"
					}
				case rdl.BaseTypeInt32, rdl.BaseTypeInt64, rdl.BaseTypeInt16:
					prop["type"] = "integer"
					//not always supported prop["format"] = strings.ToLower(fbt.String())
				case rdl.BaseTypeStruct:
					prop["$ref"] = "#/definitions/" + string(f.Type)
				case rdl.BaseTypeMap:
					prop["type"] = "object"
					if f.Items != "" {
						fitems := string(f.Items)
						items := make(map[string]interface{})
						switch f.Items {
						case "String":
							items["type"] = strings.ToLower(fitems)
						case "Int32", "Int64", "Int16":
							items["type"] = "integer"
							items["format"] = strings.ToLower(fitems)
						default:
							items["$ref"] = "#/definitions/" + fitems
						}
						prop["additionalProperties"] = items
					}
				case rdl.BaseTypeEnum:
					prop["$ref"] = "#/definitions/" + string(f.Type)
				default:
					panic("not yet implemented: " + f.Type)
				}
				props[string(f.Name)] = prop
			}
		}
		st["properties"] = props
		if len(required) > 0 {
			st["required"] = required
		}
	case rdl.TypeVariantMapTypeDef:
		typedef := t.MapTypeDef
		st["type"] = "object"
		if typedef.Items != "Any" {
			items := make(map[string]interface{})
			switch reg.FindBaseType(typedef.Items) {
			case rdl.BaseTypeString:
				items["type"] = strings.ToLower(string(typedef.Items))
			case rdl.BaseTypeInt32, rdl.BaseTypeInt64, rdl.BaseTypeInt16:
				items["type"] = "integer"
				items["format"] = strings.ToLower(string(typedef.Items))
			default:
				items["$ref"] = "#/definitions/" + string(typedef.Items)
			}
			st["additionalProperties"] = items
		}
	case rdl.TypeVariantArrayTypeDef:
		typedef := t.ArrayTypeDef
		st["type"] = "array"
		if typedef.Items != "Any" {
			items := make(map[string]interface{})
			switch reg.FindBaseType(typedef.Items) {
			case rdl.BaseTypeString:
				items["type"] = strings.ToLower(string(typedef.Items))
			case rdl.BaseTypeInt32, rdl.BaseTypeInt64, rdl.BaseTypeInt16:
				items["type"] = "integer"
				items["format"] = strings.ToLower(string(typedef.Items))
			default:
				items["$ref"] = "#/definitions/" + string(typedef.Items)
			}
			st["items"] = items
			if typedef.Size != nil {
				st["minItems"] = *typedef.Size
				st["maxItems"] = *typedef.Size
			} else {
				if typedef.MinSize != nil {
					st["minItems"] = *typedef.MinSize
				}
				if typedef.MaxSize != nil {
					st["minItems"] = *typedef.MaxSize
				}
			}
		}
	case rdl.TypeVariantEnumTypeDef:
		typedef := t.EnumTypeDef
		var tmp []string
		for _, el := range typedef.Elements {
			tmp = append(tmp, string(el.Symbol))
		}
		st["enum"] = tmp
	case rdl.TypeVariantUnionTypeDef:
		typedef := t.UnionTypeDef
		fmt.Println("[" + typedef.Name + ": Unions not supported]")
	default:
		switch bt {
		case rdl.BaseTypeString:
			if t.StringTypeDef != nil {
				typedef := t.StringTypeDef
				st["type"] = "string"
				if typedef.MaxSize != nil {
					st["maxLength"] = *typedef.MaxSize
				}
				if typedef.MinSize != nil {
					st["minLength"] = *typedef.MinSize
				}
				if typedef.Pattern != "" {
					st["pattern"] = typedef.Pattern
				}
			} else {
				return nil
			}
		case rdl.BaseTypeInt16, rdl.BaseTypeInt32, rdl.BaseTypeInt64, rdl.BaseTypeFloat32, rdl.BaseTypeFloat64:
			return nil
		case rdl.BaseTypeStruct:
			st["type"] = "object"
		default:
			panic(fmt.Sprintf("whoops: %v", t))
		}
	}
	return st
}

func TypeDefs(js JSONSchema) map[string]map[string]interface{} {
	if v, ok := js["definitions"]; ok {
		if defs, ok := v.(map[string]map[string]interface{}); ok {
			return defs
		} else {
			fmt.Println("what?!")
		}
	}
	return nil
}
