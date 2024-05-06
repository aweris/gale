package expression

import "github.com/rhysd/actionlint"

func getInterpreterFromNode(node actionlint.ExprNode) Interpreter {
	switch n := node.(type) {
	case *actionlint.NullNode: // literal: null
		return NullNode(*n)
	case *actionlint.BoolNode: // literal: boolean
		return BoolNode(*n)
	case *actionlint.IntNode: // literal: integer
		return IntNode(*n)
	case *actionlint.FloatNode: // literal: float
		return FloatNode(*n)
	case *actionlint.StringNode: // literal: string
		return StringNode(*n)
	case *actionlint.VariableNode: // variable access
		return VariableNode(*n)
	case *actionlint.ObjectDerefNode: // property dereference of object like 'foo.bar'
		return ObjectDerefNode(*n)
	case *actionlint.ArrayDerefNode:
		return ArrayDerefNode(*n)
	case *actionlint.IndexAccessNode:
		return IndexAccessNode(*n)
	case *actionlint.NotOpNode:
		return NotOpNode(*n)
	case *actionlint.CompareOpNode:
		return CompareOpNode(*n)
	case *actionlint.LogicalOpNode:
		return LogicalOpNode(*n)
	case *actionlint.FuncCallNode:
		return FuncCallNode(*n)
	default:
		panic("unknown node type")
	}
}
