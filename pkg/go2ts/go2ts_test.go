package go2ts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/limbicnode/go2ts/pkg/go2ts"
)

func TestConvertCreatesOutput(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "types.ts")

	testInputDir := filepath.Join("..", "..", "test", "testdata", "model")

	err := go2ts.Convert(testInputDir, outputFile)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Output file not created: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("Output file is empty")
	}
}

func TestConvert_ParseGoFilesError_Concrete(t *testing.T) {
	tempDir := t.TempDir()

	badGoFile := `package main

	func thisIsBad {` // syntax error

	err := os.WriteFile(filepath.Join(tempDir, "bad.go"), []byte(badGoFile), 0644)
	if err != nil {
		t.Fatalf("failed to write bad.go: %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "out.ts")

	// parser.ParseGoFiles will fail on Convert due to syntax errors
	err = go2ts.Convert(tempDir, outputFile)
	if err == nil {
		t.Fatal("expected error due to bad Go source, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse Go files") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestConvert_GenerateTypeScriptError(t *testing.T) {
	// correct directory input
	inputDir := filepath.Join("..", "..", "test", "testdata", "model")
	//  Output path may be unwritable or invalid
	outputFile := "/root/forbidden/types.ts"

	err := go2ts.Convert(inputDir, outputFile)
	if err == nil || !strings.Contains(err.Error(), "failed to generate TypeScript") {
		t.Errorf("expected generate error, got %v", err)
	}
}
