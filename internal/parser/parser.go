// Package parser - parser parses Go source files to extract structs
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// StructField represents a field in a Go struct.
type StructField struct {
	Name string
	Type string
	Tags string
}

// GoStruct represents a Go struct definition.
type GoStruct struct {
	Name       string
	Fields     []StructField
	TypeParams []string // generic type parameters
}

// TypeAlias represents a Go type alias definition.
type TypeAlias struct {
	Name       string
	TypeParams []string // generic type parameters names
	Underlying string   // underlying type expression as string
}

// GoFileData contains parsed Go file information.
type GoFileData struct {
	Structs []GoStruct
	Aliases []TypeAlias
}

// StructInfo contains information about a Go struct.
type StructInfo struct {
	Name       string
	TypeParams []string
	Fields     []FieldInfo
}

// FieldInfo contains information about a struct field.
type FieldInfo struct {
	Name string
	Type string
	Tags string
}

var genericTypePattern = regexp.MustCompile(`[a-zA-Z0-9_]+\[.*\]`)

// ParseGoFiles recursively parses all .go files (except *_test.go) under the given directory.
// It extracts struct and type alias definitions along with generic type parameters.
func ParseGoFiles(dir string) (GoFileData, error) {
	var data GoFileData
	fset := token.NewFileSet()

	err := filepath.Walk(dir, func(path string, _ os.FileInfo, _ error) error {
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)

		if parseErr != nil {
			return parseErr
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec := spec.(*ast.TypeSpec)

				var typeParams []string
				if typeSpec.TypeParams != nil {
					for _, field := range typeSpec.TypeParams.List {
						for _, name := range field.Names {
							typeParams = append(typeParams, name.Name)
						}
					}
				}

				// If it's a struct type, extract fields
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					var fields []StructField
					for _, field := range structType.Fields.List {
						fieldType := ExprToString(field.Type)
						tag := ""
						if field.Tag != nil {
							tag = strings.Trim(field.Tag.Value, "`")
						}
						for _, name := range field.Names {
							fields = append(fields, StructField{
								Name: name.Name,
								Type: fieldType,
								Tags: tag,
							})
						}
					}
					data.Structs = append(data.Structs, GoStruct{
						Name:       typeSpec.Name.Name,
						Fields:     fields,
						TypeParams: typeParams,
					})
					continue
				}

				// Otherwise treat as type alias with underlying type
				underlying := ExprToString(typeSpec.Type)
				data.Aliases = append(data.Aliases, TypeAlias{
					Name:       typeSpec.Name.Name,
					TypeParams: typeParams,
					Underlying: underlying,
				})
			}
		}
		return nil
	})

	return data, err
}

// ExprToString converts a Go AST expression to its string representation.
func ExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + ExprToString(t.X)
	case *ast.SelectorExpr:
		return ExprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + ExprToString(t.Elt)
	case *ast.MapType:
		return "map[" + ExprToString(t.Key) + "]" + ExprToString(t.Value)
	case *ast.IndexExpr:
		return ExprToString(t.X) + "[" + ExprToString(t.Index) + "]"
	case *ast.IndexListExpr:
		var indexes []string
		for _, idx := range t.Indices {
			indexes = append(indexes, ExprToString(idx))
		}
		return ExprToString(t.X) + "[" + strings.Join(indexes, ", ") + "]"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.StructType:
		if t == nil || t.Fields == nil || len(t.Fields.List) == 0 {
			return "struct{}"
		}
		var parts []string
		for _, field := range t.Fields.List {
			typStr := ExprToString(field.Type)
			if len(field.Names) == 0 {
				// embedded field without explicit name
				parts = append(parts, typStr)
				continue
			}
			var names []string
			for _, n := range field.Names {
				if n != nil {
					names = append(names, n.Name)
				}
			}
			parts = append(parts, fmt.Sprintf("%s %s", strings.Join(names, ", "), typStr))
		}
		return "struct{ " + strings.Join(parts, "; ") + " }"
	default:
		return ""
	}
}

// GoTypeToTSType converts a Go type string into a corresponding TypeScript type.
func GoTypeToTSType(
	goType string,
	aliasMap map[string]string,
	typeParams []string,
	structMap map[string]StructInfo,
	typeParamMapping map[string]string,
	visited map[string]bool,
) string {
	goType = strings.TrimSpace(goType)

	if visited[goType] {
		return "any" // circular reference prevention
	}

	visited[goType] = true
	defer delete(visited, goType)

	// Preventing circular references
	if mapped, ok := typeParamMapping[goType]; ok {
		return mapped
	}

	if goType == "" {
		return ""
	}

	if special := checkSpecialCases(goType); special != "" {
		return special
	}

	// return generic type params
	for _, tp := range typeParams {
		if goType == tp {
			return tp
		}
	}

	const ptrPrefix = len("*")
	const slicePrefix = len("[]")

	if strings.HasPrefix(goType, "*") {
		inner := GoTypeToTSType(goType[ptrPrefix:], aliasMap, typeParams, structMap, typeParamMapping, visited)
		return inner + " | null"
	}

	if strings.HasPrefix(goType, "[]") {
		elem := GoTypeToTSType(goType[slicePrefix:], aliasMap, typeParams, structMap, typeParamMapping, visited)
		if strings.HasPrefix(elem, "{ [key:") && !strings.HasPrefix(elem, "(") {
			elem = "(" + elem + ")"
		}
		return elem + "[]"
	}

	if strings.HasPrefix(goType, "map[") {
		return parseMapType(goType,
			aliasMap,
			typeParams,
			structMap,
			typeParamMapping,
			visited)
	}

	if strings.HasPrefix(goType, "struct{") {
		return ParseStructType(goType,
			aliasMap,
			typeParams,
			structMap,
			typeParamMapping,
			visited)
	}

	if genericTypePattern.MatchString(goType) {
		return CheckGenericPatterns(goType,
			aliasMap,
			typeParams,
			structMap,
			typeParamMapping,
			visited)
	}

	if aliasResult := checkAliasTypes(goType,
		aliasMap,
		typeParams,
		structMap,
		typeParamMapping,
		visited); aliasResult != "" {
		return aliasResult
	}

	if basicResult := checkBasicTypes(goType); basicResult != goType {
		return basicResult
	}

	if complexResult := checkComplexTypes(goType); complexResult != "" {
		return complexResult
	}

	if IsAliasName(goType) {
		if IsUserDefinedStruct(goType, structMap) {
			return goType
		}
		return goType
	}
	return goType
}

func checkSpecialCases(goType string) string {
	switch goType {
	case "[]byte":
		return "Uint8Array"
	case "struct{}":
		return "any"
	case "func":
		return "(...args: any[]) => any"
	case "*time.Time", "*url.URL":
		return "string"
	}
	return ""
}

// IsUserDefinedStruct checks whether the type name is a user-defined struct.
func IsUserDefinedStruct(name string, structMap map[string]StructInfo) bool {
	_, ok := structMap[name]
	return ok
}

// IsAliasName - Checks if the type name starts with an uppercase letter
func IsAliasName(typeName string) bool {
	if typeName == "" {
		return false
	}
	r := rune(typeName[0])
	return r >= 'A' && r <= 'Z'
}

// SplitGenericType - parses a Go generic type string (e.g. "Result[T, E]")
// Returns the base type and a slice of generic type parameters.
// e.g. "Result[T, E]" -> base = "Result", params = ["T","E"]
// 1. Find the first '[' and last ']' to extract the generic parameter string.
// 2. Initialize a buffer to temporarily store characters for each parameter.
// 3. Use 'depth' to track nested generics:
//   - ',' at depth 0 marks the end of a parameter and triggers flush
//   - '[' increases depth (nested generic starts)
//   - ']' decreases depth (nested generic ends)
//
// 4. All other characters are appended to the buffer.
// 5. After iteration, flush the last parameter remaining in the buffer.
func SplitGenericType(goType string) (base string, params []string) {
	lidx := strings.Index(goType, "[")
	ridx := strings.LastIndex(goType, "]")
	if lidx < 0 || ridx <= lidx {
		return goType, nil // no generics
	}

	base = goType[:lidx]
	paramStr := goType[lidx+1 : ridx]
	if paramStr == "" {
		return base, []string{""}
	}

	var parts []string
	var buf strings.Builder
	depth := 0

	flush := func() {
		part := strings.TrimSpace(buf.String())
		if part != "" {
			parts = append(parts, part)
		}
		buf.Reset()
	}

	for _, r := range paramStr {
		switch r {
		case ',':
			if depth == 0 {
				flush()
				continue
			}
		case '[':
			depth++
		case ']':
			depth--
		}
		buf.WriteRune(r)
	}

	flush() // flush last parameter
	return base, parts
}

// CheckGenericPatterns - Converts a Go generic type to a TypeScript generic type, handling params and aliases.
func CheckGenericPatterns(
	goType string,
	aliasMap map[string]string,
	typeParams []string,
	structMap map[string]StructInfo,
	typeParamMapping map[string]string,
	visited map[string]bool,
) string {
	// Split base type and type parameters (e.g., "Result[T, E]" → base:"Result", params:["T","E"]
	base, params := SplitGenericType(goType)

	// Recursively convert all type parameters into TypeScript types
	tsParams := make([]string, 0, len(params))
	for _, p := range params {
		tsParam := GoTypeToTSType(
			p,
			aliasMap,
			typeParams,
			structMap,
			typeParamMapping,
			visited,
		)
		tsParams = append(tsParams, tsParam)
	}

	// If base type has an alias mapping, replace it (e.g., "int" → "number")
	if baseAlias, ok := aliasMap[base]; ok && baseAlias != base {
		base = GoTypeToTSType(
			baseAlias,
			aliasMap,
			typeParams,
			structMap,
			typeParamMapping,
			visited,
		)
	}

	// Return TypeScript generic type string, e.g., "PromiseResult<T, E>"
	return base + "<" + strings.Join(tsParams, ", ") + ">"
}

func checkBasicTypes(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"byte", "rune",
		"float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time.Time":
		return "string"
	case "url.URL":
		return "string"
	case "interface{}", "*interface{}", "interface {}", "*interface {}":
		return "any"
	case "complex64", "complex128":
		return "any"
	case "decimal.Decimal", "primitive.ObjectID", "primitive.Decimal128",
		"uuid.UUID", "pgtype.UUID":
		return "string"
	case "sql.NullString":
		return "string | null"
	case "sql.NullInt64":
		return "number | null"
	case "pq.NullTime":
		return "string | null"
	case "sql.NullBool":
		return "boolean | null"
	case "unsafe.Pointer":
		return "any"
	case "error":
		return "Error"
	}

	return goType
}

func checkAliasTypes(goType string,
	aliasMap map[string]string,
	typeParams []string,
	structMap map[string]StructInfo,
	typeParamMapping map[string]string,
	visited map[string]bool) string {
	if base, ok := aliasMap[goType]; ok {
		if base == goType {
			return "any"
		}
		return GoTypeToTSType(base, aliasMap, typeParams, structMap, typeParamMapping, visited)
	}
	return ""
}

func checkComplexTypes(goType string) string {
	if strings.Contains(goType, ".") {
		return "any"
	}
	if strings.ContainsAny(goType, "*[]") {
		return "any"
	}
	return ""
}

func parseMapType(
	goType string,
	aliasMap map[string]string,
	typeParams []string,
	structMap map[string]StructInfo,
	typeParamMapping map[string]string,
	visited map[string]bool,
) string {
	const mapTypeSplitLimit = 2

	inner := goType[len("map["):]
	parts := strings.SplitN(inner, "]", mapTypeSplitLimit)
	if len(parts) != mapTypeSplitLimit {
		return "any"
	}
	rawKey := strings.TrimSpace(parts[0])
	rawVal := strings.TrimSpace(parts[1])

	var keyTS string
	if strings.HasPrefix(rawKey, "struct{") {
		keyTS = "string"
	} else {
		switch rawKey {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64":
			keyTS = "number"
		default:
			keyResolved := rawKey
			visitedKeys := map[string]bool{}
			for {
				if visitedKeys[keyResolved] {
					break
				}
				visitedKeys[keyResolved] = true
				if base, ok := aliasMap[keyResolved]; ok && base != keyResolved {
					keyResolved = base
				} else {
					break
				}
			}
			keyTS = GoTypeToTSType(keyResolved, aliasMap, typeParams, structMap, typeParamMapping, visited)
			if keyTS != "string" && keyTS != "number" && keyTS != "symbol" {
				keyTS = "string"
			}
		}
	}

	valTS := GoTypeToTSType(rawVal,
		aliasMap,
		typeParams,
		structMap,
		typeParamMapping,
		visited)

	if strings.Contains(valTS, "|") && !strings.HasSuffix(valTS, "[]") && !strings.HasPrefix(valTS, "(") {
		valTS = "(" + valTS + ")"
	}
	return "{ [key: " + keyTS + "]: " + valTS + " }"
}

const minFieldParts = 2 // name, type

// ParseStructType converts an inline Go anonymous struct into a TypeScript object type string.
// e.g. ["Foo: number", "Bar: string"] → "{ Foo: number; Bar: string }"
func ParseStructType(
	goType string,
	aliasMap map[string]string,
	typeParams []string,
	structMap map[string]StructInfo,
	typeParamMapping map[string]string,
	visited map[string]bool,
) string {
	body := strings.TrimPrefix(goType, "struct{")
	body = strings.TrimSuffix(body, "}")
	fields := strings.Split(body, ";")
	var tsFields []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		parts := strings.Fields(f)
		if len(parts) >= minFieldParts {
			tsFields = append(tsFields, fmt.Sprintf("%s: %s",
				parts[0],
				GoTypeToTSType(strings.Join(parts[1:], " "), aliasMap, typeParams, structMap, typeParamMapping, visited)))
		} else {
			tsFields = append(tsFields, "unknown: any")
		}
	}
	return "{ " + strings.Join(tsFields, "; ") + " }"
}
