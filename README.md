# Go Oso Hierarchy Test

This is a small program designed to test building recursive team hierarchies using Oso.

## Run

```bash
make run
```

## Problem

Currently, running the program results in the following error:

```go
2023/01/27 22:19:45 User cannot read subTeam: Error: types.ErrorKindRuntime{RuntimeErrorVariant:types.RuntimeErrorUnhandledPartial{Var:"_parent_487", Term:types.Term{Value:types.Value{ValueVariant:types.ValueExpression{Operator:types.Operator{OperatorVariant:types.OperatorAnd{}}, Args:[]types.Term{types.Term{Value:types.Value{ValueVariant:types.ValueExpression{Operator:types.Operator{OperatorVariant:types.OperatorIsa{}}, Args:[]types.Term{types.Term{Value:types.Value{ValueVariant:"_parent_487"}}, types.Term{Value:types.Value{ValueVariant:types.ValuePattern{PatternVariant:types.PatternInstance{Tag:"Team", Fields:types.Dictionary{Fields:map[types.Symbol]types.Term{}}}}}}}}}}, types.Term{Value:types.Value{ValueVariant:types.ValueExpression{Operator:types.Operator{OperatorVariant:types.OperatorUnify{}}, Args:[]types.Term{types.Term{Value:types.Value{ValueVariant:types.ValueNumber{NumericVariant:1}}}, types.Term{Value:types.Value{ValueVariant:types.ValueExpression{Operator:types.Operator{OperatorVariant:types.OperatorDot{}}, Args:[]types.Term{types.Term{Value:types.Value{ValueVariant:"_parent_487"}}, types.Term{Value:types.Value{ValueVariant:"ID"}}}}}}}}}}}}}}}}
Found an unhandled partial in the query result: _parent_487

This can happen when there is a variable used inside a rule
which is not related to any of the query inputs.

For example: f(_x) if y.a = 1 and y.b = 2;

In this example, the variable `y` is constrained by `a = 1 and b = 2`,
but we cannot resolve these constraints without further information.

The unhandled partial is for variable _parent_487.
The expression is: _parent_487 matches Team{} and 1 = _parent_487.ID
 at line 28, column 14 of file main.polar:
        028: has_relation(parent: Team, "parent", team: Team) if
                          ^
```
