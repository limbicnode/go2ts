package parser_test

import (
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/limbicnode/go2ts/internal/parser"
)

func TestParseGoFiles_EdgeCases(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "skip_test.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write skip_test.go: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bad.go"), []byte("package main; func {"), 0644); err != nil {
		t.Fatalf("failed to write bad.go: %v", err)
	}

	if _, err := parser.ParseGoFiles(dir); err == nil {
		t.Errorf("expected parse error for bad.go")
	}
}

func TestParseWithExampleModel(t *testing.T) {
	dir := filepath.Join("..", "..", "test", "testdata", "model")

	data, err := parser.ParseGoFiles(dir)
	if err != nil {
		t.Fatalf("ParseGoFiles failed: %v", err)
	}

	structMap := make(map[string]parser.StructInfo)
	for _, s := range data.Structs {
		fields := make([]parser.FieldInfo, len(s.Fields))
		for i, f := range s.Fields {
			fields[i] = parser.FieldInfo(f)
		}
		structMap[s.Name] = parser.StructInfo{
			Name:       s.Name,
			TypeParams: s.TypeParams,
			Fields:     fields,
		}
	}

	totalStructs := len(data.Structs)
	totalAliases := len(data.Aliases)

	if totalStructs == 0 && totalAliases == 0 {
		t.Fatal("no structs or type aliases found in directory")
	}

	aliasMap := map[string]string{}
	for _, alias := range data.Aliases {
		aliasMap[alias.Name] = alias.Underlying
	}

	structsPassed := 0
	anyCount := 0

NextStructLoop:
	for _, st := range data.Structs {
		if st.Name == "" {
			t.Logf("Struct with empty name: %+v", st)
			continue
		}

		for _, f := range st.Fields {
			if f.Name == "" || f.Type == "" {
				continue
			}

			typeParamMapping := map[string]string{}

			tsType := parser.GoTypeToTSType(f.Type, aliasMap, st.TypeParams, structMap, typeParamMapping, map[string]bool{})

			if tsType == "" {
				t.Logf("Struct '%s' has field with unsupported type: %+v (GoType: %q)", st.Name, f, f.Type)
				continue NextStructLoop
			}

			if tsType == "any" || strings.Contains(tsType, "any") {
				// t.Logf("any detected: %q (field %s.%s)", tsType, st.Name, f.Name)
				anyCount++
			}
		}
		structsPassed++
	}

	aliasesPassed := 0
	for _, alias := range data.Aliases {
		if alias.Name == "" || alias.Underlying == "" {
			t.Logf("Alias invalid: %+v", alias)
			continue
		}
		aliasesPassed++
	}

	structsFailed := totalStructs - structsPassed
	aliasesFailed := totalAliases - aliasesPassed

	t.Logf("Structs: total=%d, passed=%d, failed=%d, success=%.2f%%",
		totalStructs, structsPassed, structsFailed,
		float64(structsPassed)/float64(totalStructs)*100)

	t.Logf("Type Aliases: total=%d, passed=%d, failed=%d, success=%.2f%%",
		totalAliases, aliasesPassed, aliasesFailed,
		float64(aliasesPassed)/float64(totalAliases)*100)

	t.Logf("Fields converted to 'any': %d", anyCount)

	if structsPassed == 0 && aliasesPassed == 0 {
		t.Fatal("all structs and type aliases failed validation")
	}
}

func TestExprToString(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{"Ident", &ast.Ident{Name: "MyType"}, "MyType"},
		{"StarExpr", &ast.StarExpr{X: &ast.Ident{Name: "MyType"}}, "*MyType"},
		{"SelectorExpr", &ast.SelectorExpr{X: &ast.Ident{Name: "pkg"}, Sel: &ast.Ident{Name: "Type"}}, "pkg.Type"},
		{"ArrayType", &ast.ArrayType{Elt: &ast.Ident{Name: "int"}}, "[]int"},
		{"MapType", &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}}, "map[string]int"},
		{"IndexExpr", &ast.IndexExpr{X: &ast.Ident{Name: "MyType"}, Index: &ast.Ident{Name: "T"}}, "MyType[T]"},
		{"IndexListExpr", &ast.IndexListExpr{
			X:       &ast.Ident{Name: "MyType"},
			Indices: []ast.Expr{&ast.Ident{Name: "T"}, &ast.Ident{Name: "K"}},
		}, "MyType[T, K]"},
		{"InterfaceType", &ast.InterfaceType{}, "interface{}"},
		{"EmptyStructType", &ast.StructType{}, "struct{}"},
		{"StructWithFields", &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "Field1"}}, Type: &ast.Ident{Name: "int"}},
					{Names: []*ast.Ident{{Name: "Field2"}, {Name: "Field3"}}, Type: &ast.Ident{Name: "string"}},
				},
			},
		}, "struct{ Field1 int; Field2, Field3 string }"},
		{"StructEmbeddedField", &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{Names: nil, Type: &ast.Ident{Name: "MyEmbeddedType"}},
				},
			},
		}, "struct{ MyEmbeddedType }"},
		{"FuncType", &ast.FuncType{}, "func"},
		{"UnknownExpr", &ast.BadExpr{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.ExprToString(tt.expr); got != tt.want {
				t.Errorf("ExprToString(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestGoTypeToTSType(t *testing.T) {
	aliasMap := map[string]string{
		"MyString":   "string",
		"MyInt":      "int",
		"MyAlias":    "MyString",
		"Alias2":     "MyAlias",
		"Alias3":     "Alias2",
		"SelfRef":    "SelfRef",
		"Nested":     "map[string][]*MyAlias",
		"AliasInt":   "int",
		"AliasMap":   "map[string]string",
		"AliasLoop1": "AliasLoop2", // loop test
		"AliasLoop2": "AliasLoop1", // loop test
	}

	typeParams := []string{"T"}
	emptyStructMap := map[string]parser.StructInfo{
		// "CustomType": {},
	}
	typeParamMapping := map[string]string{}

	tests := []struct {
		goType string
		want   string
	}{
		{"", ""},
		{"malformed[", "any"},
		{"SelfRef", "any"},
		{"*int", "number | null"},
		{"[][]map[int]string", "({ [key: number]: string })[][]"},
		{"map[string][]*MyAlias", "{ [key: string]: string | null[] }"},
		{"Alias3", "string"},
		{"MyType[T]", "MyType<T>"},
		{"Result[K, V]", "Result<K, V>"},
		{"[]Custom[T]", "Custom<T>[]"},
		{"*Option[string]", "Option<string> | null"},
		{"UnknownAlias", "UnknownAlias"},
		{"*Custom", "Custom | null"},
		{"[]Custom", "Custom[]"},
		{"map[string", "any"},
		{"string", "string"},
		{"int", "number"},
		{"bool", "boolean"},
		{"[]byte", "Uint8Array"},
		{"pkg.Type", "any"},
		{"map[struct{ X, Y int }]string", "{ [key: string]: string }"},
		{"struct{ X int; Y int }", "{ X: number; Y: number }"},
		{"float64", "number"},
		{"time.Time", "string"},
		{"url.URL", "string"},
		{"unsafe.Pointer", "any"},
		{"interface{}", "any"},
		{"[]string", "string[]"},
		{"map[string]int", "{ [key: string]: number }"},
		{"struct{ Field1 int; Field2 string }", "{ Field1: number; Field2: string }"},
		{"AliasInt", "number"},
		{"AliasMap", "{ [key: string]: string }"},
		{"T", "T"},
		{"*pkg.Custom", "any | null"},
		{"[]pkg.Custom", "any[]"},
		{"MyType[T]", "MyType<T>"},
		{"Result[K, V]", "Result<K, V>"},
		{"struct{ FieldOnly }", "{ unknown: any }"},
		{"map[UnknownStruct]string", "{ [key: string]: string }"},
		{"struct{}", "any"},
		{"error", "Error"},
		//
		{"*time.Time", "string"},
		{"*url.URL", "string"},
		{"decimal.Decimal", "string"},
		{"primitive.ObjectID", "string"},
		{"primitive.Decimal128", "string"},
		{"uuid.UUID", "string"},
		{"pgtype.UUID", "string"},
		{"sql.NullString", "string | null"},
		{"sql.NullInt64", "number | null"},
		{"pq.NullTime", "string | null"},
		{"sql.NullBool", "boolean | null"},
		{"complex64", "any"},
		{"complex128", "any"},
		{"CustomType", "CustomType"},
		{"map[AliasLoop1]string", "{ [key: string]: string }"},
	}

	for _, tc := range tests {
		if got := parser.GoTypeToTSType(tc.goType,
			aliasMap,
			typeParams,
			emptyStructMap,
			typeParamMapping,
			map[string]bool{}); got != tc.want {
			t.Errorf("GoTypeToTSType(%q) = %q, want %q", tc.goType, got, tc.want)
		}
	}
}

func TestIsAliasName(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"MyType", true},
		{"Alias123", true},
		{"aLowercase", false},
		{"", false},
		{"_Underscore", false},
		{"9Number", false},
	}

	for _, tt := range tests {
		if got := parser.IsAliasName(tt.input); got != tt.expected {
			t.Errorf("IsAliasName(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSplitGenericType(t *testing.T) {
	tests := []struct {
		name       string
		goType     string
		wantBase   string
		wantParams []string
	}{
		{
			name:       "SingleTypeParameter",
			goType:     "MyType[T]",
			wantBase:   "MyType",
			wantParams: []string{"T"},
		},
		{
			name:       "MultipleTypeParameters",
			goType:     "Result[K, V]",
			wantBase:   "Result",
			wantParams: []string{"K", "V"},
		},
		{
			name:       "WithSpaces",
			goType:     "MapEntry[  KeyType , ValueType  ]",
			wantBase:   "MapEntry",
			wantParams: []string{"KeyType", "ValueType"},
		},
		{
			name:       "NoGeneric",
			goType:     "PlainType",
			wantBase:   "PlainType",
			wantParams: nil,
		},
		{
			name:       "MalformedBrackets",
			goType:     "BrokenType[Param",
			wantBase:   "BrokenType[Param",
			wantParams: nil,
		},
		{
			name:       "EmptyParameter",
			goType:     "EmptyType[]",
			wantBase:   "EmptyType",
			wantParams: []string{""}, // or nil depending on design
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotParams := parser.SplitGenericType(tt.goType)
			if gotBase != tt.wantBase {
				t.Errorf("Base = %q, want %q", gotBase, tt.wantBase)
			}
			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("Params = %#v, want %#v", gotParams, tt.wantParams)
			}
		})
	}
}

func TestCheckGenericPatterns(t *testing.T) {
	aliasMap := map[string]string{
		"MyAlias": "string",
	}
	typeParams := []string{"T"}
	structMap := map[string]parser.StructInfo{}
	typeParamMapping := map[string]string{"T": "T"}
	visited := map[string]bool{}

	tests := []struct {
		name     string
		goType   string
		expected string
	}{
		{
			name:     "Simple generic with single param",
			goType:   "MyGeneric[int]",
			expected: "MyGeneric<number>",
		},
		{
			name:     "Generic with multiple params",
			goType:   "Pair[string, int]",
			expected: "Pair<string, number>",
		},
		{
			name:     "Generic with alias base",
			goType:   "MyAlias[T]",
			expected: "string<T>",
		},
		{
			name:     "Generic with type param",
			goType:   "Container[T]",
			expected: "Container<T>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.CheckGenericPatterns(tt.goType, aliasMap, typeParams, structMap, typeParamMapping, visited)

			// remove empty string
			gotClean := strings.ReplaceAll(got, " ", "")
			expectedClean := strings.ReplaceAll(tt.expected, " ", "")

			if gotClean != expectedClean {
				t.Errorf("checkGenericPatterns(%q) = %q, want %q", tt.goType, got, tt.expected)
			}
		})
	}
}

func TestParseStructType(t *testing.T) {
	aliasMap := map[string]string{}
	typeParams := []string{}
	structMap := map[string]parser.StructInfo{}
	typeParamMapping := map[string]string{}
	visited := map[string]bool{}

	tests := []struct {
		name     string
		goType   string
		expected string
	}{
		{
			name:     "simple struct with int and string",
			goType:   "struct{Id int; Name string}",
			expected: "{ Id: number; Name: string }",
		},
		{
			name:     "struct with bool and float64",
			goType:   "struct{Active bool; Score float64}",
			expected: "{ Active: boolean; Score: number }",
		},
		{
			name:     "struct with time.Time and sql.NullString",
			goType:   "struct{CreatedAt time.Time; Title sql.NullString}",
			expected: "{ CreatedAt: string; Title: string | null }",
		},
		{
			name:     "empty struct",
			goType:   "struct{}",
			expected: "{  }",
		},
		{
			name:     "struct with unsupported format (too few fields)",
			goType:   "struct{JustField}", //  The length of parts is less than minFieldParts
			expected: "{ unknown: any }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.ParseStructType(
				tt.goType,
				aliasMap,
				typeParams,
				structMap,
				typeParamMapping,
				visited,
			)

			if got != tt.expected {
				t.Errorf("parseStructType(%q) = %q, want %q", tt.goType, got, tt.expected)
			}
		})
	}
}

func TestGoTypeToTSType_MissedBranches(t *testing.T) {
	aliasMap := map[string]string{}
	typeParams := []string{}
	structMap := map[string]parser.StructInfo{}
	typeParamMapping := map[string]string{"X": "Xtype"}
	visited := map[string]bool{"loop": true}

	// Test Circular reference
	// If a type is already visited (to prevent infinite recursion),
	// GoTypeToTSType should return "any"
	if got := parser.GoTypeToTSType("loop",
		aliasMap,
		typeParams,
		structMap,
		typeParamMapping,
		visited); got != "any" {
		t.Errorf("Circular goType should yield 'any', got %q", got)
	}

	// Type parameter mapping
	// If the type is in typeParamMapping (generic type parameter),
	// it should return the mapped TypeScript type
	if got := parser.GoTypeToTSType("X",
		aliasMap,
		typeParams,
		structMap,
		typeParamMapping,
		map[string]bool{}); got != "Xtype" {
		t.Errorf("Type param mapping failed, got %q", got)
	}

	// Fallback for unknown types
	// If the type is not found in any map,
	// GoTypeToTSType should return the original type name
	if got := parser.GoTypeToTSType("NonMatchingType", aliasMap,
		typeParams,
		structMap,
		typeParamMapping,
		map[string]bool{}); got != "NonMatchingType" {
		t.Errorf("Fallback for unknown type failed, got %q", got)
	}
}

func TestIsUserDefinedStruct(t *testing.T) {
	structMap := map[string]parser.StructInfo{
		"User": {Name: "User"},
	}
	if !parser.IsUserDefinedStruct("User", structMap) {
		t.Errorf("expected true for defined struct")
	}
	if parser.IsUserDefinedStruct("Unknown", structMap) {
		t.Errorf("expected false for undefined struct")
	}
}
