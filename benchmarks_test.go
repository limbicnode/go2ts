package go2ts

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/limbicnode/go2ts/internal/generator"
	"github.com/limbicnode/go2ts/internal/parser"
)

// BenchmarkGenerateTypeScript measures code generation performance
func BenchmarkGenerateTypeScript(b *testing.B) {
	runBenchmark(b, false)
}

// BenchmarkGenerateTypeScriptMem measures memory usage
func BenchmarkGenerateTypeScriptMem(b *testing.B) {
	runBenchmark(b, true)
}

// runBenchmark is the common benchmark logic
func runBenchmark(b *testing.B, reportAllocs bool) {
	if reportAllocs {
		b.ReportAllocs()
	}

	dir := filepath.Join("..", "..", "test", "testdata", "model")
	data, err := parser.ParseGoFiles(dir)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tmpFile := filepath.Join(b.TempDir(), "types.ts")
		err := generator.GenerateTypeScript(data, tmpFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseGoFiles measures parsing performance
func BenchmarkParseGoFiles(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseGoFiles("./testdata")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseGoFilesMem measures memory usage
func BenchmarkParseGoFilesMem(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseGoFiles("./testdata")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtractJSONTag(b *testing.B) {
	tag := `json:"example,omitempty"`
	for i := 0; i < b.N; i++ {
		_ = reflect.StructTag(tag).Get("json")
	}
}

// BenchmarkGoTypeToTSType measures type conversion performance
func BenchmarkGoTypeToTSType(b *testing.B) {
	aliasMap := make(map[string]string)
	typeParams := []string{"T", "U"}
	structMap := make(map[string]parser.StructInfo)
	typeParamMapping := make(map[string]string)
	visited := make(map[string]bool)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.GoTypeToTSType("map[string]int", aliasMap, typeParams, structMap, typeParamMapping, visited)
	}
}

func BenchmarkWriteFile(b *testing.B) {
	content := []byte("long string ...")
	for i := 0; i < b.N; i++ {
		tmp := b.TempDir() + "/tmp.ts"
		if err := os.WriteFile(tmp, content, 0644); err != nil {
			b.Fatal(err)
		}
	}
}
