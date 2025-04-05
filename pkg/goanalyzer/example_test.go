package goanalyzer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cred.com/hack25/backend/pkg/goanalyzer"
)

// ExampleAnalyzeFile demonstrates how to use the AnalyzeFile function
func ExampleAnalyzeFile() {
	// Create a temporary Go file for testing
	tempDir, err := os.MkdirTemp("", "goanalyzer-example")
	if err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go file
	testFilePath := filepath.Join(tempDir, "example.go")
	testFileContent := `package example

import (
	"fmt"
	"strings"
)

const (
	Version = "1.0.0"
	debug   = false
)

var DefaultOptions = Options{
	Timeout: 30,
	Retries: 3,
}

type Options struct {
	Timeout int
	Retries int
}

func (o Options) String() string {
	return fmt.Sprintf("Timeout: %d, Retries: %d", o.Timeout, o.Retries)
}

func Process(input string, opts Options) (string, error) {
	if debug {
		fmt.Println("Processing with options:", opts.String())
	}
	
	result := strings.ToUpper(input)
	return result, nil
}
`

	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing test file: %v\n", err)
		return
	}

	// Analyze the file
	symbols, err := goanalyzer.AnalyzeFile(testFilePath)
	if err != nil {
		fmt.Printf("Error analyzing file: %v\n", err)
		return
	}

	// Print summary of symbols found
	fmt.Println("Symbols found:")

	// Count symbols by kind
	kindCount := make(map[string]int)
	for _, symbol := range symbols {
		kindCount[symbol.Kind]++
	}

	for kind, count := range kindCount {
		fmt.Printf("- %s: %d\n", kind, count)
	}

	// Find exported symbols
	fmt.Println("\nExported symbols:")
	for _, symbol := range symbols {
		if symbol.Exported {
			fmt.Printf("- %s (%s)\n", symbol.Name, symbol.Kind)
		}
	}

	// Find function calls
	fmt.Println("\nFunction calls:")
	for _, symbol := range symbols {
		if symbol.Kind == "func" || symbol.Kind == "method" {
			if len(symbol.Calls) > 0 {
				fmt.Printf("- %s calls:\n", symbol.Name)
				for _, call := range symbol.Calls {
					callee := call.Callee
					if call.Package != "" {
						callee = call.Package + "." + callee
					}
					fmt.Printf("  - %s\n", callee)
				}
			}
		}
	}

	// Output:
	// Symbols found:
	// - package: 1
	// - import: 2
	// - const: 2
	// - var: 1
	// - struct: 1
	// - method: 1
	// - func: 1
	//
	// Exported symbols:
	// - Version (const)
	// - DefaultOptions (var)
	// - Options (struct)
	// - String (method)
	// - Process (func)
	//
	// Function calls:
	// - String calls:
	//   - fmt.Sprintf
	// - Process calls:
	//   - fmt.Println
	//   - opts.String
	//   - strings.ToUpper
}

func TestAnalyzeFile(t *testing.T) {
	// Create a temporary Go file for testing
	tempDir, err := os.MkdirTemp("", "goanalyzer-test")
	if err != nil {
		t.Fatalf("Error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go file
	testFilePath := filepath.Join(tempDir, "test.go")
	testFileContent := `package test

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}

func Calculate(a, b int) int {
	sum := Add(a, b)
	return Multiply(sum, 2)
}
`

	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	if err != nil {
		t.Fatalf("Error writing test file: %v", err)
	}

	// Analyze the file
	symbols, err := goanalyzer.AnalyzeFile(testFilePath)
	if err != nil {
		t.Fatalf("Error analyzing file: %v", err)
	}

	// Verify package
	if len(symbols) == 0 || symbols[0].Kind != "package" || symbols[0].Name != "test" {
		t.Errorf("Expected package 'test', got %+v", symbols[0])
	}

	// Find Calculate function
	var calculateFunc *goanalyzer.CodeSymbol
	for i, symbol := range symbols {
		if symbol.Kind == "func" && symbol.Name == "Calculate" {
			calculateFunc = &symbols[i]
			break
		}
	}

	if calculateFunc == nil {
		t.Fatal("Calculate function not found")
	}

	// Verify Calculate calls Add and Multiply
	if len(calculateFunc.Calls) != 2 {
		t.Errorf("Expected Calculate to have 2 calls, got %d", len(calculateFunc.Calls))
	}

	callees := make(map[string]bool)
	for _, call := range calculateFunc.Calls {
		callees[call.Callee] = true
	}

	if !callees["Add"] || !callees["Multiply"] {
		t.Errorf("Expected Calculate to call Add and Multiply, got calls to %v", callees)
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(symbols)
	if err != nil {
		t.Fatalf("Error marshaling symbols to JSON: %v", err)
	}

	var unmarshaledSymbols []goanalyzer.CodeSymbol
	err = json.Unmarshal(jsonData, &unmarshaledSymbols)
	if err != nil {
		t.Fatalf("Error unmarshaling symbols from JSON: %v", err)
	}

	if len(unmarshaledSymbols) != len(symbols) {
		t.Errorf("Expected %d symbols after JSON round-trip, got %d", len(symbols), len(unmarshaledSymbols))
	}
}
