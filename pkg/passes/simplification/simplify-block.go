package simplification

import (
	"github.com/shoriwe/gplasma/pkg/ast"
	"github.com/shoriwe/gplasma/pkg/ast2"
)

func simplifyBlock(block *ast.BlockStatement) *ast2.Block {
	body := make([]ast2.Node, 0, len(block.Body))
	for _, node := range block.Body {
		body = append(body, simplifyNode(node))
	}
	return &ast2.Block{
		Body: body,
	}
}
