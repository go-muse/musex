package meter

// Form is the form of a time signature record. The zero value and values
// outside the constants are invalid: IsValid reports false for them.
type Form uint8

// Forms of a time signature record.
const (
	// FormSimple is a simple signature, such as 3/4.
	FormSimple Form = iota + 1
	// FormSymbol is a symbol: C or ¢.
	FormSymbol
	// FormAdditive is a compound numerator, such as (3+2+3)/8.
	FormAdditive
	// FormSum is a sum of signatures, such as 2/4+3/8.
	FormSum
)

// IsValid reports whether f is one of the Form constants.
func (f Form) IsValid() bool {
	return f >= FormSimple && f <= FormSum
}

// Equal reports whether f and other are the same value. It is the same as
// f == other, including invalid values.
func (f Form) Equal(other Form) bool {
	return f == other
}

// String returns the technical text of f: "simple", "symbol", "additive", or
// "sum", and "Form(invalid)" for an invalid value. The text is not a name of
// the form.
func (f Form) String() string {
	switch f {
	case FormSimple:
		return "simple"
	case FormSymbol:
		return "symbol"
	case FormAdditive:
		return "additive"
	case FormSum:
		return "sum"
	}
	return "Form(invalid)"
}

// Symbol is a time signature symbol. The zero value and values outside the
// constants are invalid: IsValid reports false for them.
type Symbol uint8

// Time signature symbols.
const (
	// SymbolCommon is C, the structure of 4/4.
	SymbolCommon Symbol = iota + 1
	// SymbolCut is ¢, the structure of 2/2. Its text is "C|".
	SymbolCut
)

// IsValid reports whether s is one of the Symbol constants.
func (s Symbol) IsValid() bool {
	return s == SymbolCommon || s == SymbolCut
}

// Equal reports whether s and other are the same value. It is the same as
// s == other, including invalid values.
func (s Symbol) Equal(other Symbol) bool {
	return s == other
}

// String returns the text of s in the record of a meter: "C" or "C|" (the
// ASCII text of ¢), and "Symbol(invalid)" for an invalid value. The text is
// not a name of the symbol; see Symbol.Name.
func (s Symbol) String() string {
	switch s {
	case SymbolCommon:
		return "C"
	case SymbolCut:
		return "C|"
	}
	return "Symbol(invalid)"
}

// numerator returns the numerator of the structure of a valid symbol: 4 for
// C and 2 for ¢. It equals the denominator.
func (s Symbol) numerator() int {
	if s == SymbolCut {
		return 2
	}
	return 4
}

// Kind is the kind of a meter in the English-language tradition: simple,
// compound, asymmetrical, or undetermined. The zero value and values outside
// the constants are invalid: IsValid reports false for them.
type Kind uint8

// Kinds of a meter in the English-language tradition.
const (
	// KindSimple is a simple meter: each beat divides into two parts.
	KindSimple Kind = iota + 1
	// KindCompound is a compound meter: each beat divides into three parts.
	KindCompound
	// KindAsymmetrical is an asymmetrical meter: its beats are unequal.
	KindAsymmetrical
	// KindUndetermined is the result when no rule of the tradition applies,
	// such as for 1/4. It is a result, not an error.
	KindUndetermined
)

// IsValid reports whether k is one of the Kind constants.
func (k Kind) IsValid() bool {
	return k >= KindSimple && k <= KindUndetermined
}

// Equal reports whether k and other are the same value. It is the same as
// k == other, including invalid values.
func (k Kind) Equal(other Kind) bool {
	return k == other
}

// String returns the technical text of k: "simple", "compound",
// "asymmetrical", or "undetermined", and "Kind(invalid)" for an invalid
// value. The text is not a name of the kind; see Kind.Name.
func (k Kind) String() string {
	switch k {
	case KindSimple:
		return "simple"
	case KindCompound:
		return "compound"
	case KindAsymmetrical:
		return "asymmetrical"
	case KindUndetermined:
		return "undetermined"
	}
	return "Kind(invalid)"
}

// RussianKind is the kind of a meter in the Russian tradition: simple,
// compound uniform, compound mixed, or undetermined. It is independent of
// [Kind]: the two traditions classify meters by different rules. The zero
// value and values outside the constants are invalid: IsValid reports false
// for them.
type RussianKind uint8

// Kinds of a meter in the Russian tradition.
const (
	// RussianSimple is a simple meter: one group of two or three units.
	RussianSimple RussianKind = iota + 1
	// RussianCompoundUniform is a compound uniform meter: several equal
	// groups of two or three units.
	RussianCompoundUniform
	// RussianCompoundMixed is a compound mixed meter: unequal groups of two
	// or three units.
	RussianCompoundMixed
	// RussianUndetermined is the result when a group has neither two nor
	// three units, such as for 1/4. It is a result, not an error.
	RussianUndetermined
)

// IsValid reports whether k is one of the RussianKind constants.
func (k RussianKind) IsValid() bool {
	return k >= RussianSimple && k <= RussianUndetermined
}

// Equal reports whether k and other are the same value. It is the same as
// k == other, including invalid values.
func (k RussianKind) Equal(other RussianKind) bool {
	return k == other
}

// String returns the technical text of k: "simple", "compound-uniform",
// "compound-mixed", or "undetermined", and "RussianKind(invalid)" for an
// invalid value. The text is not a name of the kind; see
// RussianKind.RussianName.
func (k RussianKind) String() string {
	switch k {
	case RussianSimple:
		return "simple"
	case RussianCompoundUniform:
		return "compound-uniform"
	case RussianCompoundMixed:
		return "compound-mixed"
	case RussianUndetermined:
		return "undetermined"
	}
	return "RussianKind(invalid)"
}

// Weight is the metric weight of a point of a measure: strong, relatively
// strong, weak, or off-unit. The zero value and values outside the constants
// are invalid: IsValid reports false for them.
type Weight uint8

// Metric weights.
const (
	// WeightStrong is the weight of the first unit of the first group.
	WeightStrong Weight = iota + 1
	// WeightRelativelyStrong is the weight of the first unit of each other
	// group.
	WeightRelativelyStrong
	// WeightWeak is the weight of every other unit.
	WeightWeak
	// WeightOffUnit is the result for a point that is not the start of a
	// unit. It is a result, not an error.
	WeightOffUnit
)

// IsValid reports whether w is one of the Weight constants.
func (w Weight) IsValid() bool {
	return w >= WeightStrong && w <= WeightOffUnit
}

// Equal reports whether w and other are the same value. It is the same as
// w == other, including invalid values.
func (w Weight) Equal(other Weight) bool {
	return w == other
}

// String returns the technical text of w: "strong", "relatively-strong",
// "weak", or "off-unit", and "Weight(invalid)" for an invalid value. The text
// is not a name of the weight; see Weight.RussianName.
func (w Weight) String() string {
	switch w {
	case WeightStrong:
		return "strong"
	case WeightRelativelyStrong:
		return "relatively-strong"
	case WeightWeak:
		return "weak"
	case WeightOffUnit:
		return "off-unit"
	}
	return "Weight(invalid)"
}
