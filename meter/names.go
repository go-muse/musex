package meter

// The names below come only from the reference table of the package. Where
// the table has no name, the name methods return ok == false without an
// error: "no name" is a result, not a failure. The identifiers of the
// constants are Go names, not names of the values.

// nameResult returns (name, name != "", nil) for a valid value and an
// ErrInvalidArgument error for an invalid one.
func nameResult(valid bool, op, name string) (string, bool, error) {
	if !valid {
		return "", false, invalidError(op, "invalid value")
	}
	return name, name != "", nil
}

// Name returns the English name of s: "common time" for C and "cut time" for
// ¢.
//
// Errors: ErrInvalidArgument for an invalid value.
func (s Symbol) Name() (name string, ok bool, err error) {
	var n string
	switch s {
	case SymbolCommon:
		n = "common time"
	case SymbolCut:
		n = "cut time"
	}
	return nameResult(s.IsValid(), "Symbol.Name", n)
}

// MustName is like Name but panics with the error of Name.
func (s Symbol) MustName() (name string, ok bool) {
	return must2(s.Name())
}

// RussianName returns the Russian name of s: "alla breve" for ¢. C has no
// Russian name: the result is ("", false, nil).
//
// Errors: ErrInvalidArgument for an invalid value.
func (s Symbol) RussianName() (name string, ok bool, err error) {
	var n string
	if s == SymbolCut {
		n = "alla breve"
	}
	return nameResult(s.IsValid(), "Symbol.RussianName", n)
}

// MustRussianName is like RussianName but panics with the error of
// RussianName.
func (s Symbol) MustRussianName() (name string, ok bool) {
	return must2(s.RussianName())
}

// Name returns the English name of k: "simple", "compound", or
// "asymmetrical". KindUndetermined has no name: the result is ("", false,
// nil). For the full name with the number of beats, such as "simple triple",
// use Meter.KindName.
//
// Errors: ErrInvalidArgument for an invalid value.
func (k Kind) Name() (name string, ok bool, err error) {
	var n string
	switch k {
	case KindSimple:
		n = "simple"
	case KindCompound:
		n = "compound"
	case KindAsymmetrical:
		n = "asymmetrical"
	}
	return nameResult(k.IsValid(), "Kind.Name", n)
}

// MustName is like Name but panics with the error of Name.
func (k Kind) MustName() (name string, ok bool) {
	return must2(k.Name())
}

// RussianName returns the Russian name of k: "простой", "сложный", or
// "асимметричный". KindUndetermined has no name: the result is ("", false,
// nil).
//
// Errors: ErrInvalidArgument for an invalid value.
func (k Kind) RussianName() (name string, ok bool, err error) {
	var n string
	switch k {
	case KindSimple:
		n = "простой"
	case KindCompound:
		n = "сложный"
	case KindAsymmetrical:
		n = "асимметричный"
	}
	return nameResult(k.IsValid(), "Kind.RussianName", n)
}

// MustRussianName is like RussianName but panics with the error of
// RussianName.
func (k Kind) MustRussianName() (name string, ok bool) {
	return must2(k.RussianName())
}

// Name returns the English name of k. No kind of the Russian tradition has an
// English name: for every valid value the result is ("", false, nil).
//
// Errors: ErrInvalidArgument for an invalid value.
func (k RussianKind) Name() (name string, ok bool, err error) {
	return nameResult(k.IsValid(), "RussianKind.Name", "")
}

// MustName is like Name but panics with the error of Name.
func (k RussianKind) MustName() (name string, ok bool) {
	return must2(k.Name())
}

// RussianName returns the Russian name of k: "простой", "сложный
// однородный", or "сложный смешанный". RussianUndetermined has no name: the
// result is ("", false, nil).
//
// Errors: ErrInvalidArgument for an invalid value.
func (k RussianKind) RussianName() (name string, ok bool, err error) {
	var n string
	switch k {
	case RussianSimple:
		n = "простой"
	case RussianCompoundUniform:
		n = "сложный однородный"
	case RussianCompoundMixed:
		n = "сложный смешанный"
	}
	return nameResult(k.IsValid(), "RussianKind.RussianName", n)
}

// MustRussianName is like RussianName but panics with the error of
// RussianName.
func (k RussianKind) MustRussianName() (name string, ok bool) {
	return must2(k.RussianName())
}

// Name returns the English name of w. No metric weight has an English name:
// for every valid value the result is ("", false, nil).
//
// Errors: ErrInvalidArgument for an invalid value.
func (w Weight) Name() (name string, ok bool, err error) {
	return nameResult(w.IsValid(), "Weight.Name", "")
}

// MustName is like Name but panics with the error of Name.
func (w Weight) MustName() (name string, ok bool) {
	return must2(w.Name())
}

// RussianName returns the Russian name of w: "сильная", "относительно
// сильная", or "слабая". WeightOffUnit has no name: the result is ("", false,
// nil).
//
// Errors: ErrInvalidArgument for an invalid value.
func (w Weight) RussianName() (name string, ok bool, err error) {
	var n string
	switch w {
	case WeightStrong:
		n = "сильная"
	case WeightRelativelyStrong:
		n = "относительно сильная"
	case WeightWeak:
		n = "слабая"
	}
	return nameResult(w.IsValid(), "Weight.RussianName", n)
}

// MustRussianName is like RussianName but panics with the error of
// RussianName.
func (w Weight) MustRussianName() (name string, ok bool) {
	return must2(w.RussianName())
}
