package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer"
	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

func main() {
	// Parse command line arguments
	dirFlag := flag.String("dir", "", "Directory to analyze (analyzes all Go files recursively)")
	fileFlag := flag.String("file", "", "Single Go file to analyze")
	jsonOutputFlag := flag.Bool("json", false, "Output results as JSON")
	showCodeFlag := flag.Bool("code", false, "Show code blocks for functions")
	showCallsFlag := flag.Bool("calls", false, "Show function call hierarchy")
	showRefsFlag := flag.Bool("refs", false, "Show symbol references")
	showStatementsFlag := flag.Bool("statements", false, "Show detailed statement analysis")
	functionNameFlag := flag.String("function", "", "Focus on a specific function (used with -statements)")
	flag.Parse()

	if *dirFlag == "" && *fileFlag == "" {
		fmt.Println("Error: Either -dir or -file must be specified")
		flag.Usage()
		os.Exit(1)
	}

	// Create a new analyzer
	analyzer := goanalyzer.New()

	var results []models.FileAnalysis

	// Analyze directory or file
	if *dirFlag != "" {
		// Make sure the path is absolute
		absPath, err := filepath.Abs(*dirFlag)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Analyzing Go files in directory: %s\n", absPath)
		results, err = analyzer.AnalyzeDirectory(absPath)
		if err != nil {
			fmt.Printf("Error analyzing directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Make sure the path is absolute
		absPath, err := filepath.Abs(*fileFlag)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Analyzing Go file: %s\n", absPath)
		fileAnalysis, err := analyzer.AnalyzeFile(absPath)
		if err != nil {
			fmt.Printf("Error analyzing file: %v\n", err)
			os.Exit(1)
		}
		results = append(results, *fileAnalysis)
	}

	// Output results
	if *jsonOutputFlag {
		// Output as JSON
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		// Output in human-readable format
		printResults(results, *showCodeFlag, *showCallsFlag, *showRefsFlag, *showStatementsFlag, *functionNameFlag, analyzer)
	}
}

// printResults prints the analysis results in a human-readable format
func printResults(results []models.FileAnalysis, showCode, showCalls, showRefs, showStatements bool,
	functionName string, analyzer *goanalyzer.Analyzer) {
	for _, file := range results {
		fmt.Printf("\n=== File: %s ===\n", file.FilePath)
		fmt.Printf("Package: %s\n", file.Package)

		if len(file.Imports) > 0 {
			fmt.Println("\nImports:")
			for _, imp := range file.Imports {
				fmt.Printf("  %s: %s\n", imp.Name, imp.Type)
			}
		}

		if len(file.Constants) > 0 {
			fmt.Println("\nConstants:")
			for _, c := range file.Constants {
				fmt.Printf("  %s: %s = %s (exported: %t)\n", c.Name, c.Type, c.Value, c.Exported)
			}
		}

		if len(file.Variables) > 0 {
			fmt.Println("\nVariables:")
			for _, v := range file.Variables {
				fmt.Printf("  %s: %s = %s (exported: %t)\n", v.Name, v.Type, v.Value, v.Exported)
			}
		}

		if len(file.Types) > 0 {
			fmt.Println("\nTypes:")
			for _, t := range file.Types {
				fmt.Printf("  %s: %s (exported: %t)\n", t.Name, t.Type, t.Exported)
			}
		}

		if len(file.Structs) > 0 {
			fmt.Println("\nStructs:")
			for _, s := range file.Structs {
				fmt.Printf("  %s (exported: %t)\n", s.Name, s.Exported)
				if len(s.Fields) > 0 {
					fmt.Println("    Fields:")
					for _, f := range s.Fields {
						fmt.Printf("      %s: %s (exported: %t)\n", f.Name, f.Type, f.Exported)
					}
				}
			}
		}

		if len(file.Interfaces) > 0 {
			fmt.Println("\nInterfaces:")
			for _, i := range file.Interfaces {
				fmt.Printf("  %s (exported: %t)\n", i.Name, i.Exported)
				if len(i.Methods) > 0 {
					fmt.Println("    Methods:")
					for _, m := range i.Methods {
						fmt.Printf("      %s\n", m)
					}
				}
			}
		}

		if len(file.Functions) > 0 {
			fmt.Println("\nFunctions:")
			for _, f := range file.Functions {
				if f.Kind == "method" {
					fmt.Printf("  Method: %s on %s (exported: %t)\n", f.Name, f.Receiver, f.Exported)
				} else {
					fmt.Printf("  Function: %s (exported: %t)\n", f.Name, f.Exported)
				}

				// Display code block if requested
				if showCode && f.CodeBlock != "" {
					fmt.Println("    Code:")
					for _, line := range strings.Split(f.CodeBlock, "\n") {
						fmt.Printf("      %s\n", line)
					}
				}

				// Display statement analysis if requested and matches function filter
				if showStatements && (functionName == "" || functionName == f.Name) {
					// Get AST statements if available
					if len(f.Statements) > 0 {
						fmt.Println("    Statement Analysis:")
						// We need to call the analyzer to analyze the statements
						// Since we don't have direct access to the Analyzer.AnalyzeStatements method,
						// we'll note this in the output
						fmt.Println("      Statement analysis available but requires internal access to AnalyzeStatements")
						fmt.Println("      Call hierarchy can provide insights in the meantime:")
					}
				}

				// Display call hierarchy if requested
				if showCalls {
					calls := analyzer.GetCallHierarchy(file.FilePath, f.Name)
					if len(calls) > 0 {
						fmt.Println("    Calls:")
						for _, call := range calls {
							fmt.Printf("      -> %s (line %d)\n", call.Callee, call.Position.Line)
						}
					}
				}

				if len(f.Parameters) > 0 {
					fmt.Println("    Parameters:")
					for _, p := range f.Parameters {
						if p.Name == "" {
							fmt.Printf("      %s\n", p.Type)
						} else {
							fmt.Printf("      %s: %s\n", p.Name, p.Type)
						}
					}
				}

				if len(f.Results) > 0 {
					fmt.Println("    Returns:")
					for _, r := range f.Results {
						if r.Name == "" {
							fmt.Printf("      %s\n", r.Type)
						} else {
							fmt.Printf("      %s: %s\n", r.Name, r.Type)
						}
					}
				}
			}
		}

		// Show references if requested
		if showRefs && len(file.References) > 0 {
			fmt.Println("\nReferences:")
			refsBySymbol := make(map[string][]models.ReferenceInfo)

			// Group references by symbol
			for _, ref := range file.References {
				refsBySymbol[ref.Symbol] = append(refsBySymbol[ref.Symbol], ref)
			}

			// Display references
			for symbol, refs := range refsBySymbol {
				fmt.Printf("  %s:\n", symbol)

				// Count declarations and usages
				var declarations, usages int
				for _, ref := range refs {
					if ref.RefType == "declaration" {
						declarations++
					} else {
						usages++
					}
				}

				fmt.Printf("    %d declaration(s), %d usage(s)\n", declarations, usages)

				// Show first few usages
				count := 0
				for _, ref := range refs {
					if ref.RefType == "usage" && count < 5 { // Limit to first 5 usages to avoid clutter
						fmt.Printf("    - Used at line %d\n", ref.Position.Line)
						count++
					}
				}

				if count < usages {
					fmt.Printf("    - ... and %d more usages\n", usages-count)
				}
			}
		}
	}
}
