package pcseq

import "github.com/go-muse/musex/pcset"

// RotationCombinatoriality lists inversion labels complementary to a rotation.
type RotationCombinatoriality struct {
	Rotation int
	Forms    []FormLabel
}

// HexachordAnalysis lists three kinds of hexachord combinatoriality.
type HexachordAnalysis struct {
	Transpositions       []int
	Inversions           []int
	RetrogradeInversions []int
}

// AnalyzeHexachord analyzes a six-class set of modulus 12.
func AnalyzeHexachord(content pcset.Set) (HexachordAnalysis, error) {
	return HexachordAnalysis{}, nil
}

// IsAllCombinatorial reports whether all three lists are nonempty.
func (a HexachordAnalysis) IsAllCombinatorial() bool { return false }

// IsInversionOnly reports whether only inversion combinatoriality is present.
func (a HexachordAnalysis) IsInversionOnly() bool { return false }

// IsTranspositionOnly reports whether only transposition combinatoriality is present.
func (a HexachordAnalysis) IsTranspositionOnly() bool { return false }

// IsRetrogradeInversionOnly reports whether only RI combinatoriality is present.
func (a HexachordAnalysis) IsRetrogradeInversionOnly() bool { return false }
