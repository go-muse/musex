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

func firstHalfContent(row Row) (pcset.Set, error) {
	half, err := row.Sequence().Segment(0, row.modulus/2)
	if err != nil {
		return pcset.Set{}, err
	}
	return half.Content()
}

func combinatorialContent(a, b Row) (bool, error) {
	if a.modulus%2 != 0 {
		return false, nil
	}
	aContent, err := firstHalfContent(a)
	if err != nil {
		return false, err
	}
	bContent, err := firstHalfContent(b)
	if err != nil {
		return false, err
	}
	return aContent.Equal(bContent.Complement()), nil
}

// IsCombinatorialPair reports whether two forms have complementary first-half content.
func (r Row) IsCombinatorialPair(a, b FormLabel) (bool, error) {
	if !r.IsValid() {
		return false, invalidValue()
	}
	aForm, err := r.Form(a)
	if err != nil {
		return false, err
	}
	bForm, err := r.Form(b)
	if err != nil {
		return false, err
	}
	return combinatorialContent(aForm, bForm)
}

// CombinatorialForms returns combinatorial P, I, and RI labels.
func (r Row) CombinatorialForms() ([]FormLabel, error) {
	if !r.IsValid() {
		return nil, invalidValue()
	}
	if err := checkModulusLimit(r.modulus, MaxAnalysisModulus); err != nil {
		return nil, err
	}
	forms := make([]FormLabel, 0)
	if r.modulus%2 != 0 {
		return forms, nil
	}
	for _, kind := range []FormKind{P, I, RI} {
		for n := range r.modulus {
			label := FormLabel{kind: kind, index: n, modulus: r.modulus}
			candidate := formUnchecked(r, label)
			matches := must(combinatorialContent(r, candidate))
			if matches {
				forms = append(forms, label)
			}
		}
	}
	return forms, nil
}

// CombinatorialRotations returns inversion-combinatorial rotated forms.
func (r Row) CombinatorialRotations() ([]RotationCombinatoriality, error) {
	if !r.IsValid() {
		return nil, invalidValue()
	}
	if err := checkModulusLimit(r.modulus, MaxAnalysisModulus); err != nil {
		return nil, err
	}
	result := make([]RotationCombinatoriality, 0)
	if r.modulus%2 != 0 {
		return result, nil
	}
	for rotation := range r.modulus {
		rotated := r.MustRotate(rotation)
		forms := make([]FormLabel, 0)
		for n := range r.modulus {
			label := FormLabel{kind: I, index: n, modulus: r.modulus}
			candidate := formUnchecked(rotated, label)
			matches := must(combinatorialContent(rotated, candidate))
			if matches {
				forms = append(forms, label)
			}
		}
		if len(forms) > 0 {
			result = append(result, RotationCombinatoriality{Rotation: rotation, Forms: forms})
		}
	}
	return result, nil
}

// AnalyzeHexachord analyzes a six-class set of modulus 12.
func AnalyzeHexachord(content pcset.Set) (HexachordAnalysis, error) {
	if !content.IsValid() {
		return HexachordAnalysis{}, invalidValue()
	}
	if content.Modulus() != 12 {
		return HexachordAnalysis{}, ErrNotModulus12
	}
	if content.Cardinality() != 6 {
		return HexachordAnalysis{}, invalidSegmentContent()
	}
	result := HexachordAnalysis{
		Transpositions:       make([]int, 0),
		Inversions:           make([]int, 0),
		RetrogradeInversions: make([]int, 0),
	}
	complement := content.Complement()
	for n := range 12 {
		if content.Transpose(n).Equal(complement) {
			result.Transpositions = append(result.Transpositions, n)
		}
		if content.Invert(n).Equal(complement) {
			result.Inversions = append(result.Inversions, n)
		}
		if complement.Invert(n).Equal(complement) {
			result.RetrogradeInversions = append(result.RetrogradeInversions, n)
		}
	}
	return result, nil
}

// HexachordAnalysis analyzes the first-half hexachord of a twelve-tone row.
func (r Row) HexachordAnalysis() (HexachordAnalysis, error) {
	if !r.IsValid() {
		return HexachordAnalysis{}, invalidValue()
	}
	return AnalyzeHexachord(must(firstHalfContent(r)))
}

// IsAllCombinatorial reports whether all three lists are nonempty.
func (a HexachordAnalysis) IsAllCombinatorial() bool {
	return len(a.Transpositions) > 0 && len(a.Inversions) > 0 && len(a.RetrogradeInversions) > 0
}

// IsInversionOnly reports whether only inversion combinatoriality is present.
func (a HexachordAnalysis) IsInversionOnly() bool {
	return len(a.Transpositions) == 0 && len(a.Inversions) > 0 && len(a.RetrogradeInversions) == 0
}

// IsTranspositionOnly reports whether only transposition combinatoriality is present.
func (a HexachordAnalysis) IsTranspositionOnly() bool {
	return len(a.Transpositions) > 0 && len(a.Inversions) == 0 && len(a.RetrogradeInversions) == 0
}

// IsRetrogradeInversionOnly reports whether only RI combinatoriality is present.
func (a HexachordAnalysis) IsRetrogradeInversionOnly() bool {
	return len(a.Transpositions) == 0 && len(a.Inversions) == 0 && len(a.RetrogradeInversions) > 0
}
