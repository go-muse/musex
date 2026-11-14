package meter

// unsetText is the string of the unset meter.
const unsetText = "Meter(unset)"

// String returns the canonical text of m without spaces, such as "3/4",
// "5/8[2+3]", "C", "C|", "C[3+1]", "(3+2+3)/8", or "2/4+3/8", and
// "Meter(unset)" for the unset meter. Parse reads every text that String
// returns for a set meter.
func (m Meter) String() string {
	if !m.IsSet() {
		return unsetText
	}
	return m.rec
}

// MarshalText returns the canonical text of m, as String does.
//
// Errors: ErrUnset for the unset meter; an unset field of a struct is not
// encoded (use the omitzero tag for an optional field).
func (m Meter) MarshalText() ([]byte, error) {
	if !m.IsSet() {
		return nil, unsetError("Meter.MarshalText", "receiver")
	}
	return []byte(m.rec), nil
}

// AppendText appends the canonical text of m to dst and returns the extended
// slice. It does not change dst[:len(dst)].
//
// Errors: ErrUnset for the unset meter; dst is then returned unchanged.
func (m Meter) AppendText(dst []byte) ([]byte, error) {
	if !m.IsSet() {
		return dst, unsetError("Meter.AppendText", "receiver")
	}
	return append(dst, m.rec...), nil
}

// UnmarshalText sets *m to the meter of text, as Parse does. It does not
// keep or change text. On an error *m is unchanged.
//
// Errors: ErrInvalidArgument for a nil receiver; otherwise the errors of
// Parse.
func (m *Meter) UnmarshalText(text []byte) error {
	if m == nil {
		return invalidError("Meter.UnmarshalText", "nil receiver")
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*m = v
	return nil
}

// Parse returns the meter of text in the grammar of the package (ABNF):
//
//	meter    = numeric [ grouping ] / symbol [ grouping ] / additive
//	         / part 1*( "+" part )
//	part     = numeric / additive
//	numeric  = count "/" count
//	additive = "(" count 1*( "+" count ) ")/" count
//	symbol   = "C|" / "C" / %xC2.A2
//	grouping = "[" count *( "+" count ) "]"
//	count    = nzdigit *digit
//
// The text is exact: no spaces, signs, leading zeros, other letter case, or
// names. Besides canonical texts, Parse accepts ¢ (U+00A2) as C| and an
// explicit grouping equal to the standard one ("6/8[3+3]" is 6/8). A compound
// numerator and a sum take no grouping, and a part of a sum is not a symbol.
//
// Errors, in this order: ErrInvalidText for text outside the grammar,
// including "Meter(unset)"; ErrOutOfRange without other sentinels for the
// first denominator from the left greater than math.MaxInt; then the errors
// of building the value in the order of the text: ErrTooMany for more than
// MaxUnits units in a part (including a number greater than math.MaxInt),
// then for the sum; ErrOutOfRange together with fraction.ErrOutOfRange for a
// sum outside the range; ErrInvalidArgument for an explicit grouping with a
// group greater than math.MaxInt or a sum that differs from the numerator.
//
// An error message quotes the rejected text: in full up to 64 bytes,
// otherwise its longest prefix of whole UTF-8 decoding units within 64 bytes
// followed by "...". The message consists only of constant sentinel texts and
// this excerpt, is at most 512 bytes long, and for one class of error depends
// on the text only through its first 67 bytes. Parsing takes time linear in
// the length of text.
//
// Example: Parse("5/8[2+3]") is MustNew(5, 8).MustWithGrouping(2, 3);
// Parse("3/99999999999999999999") fails with ErrOutOfRange.
func Parse(text string) (Meter, error) {
	x, ok := scan(text)
	if !ok {
		return Meter{}, parseError(classInvalidText, text)
	}
	if x.symbol != 0 {
		n := x.symbol.numerator()
		r := record{symbol: x.symbol, parts: []part{{addends: []int{n}, den: n}}}
		return parseGrouping(text, r, x.grouping)
	}
	r := record{parts: make([]part, len(x.parts))}
	for i, sp := range x.parts {
		d, ok := atoi(sp.den)
		if !ok {
			return Meter{}, parseError(classDenominator, text)
		}
		r.parts[i].den = d
	}
	counts := make([]int, 0, len(x.parts))
	for i, sp := range x.parts {
		addends := make([]int, len(sp.addends))
		for j, s := range sp.addends {
			a, ok := atoi(s)
			if !ok {
				return Meter{}, parseError(classTooMany, text)
			}
			addends[j] = a
		}
		if !unitsFit(addends) {
			return Meter{}, parseError(classTooMany, text)
		}
		r.parts[i].addends = addends
		counts = append(counts, r.parts[i].count())
	}
	if !unitsFit(counts) {
		return Meter{}, parseError(classTooMany, text)
	}
	return parseGrouping(text, r, x.grouping)
}

// parseGrouping applies the explicit grouping of text, if any, to r and
// returns the meter. The grammar gives a grouping only to a simple signature
// or a symbol; a sum, which has none, is checked for the range here.
func parseGrouping(text string, r record, grouping []string) (Meter, error) {
	if grouping != nil {
		groups := make([]int, len(grouping))
		for i, s := range grouping {
			g, ok := atoi(s)
			if !ok {
				return Meter{}, parseError(classInvalidArgument, text)
			}
			groups[i] = g
		}
		if !sumEquals(groups, r.numerator()) {
			return Meter{}, parseError(classInvalidArgument, text)
		}
		r = r.withGrouping(groups)
	}
	m, err := build("meter.Parse", r)
	if err != nil {
		return Meter{}, parseError(classFractionRange, text)
	}
	return m, nil
}

// MustParse is like Parse but panics with the error of Parse.
func MustParse(text string) Meter {
	return must(Parse(text))
}
