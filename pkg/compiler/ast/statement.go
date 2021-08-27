package ast

import (
	"github.com/shoriwe/gplasma/pkg/compiler/lexer"
	"github.com/shoriwe/gplasma/pkg/errors"
	"github.com/shoriwe/gplasma/pkg/vm"
	"reflect"
)

type Statement interface {
	S()
	Node
}

func compileClassBody(body []Node) ([]vm.Code, *errors.Error) {
	foundInitialize := false
	var isInitialize bool
	var nodeCode []vm.Code
	var compilationError *errors.Error
	var result []vm.Code
	for _, node := range body {
		switch node.(type) {
		case IExpression:
			nodeCode, compilationError = node.(IExpression).CompilePush(true)
		case Statement:
			if _, ok := node.(*FunctionDefinitionStatement); ok {
				nodeCode, compilationError, isInitialize = node.(*FunctionDefinitionStatement).CompileAsClassFunction()
				if isInitialize && !foundInitialize {
					foundInitialize = true
				}
			} else {
				nodeCode, compilationError = node.(Statement).Compile()
			}
		}
		if compilationError != nil {
			return nil, compilationError
		}
		result = append(result, nodeCode...)
	}
	if !foundInitialize {
		initFunction := &FunctionDefinitionStatement{
			Name: &Identifier{
				Token: &lexer.Token{
					String: vm.Initialize,
				},
			},
			Arguments: nil,
			Body:      nil,
		}
		nodeCode, _, _ = initFunction.CompileAsClassFunction()
		result = append(result, nodeCode...)
	}
	return result, nil
}

type AssignStatement struct {
	Statement
	LeftHandSide   IExpression // Identifiers or Selectors
	AssignOperator *lexer.Token
	RightHandSide  IExpression
}

func compileAssignStatementMiddleBinaryExpression(leftHandSide IExpression, assignOperator *lexer.Token) ([]vm.Code, *errors.Error) {
	result, leftHandSideCompilationError := leftHandSide.CompilePush(true)
	if leftHandSideCompilationError != nil {
		return nil, leftHandSideCompilationError
	}
	// Finally decide the instruction to use
	var operation uint8
	switch assignOperator.DirectValue {
	case lexer.AddAssign:
		operation = vm.AddOP
	case lexer.SubAssign:
		operation = vm.SubOP
	case lexer.StarAssign:
		operation = vm.MulOP
	case lexer.DivAssign:
		operation = vm.DivOP
	case lexer.ModulusAssign:
		operation = vm.ModOP
	case lexer.PowerOfAssign:
		operation = vm.PowOP
	case lexer.BitwiseXorAssign:
		operation = vm.BitXorOP
	case lexer.BitWiseAndAssign:
		operation = vm.BitAndOP
	case lexer.BitwiseOrAssign:
		operation = vm.BitOrOP
	case lexer.BitwiseLeftAssign:
		operation = vm.BitLeftOP
	case lexer.BitwiseRightAssign:
		operation = vm.BitRightOP
	default:
		panic(errors.NewUnknownVMOperationError(operation))
	}
	return append(result, vm.NewCode(operation, assignOperator.Line, nil)), nil
}

func compileIdentifierAssign(identifier *Identifier) ([]vm.Code, *errors.Error) {
	return []vm.Code{vm.NewCode(vm.AssignIdentifierOP, identifier.Token.Line, identifier.Token.String)}, nil
}

func compileSelectorAssign(selectorExpression *SelectorExpression) ([]vm.Code, *errors.Error) {
	result, sourceCompilationError := selectorExpression.X.CompilePush(true)
	if sourceCompilationError != nil {
		return nil, sourceCompilationError
	}
	return append(result, vm.NewCode(vm.AssignSelectorOP, selectorExpression.Identifier.Token.Line, selectorExpression.Identifier.Token.String)), nil
}

func compileIndexAssign(indexExpression *IndexExpression) ([]vm.Code, *errors.Error) {
	result, sourceCompilationError := indexExpression.Source.CompilePush(true)
	if sourceCompilationError != nil {
		return nil, sourceCompilationError
	}
	index, indexCompilationError := indexExpression.Index.CompilePush(true)
	if indexCompilationError != nil {
		return nil, indexCompilationError
	}
	result = append(result, index...)
	return append(result, vm.NewCode(vm.AssignIndexOP, errors.UnknownLine, nil)), nil
}

func (assignStatement *AssignStatement) Compile() ([]vm.Code, *errors.Error) {
	result, valueCompilationError := assignStatement.RightHandSide.CompilePush(true)
	if valueCompilationError != nil {
		return nil, valueCompilationError
	}
	if assignStatement.AssignOperator.DirectValue != lexer.Assign {
		// Do something here to evaluate the operation
		assignOperation, middleOperationCompilationError := compileAssignStatementMiddleBinaryExpression(assignStatement.LeftHandSide, assignStatement.AssignOperator)
		if middleOperationCompilationError != nil {
			return nil, middleOperationCompilationError
		}
		result = append(result, assignOperation...)
		result = append(result, vm.NewCode(vm.PushOP, errors.UnknownLine, nil))
	}
	var leftHandSide []vm.Code
	var leftHandSideCompilationError *errors.Error
	switch assignStatement.LeftHandSide.(type) {
	case *Identifier:
		leftHandSide, leftHandSideCompilationError = compileIdentifierAssign(assignStatement.LeftHandSide.(*Identifier))
	case *SelectorExpression:
		leftHandSide, leftHandSideCompilationError = compileSelectorAssign(assignStatement.LeftHandSide.(*SelectorExpression))
	case *IndexExpression:
		leftHandSide, leftHandSideCompilationError = compileIndexAssign(assignStatement.LeftHandSide.(*IndexExpression))
	default:
		panic(reflect.TypeOf(assignStatement.LeftHandSide))
	}
	if leftHandSideCompilationError != nil {
		return nil, leftHandSideCompilationError
	}
	return append(result, leftHandSide...), nil
}

type DoWhileStatement struct {
	Statement
	Condition IExpression
	Body      []Node
}

func (doWhileStatement *DoWhileStatement) Compile() ([]vm.Code, *errors.Error) {
	condition, conditionCompilationError := doWhileStatement.Condition.CompilePush(true)
	if conditionCompilationError != nil {
		return nil, conditionCompilationError
	}
	conditionLength := len(condition)
	body, bodyCompilationError := compileBody(doWhileStatement.Body)
	if bodyCompilationError != nil {
		return nil, bodyCompilationError
	}
	bodyLength := len(body)
	for index, instruction := range body {
		if instruction.Instruction.OpCode == vm.BreakOP && instruction.Value == nil {
			body[index].Value = (bodyLength - index) + conditionLength
		} else if instruction.Instruction.OpCode == vm.ContinueOP && instruction.Value == nil {
			body[index].Value = (bodyLength - index) - 1
		} else if instruction.Instruction.OpCode == vm.RedoOP && instruction.Value == nil {
			body[index].Value = -(index + 1)
		}
	}
	result := body
	result = append(result, condition...)
	result = append(result,
		vm.NewCode(vm.UnlessJumpOP, errors.UnknownLine, -(bodyLength+conditionLength+1)),
	)
	return result, nil
}

type WhileLoopStatement struct {
	Statement
	Condition IExpression
	Body      []Node
}

func (whileStatement *WhileLoopStatement) Compile() ([]vm.Code, *errors.Error) {
	condition, conditionCompilationError := whileStatement.Condition.CompilePush(true)
	if conditionCompilationError != nil {
		return nil, conditionCompilationError
	}
	conditionLength := len(condition)
	body, bodyCompilationError := compileBody(whileStatement.Body)
	if bodyCompilationError != nil {
		return nil, bodyCompilationError
	}
	bodyLength := len(body)
	for index, instruction := range body {
		if instruction.Instruction.OpCode == vm.BreakOP && instruction.Value == nil {
			body[index].Value = bodyLength - index
		} else if instruction.Instruction.OpCode == vm.ContinueOP && instruction.Value == nil {
			body[index].Value = -(conditionLength + index + 2)
		} else if instruction.Instruction.OpCode == vm.RedoOP && instruction.Value == nil {
			body[index].Value = -(index + 1)
		}
	}
	result := condition
	result = append(result, vm.NewCode(vm.IfJumpOP, errors.UnknownLine, bodyLength+1))
	result = append(result, body...)
	result = append(result,
		vm.NewCode(vm.ContinueOP, errors.UnknownLine,
			-(conditionLength+1+bodyLength+1),
		),
	)
	return result, nil
}

type UntilLoopStatement struct {
	Statement
	Condition IExpression
	Body      []Node
}

func (untilLoop *UntilLoopStatement) Compile() ([]vm.Code, *errors.Error) {
	condition, conditionCompilationError := untilLoop.Condition.CompilePush(true)
	conditionLength := len(condition)
	if conditionCompilationError != nil {
		return nil, conditionCompilationError
	}
	body, bodyCompilationError := compileBody(untilLoop.Body)
	if bodyCompilationError != nil {
		return nil, bodyCompilationError
	}
	bodyLength := len(body)
	for index, instruction := range body {
		if instruction.Instruction.OpCode == vm.BreakOP && instruction.Value == nil {
			body[index].Value = bodyLength - index
		} else if instruction.Instruction.OpCode == vm.ContinueOP && instruction.Value == nil {
			body[index].Value = -(conditionLength + index + 2)
		} else if instruction.Instruction.OpCode == vm.RedoOP && instruction.Value == nil {
			body[index].Value = -(index + 1)
		}
	}
	result := condition
	result = append(result, vm.NewCode(vm.UnlessJumpOP, errors.UnknownLine, bodyLength+1))
	result = append(result, body...)
	result = append(result,
		vm.NewCode(vm.ContinueOP, errors.UnknownLine,
			-(conditionLength+1+bodyLength+1),
		),
	)
	return result, nil
}

type ForLoopStatement struct {
	Statement
	Receivers []*Identifier
	Source    IExpression
	Body      []Node
}

func (forStatement *ForLoopStatement) Compile() ([]vm.Code, *errors.Error) {
	var result []vm.Code
	source, compilationError := forStatement.Source.CompilePush(true)
	if compilationError != nil {
		return nil, compilationError
	}
	result = append(result, source...)
	body, bodyCompilationError := compileBody(forStatement.Body)
	if bodyCompilationError != nil {
		return nil, bodyCompilationError
	}
	result = append(result, body...)
	var receivers []string
	for _, receiver := range forStatement.Receivers {
		receivers = append(
			receivers,
			receiver.Token.String,
		)
	}
	result = append(
		result,
		vm.NewCode(
			vm.ForLoopOP,
			errors.UnknownLine,
			receivers,
		),
	)
	return result, nil
}

type ElifBlock struct {
	Condition IExpression
	Body      []Node
}

type ElifInformation struct {
	Condition       []vm.Code
	ConditionLength int
	Body            []vm.Code
	BodyLength      int
}

type IfStatement struct {
	Statement
	Condition  IExpression
	Body       []Node
	ElifBlocks []*ElifBlock
	Else       []Node
}

func (ifStatement *IfStatement) Compile() ([]vm.Code, *errors.Error) {
	panic(1)
}

type UnlessStatement struct {
	Statement
	Condition  IExpression
	Body       []Node
	ElifBlocks []*ElifBlock
	Else       []Node
}

func (unlessStatement *UnlessStatement) Compile() ([]vm.Code, *errors.Error) {
	panic(1)
}

type CaseBlock struct {
	Cases []IExpression
	Body  []Node
}

type SwitchStatement struct {
	Statement
	Target     IExpression
	CaseBlocks []*CaseBlock
	Default    []Node
}

func (switchStatement *SwitchStatement) Compile() ([]vm.Code, *errors.Error) {
	panic(1)
}

type ModuleStatement struct {
	Statement
	Name *Identifier
	Body []Node
}

func (moduleStatement *ModuleStatement) Compile() ([]vm.Code, *errors.Error) {
	panic(1)
}

type FunctionDefinitionStatement struct {
	Statement
	Name      *Identifier
	Arguments []*Identifier
	Body      []Node
}

func (functionDefinition *FunctionDefinitionStatement) Compile() ([]vm.Code, *errors.Error) {
	functionCode, functionDefinitionBodyCompilationError := compileBody(functionDefinition.Body)
	if functionDefinitionBodyCompilationError != nil {
		return nil, functionDefinitionBodyCompilationError
	}
	var result []vm.Code
	result = append(result, vm.NewCode(vm.NewFunctionOP, errors.UnknownLine, [2]int{len(functionCode) + 2, len(functionDefinition.Arguments)}))
	var arguments []string
	for _, argument := range functionDefinition.Arguments {
		arguments = append(arguments, argument.Token.String)
	}
	result = append(result, vm.NewCode(vm.LoadFunctionArgumentsOP, errors.UnknownLine, arguments))
	result = append(result, functionCode...)
	result = append(result, vm.NewCode(vm.ReturnOP, errors.UnknownLine, 0))
	return append(result, vm.NewCode(vm.AssignIdentifierOP, functionDefinition.Name.Token.Line, functionDefinition.Name.Token.String)), nil
}

func (functionDefinition *FunctionDefinitionStatement) CompileAsClassFunction() ([]vm.Code, *errors.Error, bool) {
	functionCode, functionDefinitionBodyCompilationError := compileBody(functionDefinition.Body)
	if functionDefinitionBodyCompilationError != nil {
		return nil, functionDefinitionBodyCompilationError, false
	}
	var result []vm.Code
	result = append(result, vm.NewCode(vm.NewClassFunctionOP, errors.UnknownLine, [2]int{len(functionCode) + 2, len(functionDefinition.Arguments)}))
	var arguments []string
	for _, argument := range functionDefinition.Arguments {
		arguments = append(arguments, argument.Token.String)
	}
	result = append(result, vm.NewCode(vm.LoadFunctionArgumentsOP, errors.UnknownLine, arguments))
	result = append(result, functionCode...)
	result = append(result, vm.NewCode(vm.ReturnOP, errors.UnknownLine, 0))
	result = append(result, vm.NewCode(vm.AssignIdentifierOP, functionDefinition.Name.Token.Line, functionDefinition.Name.Token.String))
	return result, nil, functionDefinition.Name.Token.String == vm.Initialize
}

type InterfaceStatement struct {
	Statement
	Name              *Identifier
	Bases             []IExpression
	MethodDefinitions []*FunctionDefinitionStatement
}

func (interfaceStatement *InterfaceStatement) Compile() ([]vm.Code, *errors.Error) {
	panic(1)
}

type ClassStatement struct {
	Statement
	Name  *Identifier
	Bases []IExpression // Identifiers and selectors
	Body  []Node
}

func (classStatement *ClassStatement) Compile() ([]vm.Code, *errors.Error) {
	panic(1)
}

type ExceptBlock struct {
	Targets     []IExpression
	CaptureName *Identifier
	Body        []Node
}

type RaiseStatement struct {
	Statement
	X IExpression
}

func (raise *RaiseStatement) Compile() ([]vm.Code, *errors.Error) {
	result, expressionCompilationError := raise.X.CompilePush(true)
	if expressionCompilationError != nil {
		return nil, expressionCompilationError
	}
	result = append(result, vm.NewCode(vm.RaiseOP, errors.UnknownLine, nil))
	return result, nil
}

type exceptBlock struct {
	Targets       []vm.Code
	TargetsLength int
	Receiver      string
	Body          []vm.Code
	BodyLength    int
}

type TryStatement struct {
	Statement
	Body         []Node
	ExceptBlocks []*ExceptBlock
	Else         []Node
	Finally      []Node
}

func (tryStatement *TryStatement) compileTryStatement() ([]vm.Code, *errors.Error) {
	panic("IMPLEMENT ME!!!")
}

type BeginStatement struct {
	Statement
	Body []Node
}

type EndStatement struct {
	Statement
	Body []Node
}

type ReturnStatement struct {
	Statement
	Results []IExpression
}

func (returnStatement *ReturnStatement) Compile() ([]vm.Code, *errors.Error) {
	numberOfResults := len(returnStatement.Results)
	var result []vm.Code
	for i := numberOfResults - 1; i > -1; i-- {
		returnResult, resultCompilationError := returnStatement.Results[i].CompilePush(true)
		if resultCompilationError != nil {
			return nil, resultCompilationError
		}
		result = append(result, returnResult...)
	}
	return append(result, vm.NewCode(vm.ReturnOP, errors.UnknownLine, numberOfResults)), nil
}

type YieldStatement struct {
	Statement
	Results []IExpression
}

type SuperInvocationStatement struct {
	Statement
	Arguments []IExpression
}

type ContinueStatement struct {
	Statement
}

type BreakStatement struct {
	Statement
}

func (_ *BreakStatement) Compile() ([]vm.Code, *errors.Error) {
	return []vm.Code{vm.NewCode(vm.BreakOP, errors.UnknownLine, nil)}, nil
}

type RedoStatement struct {
	Statement
}

func (_ *RedoStatement) Compile() ([]vm.Code, *errors.Error) {
	return []vm.Code{vm.NewCode(vm.RedoOP, errors.UnknownLine, nil)}, nil
}

type PassStatement struct {
	Statement
}

func (_ *PassStatement) compilePassStatement() ([]vm.Code, *errors.Error) {
	return []vm.Code{vm.NewCode(vm.NOP, errors.UnknownLine, nil)}, nil
}
