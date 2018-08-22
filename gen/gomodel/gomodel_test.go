package gomodel

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ardielle/ardielle-go/rdl"
)

func pretty(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Sprintf("*** %v", err)
	}
	return string(b)
}

const defaultGoLibRdl = "github.com/ardielle/ardielle-go/rdl"

func generate(infile, outdir string) error {
	schema, err := rdl.ParseRDLFile("../../testdata/"+infile, false, false, true)
	if err != nil {
		return nil
	}
	err = Generate(schema, &GeneratorParams{
		Outdir: outdir,
		Banner: "gomodel_test",
		Namespace: "rdl",
		UntaggedUnions: nil,
		LibRdl: DefaultLibRdl,
		PrefixEnums: true,
		PreciseTypes: true,
	})
	if err != nil {
		return err
	}
	//now compile it, or at least create a shell script to invoke the compiler
	//should also test, for each model type, that we can read/write JSON
	//need a model test with all the types in it.
	return nil
}

func TestModelGen(test *testing.T) {
	outdir := "/tmp/gomodel_gen"
	err := generate("rdl.rdl", outdir)
	if err != nil {
		test.Errorf("TestModelGen: %v", err)
		return
	}
}

