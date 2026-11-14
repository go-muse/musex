package meter

import (
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// part is one part of a record: a numerator of one or more addends over a
// denominator. A simple numerator has one addend.
type part struct {
	addends []int
	den     int
}

// count returns the number of units of p: the sum of its addends.
func (p part) count() int {
	n := 0
	for _, a := range p.addends {
		n += a
	}
	return n
}

// grouping returns the grouping of p as a meter of its own without an
// explicit grouping: the addends of a compound numerator, otherwise the
// standard grouping of the numerator, which may be nil.
func (p part) grouping() []int {
	if len(p.addends) > 1 {
		return slices.Clone(p.addends)
	}
	return standardGrouping(p.addends[0])
}

// write appends the text of p to b: "n/d" or "(a+b)/d".
func (p part) write(b *strings.Builder) {
	if len(p.addends) == 1 {
		b.WriteString(strconv.Itoa(p.addends[0]))
	} else {
		b.WriteByte('(')
		writeInts(b, p.addends)
		b.WriteByte(')')
	}
	b.WriteByte('/')
	b.WriteString(strconv.Itoa(p.den))
}

// writeInts appends the numbers joined with "+" to b.
func writeInts(b *strings.Builder, ns []int) {
	for i, n := range ns {
		if i > 0 {
			b.WriteByte('+')
		}
		b.WriteString(strconv.Itoa(n))
	}
}

// record is the decoded record of a set meter. A record is valid by
// construction: every number is at least 1, there are at most MaxUnits units,
// a compound numerator has at least two addends, a sum has at least two
// parts, and an explicit grouping belongs only to a simple signature or a
// symbol, sums to its numerator, and differs from the standard grouping.
type record struct {
	// symbol is the symbol of a FormSymbol record and 0 otherwise.
	symbol Symbol
	// parts are the parts of the record. A symbol has the single part of its
	// structure, 4/4 or 2/2.
	parts []part
	// explicit is the explicit grouping, or nil if there is none.
	explicit []int
}

// form returns the form of r.
func (r record) form() Form {
	switch {
	case r.symbol != 0:
		return FormSymbol
	case len(r.parts) > 1:
		return FormSum
	case len(r.parts[0].addends) > 1:
		return FormAdditive
	}
	return FormSimple
}

// numerator returns the number of units of r.
func (r record) numerator() int {
	n := 0
	for _, p := range r.parts {
		n += p.count()
	}
	return n
}

// commonDen returns the denominator shared by all parts of r and true, or 0
// and false if the parts have different denominators.
func (r record) commonDen() (int, bool) {
	d := r.parts[0].den
	for _, p := range r.parts[1:] {
		if p.den != d {
			return 0, false
		}
	}
	return d, true
}

// grouping returns the grouping of r as a new slice, or nil if r has none.
func (r record) grouping() []int {
	switch r.form() {
	case FormSimple, FormSymbol:
		if r.explicit != nil {
			return slices.Clone(r.explicit)
		}
		return standardGrouping(r.numerator())
	case FormAdditive:
		return slices.Clone(r.parts[0].addends)
	}
	var g []int
	for _, p := range r.parts {
		pg := p.grouping()
		if pg == nil {
			return nil
		}
		g = append(g, pg...)
	}
	return g
}

// unitDens returns the denominator of every unit of r in order.
func (r record) unitDens() []int {
	dens := make([]int, 0, r.numerator())
	for _, p := range r.parts {
		for range p.count() {
			dens = append(dens, p.den)
		}
	}
	return dens
}

// text returns the canonical text of r.
func (r record) text() string {
	var b strings.Builder
	if r.symbol != 0 {
		b.WriteString(r.symbol.String())
	} else {
		for i, p := range r.parts {
			if i > 0 {
				b.WriteByte('+')
			}
			p.write(&b)
		}
	}
	if r.explicit != nil {
		b.WriteByte('[')
		writeInts(&b, r.explicit)
		b.WriteByte(']')
	}
	return b.String()
}

// withGrouping returns r with the grouping groups, which must sum to the
// numerator of a simple signature or a symbol. A grouping equal to the
// standard grouping is not stored.
func (r record) withGrouping(groups []int) record {
	if slices.Equal(groups, standardGrouping(r.numerator())) {
		r.explicit = nil
	} else {
		r.explicit = slices.Clone(groups)
	}
	return r
}

// standardGrouping returns the standard grouping of n units: one group for 1,
// 2, and 3 units, 2+2 for 4 units, and groups of three for multiples of three
// from 6. Other numbers have no standard grouping, and the result is nil.
func standardGrouping(n int) []int {
	switch {
	case n >= 1 && n <= 3:
		return []int{n}
	case n == 4:
		return []int{2, 2}
	case n >= 6 && n%3 == 0:
		g := make([]int, n/3)
		for i := range g {
			g[i] = 3
		}
		return g
	}
	return nil
}

// layout returns the exact start of every unit of r and the size of r.
func (r record) layout() (starts []*big.Rat, size *big.Rat) {
	starts = make([]*big.Rat, 0, r.numerator())
	size = new(big.Rat)
	for _, p := range r.parts {
		d := big.NewInt(int64(p.den))
		for k := range p.count() {
			s := new(big.Rat).SetFrac(big.NewInt(int64(k)), d)
			starts = append(starts, s.Add(s, size))
		}
		size.Add(size, new(big.Rat).SetFrac(big.NewInt(int64(p.count())), d))
	}
	return starts, size
}

// position converts an exact value that is known to be within the range of
// common/fraction and not negative into a position.
func position(r *big.Rat) duration.Position {
	return duration.MustPositionFromFraction(fraction.MustFromBigRat(r))
}

// bigRat returns the exact value of a set position.
func bigRat(p duration.Position) *big.Rat {
	r, _ := p.MustFraction().BigRat()
	return r
}

// build returns the meter of r. For a record of several parts it first checks
// that the size and the start of every unit are within the range of
// common/fraction: other operations then never leave the range when they read
// the structure. For a record of one part both hold by construction: the
// starts are k/d with k at most MaxUnits and d at most math.MaxInt.
func build(op string, r record) (Meter, error) {
	if len(r.parts) == 1 {
		p := r.parts[0]
		return Meter{rec: r.text(), size: duration.MustNew(p.count(), p.den)}, nil
	}
	starts, size := r.layout()
	for _, s := range starts {
		if _, err := fraction.FromBigRat(s); err != nil {
			return Meter{}, causedRangeError(op, err)
		}
	}
	f, err := fraction.FromBigRat(size)
	if err != nil {
		return Meter{}, causedRangeError(op, err)
	}
	return Meter{rec: r.text(), size: duration.MustFromFraction(f)}, nil
}

// syntaxPart is a part of text accepted by the grammar: digit strings without
// leading zeros.
type syntaxPart struct {
	addends []string
	den     string
}

// syntax is text accepted by the grammar before the numbers are checked.
type syntax struct {
	symbol   Symbol
	parts    []syntaxPart
	grouping []string
}

// scan checks text against the grammar of the package and returns its parts.
// It reports false for text outside the grammar. It runs in time linear in
// the length of text.
func scan(text string) (syntax, bool) {
	var x syntax
	i := 0
	switch {
	case strings.HasPrefix(text, "C|"):
		x.symbol, i = SymbolCut, 2
	case strings.HasPrefix(text, "C"):
		x.symbol, i = SymbolCommon, 1
	case strings.HasPrefix(text, "¢"):
		x.symbol, i = SymbolCut, len("¢")
	}
	if x.symbol != 0 {
		if i == len(text) {
			return x, true
		}
		g, j, ok := scanGrouping(text, i)
		if !ok || j != len(text) {
			return syntax{}, false
		}
		x.grouping = g
		return x, true
	}
	for {
		p, j, ok := scanPart(text, i)
		if !ok {
			return syntax{}, false
		}
		x.parts = append(x.parts, p)
		i = j
		if i == len(text) {
			return x, true
		}
		switch {
		case text[i] == '+':
			i++
		case text[i] == '[' && len(x.parts) == 1 && len(p.addends) == 1:
			g, j, ok := scanGrouping(text, i)
			if !ok || j != len(text) {
				return syntax{}, false
			}
			x.grouping = g
			return x, true
		default:
			return syntax{}, false
		}
	}
}

// scanCount scans a number without a leading zero at text[i:] and returns it
// and the index after it.
func scanCount(text string, i int) (string, int, bool) {
	if i >= len(text) || text[i] < '1' || text[i] > '9' {
		return "", i, false
	}
	j := i + 1
	for j < len(text) && text[j] >= '0' && text[j] <= '9' {
		j++
	}
	return text[i:j], j, true
}

// scanPart scans "n/d" or "(a+b+...)/d" at text[i:].
func scanPart(text string, i int) (syntaxPart, int, bool) {
	var p syntaxPart
	if i < len(text) && text[i] == '(' {
		i++
		for {
			a, j, ok := scanCount(text, i)
			if !ok {
				return syntaxPart{}, i, false
			}
			p.addends = append(p.addends, a)
			i = j
			if i < len(text) && text[i] == '+' {
				i++
				continue
			}
			break
		}
		if len(p.addends) < 2 || !strings.HasPrefix(text[i:], ")/") {
			return syntaxPart{}, i, false
		}
		i += 2
	} else {
		n, j, ok := scanCount(text, i)
		if !ok || j >= len(text) || text[j] != '/' {
			return syntaxPart{}, i, false
		}
		p.addends = []string{n}
		i = j + 1
	}
	d, j, ok := scanCount(text, i)
	if !ok {
		return syntaxPart{}, i, false
	}
	p.den = d
	return p, j, true
}

// scanGrouping scans "[a+b+...]" at text[i:].
func scanGrouping(text string, i int) ([]string, int, bool) {
	if i >= len(text) || text[i] != '[' {
		return nil, i, false
	}
	i++
	var g []string
	for {
		c, j, ok := scanCount(text, i)
		if !ok {
			return nil, i, false
		}
		g = append(g, c)
		i = j
		if i < len(text) && text[i] == '+' {
			i++
			continue
		}
		break
	}
	if i >= len(text) || text[i] != ']' {
		return nil, i, false
	}
	return g, i + 1, true
}

// maxIntDigits is the number of decimal digits of math.MaxInt; the project
// supports only 64-bit platforms.
const maxIntDigits = 19

// atoi converts a digit string without a sign or a leading zero. It reports
// false for a number greater than math.MaxInt. A string longer than
// maxIntDigits is rejected without conversion.
func atoi(s string) (int, bool) {
	if len(s) > maxIntDigits {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

// decode returns the record of a set meter from its canonical text, which is
// always accepted by the grammar and holds numbers within int.
func decode(rec string) record {
	x, _ := scan(rec)
	var r record
	if x.symbol != 0 {
		n := x.symbol.numerator()
		r.symbol = x.symbol
		r.parts = []part{{addends: []int{n}, den: n}}
	} else {
		r.parts = make([]part, len(x.parts))
		for i, sp := range x.parts {
			r.parts[i] = part{addends: decodeInts(sp.addends), den: decodeInt(sp.den)}
		}
	}
	if x.grouping != nil {
		r.explicit = decodeInts(x.grouping)
	}
	return r
}

// decodeInt converts a digit string of a canonical text.
func decodeInt(s string) int {
	n, _ := atoi(s)
	return n
}

// decodeInts converts the digit strings of a canonical text.
func decodeInts(ss []string) []int {
	ns := make([]int, len(ss))
	for i, s := range ss {
		ns[i] = decodeInt(s)
	}
	return ns
}
