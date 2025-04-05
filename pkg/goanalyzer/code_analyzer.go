package goanalyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

// CodeSymbol represents a symbol in a Go file
type CodeSymbol struct {
	Name       string          `json:"name"`
	Kind       string          `json:"kind"` // "package", "import", "const", "var", "type", "func", "struct", "interface", etc.
	Line       int             `json:"line"`
	Exported   bool            `json:"exported"`   // whether the symbol is exported (starts with uppercase)
	Receiver   string          `json:"receiver"`   // for methods, the receiver type
	Type       string          `json:"type"`       // type information if available
	Fields     []CodeSymbol    `json:"fields"`     // for structs and interfaces
	Methods    []CodeSymbol    `json:"methods"`    // for types
	Params     []CodeSymbol    `json:"params"`     // for functions
	Results    []CodeSymbol    `json:"results"`    // for functions
	References []CodeReference `json:"references"` // references to this symbol
	Calls      []CodeCall      `json:"calls"`      // for functions, what other functions it calls
}

// CodeReference represents a reference to a symbol
type CodeReference struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Context string `json:"context"` // surrounding code
}

// CodeCall represents a function call
type CodeCall struct {
	Callee    string       `json:"callee"`    // name of the called function
	Package   string       `json:"package"`   // package of the called function if not in the same package
	Line      int          `json:"line"`      // line where the call occurs
	Arguments []CodeSymbol `json:"arguments"` // arguments passed to the call
}

// AnalyzeFile analyzes a Go file and returns its symbols
func AnalyzeFile(filename string) ([]CodeSymbol, error) {
	fset := token.NewFileSet()

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filename, err)
	}

	file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing file %s: %w", filename, err)
	}

	symbols := []CodeSymbol{}

	// Package
	symbols = append(symbols, CodeSymbol{
		Name: file.Name.Name,
		Kind: "package",
		Line: fset.Position(file.Name.Pos()).Line,
	})

	// Imports
	for _, imp := range file.Imports {
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			// Extract last part of the import path as the package name
			path := strings.Trim(imp.Path.Value, "\"")
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}

		symbols = append(symbols, CodeSymbol{
			Name: name,
			Kind: "import",
			Type: strings.Trim(imp.Path.Value, "\""),
			Line: fset.Position(imp.Pos()).Line,
		})
	}

	// Declarations (consts, vars, types, funcs)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.CONST:
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					for i, name := range vs.Names {
						typeStr := ""
						if vs.Type != nil {
							typeStr = getTypeString(vs.Type)
						}

						// Value string can be used to set the Type if needed
						if i < len(vs.Values) && vs.Type == nil {
							vs.Type = vs.Values[i]
						}

						symbols = append(symbols, CodeSymbol{
							Name:     name.Name,
							Kind:     "const",
							Line:     fset.Position(name.Pos()).Line,
							Exported: ast.IsExported(name.Name),
							Type:     typeStr,
						})
					}
				}

			case token.VAR:
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					for i, name := range vs.Names {
						typeStr := ""
						if vs.Type != nil {
							typeStr = getTypeString(vs.Type)
						}

						// Value string can be used to set the Type if needed
						if i < len(vs.Values) && vs.Type == nil {
							vs.Type = vs.Values[i]
						}

						symbols = append(symbols, CodeSymbol{
							Name:     name.Name,
							Kind:     "var",
							Line:     fset.Position(name.Pos()).Line,
							Exported: ast.IsExported(name.Name),
							Type:     typeStr,
						})
					}
				}

			case token.TYPE:
				for _, spec := range d.Specs {
					ts := spec.(*ast.TypeSpec)
					symbol := CodeSymbol{
						Name:     ts.Name.Name,
						Kind:     "type",
						Line:     fset.Position(ts.Pos()).Line,
						Exported: ast.IsExported(ts.Name.Name),
						Type:     getTypeString(ts.Type),
					}

					// If it's a struct or interface, extract fields and methods
					switch t := ts.Type.(type) {
					case *ast.StructType:
						symbol.Kind = "struct"
						if t.Fields != nil {
							for _, field := range t.Fields.List {
								for _, name := range field.Names {
									symbol.Fields = append(symbol.Fields, CodeSymbol{
										Name:     name.Name,
										Kind:     "field",
										Line:     fset.Position(name.Pos()).Line,
										Exported: ast.IsExported(name.Name),
										Type:     getTypeString(field.Type),
									})
								}
							}
						}

					case *ast.InterfaceType:
						symbol.Kind = "interface"
						if t.Methods != nil {
							for _, method := range t.Methods.List {
								for _, name := range method.Names {
									symbol.Methods = append(symbol.Methods, CodeSymbol{
										Name:     name.Name,
										Kind:     "method",
										Line:     fset.Position(name.Pos()).Line,
										Exported: ast.IsExported(name.Name),
										Type:     getTypeString(method.Type),
									})
								}
							}
						}
					}

					symbols = append(symbols, symbol)
				}
			}

		case *ast.FuncDecl:
			// Function or method
			symbol := CodeSymbol{
				Name:     d.Name.Name,
				Kind:     "func",
				Line:     fset.Position(d.Pos()).Line,
				Exported: ast.IsExported(d.Name.Name),
			}

			// If it has a receiver, it's a method
			if d.Recv != nil {
				symbol.Kind = "method"
				for _, field := range d.Recv.List {
					symbol.Receiver = getTypeString(field.Type)
				}
			}

			// Extract parameters
			if d.Type.Params != nil {
				for _, field := range d.Type.Params.List {
					typeStr := getTypeString(field.Type)

					if len(field.Names) == 0 {
						// Anonymous parameter
						symbol.Params = append(symbol.Params, CodeSymbol{
							Name: "",
							Kind: "param",
							Type: typeStr,
							Line: fset.Position(field.Pos()).Line,
						})
					} else {
						for _, name := range field.Names {
							symbol.Params = append(symbol.Params, CodeSymbol{
								Name: name.Name,
								Kind: "param",
								Type: typeStr,
								Line: fset.Position(name.Pos()).Line,
							})
						}
					}
				}
			}

			// Extract results
			if d.Type.Results != nil {
				for _, field := range d.Type.Results.List {
					typeStr := getTypeString(field.Type)

					if len(field.Names) == 0 {
						// Anonymous result
						symbol.Results = append(symbol.Results, CodeSymbol{
							Name: "",
							Kind: "result",
							Type: typeStr,
							Line: fset.Position(field.Pos()).Line,
						})
					} else {
						for _, name := range field.Names {
							symbol.Results = append(symbol.Results, CodeSymbol{
								Name: name.Name,
								Kind: "result",
								Type: typeStr,
								Line: fset.Position(name.Pos()).Line,
							})
						}
					}
				}
			}

			// Analyze function body for calls
			if d.Body != nil {
				ast.Inspect(d.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						callSymbol := CodeCall{
							Line: fset.Position(call.Pos()).Line,
						}

						// Extract callee
						switch fun := call.Fun.(type) {
						case *ast.Ident:
							callSymbol.Callee = fun.Name

						case *ast.SelectorExpr:
							if x, ok := fun.X.(*ast.Ident); ok {
								callSymbol.Package = x.Name
							}
							callSymbol.Callee = fun.Sel.Name
						}

						// Extract arguments
						for _, arg := range call.Args {
							callSymbol.Arguments = append(callSymbol.Arguments, CodeSymbol{
								Type: getExprString(arg),
							})
						}

						symbol.Calls = append(symbol.Calls, callSymbol)
					}

					return true
				})
			}

			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

// getTypeString returns a string representation of a type
func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getExprString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt)
		}
		return "[" + getExprString(t.Len) + "]" + getTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + getTypeString(t.Value)
		case ast.RECV:
			return "<-chan " + getTypeString(t.Value)
		default:
			return "chan " + getTypeString(t.Value)
		}
	case *ast.FuncType:
		return "func" + getFuncTypeParams(t.Params) + getFuncTypeResults(t.Results)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.Ellipsis:
		return "..." + getTypeString(t.Elt)
	}

	return fmt.Sprintf("%T", expr)
}

// getExprString returns a string representation of an expression
func getExprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return getExprString(e.X) + "." + e.Sel.Name
	case *ast.CallExpr:
		args := []string{}
		for _, arg := range e.Args {
			args = append(args, getExprString(arg))
		}
		return getExprString(e.Fun) + "(" + strings.Join(args, ", ") + ")"
	case *ast.BinaryExpr:
		return getExprString(e.X) + " " + e.Op.String() + " " + getExprString(e.Y)
	case *ast.UnaryExpr:
		return e.Op.String() + getExprString(e.X)
	case *ast.CompositeLit:
		return getTypeString(e.Type) + "{}"
	}

	return fmt.Sprintf("%T", expr)
}

// getFuncTypeParams returns a string representation of function parameters
func getFuncTypeParams(fields *ast.FieldList) string {
	if fields == nil {
		return "()"
	}

	params := []string{}
	for _, field := range fields.List {
		typeStr := getTypeString(field.Type)

		if len(field.Names) == 0 {
			params = append(params, typeStr)
		} else {
			names := []string{}
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			params = append(params, strings.Join(names, ", ")+" "+typeStr)
		}
	}

	return "(" + strings.Join(params, ", ") + ")"
}

// getFuncTypeResults returns a string representation of function results
func getFuncTypeResults(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}

	results := []string{}
	for _, field := range fields.List {
		typeStr := getTypeString(field.Type)

		if len(field.Names) == 0 {
			results = append(results, typeStr)
		} else {
			names := []string{}
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			results = append(results, strings.Join(names, ", ")+" "+typeStr)
		}
	}

	if len(results) == 1 && !strings.Contains(results[0], " ") {
		return " " + results[0]
	}

	return " (" + strings.Join(results, ", ") + ")"
}

// FindCallHierarchy analyzes the call hierarchy for a given function
func FindCallHierarchy(filename string, functionName string) ([]string, error) {
	symbols, err := AnalyzeFile(filename)
	if err != nil {
		return nil, err
	}

	// Will hold the call paths (currently we're just returning direct callers)

	// Find the target function
	var targetFunc *CodeSymbol
	for i, symbol := range symbols {
		if (symbol.Kind == "func" || symbol.Kind == "method") && symbol.Name == functionName {
			targetFunc = &symbols[i]
			break
		}
	}

	if targetFunc == nil {
		return nil, fmt.Errorf("function %s not found in %s", functionName, filename)
	}

	// Find functions that call the target function
	callers := []string{}
	for _, symbol := range symbols {
		if symbol.Kind != "func" && symbol.Kind != "method" {
			continue
		}

		for _, call := range symbol.Calls {
			if call.Callee == functionName {
				callers = append(callers, symbol.Name)
				break
			}
		}
	}

	return callers, nil
}
