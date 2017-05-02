package codegen

import (
	"github.com/ardielle/ardielle-go/rdl"
	TestA "github.com/ardielle/ardielle-go/rdl/_gen/A"
	"testing"
)

//go:generate go run generator.go

func BenchmarkA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rdl.Validate(TestA.CodegenSchema(), "StringStruct", TestA.Example)
	}
}

func BenchmarkB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rdl.Validate(TestA.CodegenSchema(), "StringStruct", TestA.Example)
	}
}

func TestCodeGenModel(test *testing.T) {
	v := rdl.Validate(TestA.CodegenSchema(), "StringStruct", TestA.Example)
	if !v.Valid {
		test.Errorf("Validation error: %v, validation", v)
	}
}
