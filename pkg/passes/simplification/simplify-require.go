package simplification

import (
	"github.com/shoriwe/gplasma/pkg/ast"
	"github.com/shoriwe/gplasma/pkg/ast2"
)

func simplifyRequire(require *ast.RequireStatement) *ast2.Require {
	return &ast2.Require{
		X: simplifyExpression(require.X),
	}
}
