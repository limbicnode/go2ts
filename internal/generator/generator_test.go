package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/limbicnode/go2ts/internal/generator"
	"github.com/limbicnode/go2ts/internal/parser"
)

func TestGenerateTypeScriptFromModel(t *testing.T) {
	dir := filepath.Join("..", "..", "test", "testdata", "model")

	data, err := parser.ParseGoFiles(dir)
	if err != nil {
		t.Fatalf("ParseGoFiles failed: %v", err)
	}

	if len(data.Structs) == 0 {
		t.Fatalf("No structs parsed from %s", dir)
	}

	outPath := filepath.Join("..", "..", "types.ts")
	// defer os.Remove(outPath)

	err = generator.GenerateTypeScript(data, outPath)
	if err != nil {
		t.Fatalf("GenerateTypeScript failed: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created: %s", outPath)
	}

	t.Logf("✅ Generated TS file: %s", outPath)
}

func TestGenerateTypeScriptAdditionalCases(t *testing.T) {
	data := parser.GoFileData{
		Aliases: []parser.TypeAlias{
			{Name: "MyAny", Underlying: "interface{}", TypeParams: nil},
		},
		Structs: []parser.GoStruct{
			{
				Name: "MyStruct",
				Fields: []parser.StructField{
					{Name: "Field1", Type: "string"},
					{Name: "Field2", Type: "UnknownType"}, // GoTypeToTSType returns ""
				},
			},
		},
	}
	outPath := filepath.Join(os.TempDir(), "types2.ts")
	defer os.Remove(outPath)

	if err := generator.GenerateTypeScript(data, outPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// induce error with a path that has no write permission
	err := generator.GenerateTypeScript(data, "/root/no_permission.ts")
	if err == nil {
		t.Error("expected error due to write permission, got nil")
	}
}
func TestGenerateTypeScript_EmptyData(t *testing.T) {
	// when both Aliases and Structs are empty
	data := parser.GoFileData{}
	outPath := filepath.Join(os.TempDir(), "empty.ts")
	defer os.Remove(outPath)

	if err := generator.GenerateTypeScript(data, outPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractJSONTag(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{`json:"name"`, "name"},
		{`json:"name,omitempty"`, "name"},
		{`json:"-"`, ""},
		{`json:"name,omitempty" xml:"xmlName"`, "name"},
		{`xml:"xmlName" json:"name"`, "name"},
		{``, ""},
		{`xml:"xmlName"`, ""},
		{`json:""`, ""},
		{`json:"name,omitempty" bson:"bsonName"`, "name"},
		{`json:"name,omitempty" json:"other"`, "name"},
		{`json:"-" xml:"-"`, ""},
		{`   `, ""},
		{`json:"",xml:"x"`, ""},       // cover "name == \"\""
		{`json:"foo,bar,baz"`, "foo"}, // cover comma case
		{`json:"noendingquote`, ""},   // cover end == -1
	}

	for _, tt := range tests {
		got := generator.ExtractJSONTag(tt.tag)
		if got != tt.expected {
			t.Errorf("ExtractJSONTag(%q) = %q; want %q", tt.tag, got, tt.expected)
		}
	}
}

func TestGenerateTypeScript_MissingBranches(t *testing.T) {
	data := parser.GoFileData{
		Aliases: []parser.TypeAlias{
			// duplicate alias → triggers seenAliases branch
			{Name: "Alias1", Underlying: "string"},
			{Name: "Alias1", Underlying: "string"},
			{Name: "GenericAlias", TypeParams: []string{"T", "U"}, Underlying: "T"}, // alias with typeParams
			{Name: "EmptyTypeAlias", Underlying: ""},                                // GoTypeToTSType → "" → "any"
		},
		Structs: []parser.GoStruct{
			{
				Name:       "GenericStruct",
				TypeParams: []string{"X"},
				Fields: []parser.StructField{
					{Name: "NoTagField", Type: ""},                               // empty type → any
					{Name: "WithTag", Type: "string", Tags: `json:"customName"`}, // applies JSON tag
				},
			},
		},
	}

	outPath := filepath.Join(os.TempDir(), "branches.ts")
	defer os.Remove(outPath)

	if err := generator.GenerateTypeScript(data, outPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
