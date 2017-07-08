//
// This file generated by rdl 1.4.13
//

package rdl

var schema *Schema

func init() {
	sb := NewSchemaBuilder("rdl")
	sb.Version(3)
	sb.Comment("This defines the schema for a schema, the output of the RDL parser. This can be used to represent schemas in JSON, Protobuf, Avro, etc, from a single definition.")

	tIdentifier := NewStringTypeBuilder("Identifier")
	tIdentifier.Comment("All names need to be of this restricted string type")
	tIdentifier.Pattern("[a-zA-Z_]+[a-zA-Z_0-9]*")
	sb.AddType(tIdentifier.Build())

	tNamespacedIdentifier := NewStringTypeBuilder("NamespacedIdentifier")
	tNamespacedIdentifier.Comment("A Namespace is a dotted compound name, using reverse domain name order (i.e. \"com.yahoo.auth\")")
	tNamespacedIdentifier.Pattern("([a-zA-Z_]+[a-zA-Z_0-9]*)(\\.[a-zA-Z_]+[a-zA-Z_0-9])*")
	sb.AddType(tNamespacedIdentifier.Build())

	tTypeName := NewAliasTypeBuilder("Identifier", "TypeName")
	tTypeName.Comment("The identifier for an already-defined type")
	sb.AddType(tTypeName.Build())

	tTypeRef := NewAliasTypeBuilder("NamespacedIdentifier", "TypeRef")
	tTypeRef.Comment("A type reference can be a simple name, or also a namespaced name.")
	sb.AddType(tTypeRef.Build())

	tBaseType := NewEnumTypeBuilder("Enum", "BaseType")
	tBaseType.Element("Bool", "")
	tBaseType.Element("Int8", "")
	tBaseType.Element("Int16", "")
	tBaseType.Element("Int32", "")
	tBaseType.Element("Int64", "")
	tBaseType.Element("Float32", "")
	tBaseType.Element("Float64", "")
	tBaseType.Element("Bytes", "")
	tBaseType.Element("String", "")
	tBaseType.Element("Timestamp", "")
	tBaseType.Element("Symbol", "")
	tBaseType.Element("UUID", "")
	tBaseType.Element("Array", "")
	tBaseType.Element("Map", "")
	tBaseType.Element("Struct", "")
	tBaseType.Element("Enum", "")
	tBaseType.Element("Union", "")
	tBaseType.Element("Any", "")
	sb.AddType(tBaseType.Build())

	tExtendedAnnotation := NewStringTypeBuilder("ExtendedAnnotation")
	tExtendedAnnotation.Comment("ExtendedAnnotation - parsed and preserved, but has no defined meaning in RDL. Such annotations must begin with \"x_\", and may have an associated string literal value (the value will be \"\" if the annotation is just a flag).")
	tExtendedAnnotation.Pattern("x_[a-zA-Z_0-9]*")
	sb.AddType(tExtendedAnnotation.Build())

	tTypeDef := NewStructTypeBuilder("Struct", "TypeDef")
	tTypeDef.Comment("TypeDef is the basic type definition.")
	tTypeDef.Field("type", "TypeRef", false, nil, "The type this type is derived from. For base types, it is the same as the name")
	tTypeDef.Field("name", "TypeName", false, nil, "The name of the type")
	tTypeDef.Field("comment", "String", true, nil, "The comment for the type")
	tTypeDef.MapField("annotations", "ExtendedAnnotation", "String", true, "additional annotations starting with \"x_\"")
	sb.AddType(tTypeDef.Build())

	tAliasTypeDef := NewStructTypeBuilder("TypeDef", "AliasTypeDef")
	tAliasTypeDef.Comment("AliasTypeDef is used for type definitions that add no additional attributes, and thus just create an alias")
	sb.AddType(tAliasTypeDef.Build())

	tBytesTypeDef := NewStructTypeBuilder("TypeDef", "BytesTypeDef")
	tBytesTypeDef.Comment("Bytes allow the restriction by fixed size, or min/max size.")
	tBytesTypeDef.Field("size", "Int32", true, nil, "Fixed size")
	tBytesTypeDef.Field("minSize", "Int32", true, nil, "Min size")
	tBytesTypeDef.Field("maxSize", "Int32", true, nil, "Max size")
	sb.AddType(tBytesTypeDef.Build())

	tStringTypeDef := NewStructTypeBuilder("TypeDef", "StringTypeDef")
	tStringTypeDef.Comment("Strings allow the restriction by regular expression pattern or by an explicit set of values. An optional maximum size may be asserted")
	tStringTypeDef.Field("pattern", "String", true, nil, "A regular expression that must be matched. Mutually exclusive with values")
	tStringTypeDef.ArrayField("values", "String", true, "A set of allowable values")
	tStringTypeDef.Field("minSize", "Int32", true, nil, "Min size")
	tStringTypeDef.Field("maxSize", "Int32", true, nil, "Max size")
	sb.AddType(tStringTypeDef.Build())

	tNumber := NewUnionTypeBuilder("Union", "Number")
	tNumber.Comment("A numeric is any of the primitive numeric types")
	tNumber.Variant("Int8")
	tNumber.Variant("Int16")
	tNumber.Variant("Int32")
	tNumber.Variant("Int64")
	tNumber.Variant("Float32")
	tNumber.Variant("Float64")
	sb.AddType(tNumber.Build())

	tNumberTypeDef := NewStructTypeBuilder("TypeDef", "NumberTypeDef")
	tNumberTypeDef.Comment("A number type definition allows the restriction of numeric values.")
	tNumberTypeDef.Field("min", "Number", true, nil, "Min value")
	tNumberTypeDef.Field("max", "Number", true, nil, "Max value")
	sb.AddType(tNumberTypeDef.Build())

	tArrayTypeDef := NewStructTypeBuilder("TypeDef", "ArrayTypeDef")
	tArrayTypeDef.Comment("Array types can be restricted by item type and size")
	tArrayTypeDef.Field("items", "TypeRef", false, "Any", "The type of the items, default to any type")
	tArrayTypeDef.Field("size", "Int32", true, nil, "If present, indicate the fixed size.")
	tArrayTypeDef.Field("minSize", "Int32", true, nil, "If present, indicate the min size")
	tArrayTypeDef.Field("maxSize", "Int32", true, nil, "If present, indicate the max size")
	sb.AddType(tArrayTypeDef.Build())

	tMapTypeDef := NewStructTypeBuilder("TypeDef", "MapTypeDef")
	tMapTypeDef.Comment("Map types can be restricted by key type, item type and size")
	tMapTypeDef.Field("keys", "TypeRef", false, "String", "The type of the keys, default to String.")
	tMapTypeDef.Field("items", "TypeRef", false, "Any", "The type of the items, default to Any type")
	tMapTypeDef.Field("size", "Int32", true, nil, "If present, indicates the fixed size.")
	tMapTypeDef.Field("minSize", "Int32", true, nil, "If present, indicate the min size")
	tMapTypeDef.Field("maxSize", "Int32", true, nil, "If present, indicate the max size")
	sb.AddType(tMapTypeDef.Build())

	tStructFieldDef := NewStructTypeBuilder("Struct", "StructFieldDef")
	tStructFieldDef.Comment("Each field in a struct_field_spec is defined by this type")
	tStructFieldDef.Field("name", "Identifier", false, nil, "The name of the field")
	tStructFieldDef.Field("type", "TypeRef", false, nil, "The type of the field")
	tStructFieldDef.Field("optional", "Bool", false, false, "The field may be omitted even if specified")
	tStructFieldDef.Field("default", "Any", true, nil, "If field is absent, what default value should be assumed.")
	tStructFieldDef.Field("comment", "String", true, nil, "The comment for the field")
	tStructFieldDef.Field("items", "TypeRef", true, nil, "For map or array fields, the type of the items")
	tStructFieldDef.Field("keys", "TypeRef", true, nil, "For map type fields, the type of the keys")
	tStructFieldDef.MapField("annotations", "ExtendedAnnotation", "String", true, "additional annotations starting with \"x_\"")
	sb.AddType(tStructFieldDef.Build())

	tStructTypeDef := NewStructTypeBuilder("TypeDef", "StructTypeDef")
	tStructTypeDef.Comment("A struct can restrict specific named fields to specific types. By default, any field not specified is allowed, and can be of any type. Specifying closed means only those fields explicitly")
	tStructTypeDef.ArrayField("fields", "StructFieldDef", false, "The fields in this struct. By default, open Structs can have any fields in addition to these")
	tStructTypeDef.Field("closed", "Bool", false, false, "indicates that only the specified fields are acceptable. Default is open (any fields)")
	sb.AddType(tStructTypeDef.Build())

	tEnumElementDef := NewStructTypeBuilder("Struct", "EnumElementDef")
	tEnumElementDef.Comment("EnumElementDef defines one of the elements of an Enum")
	tEnumElementDef.Field("symbol", "Identifier", false, nil, "The identifier representing the value")
	tEnumElementDef.Field("comment", "String", true, nil, "the comment for the element")
	tEnumElementDef.MapField("annotations", "ExtendedAnnotation", "String", true, "additional annotations starting with \"x_\"")
	sb.AddType(tEnumElementDef.Build())

	tEnumTypeDef := NewStructTypeBuilder("TypeDef", "EnumTypeDef")
	tEnumTypeDef.Comment("Define an enumerated type. Each value of the type is represented by a symbolic identifier.")
	tEnumTypeDef.ArrayField("elements", "EnumElementDef", false, "The enumeration of the possible elements")
	sb.AddType(tEnumTypeDef.Build())

	tUnionTypeDef := NewStructTypeBuilder("TypeDef", "UnionTypeDef")
	tUnionTypeDef.Comment("Define a type as one of any other specified type.")
	tUnionTypeDef.ArrayField("variants", "TypeRef", false, "The type names of constituent types. Union types get expanded, this is a flat list")
	sb.AddType(tUnionTypeDef.Build())

	tType := NewUnionTypeBuilder("Union", "Type")
	tType.Comment("A Type can be specified by any of the above specialized Types, determined by the value of the the 'type' field")
	tType.Variant("BaseType")
	tType.Variant("StructTypeDef")
	tType.Variant("MapTypeDef")
	tType.Variant("ArrayTypeDef")
	tType.Variant("EnumTypeDef")
	tType.Variant("UnionTypeDef")
	tType.Variant("StringTypeDef")
	tType.Variant("BytesTypeDef")
	tType.Variant("NumberTypeDef")
	tType.Variant("AliasTypeDef")
	sb.AddType(tType.Build())

	tResourceInput := NewStructTypeBuilder("Struct", "ResourceInput")
	tResourceInput.Comment("ResourceOutput defines input characteristics of a Resource")
	tResourceInput.Field("name", "Identifier", false, nil, "the formal name of the input")
	tResourceInput.Field("type", "TypeRef", false, nil, "The type of the input")
	tResourceInput.Field("comment", "String", true, nil, "The optional comment")
	tResourceInput.Field("pathParam", "Bool", false, false, "true of this input is a path parameter")
	tResourceInput.Field("queryParam", "String", true, nil, "if present, the name of the query param name")
	tResourceInput.Field("header", "String", true, nil, "If present, the name of the header the input is associated with")
	tResourceInput.Field("pattern", "String", true, nil, "If present, the pattern associated with the pathParam (i.e. wildcard path matches)")
	tResourceInput.Field("default", "Any", true, nil, "If present, the default value for optional params")
	tResourceInput.Field("optional", "Bool", false, false, "If present, indicates that the input is optional")
	tResourceInput.Field("flag", "Bool", false, false, "If present, indicates the queryparam is of flag style (no value)")
	tResourceInput.Field("context", "String", true, nil, "If present, indicates the parameter comes form the implementation context")
	tResourceInput.MapField("annotations", "ExtendedAnnotation", "String", true, "additional annotations starting with \"x_\"")
	sb.AddType(tResourceInput.Build())

	tResourceOutput := NewStructTypeBuilder("Struct", "ResourceOutput")
	tResourceOutput.Comment("ResourceOutput defines output characteristics of a Resource")
	tResourceOutput.Field("name", "Identifier", false, nil, "the formal name of the output")
	tResourceOutput.Field("type", "TypeRef", false, nil, "The type of the output")
	tResourceOutput.Field("header", "String", false, nil, "the name of the header associated with this output")
	tResourceOutput.Field("comment", "String", true, nil, "The optional comment for the output")
	tResourceOutput.Field("optional", "Bool", false, false, "If present, indicates that the output is optional (the server decides)")
	tResourceOutput.MapField("annotations", "ExtendedAnnotation", "String", true, "additional annotations starting with \"x_\"")
	sb.AddType(tResourceOutput.Build())

	tResourceAuth := NewStructTypeBuilder("Struct", "ResourceAuth")
	tResourceAuth.Comment("ResourceAuth defines authentication and authorization attributes of a resource. Presence of action, resource, or domain implies authentication; the authentication flag alone is required only when no authorization is done.")
	tResourceAuth.Field("authenticate", "Bool", false, false, "if present and true, then the requester must be authenticated")
	tResourceAuth.Field("action", "String", true, nil, "the action to authorize access to. This forces authentication")
	tResourceAuth.Field("resource", "String", true, nil, "the resource identity to authorize access to")
	tResourceAuth.Field("domain", "String", true, nil, "if present, the alternate domain to check access to. This is rare.")
	sb.AddType(tResourceAuth.Build())

	tExceptionDef := NewStructTypeBuilder("Struct", "ExceptionDef")
	tExceptionDef.Comment("ExceptionDef describes the exception a symbolic response code maps to.")
	tExceptionDef.Field("type", "String", false, nil, "The type of the exception")
	tExceptionDef.Field("comment", "String", true, nil, "the optional comment for the exception")
	sb.AddType(tExceptionDef.Build())

	tResource := NewStructTypeBuilder("Struct", "Resource")
	tResource.Comment("A Resource of a REST service")
	tResource.Field("type", "TypeRef", false, nil, "The type of the resource")
	tResource.Field("method", "String", false, nil, "The method for the action (typically GET, POST, etc for HTTP access)")
	tResource.Field("path", "String", false, nil, "The resource path template")
	tResource.Field("comment", "String", true, nil, "The optional comment")
	tResource.ArrayField("inputs", "ResourceInput", true, "An Array named inputs")
	tResource.ArrayField("outputs", "ResourceOutput", true, "An Array of named outputs")
	tResource.Field("auth", "ResourceAuth", true, nil, "The optional authentication or authorization directive")
	tResource.Field("expected", "String", false, "OK", "The expected symbolic response code")
	tResource.ArrayField("alternatives", "String", true, "The set of alternative but non-error response codes")
	tResource.MapField("exceptions", "String", "ExceptionDef", true, "A map of symbolic response code to Exception definitions")
	tResource.Field("async", "Bool", true, nil, "A hint to server implementations that this resource would be better implemented with async I/O")
	tResource.MapField("annotations", "ExtendedAnnotation", "String", true, "additional annotations starting with \"x_\"")
	tResource.ArrayField("consumes", "String", true, "Optional hint for resource acceptable input types")
	tResource.ArrayField("produces", "String", true, "Optional hint for resource output content types")
	tResource.Field("name", "Identifier", true, nil, "The optional name of the resource")
	sb.AddType(tResource.Build())

	tSchema := NewStructTypeBuilder("Struct", "Schema")
	tSchema.Comment("A Schema is a container for types and resources. It is self-contained (no external references). and is the output of the RDL parser.")
	tSchema.Field("namespace", "NamespacedIdentifier", true, nil, "The namespace for the schema")
	tSchema.Field("name", "Identifier", true, nil, "The name of the schema")
	tSchema.Field("version", "Int32", true, nil, "The version of the schema")
	tSchema.Field("comment", "String", true, nil, "The comment for the entire schema")
	tSchema.ArrayField("types", "Type", true, "The types this schema defines.")
	tSchema.ArrayField("resources", "Resource", true, "The resources for a service this schema defines")
	tSchema.Field("base", "String", true, nil, "the base path for resources in the schema.")
	sb.AddType(tSchema.Build())

	schema = sb.Build()
}

func RdlSchema() *Schema {
	return schema
}
