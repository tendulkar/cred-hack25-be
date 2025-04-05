package analyzer

import (
	"fmt"
	"go/ast"
	"strings"
)

// formatNode formats an AST node as a string
func (a *Analyzer) formatNode(node ast.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.BasicLit:
		return n.Value
	case *ast.SelectorExpr:
		return a.formatNode(n.X) + "." + n.Sel.Name
	case *ast.StarExpr:
		return "*" + a.formatNode(n.X)
	case *ast.ArrayType:
		if n.Len == nil {
			return "[]" + a.formatNode(n.Elt)
		}
		return "[" + a.formatNode(n.Len) + "]" + a.formatNode(n.Elt)
	case *ast.MapType:
		return "map[" + a.formatNode(n.Key) + "]" + a.formatNode(n.Value)
	case *ast.StructType:
		return "struct{...}"
	case *ast.InterfaceType:
		return "interface{...}"
	case *ast.FuncType:
		return "func(...) ..."
	case *ast.ChanType:
		var dir string
		switch n.Dir {
		case ast.SEND:
			dir = "chan<- "
		case ast.RECV:
			dir = "<-chan "
		default:
			dir = "chan "
		}
		return dir + a.formatNode(n.Value)
	case *ast.BinaryExpr:
		return a.formatNode(n.X) + " " + n.Op.String() + " " + a.formatNode(n.Y)
	case *ast.UnaryExpr:
		return n.Op.String() + a.formatNode(n.X)
	case *ast.ParenExpr:
		return "(" + a.formatNode(n.X) + ")"
	case *ast.CallExpr:
		var args []string
		for _, arg := range n.Args {
			args = append(args, a.formatNode(arg))
		}
		return a.formatNode(n.Fun) + "(" + strings.Join(args, ", ") + ")"
	default:
		return fmt.Sprintf("<%T>", node)
	}
}
