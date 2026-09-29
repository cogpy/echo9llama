package model

const (
	TokenTypeNormal = iota + 1
	TokenTypeUnknown
	TokenTypeControl
	TokenTypeUserDefined
	TokenTypeUnused
	TokenTypeByte
)

type TextProcessor interface {
	Encode(s string, addSpecial bool) ([]int32, error)
	Decode([]int32) (string, error)
	Is(int32, Special) bool
	Vocabulary() *Vocabulary
}
