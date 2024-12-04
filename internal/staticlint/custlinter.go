package custlinter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "customlinter",
	Doc:  "check for os.Exit in main",
	Run:  custLinter,
}

func custLinter(pass *analysis.Pass) (interface{}, error) {
	// Function for checkin code contain os.Exit in main function.
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Reportf(node.Pos(), "os.Exit in main")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
