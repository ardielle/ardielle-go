package rdl

import (
	"encoding/json"
	"fmt"
	"testing"
)

func getExampleFromAnnotation(anno map[ExtendedAnnotation]string) (interface{}, bool) {
	expectFail := false
	if anno != nil {
		if b, ok := anno["x_expectfail"]; ok {
			expectFail = "true" == b
		}
		if s, ok := anno["x_example"]; ok {
			var value interface{}
			err := json.Unmarshal([]byte(s), &value)
			if err != nil {
				fmt.Printf("[warning: non-JSON value for x_example in schema %q, string type assumed: %q]\n", schema.Name, s)
			}
			return value, expectFail
		}
	}
	return nil, false
}

func getSchemaExample(schema *Schema) (interface{}, bool) {
	return getExampleFromAnnotation(schema.Annotations)
}

func getTypeExample(t *Type, typename TypeName) (interface{}, bool) {
	if t == nil {
		return nil, false
	}
	var anno map[ExtendedAnnotation]string
	switch t.Variant {
	case TypeVariantAliasTypeDef:
		if t.AliasTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.AliasTypeDef.Annotations
	case TypeVariantStringTypeDef:
		if t.StringTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.StringTypeDef.Annotations
	case TypeVariantNumberTypeDef:
		if t.NumberTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.NumberTypeDef.Annotations
	case TypeVariantArrayTypeDef:
		if t.ArrayTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.ArrayTypeDef.Annotations
	case TypeVariantMapTypeDef:
		if t.MapTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.MapTypeDef.Annotations
	case TypeVariantStructTypeDef:
		if t.StructTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.StructTypeDef.Annotations
	case TypeVariantBytesTypeDef:
		if t.BytesTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.BytesTypeDef.Annotations
	case TypeVariantEnumTypeDef:
		if t.EnumTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.EnumTypeDef.Annotations
	case TypeVariantUnionTypeDef:
		if t.UnionTypeDef.Annotations == nil {
			return nil, false
		}
		anno = t.UnionTypeDef.Annotations
	case TypeVariantBaseType:
		return nil, false
	}
	return getExampleFromAnnotation(anno)
	/*
		expectFail := false
		if b, ok := anno["x_expectfail"]; ok {
			expectFail = "true" == b
		}
		if s, ok := anno["x_example"]; ok {
			var value interface{}
			err := json.Unmarshal([]byte(s), &value)
			if err != nil {
				fmt.Printf("[warning: non-JSON value for x_example in type %q, string type assumed: %q]\n", typename, s)
				return s, expectFail
			}
			return value, expectFail
		}
		return nil, false
	*/
}

func getFieldExample(field *StructFieldDef, typename TypeName, fieldName string) (interface{}, bool) {
	if field.Annotations != nil {
		fmt.Println("hey, here is one:", field)
	}
	return nil, false
}

func TestExample(test *testing.T) {
	schema := loadTestSchema(test, "exampletest.rdl")
	if schema == nil {
		return
	}
	example, _ := getSchemaExample(schema)
	if example != nil {
		fmt.Printf("[warning: x_example at the schema level ignored]\n")
	}
	reg := NewTypeRegistry(schema)
	for _, t := range schema.Types {
		typename, _, _ := TypeInfo(t)
		example, expectFail := getTypeExample(t, typename)
		if example != nil {
			validation := Validate(schema, string(typename), example)
			if validation.Valid == expectFail {
				if expectFail {
					test.Errorf("Example for type %q should have failed validation but didn't\n", typename)
				} else {
					test.Errorf("Example for type %q failed to validate: %v\n", typename, validation)
				}
			}
			base := reg.BaseType(t)
			switch base {
			case BaseTypeStruct:
				fields := flattenedFields(reg, t)
				for _, field := range fields {
					if field.Annotations != nil {
						example, expectFail := getExampleFromAnnotation(field.Annotations)
						if example != nil {
							validation := Validate(schema, string(field.Type), example)
							if validation.Valid == expectFail {
								if expectFail {
									test.Errorf("Example for type %s.%s should have failed validation but didn't\n", typename, field.Name)
								} else {
									test.Errorf("Example for type %s.%s failed to validate: %v\n", typename, field.Name, validation)
								}
							}
						}
					}
				}
			}

		}
		//fmt.Printf("example for type %s has been validated as expected: %v\n", typename, validation)
	}

}
