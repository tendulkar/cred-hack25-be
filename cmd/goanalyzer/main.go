package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer"
)

func main() {
	var filePath string
	var recursive bool
	var format string
	var outputFile string

	// Parse command-line arguments
	flag.StringVar(&filePath, "path", "", "Path to a Go file or directory")
	flag.BoolVar(&recursive, "recursive", false, "Recursively analyze directories")
	flag.StringVar(&format, "format", "json", "Output format (json, text)")
	flag.StringVar(&outputFile, "output", "", "Output file (default: stdout)")
	flag.Parse()

	if filePath == "" {
		fmt.Println("Please provide a file or directory path using -path flag")
		flag.Usage()
		os.Exit(1)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("Error accessing path %s: %v", filePath, err)
	}

	// Prepare output writer
	var output = os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Error creating output file %s: %v", outputFile, err)
		}
		defer f.Close()
		output = f
	}

	// Analyze files
	if fileInfo.IsDir() {
		analyzeDirectory(filePath, recursive, format, output)
	} else {
		analyzeFile(filePath, format, output)
	}
}

func analyzeFile(filePath, format string, output *os.File) {
	if !strings.HasSuffix(filePath, ".go") {
		fmt.Fprintf(output, "Skipping non-Go file: %s\n", filePath)
		return
	}

	symbols, err := goanalyzer.AnalyzeFile(filePath)
	if err != nil {
		log.Printf("Error analyzing file %s: %v", filePath, err)
		return
	}

	// Output results based on format
	switch format {
	case "json":
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(map[string]interface{}{
			"file":    filePath,
			"symbols": symbols,
		}); err != nil {
			log.Printf("Error encoding JSON: %v", err)
		}

	case "text":
		fmt.Fprintf(output, "# File: %s\n\n", filePath)

		// Group symbols by kind
		symbolsByKind := make(map[string][]goanalyzer.CodeSymbol)
		for _, symbol := range symbols {
			symbolsByKind[symbol.Kind] = append(symbolsByKind[symbol.Kind], symbol)
		}

		// Output each group
		for kind, syms := range symbolsByKind {
			fmt.Fprintf(output, "## %s\n\n", strings.ToUpper(kind))
			for _, sym := range syms {
				printSymbol(sym, output, 0)
			}
			fmt.Fprintln(output)
		}

		// Output call hierarchy
		fmt.Fprintf(output, "## CALL HIERARCHY\n\n")
		for _, sym := range symbols {
			if sym.Kind == "func" || sym.Kind == "method" {
				if len(sym.Calls) > 0 {
					fmt.Fprintf(output, "- %s calls:\n", sym.Name)
					for _, call := range sym.Calls {
						caller := call.Callee
						if call.Package != "" {
							caller = call.Package + "." + caller
						}
						fmt.Fprintf(output, "  - %s (line %d)\n", caller, call.Line)
					}
					fmt.Fprintln(output)
				}
			}
		}

	default:
		log.Printf("Unsupported format: %s", format)
	}
}

func analyzeDirectory(dirPath string, recursive bool, format string, output *os.File) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and .git directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}

		// If not recursive and it's not the root directory, skip subdirectories
		if !recursive && info.IsDir() && path != dirPath {
			return filepath.SkipDir
		}

		// Analyze Go files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			analyzeFile(path, format, output)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", dirPath, err)
	}
}

func printSymbol(symbol goanalyzer.CodeSymbol, output *os.File, indent int) {
	indentation := strings.Repeat("  ", indent)

	exportMark := " "
	if symbol.Exported {
		exportMark = "*"
	}

	typePart := ""
	if symbol.Type != "" {
		typePart = ": " + symbol.Type
	}

	fmt.Fprintf(output, "%s%s %s%s (line %d)\n", indentation, exportMark, symbol.Name, typePart, symbol.Line)

	// Print fields for structs
	if len(symbol.Fields) > 0 {
		fmt.Fprintf(output, "%s  Fields:\n", indentation)
		for _, field := range symbol.Fields {
			printSymbol(field, output, indent+2)
		}
	}

	// Print methods for types
	if len(symbol.Methods) > 0 {
		fmt.Fprintf(output, "%s  Methods:\n", indentation)
		for _, method := range symbol.Methods {
			printSymbol(method, output, indent+2)
		}
	}

	// Print parameters and results for functions
	if len(symbol.Params) > 0 {
		fmt.Fprintf(output, "%s  Parameters:\n", indentation)
		for _, param := range symbol.Params {
			printSymbol(param, output, indent+2)
		}
	}

	if len(symbol.Results) > 0 {
		fmt.Fprintf(output, "%s  Results:\n", indentation)
		for _, result := range symbol.Results {
			printSymbol(result, output, indent+2)
		}
	}
}
