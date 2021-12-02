package tiger

type Level interface {
	IsLevel()
}

type TopLevel struct{}

func (t TopLevel) IsLevel() {}

type ChildLevel struct {
	parent Level
	frame  *MipsFrame
}

func (t ChildLevel) IsLevel() {}

type TranslateAccess struct {
	level  *Level
	access *Access
}

type TranslateExp interface {
	IsTranslateExp()
}

type Ex struct {
	ex TreeExp
}

func (ex *Ex) IsTranslateExp() {}

type Nx struct {
	stm TreeStm
}

func (ex *Ex) IsTranslateStm() {}

type Cx struct {
	cx func(Label, Label) TreeStm
}

var (
	outermost = TopLevel{}
	fragments []*Frag
	errExp = Ex{ex: Const}
)

func resetFragments () {
	fragments = nil
}

func newLevel(parent Level, name Label, formals []bool) Level {
	return ChildLevel{
		parent: parent,
		frame:  &MipsFrame{
			name:    name,
			formals: nil,
			locals:  0,
			instrs:  nil,
		},
	}
}
