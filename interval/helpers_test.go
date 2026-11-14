package interval_test

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
)

// The helpers build every expectation from business-spec.md and
// architecture.md of task 18.1 through the public API of the dependencies,
// never through results of package interval (plan, "Approach to tests").

// mag is M = alteration.MaxMagnitude, the bound of an alteration value and
// of a deviation of quality (architecture, notation).
const mag = alteration.MaxMagnitude

// maxSizeSimple is the simple size of an interval of size MaxSize:
// math.MaxInt = 2^63 - 1 is divisible by 7, so ((MaxInt - 2) mod 7) + 2 = 7.
const maxSizeSimple = 7

// ---------------------------------------------------------------------------
// Errors (architecture 4.2).

type namedErr struct {
	name string
	err  error
}

// sentinels are the twelve sentinels of package interval, E-1...E-12.
var sentinels = []namedErr{
	{"ErrUnset", interval.ErrUnset},
	{"ErrInvalidArgument", interval.ErrInvalidArgument},
	{"ErrNoDirection", interval.ErrNoDirection},
	{"ErrNoInversion", interval.ErrNoInversion},
	{"ErrNoOctave", interval.ErrNoOctave},
	{"ErrNoSystem", interval.ErrNoSystem},
	{"ErrDifferentSystems", interval.ErrDifferentSystems},
	{"ErrStartRequired", interval.ErrStartRequired},
	{"ErrFractionalSteps", interval.ErrFractionalSteps},
	{"ErrParse", interval.ErrParse},
	{"ErrNotOnFifths", interval.ErrNotOnFifths},
	{"ErrOutOfRange", interval.ErrOutOfRange},
}

// assertErr checks that err matches exactly the sentinel want among the
// sentinels of package interval (none when want is nil: a "cause of system"
// or "cause of note" without a sentinel of interval, architecture 4.2) and
// matches every listed cause of a dependency.
func assertErr(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := true

	for _, s := range sentinels {
		expected := s.err == want
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.300s", s.name, !expected, expected, err)
		}
	}

	for _, c := range causes {
		if !errors.Is(err, c) {
			ok = false

			t.Errorf("errors.Is(err, %v) = false, want true; err: %.300s", c, err)
		}
	}

	return ok
}

// capturePanic runs f and reports the recovered panic value, if any.
func capturePanic(f func()) (value any, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			value, panicked = r, true
		}
	}()

	f()

	return nil, false
}

// panicError runs f, requires a panic and returns its value as an error
// (architecture 4.2, Must: the panic value is the error itself).
func panicError(t *testing.T, f func()) error {
	t.Helper()

	value, panicked := capturePanic(f)
	if !assert.True(t, panicked, "expected a panic") {
		return nil
	}

	err, ok := value.(error)
	assert.True(t, ok, "panic value is %T, want error", value)

	return err
}

// noPanic runs f and fails the test if it panics.
func noPanic(t *testing.T, f func()) {
	t.Helper()

	if value, panicked := capturePanic(f); panicked {
		t.Errorf("unexpected panic: %v", value)
	}
}

// ---------------------------------------------------------------------------
// Alterations and names.

// alt returns the alteration value num/den.
func alt(num, den int) alteration.Alteration { return alteration.MustFromFraction(num, den) }

// ai returns the whole alteration value n.
func ai(n int) alteration.Alteration { return alteration.MustFromInt(n) }

// sp returns the name with letter l and alteration num/den.
func sp(l letter.Letter, num, den int) spelling.Spelling {
	return spelling.MustNew(l, alteration.MustFromFraction(num, den))
}

// eHalfFlat is E½♭, the name E with alteration -1/2 (business-spec 1.2).
func eHalfFlat() spelling.Spelling { return sp(letter.E, -1, 2) }

// noStart is the "not given" start name of the positional layer (4.2, KR-7).
var noStart = spelling.Spelling{}

// ---------------------------------------------------------------------------
// Test notation of intervals, independent of interval.Parse.
//
// An optional "↓" for a descending interval; a sign of quality: P, M, m, n,
// A repeated k times, d repeated k times, or {±p/q} with the deviation
// itself; then the size in decimal digits. The groups and deviations are
// those of business-spec 1.3 and architecture 4.4.1.

// simpleSize is the simple size of business-spec P-5.
func simpleSize(n int) int {
	if n == 1 {
		return 1
	}

	return (n-2)%7 + 2
}

// perfectSize reports whether size n belongs to the group of "perfect"
// sizes: simple part 1, 4, 5 or 8.
func perfectSize(n int) bool {
	switch simpleSize(n) {
	case 1, 4, 5, 8:
		return true
	default:
		return false
	}
}

// devOf returns the deviation and the size written by the test notation s
// (without the direction mark).
func devOf(s string) (alteration.Alteration, int) {
	var sign, digits string

	if strings.HasPrefix(s, "{") {
		end := strings.Index(s, "}")
		sign, digits = s[:end+1], s[end+1:]
	} else {
		at := strings.IndexAny(s, "0123456789")
		sign, digits = s[:at], s[at:]
	}

	n, err := strconv.Atoi(digits)
	if err != nil {
		panic(fmt.Sprintf("test notation %q: %v", s, err))
	}

	perfect := perfectSize(n)

	switch {
	case sign == "P" || sign == "M":
		return alteration.Natural(), n
	case sign == "m":
		return ai(-1), n
	case sign == "n":
		return alt(-1, 2), n
	case strings.HasPrefix(sign, "{"):
		return parseDev(sign[1 : len(sign)-1]), n
	case strings.Trim(sign, "A") == "":
		return ai(len(sign)), n
	case strings.Trim(sign, "d") == "":
		if perfect {
			return ai(-len(sign)), n
		}

		return ai(-1 - len(sign)), n
	default:
		panic(fmt.Sprintf("test notation %q: unknown sign", s))
	}
}

// parseDev reads "+p", "-p", "+p/q" or "-p/q".
func parseDev(s string) alteration.Alteration {
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimLeft(s, "+-")
	num, den := s, "1"

	if i := strings.Index(s, "/"); i >= 0 {
		num, den = s[:i], s[i+1:]
	}

	p, err1 := strconv.Atoi(num)
	q, err2 := strconv.Atoi(den)

	if err1 != nil || err2 != nil {
		panic(fmt.Sprintf("test deviation %q", s))
	}

	if neg {
		p = -p
	}

	return alt(p, q)
}

// iv returns the interval written by the test notation s. It panics on a
// malformed notation: it is used to build tables.
func iv(s string) interval.Interval {
	down := strings.HasPrefix(s, "↓")
	d, n := devOf(strings.TrimPrefix(s, "↓"))

	i := interval.MustFromDeviation(n, d)
	if down {
		i = i.MustDown()
	}

	return i
}

// ivs maps iv over the notations.
func ivs(ss ...string) []interval.Interval {
	out := make([]interval.Interval, 0, len(ss))
	for _, s := range ss {
		out = append(out, iv(s))
	}

	return out
}

// aug returns the ascending augmented interval of size n and multiplicity k.
func aug(k, n int) interval.Interval { return interval.MustFromDeviation(n, ai(k)) }

// dim returns the ascending diminished interval of size n and multiplicity k.
func dim(k, n int) interval.Interval {
	if perfectSize(n) {
		return interval.MustFromDeviation(n, ai(-k))
	}

	return interval.MustFromDeviation(n, ai(-1-k))
}

// basicSize is the main kind of size n with deviation 0: perfect or major.
func basicSize(n int) interval.Interval { return interval.MustFromDeviation(n, alteration.Natural()) }

// both returns the intervals and their opposites.
func both(is []interval.Interval) []interval.Interval {
	out := make([]interval.Interval, 0, 2*len(is))
	for _, i := range is {
		out = append(out, i, i.MustOpposite())
	}

	return out
}

// predefinedBoth is the 57 predefined intervals in both directions (F-2).
func predefinedBoth() []interval.Interval { return both(interval.Predefined()) }

// ---------------------------------------------------------------------------
// Systems (business-spec 1.2; plan, "Systems and context").

func trad() *system.System  { return system.Traditional() }
func pos19() *system.System { return system.Positions19() }
func qt24() *system.System  { return system.QuarterTones24() }
func sf31() *system.System  { return system.SharpsAndFlats31() }

// irregular is the irregular system of the domain base: N = 12,
// C0 D3 E4 F5 G7 A9 B11, c = 1, integer alterations.
func irregular() *system.System {
	return system.MustNew(system.Definition{
		Name:        "irregular",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 3, 4, 5, 7, 9, 11},
		Chromatic:   1,
		Alterations: system.Integers(),
	})
}

// boundaryA is the traditional layout with the boundary letter A:
// A0 B2 C3 D5 E7 F8 G10 (offsets indexed by the ordinal C D E F G A B).
func boundaryA() *system.System {
	return system.MustNew(system.Definition{
		Name:        "boundary-a",
		Classes:     12,
		Boundary:    letter.A,
		Offsets:     [letter.Count]int{3, 5, 7, 8, 10, 0, 2},
		Chromatic:   1,
		Alterations: system.Integers(),
	})
}

// ---------------------------------------------------------------------------
// Pitches.

// pt is the pitch (name, ASPN octave aspn).
func pt(name spelling.Spelling, aspn int) pitch.Pitch { return pitch.MustFromASPN(name, aspn) }

// wp is the pitch with name written in the octave w of sys, such as "B♯3".
func wp(sys *system.System, name spelling.Spelling, w int) pitch.Pitch {
	return sys.MustFromWritten(name, octave.FromASPN(w))
}

// ---------------------------------------------------------------------------
// Settings and sources for notes.

func std() *settings.Settings { return settings.Standard() }

// set19 is the standard set with the 19-position system (19-EDO tuning).
func set19() *settings.Settings {
	return std().MustWithTuning(tuning.MustNew(system.Positions19(), temperament.Equal19(),
		tuning.RootFromName(spelling.C()), tuning.A4At440()))
}

// set24 is the standard set with the 24-quarter-tone system.
func set24() *settings.Settings {
	return std().MustWithTuning(tuning.MustNew(system.QuarterTones24(), temperament.Equal24(),
		tuning.RootFromName(spelling.C()), tuning.A4At440())).
		MustWithNotation(notation.MustBuiltIn("lilypond-english"))
}

// switchSource returns its current set; the set can be replaced, so a note
// can be created with one set and read with another (an unset set, or a set
// in whose system the name of the note is not admissible).
type switchSource struct {
	p atomic.Pointer[settings.Settings]
}

func newSwitch(set *settings.Settings) *switchSource {
	s := &switchSource{}
	s.p.Store(set)

	return s
}

func (s *switchSource) Settings() *settings.Settings { return s.p.Load() }

func (s *switchSource) set(set *settings.Settings) { s.p.Store(set) }

// noteAt returns the note with pitch p, correction c cents and source src.
func noteAt(t *testing.T, p pitch.Pitch, c int, src settings.Source) note.Note {
	t.Helper()

	n, err := note.FromPitch(p, centsOf(c), src)
	if err != nil {
		t.Fatalf("note.FromPitch(%v, %d, ...): %v", p, c, err)
	}

	return n
}

// noteName returns the note without octave with name, c cents and src.
func noteName(t *testing.T, name spelling.Spelling, c int, src settings.Source) note.Note {
	t.Helper()

	n, err := note.FromSpelling(name, centsOf(c), src)
	if err != nil {
		t.Fatalf("note.FromSpelling(%v, %d, ...): %v", name, c, err)
	}

	return n
}

// centsOf returns c whole cents.
func centsOf(c int) cents.Cents { return cents.MustFromInt(c) }

// ---------------------------------------------------------------------------
// Misc.

// minInt and maxInt name the edges of int.
const (
	minInt = math.MinInt
	maxInt = math.MaxInt
)
