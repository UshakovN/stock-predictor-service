package storage

type CompareTokenizer interface {
	Token() string
}

type (
	EqTokenizer  struct{}
	GtTokenizer  struct{}
	GteTokenizer struct{}
	LtTokenizer  struct{}
	LteTokenizer struct{}
)

func (EqTokenizer) Token() string {
	return "="
}

func (GtTokenizer) Token() string {
	return ">"
}

func (GteTokenizer) Token() string {
	return ">="
}

func (LtTokenizer) Token() string {
	return "<"
}

func (LteTokenizer) Token() string {
	return "<="
}
