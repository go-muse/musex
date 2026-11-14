package meter

// The kinds are rules of the product built on the reference base: the two
// traditions classify a meter independently, each by its own rules, applied
// in order until the first one that matches.

// shape is the structure of a meter: the denominator of every unit and the
// grouping, or nil if there is none.
type shape struct {
	dens     []int
	grouping []int
}

// shapeOf returns the structure of r.
func (r record) shapeOf() shape {
	return shape{dens: r.unitDens(), grouping: r.grouping()}
}

// asymmetricalWithoutGrouping reports whether a structure without a grouping
// is asymmetrical: all units have one duration, and their number is odd, at
// least 5, and not a multiple of three, as in 5/8, 7/8, and 5/8+2/8.
func (s shape) asymmetricalWithoutGrouping() bool {
	n := len(s.dens)
	for _, d := range s.dens[1:] {
		if d != s.dens[0] {
			return false
		}
	}
	return n >= 5 && n%2 == 1 && n%3 != 0
}

// groupDens returns the denominator of the units of each group; a group lies
// within one part, so its units have one denominator.
func (s shape) groupDens() []int {
	out := make([]int, len(s.grouping))
	u := 0
	for i, k := range s.grouping {
		out[i] = s.dens[u]
		u += k
	}
	return out
}

// equalGroups reports whether all groups have the same number of units and
// the same unit duration.
func (s shape) equalGroups() bool {
	dens := s.groupDens()
	for i := range s.grouping {
		if s.grouping[i] != s.grouping[0] || dens[i] != dens[0] {
			return false
		}
	}
	return true
}

// kind returns the kind of the English-language tradition, or an
// ErrGroupingRequired error when it cannot be determined without a grouping.
func (s shape) kind(op string) (Kind, error) {
	switch {
	case s.grouping == nil:
		if s.asymmetricalWithoutGrouping() {
			return KindAsymmetrical, nil
		}
		return 0, simpleError(ErrGroupingRequired, op, "kind needs a grouping")
	case standardSimple(s.dens, s.grouping):
		return KindSimple, nil
	case !s.equalGroups():
		return KindAsymmetrical, nil
	case s.grouping[0] == 3:
		return KindCompound, nil
	case s.grouping[0] == 2:
		return KindSimple, nil
	}
	return KindUndetermined, nil
}

// russianKind returns the kind of the Russian tradition, or an
// ErrGroupingRequired error when it cannot be determined without a grouping.
func (s shape) russianKind(op string) (RussianKind, error) {
	if s.grouping == nil {
		if s.asymmetricalWithoutGrouping() {
			return RussianCompoundMixed, nil
		}
		return 0, simpleError(ErrGroupingRequired, op, "kind needs a grouping")
	}
	for _, k := range s.grouping {
		if k != 2 && k != 3 {
			return RussianUndetermined, nil
		}
	}
	switch {
	case len(s.grouping) == 1:
		return RussianSimple, nil
	case s.equalGroups():
		return RussianCompoundUniform, nil
	}
	return RussianCompoundMixed, nil
}

// Kind returns the kind of m in the English-language tradition, by the first
// matching rule:
//
//  1. Without a grouping: KindAsymmetrical if all units have one duration and
//     their number is odd, at least 5, and not a multiple of three (5/8, 7/8);
//     otherwise an error.
//  2. The standard simple structure (3/4, 4/4, C): KindSimple.
//  3. Groups that are not all equal (5/8[2+3], 2/4+3/8): KindAsymmetrical.
//  4. Equal groups of three units (6/8, 12/8, 15/8): KindCompound.
//  5. Equal groups of two units (6/8[2+2+2]): KindSimple.
//  6. Otherwise (1/4, (4+4)/8): KindUndetermined.
//
// Groups are equal when they have the same number of units and the same unit
// duration. The number of beats is BeatCount; for rule 1 it is unknown.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// when rule 1 gives no kind, such as for 8/8.
func (m Meter) Kind() (Kind, error) {
	const op = "Meter.Kind"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	return r.shapeOf().kind(op)
}

// MustKind is like Kind but panics with the error of Kind.
func (m Meter) MustKind() Kind {
	return must(m.Kind())
}

// RussianKind returns the kind of m in the Russian tradition, independently
// of Kind, by the first matching rule:
//
//  1. Without a grouping: RussianCompoundMixed under the condition of rule 1
//     of Kind (5/8, 7/8); otherwise an error.
//  2. A group of neither two nor three units (1/4, 4/4[3+1]):
//     RussianUndetermined.
//  3. One group: RussianSimple (2/4, 3/4); several equal groups:
//     RussianCompoundUniform (4/4, 6/8); unequal groups: RussianCompoundMixed
//     (5/8[2+3], 2/4+2/8).
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// when rule 1 gives no kind, such as for 8/8.
func (m Meter) RussianKind() (RussianKind, error) {
	const op = "Meter.RussianKind"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	return r.shapeOf().russianKind(op)
}

// MustRussianKind is like RussianKind but panics with the error of
// RussianKind.
func (m Meter) MustRussianKind() RussianKind {
	return must(m.RussianKind())
}

// BeatDivision returns into how many equal parts a beat of m usually divides
// and subdivides: (2, 4) for KindSimple and (3, 6) for KindCompound.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// when Kind fails with it; ErrUndefined for KindAsymmetrical and
// KindUndetermined, including 5/8 and 7/8 without a grouping.
//
// Example: 6/8 gives (3, 6); 3/4 gives (2, 4).
func (m Meter) BeatDivision() (division, subdivision int, err error) {
	const op = "Meter.BeatDivision"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, 0, err
	}
	k, err := r.shapeOf().kind(op)
	if err != nil {
		return 0, 0, err
	}
	switch k {
	case KindSimple:
		return 2, 4, nil
	case KindCompound:
		return 3, 6, nil
	}
	return 0, 0, simpleError(ErrUndefined, op, "beat division of an asymmetrical or undetermined kind")
}

// MustBeatDivision is like BeatDivision but panics with the error of
// BeatDivision.
func (m Meter) MustBeatDivision() (division, subdivision int) {
	return must2(m.BeatDivision())
}

// beatsNames are the names of the number of beats 2, 3, and 4 of a simple or
// a compound kind, in English and in Russian.
var beatsNames = map[int][2]string{
	2: {"duple", "двухдольный"},
	3: {"triple", "трёхдольный"},
	4: {"quadruple", "четырёхдольный"},
}

// kindName returns the full name of the kind of m in the English-language
// tradition in the language with index lang: 0 for English, 1 for Russian.
func (m Meter) kindName(op string, lang int) (string, bool, error) {
	r, err := m.recordOf(op)
	if err != nil {
		return "", false, err
	}
	s := r.shapeOf()
	k, err := s.kind(op)
	if err != nil {
		return "", false, err
	}
	var name string
	if lang == 0 {
		name, _ = k.MustName()
	} else {
		name, _ = k.MustRussianName()
	}
	if name == "" {
		return "", false, nil
	}
	if k == KindSimple || k == KindCompound {
		g, _ := r.beatGrouping(op)
		if n, ok := beatsNames[len(g)]; ok {
			name += " " + n[lang]
		}
	}
	return name, true, nil
}

// KindName returns the full English name of the kind of m in the
// English-language tradition: the name of Kind and, for KindSimple and
// KindCompound with 2, 3, or 4 beats, the name of the number of beats:
// "simple triple" for 3/4, "compound duple" for 6/8. Otherwise only the name
// of the kind: "compound" for 15/8, "asymmetrical" for 5/8[2+3] and 5/8; the
// number of beats is BeatCount. KindUndetermined has no name: the result is
// ("", false, nil).
//
// Errors: as Kind.
func (m Meter) KindName() (name string, ok bool, err error) {
	return m.kindName("Meter.KindName", 0)
}

// MustKindName is like KindName but panics with the error of KindName.
func (m Meter) MustKindName() (name string, ok bool) {
	return must2(m.KindName())
}

// KindRussianName returns the full Russian name of the kind of m in the
// English-language tradition by the rule of KindName: "простой трёхдольный"
// for 3/4, "сложный двухдольный" for 6/8, "асимметричный" for 5/8[2+3].
// KindUndetermined has no name: the result is ("", false, nil).
//
// Errors: as Kind.
func (m Meter) KindRussianName() (name string, ok bool, err error) {
	return m.kindName("Meter.KindRussianName", 1)
}

// MustKindRussianName is like KindRussianName but panics with the error of
// KindRussianName.
func (m Meter) MustKindRussianName() (name string, ok bool) {
	return must2(m.KindRussianName())
}
