package note_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Computations by a set (architecture 6.8; OP-32...OP-38, OP-55, OP-56;
// K-1...K-8, K-19, K-24, K-29, K-30, K-32, K-34, K-35; G-1).

// key turns a result into a comparable value: exact cents, classes and octaves
// by their canonical texts, numbers and strings as they are.
func key(v any) any {
	switch x := v.(type) {
	case cents.Cents:
		return "cents " + x.String()
	case pitchclass.Class:
		return fmt.Sprintf("class %d mod %d", x.Number(), x.Modulus())
	case octave.Octave:
		return "octave " + x.String()
	default:
		return v
	}
}

// computeForm is one computing method with its forms. expect gives the result
// through the dependencies for a note (name, Octave aspn when hasOct) + c in
// the set s.
type computeForm struct {
	name        string
	needsOctave bool
	isBool      bool // IsAdmissible, IsExpressible: "no" instead of E-5
	own         func(n note.Note) (any, error)
	must        func(n note.Note) any
	in          func(n note.Note, src settings.Source) (any, error)
	mustIn      func(n note.Note, src settings.Source) any
	inT         func(n note.Note, t *tuning.Tuning) (any, error)
	mustInT     func(n note.Note, t *tuning.Tuning) any
	expect      func(s *settings.Settings, name spelling.Spelling, aspn int, hasOct bool, c cents.Cents) any
}

func wrap[T any](v T, err error) (any, error) { return v, err }

func computeForms() []computeForm {
	p := func(name spelling.Spelling, aspn int) pitch.Pitch { return pt(name, aspn) }
	w := func(s *settings.Settings, name spelling.Spelling, aspn int) octave.Octave {
		return s.System().MustWrittenOctave(p(name, aspn))
	}
	return []computeForm{
		{
			name: "IsAdmissible", isBool: true,
			own:    func(n note.Note) (any, error) { return wrap(n.IsAdmissible()) },
			must:   func(n note.Note) any { return n.MustIsAdmissible() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.IsAdmissibleIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustIsAdmissibleIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, _ int, _ bool, _ cents.Cents) any {
				return s.System().MustAdmits(name)
			},
		},
		{
			name:    "Class",
			own:     func(n note.Note) (any, error) { return wrap(n.Class()) },
			must:    func(n note.Note) any { return n.MustClass() },
			in:      func(n note.Note, s settings.Source) (any, error) { return wrap(n.ClassIn(s)) },
			mustIn:  func(n note.Note, s settings.Source) any { return n.MustClassIn(s) },
			inT:     func(n note.Note, t *tuning.Tuning) (any, error) { return wrap(n.ClassInTuning(t)) },
			mustInT: func(n note.Note, t *tuning.Tuning) any { return n.MustClassInTuning(t) },
			expect: func(s *settings.Settings, name spelling.Spelling, _ int, _ bool, _ cents.Cents) any {
				return s.System().MustClass(name)
			},
		},
		{
			name: "WrittenOctave", needsOctave: true,
			own:    func(n note.Note) (any, error) { return wrap(n.WrittenOctave()) },
			must:   func(n note.Note) any { return n.MustWrittenOctave() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.WrittenOctaveIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustWrittenOctaveIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, _ cents.Cents) any {
				return w(s, name, aspn)
			},
		},
		{
			name: "WrittenASPN", needsOctave: true,
			own:  func(n note.Note) (any, error) { return wrap(n.WrittenASPN()) },
			must: func(n note.Note) any { return n.MustWrittenASPN() },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, _ cents.Cents) any {
				return w(s, name, aspn).MustASPN()
			},
		},
		{
			name: "WrittenNumber", needsOctave: true,
			own:    func(n note.Note) (any, error) { return wrap(n.WrittenNumber()) },
			must:   func(n note.Note) any { return n.MustWrittenNumber() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.WrittenNumberIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustWrittenNumberIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, _ cents.Cents) any {
				return w(s, name, aspn).MustNumber(s.Numbering())
			},
		},
		{
			name: "WrittenName", needsOctave: true,
			own:    func(n note.Note) (any, error) { return wrap(n.WrittenName()) },
			must:   func(n note.Note) any { return n.MustWrittenName() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.WrittenNameIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustWrittenNameIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, _ cents.Cents) any {
				return w(s, name, aspn).Name(s.Catalog())
			},
		},
		{
			name: "Step", needsOctave: true,
			own:     func(n note.Note) (any, error) { return wrap(n.Step()) },
			must:    func(n note.Note) any { return n.MustStep() },
			in:      func(n note.Note, s settings.Source) (any, error) { return wrap(n.StepIn(s)) },
			mustIn:  func(n note.Note, s settings.Source) any { return n.MustStepIn(s) },
			inT:     func(n note.Note, t *tuning.Tuning) (any, error) { return wrap(n.StepInTuning(t)) },
			mustInT: func(n note.Note, t *tuning.Tuning) any { return n.MustStepInTuning(t) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, _ cents.Cents) any {
				return s.Tuning().MustStepOf(p(name, aspn))
			},
		},
		{
			name: "CentsFromReference", needsOctave: true,
			own:     func(n note.Note) (any, error) { return wrap(n.CentsFromReference()) },
			must:    func(n note.Note) any { return n.MustCentsFromReference() },
			in:      func(n note.Note, s settings.Source) (any, error) { return wrap(n.CentsFromReferenceIn(s)) },
			mustIn:  func(n note.Note, s settings.Source) any { return n.MustCentsFromReferenceIn(s) },
			inT:     func(n note.Note, t *tuning.Tuning) (any, error) { return wrap(n.CentsFromReferenceInTuning(t)) },
			mustInT: func(n note.Note, t *tuning.Tuning) any { return n.MustCentsFromReferenceInTuning(t) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, c cents.Cents) any {
				return s.Tuning().MustCentsFromReference(p(name, aspn)).MustAdd(c)
			},
		},
		{
			name: "Frequency", needsOctave: true,
			own:     func(n note.Note) (any, error) { return wrap(n.Frequency()) },
			must:    func(n note.Note) any { return n.MustFrequency() },
			in:      func(n note.Note, s settings.Source) (any, error) { return wrap(n.FrequencyIn(s)) },
			mustIn:  func(n note.Note, s settings.Source) any { return n.MustFrequencyIn(s) },
			inT:     func(n note.Note, t *tuning.Tuning) (any, error) { return wrap(n.FrequencyInTuning(t)) },
			mustInT: func(n note.Note, t *tuning.Tuning) any { return n.MustFrequencyInTuning(t) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, c cents.Cents) any {
				return s.Tuning().MustShiftedFrequency(p(name, aspn), c)
			},
		},
		{
			name: "PitchFrequency", needsOctave: true,
			own:     func(n note.Note) (any, error) { return wrap(n.PitchFrequency()) },
			must:    func(n note.Note) any { return n.MustPitchFrequency() },
			in:      func(n note.Note, s settings.Source) (any, error) { return wrap(n.PitchFrequencyIn(s)) },
			mustIn:  func(n note.Note, s settings.Source) any { return n.MustPitchFrequencyIn(s) },
			inT:     func(n note.Note, t *tuning.Tuning) (any, error) { return wrap(n.PitchFrequencyInTuning(t)) },
			mustInT: func(n note.Note, t *tuning.Tuning) any { return n.MustPitchFrequencyInTuning(t) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, _ bool, _ cents.Cents) any {
				return s.Tuning().MustFrequency(p(name, aspn))
			},
		},
		{
			name: "IsExpressible", isBool: true,
			own:    func(n note.Note) (any, error) { return wrap(n.IsExpressible()) },
			must:   func(n note.Note) any { return n.MustIsExpressible() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.IsExpressibleIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustIsExpressibleIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, hasOct bool, _ cents.Cents) any {
				if hasOct {
					return s.MustExpressesPitch(p(name, aspn))
				}
				return s.MustExpresses(name)
			},
		},
		{
			name:   "Format",
			own:    func(n note.Note) (any, error) { return wrap(n.Format()) },
			must:   func(n note.Note) any { return n.MustFormat() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.FormatIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustFormatIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, aspn int, hasOct bool, _ cents.Cents) any {
				if hasOct {
					return s.MustFormatPitch(p(name, aspn))
				}
				return s.MustFormat(name)
			},
		},
		{
			name:   "FormatName",
			own:    func(n note.Note) (any, error) { return wrap(n.FormatName()) },
			must:   func(n note.Note) any { return n.MustFormatName() },
			in:     func(n note.Note, s settings.Source) (any, error) { return wrap(n.FormatNameIn(s)) },
			mustIn: func(n note.Note, s settings.Source) any { return n.MustFormatNameIn(s) },
			expect: func(s *settings.Settings, name spelling.Spelling, _ int, _ bool, _ cents.Cents) any {
				return s.MustFormat(name)
			},
		},
	}
}

// computeNote is an input of the result table.
type computeNote struct {
	name   spelling.Spelling
	aspn   int
	hasOct bool
	c      cents.Cents
}

// TestComputeResults compares every form of every computing method with the
// dependencies on several notes and sets: the own set, an explicit source
// with an unbound note, an explicit tuning with an unbound note. Each call
// reads each needed source once (P-5) and does not change the note.
func TestComputeResults(t *testing.T) {
	t.Parallel()

	sets := map[string]*settings.Settings{
		"standard":       std(),
		"set 19":         set19(),
		"Franco-Belgian": stdFB(),
		"A4 at 415":      stdAt(415),
	}
	notes := []computeNote{
		{spelling.CSharp(), 4, true, ct(30)},
		{spelling.BSharp(), 4, true, ct(0)},
		{spelling.CFlat(), 3, true, cents.MustFromFraction(-7, 3)},
		{spelling.A(), 4, true, ct(0)},
		{spelling.E(), 4, true, cents.MustFromFloat(-13.6863)},
		{spelling.CSharp(), 0, false, ct(30)},
		{spelling.BDoubleFlat(), 0, false, ct(0)},
	}
	for _, f := range computeForms() {
		for setName, s := range sets {
			for _, in := range notes {
				if f.needsOctave && !in.hasOct {
					continue
				}
				t.Run(fmt.Sprintf("%s/%s/%v %d %v", f.name, setName, in.name, in.aspn, in.c), func(t *testing.T) {
					t.Parallel()
					want := key(f.expect(s, in.name, in.aspn, in.hasOct, in.c))
					src := newCounting(s)
					var bound, free note.Note
					if in.hasOct {
						bound = mk(t, in.name, in.aspn, in.c, src)
						free = mk(t, in.name, in.aspn, in.c, nil)
					} else {
						bound = mkName(t, in.name, in.c, src)
						free = mkName(t, in.name, in.c, nil)
					}
					before, freeBefore := bound, free
					reads := src.reads()

					got, err := f.own(bound)
					require.NoError(t, err)
					assert.Equal(t, want, key(got), "own set")
					assert.EqualValues(t, 1, src.reads()-reads, "one read")
					assert.Equal(t, want, key(f.must(bound)), "Must, own set")

					if f.in != nil {
						explicit := newCounting(s)
						got, err = f.in(free, explicit)
						require.NoError(t, err)
						assert.Equal(t, want, key(got), "explicit source")
						assert.EqualValues(t, 1, explicit.reads(), "one read of the explicit source")
						assert.Equal(t, want, key(f.mustIn(free, explicit)), "Must, explicit source")
					}
					if f.inT != nil {
						got, err = f.inT(free, s.Tuning())
						require.NoError(t, err)
						assert.Equal(t, want, key(got), "explicit tuning")
						assert.Equal(t, want, key(f.mustInT(free, s.Tuning())), "Must, explicit tuning")
					}
					assertSame(t, before, bound)
					assertSame(t, freeBefore, free)
					assert.False(t, free.IsBound(), "an explicit context does not bind the note")
				})
			}
		}
	}
}

// TestComputeFailures checks the failures of every form in the order of
// architecture 6.8: E-1, E-4, E-2, E-3, E-5; In forms: E-1 of the source,
// E-3 of the explicit set; InTuning forms: E-1 of the tuning; Must panics.
func TestComputeFailures(t *testing.T) {
	t.Parallel()

	for _, f := range computeForms() {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			s := std()
			free := mk(t, spelling.C(), 4, ct(0), nil)
			freeName := mkName(t, spelling.C(), ct(0), nil)
			boundName := mkName(t, spelling.C(), ct(0), s)
			stale := onStale(spelling.C(), 4)(t)
			inadm, _ := inadmissibleNote(t)

			check := func(label string, call func() (any, error), mustCall func(), want error, extra ...error) {
				t.Helper()
				_, err := call()
				if !assertErr(t, err, want, extra...) {
					t.Log(label)
				}
				if !assertPanicErr(t, mustCall, want, extra...) {
					t.Log(label, "(Must)")
				}
			}

			own := func(n note.Note) func() (any, error) { return func() (any, error) { return f.own(n) } }
			mustOwn := func(n note.Note) func() { return func() { f.must(n) } }

			check("unset note", own(note.Note{}), mustOwn(note.Note{}), note.ErrUnset)
			check("unbound", own(free), mustOwn(free), note.ErrNoSource)
			check("unset settings", own(stale), mustOwn(stale), note.ErrUnsetSettings, settings.ErrUnset)
			if f.needsOctave {
				check("no octave, bound", own(boundName), mustOwn(boundName), note.ErrNoOctave)
				check("no octave before no source", own(freeName), mustOwn(freeName), note.ErrNoOctave)
			} else {
				check("no octave, unbound", own(freeName), mustOwn(freeName), note.ErrNoSource)
			}
			if f.isBool {
				got, err := f.own(inadm)
				require.NoError(t, err, "an inadmissible name gives false, not E-5")
				assert.Equal(t, false, got)
			} else {
				check("inadmissible", own(inadm), mustOwn(inadm), note.ErrInadmissible, system.ErrInadmissible)
			}

			if f.in != nil {
				in := func(n note.Note, src settings.Source) func() (any, error) {
					return func() (any, error) { return f.in(n, src) }
				}
				mustIn := func(n note.Note, src settings.Source) func() { return func() { f.mustIn(n, src) } }
				check("In nil source", in(free, nil), mustIn(free, nil), note.ErrUnset)
				check("In unset note", in(note.Note{}, s), mustIn(note.Note{}, s), note.ErrUnset)
				check("In unset settings", in(free, &unsetSource{}), mustIn(free, &unsetSource{}), note.ErrUnsetSettings, settings.ErrUnset)
				_, err := f.in(stale, s)
				require.NoError(t, err, "an In form does not read the own source")
				if f.needsOctave {
					check("In no octave", in(freeName, s), mustIn(freeName, s), note.ErrNoOctave)
				}
				cc := mk(t, spelling.CDoubleSharp(), 4, ct(0), nil)
				if f.isBool {
					got, err := f.in(cc, setK25())
					require.NoError(t, err)
					assert.Equal(t, false, got)
				} else {
					check("In inadmissible", in(cc, setK25()), mustIn(cc, setK25()), note.ErrInadmissible, system.ErrInadmissible)
				}
			}
			if f.inT != nil {
				inT := func(n note.Note, tu *tuning.Tuning) func() (any, error) {
					return func() (any, error) { return f.inT(n, tu) }
				}
				mustInT := func(n note.Note, tu *tuning.Tuning) func() { return func() { f.mustInT(n, tu) } }
				check("InTuning nil", inT(free, nil), mustInT(free, nil), note.ErrUnset)
				check("InTuning unset", inT(free, &tuning.Tuning{}), mustInT(free, &tuning.Tuning{}), note.ErrUnset)
				check("InTuning unset note", inT(note.Note{}, tuning.Standard()), mustInT(note.Note{}, tuning.Standard()), note.ErrUnset)
				if f.needsOctave {
					check("InTuning no octave", inT(freeName, tuning.Standard()), mustInT(freeName, tuning.Standard()), note.ErrNoOctave)
				}
				h := mk(t, cHalf(), 4, ct(0), nil)
				check("InTuning inadmissible", inT(h, tuning.Standard()), mustInT(h, tuning.Standard()), note.ErrInadmissible, system.ErrInadmissible)
			}
		})
	}
}

// TestComputeInDoesNotReadTheOwnSource checks that the In forms read only the
// explicit source (P-7: the bond of the note does not change).
func TestComputeInDoesNotReadTheOwnSource(t *testing.T) {
	t.Parallel()

	for _, f := range computeForms() {
		if f.in == nil {
			continue
		}
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			own, explicit := newCounting(std()), newCounting(set19())
			n := mk(t, spelling.C(), 4, ct(0), own)
			before := own.reads()
			_, err := f.in(n, explicit)
			require.NoError(t, err)
			assert.Equal(t, before, own.reads())
			assert.EqualValues(t, 1, explicit.reads())
			assert.True(t, n.Source() == settings.Source(own))
		})
	}
}

// TestComputeControlExamples covers the control examples of business-spec 6.1
// for one note (frequencies to 4 places, cents to 6 places).
func TestComputeControlExamples(t *testing.T) {
	t.Parallel()

	const hz = 0.00005

	t.Run("K-1 C4", func(t *testing.T) {
		t.Parallel()
		assert.InDelta(t, 261.6256, mk(t, spelling.C(), 4, ct(0), std()).MustFrequency(), hz)
		assert.InDelta(t, 523.2511, mk(t, spelling.C(), 5, ct(0), std()).MustFrequency(), hz)
	})

	t.Run("K-2 A4 +100", func(t *testing.T) {
		t.Parallel()
		assert.InDelta(t, 466.1638, mk(t, spelling.A(), 4, ct(100), std()).MustFrequency(), hz)
	})

	t.Run("K-3 E4 -13.6863", func(t *testing.T) {
		t.Parallel()
		assert.InDelta(t, 327.0320, mk(t, spelling.E(), 4, cents.MustFromFloat(-13.6863), std()).MustFrequency(), hz)
	})

	t.Run("K-6 two octaves up multiply by 4", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(30), std())
		f := n.MustFrequency()
		n.MustShift(2)
		assert.InDelta(t, 1064.7948, n.MustFrequency(), hz)
		assert.InDelta(t, 4*f, n.MustFrequency(), 1e-9)
	})

	t.Run("K-7 steps", func(t *testing.T) {
		t.Parallel()
		s := std()
		for _, tc := range []struct {
			name spelling.Spelling
			aspn int
			want system.Step
		}{
			{spelling.A(), 4, 0}, {spelling.C(), 5, 3}, {spelling.BSharp(), 4, -9}, {spelling.C(), 4, -9},
		} {
			assert.Equal(t, tc.want, mk(t, tc.name, tc.aspn, ct(0), s).MustStep(), "%v %d", tc.name, tc.aspn)
		}
	})

	t.Run("K-8 B♯ in Octave 4", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.BSharp(), 4, ct(0), std())
		assert.True(t, n.MustWrittenOctave().Equal(oct(3)))
		assert.Equal(t, 3, n.MustWrittenASPN())
		assert.Equal(t, 3, n.MustWrittenNumber())
		assert.Equal(t, "B♯3", n.MustFormat())
		assert.Equal(t, "B♯", n.MustFormatName())
	})

	t.Run("K-34 Franco-Belgian", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.BSharp(), 4, ct(0), stdFB())
		assert.Equal(t, 2, n.MustWrittenNumber())
		assert.Equal(t, 3, n.MustWrittenASPN())
	})

	t.Run("K-19 unbound C4", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), nil)
		_, err := n.Frequency()
		assertErr(t, err, note.ErrNoSource)
		assert.InDelta(t, 261.6256, n.MustFrequencyIn(std()), hz)
		assert.InDelta(t, 261.6256, n.MustFrequencyInTuning(tuning.Standard()), hz)
		assert.False(t, n.IsBound())
	})

	t.Run("K-20 no octave", func(t *testing.T) {
		t.Parallel()
		for _, src := range []settings.Source{std(), nil} {
			n := mkName(t, spelling.CSharp(), ct(30), src)
			_, err := n.Frequency()
			assertErr(t, err, note.ErrNoOctave)
		}
	})

	t.Run("K-24 the source is retuned", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		n := mk(t, spelling.A(), 4, ct(0), src)
		assert.InDelta(t, 440.0, n.MustFrequency(), 1e-9)
		src.set(stdAt(415))
		assert.InDelta(t, 415.0, n.MustFrequency(), 1e-9)
		assert.True(t, n.Source() == settings.Source(src))
	})

	t.Run("K-29 unset settings", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		n := mk(t, spelling.C(), 4, ct(0), src)
		m := mkName(t, spelling.C(), ct(0), src)
		src.set(nil)
		_, err := n.Frequency()
		assertErr(t, err, note.ErrUnsetSettings, settings.ErrUnset)
		_, err = m.Frequency()
		assertErr(t, err, note.ErrNoOctave)
	})

	t.Run("K-30 frequency out of range", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 2000, ct(0), std())
		_, err := n.Frequency()
		assertErr(t, err, note.ErrOutOfRange, tuning.ErrOutOfRange)
		_, err = n.PitchFrequency()
		assertErr(t, err, note.ErrOutOfRange, tuning.ErrOutOfRange)
		_, err = n.FrequencyInTuning(tuning.Standard())
		assertErr(t, err, note.ErrOutOfRange, tuning.ErrOutOfRange)
	})

	t.Run("step out of range", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), math.MaxInt, ct(0), std())
		_, err := n.Step()
		assertErr(t, err, note.ErrOutOfRange, tuning.ErrOutOfRange)
	})

	t.Run("written octave out of range", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.BSharp(), math.MinInt, ct(0), std())
		_, err := n.WrittenOctave()
		assertErr(t, err, note.ErrOutOfRange, system.ErrOutOfRange)
		_, err = n.WrittenASPN()
		assertErr(t, err, note.ErrOutOfRange, system.ErrOutOfRange)
	})

	t.Run("visible number out of range", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), math.MinInt, ct(0), stdFB())
		_, err := n.WrittenNumber()
		assertErr(t, err, note.ErrOutOfRange, octave.ErrOutOfRange, number.ErrOutOfRange)
	})

	t.Run("K-32 cents from the reference", func(t *testing.T) {
		t.Parallel()
		got := mk(t, spelling.C(), 4, ct(30), std()).MustCentsFromReference()
		assert.True(t, got.Equal(ct(-870)), "got %v", got)
	})

	t.Run("K-35 expressible", func(t *testing.T) {
		t.Parallel()
		assert.False(t, mk(t, spi(letter.C, 3), 4, ct(0), std()).MustIsExpressible())
		assert.True(t, mk(t, spelling.CSharp(), 4, ct(0), std()).MustIsExpressible())
		assert.True(t, mk(t, spelling.CSharp(), 4, ct(99), std()).MustIsExpressible(), "the cents do not matter")
	})

	t.Run("G-1 zero cents is bit for bit the pitch frequency", func(t *testing.T) {
		t.Parallel()
		tu := tuning.Standard()
		for _, aspn := range []int{-3, 0, 4, 9} {
			for _, name := range []spelling.Spelling{spelling.C(), spelling.FSharp(), spelling.BFlat()} {
				n := mk(t, name, aspn, ct(0), std())
				want := math.Float64bits(tu.MustFrequency(pt(name, aspn)))
				assert.Equal(t, want, math.Float64bits(n.MustFrequency()))
				assert.Equal(t, want, math.Float64bits(n.MustPitchFrequency()))
				assert.Equal(t, want, math.Float64bits(n.MustFrequencyInTuning(tu)))
				assert.Equal(t, want, math.Float64bits(n.MustFrequencyIn(std())))
			}
		}
	})

	t.Run("cents are excluded from the pitch frequency", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.A(), 4, ct(100), std())
		assert.InDelta(t, 440.0, n.MustPitchFrequency(), 1e-9)
	})

	t.Run("format errors of notation without a note sentinel", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.CSharp(), 4, ct(0), stdNotation("romance-italian"))
		_, err := n.Format()
		assertErr(t, err, nil)
		got, err := n.FormatName()
		require.NoError(t, err, "a name-only notation formats the name")
		assert.Equal(t, stdNotation("romance-italian").MustFormat(spelling.CSharp()), got)

		x := mk(t, spi(letter.C, 3), 4, ct(0), std())
		_, err = x.Format()
		assertErr(t, err, nil)
	})

	t.Run("format of an inadmissible name", func(t *testing.T) {
		t.Parallel()
		n, _ := inadmissibleNote(t)
		_, err := n.Format()
		assertErr(t, err, note.ErrInadmissible, settings.ErrInadmissible, system.ErrInadmissible)
		_, err = n.FormatName()
		assertErr(t, err, note.ErrInadmissible, settings.ErrInadmissible, system.ErrInadmissible)
	})

	t.Run("class of a note without octave", func(t *testing.T) {
		t.Parallel()
		n := mkName(t, spelling.CSharp(), ct(0), std())
		got := n.MustClass()
		assert.True(t, got.Equal(system.Traditional().MustClass(spelling.CSharp())))
		got = n.MustClassInTuning(tuning19())
		assert.True(t, got.Equal(system.Positions19().MustClass(spelling.CSharp())))
	})
}
