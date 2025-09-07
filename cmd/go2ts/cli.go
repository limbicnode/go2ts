// Package cli provides core conversion from Go structs to TypeScript definitions by cli.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/limbicnode/go2ts/pkg/go2ts"
)

func main() {
	inputDir := flag.String("in", "./internal/model", "Directory to scan Go structs")
	outputFile := flag.String("out", "types.ts", "Output TypeScript file path")
	flag.Parse()

	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		log.Fatalf("Input directory does not exist: %s\n", *inputDir)
	}

	if err := go2ts.Convert(*inputDir, *outputFile); err != nil {
		log.Fatal(err)
	}
}
