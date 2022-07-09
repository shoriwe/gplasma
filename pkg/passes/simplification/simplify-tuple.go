package simplification

import (
	"github.com/shoriwe/gplasma/pkg/ast"
	"github.com/shoriwe/gplasma/pkg/ast2"
)

func simplifyTuple(tuple *ast.TupleExpression) *ast2.Tuple {
	values := make([]ast2.Expression, 0, len(tuple.Values))
	for _, value := range tuple.Values {
		values = append(values, simplifyExpression(value))
	}
	return &ast2.Tuple{
		Values: values,
	}
}
