package vm

type Type struct {
	*Object
	Constructor Constructor
	Name        string
}

func (t *Type) Implements(class *Type) bool {
	if t == class {
		return true
	}
	for _, subClass := range t.subClasses {
		if subClass.Implements(class) {
			return true
		}
	}
	return false
}

func (p *Plasma) NewType(typeName string, parent *SymbolTable, subclasses []*Type, constructor Constructor) *Type {
	result := &Type{
		Object:      p.NewObject(TypeName, subclasses, parent),
		Constructor: constructor,
		Name:        typeName,
	}
	result.Set(ToString,
		p.NewFunction(result.symbols,
			NewBuiltInClassFunction(result, 0,
				func(_ IObject, _ ...IObject) (IObject, *Object) {
					return p.NewString(p.PeekSymbolTable(), "Type@"+typeName), nil
				},
			),
		),
	)
	return result
}