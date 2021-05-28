package vm

import "testing"

type Test struct {
	code     []Code
	result   interface{}
	behavior uint
}

func NewTest(code []Code, result interface{}, behavior uint) Test {
	return Test{
		code:     code,
		result:   result,
		behavior: behavior,
	}
}

const (
	stringEquals uint = iota
)

func testIt(t *testing.T, tests []Test) {
	for _, test := range tests {
		vm := NewPlasmaVM()
		vm.Initialize(test.code)
		result, executionError := vm.Execute()
		if executionError != nil {
			t.Error(executionError)
			return
		}
		switch test.behavior {
		case stringEquals:
			toString, getError := result.Get(ToString)
			if getError != nil {
				t.Error(getError)
				return
			}
			if _, ok := toString.(*Function); !ok {
				t.Errorf("ToString is not a function")
				return
			}
			stringResult, callError := CallFunction(toString.(*Function), vm, result.SymbolTable())
			if callError != nil {
				t.Error(callError)
				return
			}
			if _, ok2 := stringResult.(*String); !ok2 {
				t.Errorf("ToString doesn't return a string object")
				return
			}
			if stringResult.(*String).String != test.result.(string) {
				t.Errorf("Expecting: %s but received: %s", test.result.(string), stringResult.(*String).String)
				return
			}
		}
	}
}

var stringCreationSamples = []Test{
	NewTest(
		[]Code{
			NewCode(NewStringOP, 1, "Hello"),
			NewCode(ReturnOP, 1, nil),
		},
		"Hello",
		stringEquals,
	),
	NewTest(
		[]Code{
			NewCode(NewStringOP, 1, "Carro"),
			NewCode(ReturnOP, 1, nil),
		},
		"Carro",
		stringEquals,
	),
	NewTest(
		[]Code{
			NewCode(NewStringOP, 1, "45098430958"),
			NewCode(ReturnOP, 1, nil),
		},
		"45098430958",
		stringEquals,
	),
}

func TestStringCreation(t *testing.T) {
	testIt(t, stringCreationSamples)
}

var stringBuiltInTransformationFunction = []Test{
	NewTest(
		[]Code{
			NewCode(ReturnOP, 1, nil),
			NewCode(NoOP, 1, 1),
			NewCode(NoOP, 1, false),
			NewCode(CallOP, 1, nil),
			NewCode(GetOP, 1, ToString),
			NewCode(NewStringOP, 1, "Hello"),
		},
		"Hello",
		stringEquals,
	),
}

func TestStringBuiltInTransformationFunction(t *testing.T) {
	testIt(t, stringBuiltInTransformationFunction)
}
