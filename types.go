package tiger

type SemantType interface {
	IsSemantType()
}

type RecordSemantType struct {
	symbols []Symbol
	ty      []SemantType
}
func (t *RecordSemantType) IsSemantType() {}

type NilSemantType struct{}
func (t *NilSemantType) IsSemantType() {}

type UnitSemantType struct{}
func (t *UnitSemantType) IsSemantType() {}

type IntSemantType struct{}
func (t *IntSemantType) IsSemantType() {}

type StringSemantType struct{}
func (t *StringSemantType) IsSemantType() {}

type ArraySemantType struct {
	ty SemantType
}
func (t *ArraySemantType) IsSemantType() {}

type NameSemantType struct {
	ty SemantType
	name Symbol
}
func (t *NameSemantType) IsSemantType() {}
