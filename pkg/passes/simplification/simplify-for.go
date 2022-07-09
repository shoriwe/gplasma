package simplification

import (
	"github.com/shoriwe/gplasma/pkg/ast"
	"github.com/shoriwe/gplasma/pkg/ast2"
)

func simplifyFor(for_ *ast.ForLoopStatement) *ast2.While {
	anonymousIdentifier := nextAnonIdentifier()
	hasNext := &ast2.FunctionCall{
		Function: &ast2.Selector{
			X: simplifyExpression(for_.Source),
			Identifier: &ast2.Identifier{
				Symbol: ast2.HasNextString,
			},
		},
		Arguments: nil,
	}
	next := ast2.Assignment{
		Left: anonymousIdentifier,
		Right: &ast2.FunctionCall{
			Function: &ast2.Selector{
				X: simplifyExpression(for_.Source),
				Identifier: &ast2.Identifier{
					Symbol: ast2.NextString,
				},
			},
			Arguments: nil,
		},
	}
	expand := make([]ast2.Node, 0, len(for_.Receivers))
	for index, receiver := range for_.Receivers {
		expand = append(expand, &ast2.Assignment{
			Left: simplifyIdentifier(receiver),
			Right: &ast2.Index{
				Source: anonymousIdentifier,
				Index: &ast2.Integer{
					Value: int64(index),
				},
			},
		})
	}
	body := make([]ast2.Node, 0, 1+len(expand)+len(for_.Body))
	body = append(body, next)
	body = append(body, expand...)
	for _, node := range for_.Body {
		body = append(body, simplifyNode(node))
	}
	return &ast2.While{
		Condition: hasNext,
		Body:      body,
	}
}