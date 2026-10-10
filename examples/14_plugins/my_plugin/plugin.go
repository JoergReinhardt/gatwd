package my_plugin

import (
	"fmt"
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/plugins"
)

func init() {
	plugins.Register("fetch_aws_ec2", func() evaluator.Adapter[functions.Either[error, any]] {
		inner := func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
			return func(yield func(dsltypes.ASTNode[any]) bool) {
				res := fmt.Sprintf("EC2_Instance_Mock_%v", node.Value)
				yield(dsltypes.NewASTNode[any](res, node.DSLType))
			}
		}
		return evaluator.EitherT(inner)
	})
}
