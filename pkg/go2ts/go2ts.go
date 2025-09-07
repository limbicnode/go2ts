// Package go2ts provides core conversion from Go structs to TypeScript definitions.
package go2ts

import (
	"fmt"

	"github.com/limbicnode/go2ts/internal/generator"
	"github.com/limbicnode/go2ts/internal/parser"
)

// Convert - converts Go structs in the input directory to TypeScript types in the output file.
func Convert(inputDir, outputFile string) error {
	data, err := parser.ParseGoFiles(inputDir)
	if err != nil {
		return fmt.Errorf("failed to parse Go files in %q: %w", inputDir, err)
	}
	err = generator.GenerateTypeScript(data, outputFile)
	if err != nil {
		return fmt.Errorf("failed to generate TypeScript file %q: %w", outputFile, err)
	}
	return nil
}
