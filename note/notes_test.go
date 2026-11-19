package note_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exact method-expression types prevent an unnoticed []Note return type:
// ordinary slice assignment alone would accept it without a conversion.
var (
	_ func(note.Notes) note.Notes                           = note.Notes.Clone
	_ func(note.Notes) (note.Notes, error)                  = note.Notes.SortedByPosition
	_ func(note.Notes) note.Notes                           = note.Notes.MustSortedByPosition
	_ func(note.Notes, settings.Source) (note.Notes, error) = note.Notes.SortedByPositionIn
	_ func(note.Notes, settings.Source) note.Notes          = note.Notes.MustSortedByPositionIn
	_ func(note.Notes, *tuning.Tuning) (note.Notes, error)  = note.Notes.SortedByPositionInTuning
	_ func(note.Notes, *tuning.Tuning) note.Notes           = note.Notes.MustSortedByPositionInTuning
	_ func(note.Notes) (note.Notes, error)                  = note.Notes.SortedBySound
	_ func(note.Notes) note.Notes                           = note.Notes.MustSortedBySound
	_ func(note.Notes, settings.Source) (note.Notes, error) = note.Notes.SortedBySoundIn
	_ func(note.Notes, settings.Source) note.Notes          = note.Notes.MustSortedBySoundIn
	_ func(note.Notes, *tuning.Tuning) (note.Notes, error)  = note.Notes.SortedBySoundInTuning
	_ func(note.Notes, *tuning.Tuning) note.Notes           = note.Notes.MustSortedBySoundInTuning
)

// Task 16.2, architecture 5.2: Notes is an ordinary borrowed slice until Clone.
func TestNotesSliceAndClone(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"nil", "literal empty", "make empty", "populated"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			a := mk(t, spelling.C(), 4, ct(17), src)
			b := mkName(t, spelling.DFlat(), ct(-9), nil)
			var ns note.Notes
			switch name {
			case "literal empty":
				ns = note.Notes{}
			case "make empty":
				ns = make(note.Notes, 0)
			case "populated":
				ns = note.Notes{a, {}, b, a}
			}
			before := slices.Clone([]note.Note(ns))
			reads := src.reads()
			cloned := ns.Clone()
			assert.IsType(t, note.Notes(nil), any(cloned))
			assert.Equal(t, ns == nil, cloned == nil)
			assertNotesPreserved(t, before, cloned)
			assertNotesPreserved(t, before, ns)
			assert.Equal(t, reads, src.reads(), "Clone does not read sources")
			if len(ns) > 0 {
				cloned[0].MustSetCents(ct(31))
				cloned[2] = a
				assertNotesPreserved(t, before, ns)
				copyBefore := slices.Clone([]note.Note(cloned))
				ns[0].MustSetCents(ct(42))
				ns[2] = note.Note{}
				assertNotesPreserved(t, copyBefore, cloned)
				assert.True(t, cloned[3].Source() == src)
			}
		})
	}
	t.Run("assignment conversion and subslice share elements", func(t *testing.T) {
		t.Parallel()
		a := mkName(t, spelling.C(), ct(0), nil)
		b := mkName(t, spelling.D(), ct(0), nil)
		raw := []note.Note{a, b, a}
		var ns note.Notes = raw
		var back []note.Note = ns
		converted := note.Notes(raw)
		alias := ns
		alias[0] = b
		assertNotesPreserved(t, []note.Note{b, b, a}, raw)
		sub := ns[1:]
		sub[0].MustSetCents(ct(7))
		assert.True(t, back[1].Cents().Equal(ct(7)))
		converted[2] = b
		assert.True(t, ns[2].Equal(b))
		built := make(note.Notes, 0, 2)
		built = append(built, a, b, a)
		var ranged []note.Note
		for i, n := range built {
			assert.True(t, n.Equal(built[i]))
			ranged = append(ranged, n)
		}
		assertNotesPreserved(t, []note.Note{a, b, a}, ranged)
	})
	t.Run("append to empty subslice clone cannot overwrite live input", func(t *testing.T) {
		t.Parallel()
		a := mkName(t, spelling.C(), ct(0), nil)
		b := mkName(t, spelling.D(), ct(0), nil)
		ns := note.Notes{a, b}
		cloned := ns[:0].Clone()
		require.NotNil(t, cloned)
		cloned = append(cloned, b)
		assertNotesPreserved(t, []note.Note{a, b}, ns)
		assertNotesPreserved(t, []note.Note{b}, cloned)
	})
}

// Compare record values and source identity separately. Unset notes are not
// reflexive under Note.Equal; no source is read by these observations.
func assertNotesPreserved(t *testing.T, expected, actual []note.Note) {
	t.Helper()
	require.Len(t, actual, len(expected))
	for i, want := range expected {
		assert.Equal(t, want.IsValid(), actual[i].IsValid(), "note %d", i)
		if want.IsValid() {
			assert.True(t, want.Equal(actual[i]), "record at %d changed", i)
		}
		assert.True(t, want.Source() == actual[i].Source(), "source identity at %d changed", i)
	}
}

type notesSortCase struct {
	name     string
	sound    bool
	explicit bool
	run      func(note.Notes, settings.Source, *tuning.Tuning) (note.Notes, error)
	must     func(note.Notes, settings.Source, *tuning.Tuning) note.Notes
}

func notesSortCases() []notesSortCase {
	return []notesSortCase{
		{"position", false, false,
			func(ns note.Notes, _ settings.Source, _ *tuning.Tuning) (note.Notes, error) {
				return ns.SortedByPosition()
			},
			func(ns note.Notes, _ settings.Source, _ *tuning.Tuning) note.Notes { return ns.MustSortedByPosition() }},
		{"position In", false, true,
			func(ns note.Notes, src settings.Source, _ *tuning.Tuning) (note.Notes, error) {
				return ns.SortedByPositionIn(src)
			},
			func(ns note.Notes, src settings.Source, _ *tuning.Tuning) note.Notes {
				return ns.MustSortedByPositionIn(src)
			}},
		{"position InTuning", false, true,
			func(ns note.Notes, _ settings.Source, tu *tuning.Tuning) (note.Notes, error) {
				return ns.SortedByPositionInTuning(tu)
			},
			func(ns note.Notes, _ settings.Source, tu *tuning.Tuning) note.Notes {
				return ns.MustSortedByPositionInTuning(tu)
			}},
		{"sound", true, false,
			func(ns note.Notes, _ settings.Source, _ *tuning.Tuning) (note.Notes, error) {
				return ns.SortedBySound()
			},
			func(ns note.Notes, _ settings.Source, _ *tuning.Tuning) note.Notes { return ns.MustSortedBySound() }},
		{"sound In", true, true,
			func(ns note.Notes, src settings.Source, _ *tuning.Tuning) (note.Notes, error) {
				return ns.SortedBySoundIn(src)
			},
			func(ns note.Notes, src settings.Source, _ *tuning.Tuning) note.Notes {
				return ns.MustSortedBySoundIn(src)
			}},
		{"sound InTuning", true, true,
			func(ns note.Notes, _ settings.Source, tu *tuning.Tuning) (note.Notes, error) {
				return ns.SortedBySoundInTuning(tu)
			},
			func(ns note.Notes, _ settings.Source, tu *tuning.Tuning) note.Notes {
				return ns.MustSortedBySoundInTuning(tu)
			}},
	}
}

// All twelve forms share independent oracles from note K-4, K-48 and K-50.
func TestNotesSortValuesStabilityAndOwnership(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		for _, must := range []bool{false, true} {
			suffix := " ordinary"
			if must {
				suffix = " Must"
			}
			t.Run(method.name+suffix, func(t *testing.T) {
				t.Parallel()
				a, b, explicit := newCounting(std()), newCounting(std()), newCounting(std())
				c150 := mk(t, spelling.C(), 4, ct(150), a)
				cs := mk(t, spelling.CSharp(), 4, ct(0), b)
				db := mk(t, spelling.DFlat(), 4, ct(0), a)
				e := mk(t, spelling.E(), 4, ct(-3), b)
				input := note.Notes{e, db, c150, cs, db}
				before := slices.Clone([]note.Note(input))
				want := []note.Note{c150, db, cs, db, e}
				if method.sound {
					want = []note.Note{db, cs, db, c150, e}
				}
				ar, br := a.reads(), b.reads()
				var got note.Notes
				if must {
					got = method.must(input, explicit, tuning.Standard())
				} else {
					var err error
					got, err = method.run(input, explicit, tuning.Standard())
					require.NoError(t, err)
				}
				assert.IsType(t, note.Notes(nil), any(got))
				assertNotesPreserved(t, want, got)
				assertNotesPreserved(t, before, input)
				ownReads := int64(1)
				if method.explicit {
					ownReads = 0
				}
				assert.Equal(t, ownReads, a.reads()-ar)
				assert.Equal(t, ownReads, b.reads()-br)
				explicitReads := int64(0)
				if method.name == "position In" || method.name == "sound In" {
					explicitReads = 1
				}
				assert.Equal(t, explicitReads, explicit.reads())
				got[0].MustSetCents(ct(93))
				got[1] = note.Note{}
				assertNotesPreserved(t, before, input)
				outBefore := slices.Clone([]note.Note(got))
				input[0].MustSetCents(ct(94))
				input[1] = note.Note{}
				assertNotesPreserved(t, outBefore, got)
			})
		}
	}
}

func TestNotesSortEmptyAndSingleton(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			for _, input := range []note.Notes{nil, {}, make(note.Notes, 0)} {
				src := newCounting(nil)
				got, err := method.run(input, src, tuning.Standard())
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Empty(t, got)
				must := method.must(input, src, tuning.Standard())
				assert.NotNil(t, must)
				assert.Empty(t, must)
				assert.Zero(t, src.reads())
			}
			src := newCounting(std())
			n := mk(t, spelling.A(), 4, ct(12), src)
			input := note.Notes{n}
			got, err := method.run(input, std(), tuning.Standard())
			require.NoError(t, err)
			assertNotesPreserved(t, []note.Note{n}, got)
			got[0] = note.Note{}
			assertNotesPreserved(t, []note.Note{n}, input)
			must := method.must(input, std(), tuning.Standard())
			assertNotesPreserved(t, []note.Note{n}, must)
			must[0] = note.Note{}
			assertNotesPreserved(t, []note.Note{n}, input)
		})
	}
}

func TestNotesSortExactAndFrequencyKeys(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		if !method.sound {
			continue
		}
		t.Run(method.name+" exact", func(t *testing.T) {
			t.Parallel()
			requireTinyMerges(t)
			hi := mk(t, spelling.C(), 4, tiny(), std())
			lo := mk(t, spelling.C(), 4, ct(0), std())
			input := note.Notes{hi, lo, hi}
			got, err := method.run(input, std(), tuning.Standard())
			require.NoError(t, err)
			assertNotesPreserved(t, []note.Note{lo, hi, hi}, got)
			assertNotesPreserved(t, []note.Note{lo, hi, hi}, method.must(input, std(), tuning.Standard()))
		})
	}
	t.Run("mixed tunings use one frequency key and stable ties", func(t *testing.T) {
		t.Parallel()
		requireTinyMerges(t)
		hi := mk(t, spelling.C(), 4, tiny(), std())
		lo := mk(t, spelling.C(), 4, ct(0), std())
		c19 := mk(t, spelling.C(), 4, ct(0), set19())
		cs := mk(t, spelling.CSharp(), 4, ct(0), std())
		input := note.Notes{cs, c19, hi, lo, hi}
		want := []note.Note{hi, lo, hi, c19, cs}
		got, err := input.SortedBySound()
		require.NoError(t, err)
		assertNotesPreserved(t, want, got)
		assertNotesPreserved(t, want, input.MustSortedBySound())
	})
}

func TestNotesSortSourceSnapshots(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			src := &alternatingSource{a: std(), b: set19()}
			e := mk(t, spelling.E(), 4, ct(0), src)
			c := mk(t, spelling.C(), 4, ct(0), src)
			d := mk(t, spelling.D(), 4, ct(0), src)
			input := note.Notes{e, c, d, c}
			for _, must := range []bool{false, true} {
				src.calls.Store(0)
				explicit := &alternatingSource{a: std(), b: nil}
				var got note.Notes
				if must {
					got = method.must(input, explicit, tuning.Standard())
				} else {
					var err error
					got, err = method.run(input, explicit, tuning.Standard())
					require.NoError(t, err)
				}
				assertNotesPreserved(t, []note.Note{c, c, d, e}, got)
				wantOwn := int64(1)
				if method.explicit {
					wantOwn = 0
				}
				assert.Equal(t, wantOwn, src.calls.Load())
				wantExplicit := int64(0)
				if method.name == "position In" || method.name == "sound In" {
					wantExplicit = 1
				}
				assert.Equal(t, wantExplicit, explicit.calls.Load())
			}
		})
	}
}

func TestNotesSortExplicitIgnoresOwnContext(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		if !method.explicit {
			continue
		}
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			own := newCounting(std())
			c := mk(t, spelling.C(), 4, ct(0), own)
			e := mk(t, spelling.E(), 4, ct(0), own)
			own.set = nil
			reads := own.reads()
			input := note.Notes{e, c, c}
			got, err := method.run(input, std(), tuning.Standard())
			require.NoError(t, err)
			assertNotesPreserved(t, []note.Note{c, c, e}, got)
			assertNotesPreserved(t, []note.Note{c, c, e}, method.must(input, std(), tuning.Standard()))
			assert.Equal(t, reads, own.reads(), "explicit read-only sorting ignores stale own settings")
		})
	}
}

type notesSortFailure struct {
	name  string
	input note.Notes
	src   settings.Source
	tu    *tuning.Tuning
	want  error
	extra []error
}

func TestNotesSortFailuresAndPriority(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			bound := mk(t, spelling.C(), 4, ct(0), std())
			free := mk(t, spelling.C(), 4, ct(0), nil)
			noOct := mkName(t, spelling.C(), ct(0), nil)
			stale := newCounting(std())
			staleNote := mk(t, spelling.C(), 4, ct(0), stale)
			stale.set = nil
			inadm, _ := inadmissibleNote(t)
			cases := []notesSortFailure{
				{"unset beats earlier missing octave", note.Notes{noOct, {}}, std(), tuning.Standard(), note.ErrUnset, nil},
				{"missing octave beats earlier unbound note", note.Notes{free, noOct}, std(), tuning.Standard(), note.ErrNoOctave, nil},
			}
			if !method.explicit {
				cases = append(cases,
					notesSortFailure{"no source before unset settings", note.Notes{staleNote, free}, nil, nil, note.ErrNoSource, nil},
					notesSortFailure{"unset settings before inadmissible name", note.Notes{inadm, staleNote}, nil, nil, note.ErrUnsetSettings, []error{settings.ErrUnset}},
					notesSortFailure{"inadmissible", note.Notes{bound, inadm}, nil, nil, note.ErrInadmissible, []error{system.ErrInadmissible}})
				if !method.sound {
					other := mk(t, spelling.C(), 4, ct(0), set19())
					cases = append(cases, notesSortFailure{"different systems", note.Notes{bound, other}, nil, nil, note.ErrDifferentSystems, nil})
				}
				if method.sound {
					huge := mk(t, spelling.C(), 2000, ct(0), std())
					cases = append(cases, notesSortFailure{"inadmissible before earlier frequency overflow", note.Notes{huge, inadm}, nil, nil, note.ErrInadmissible, []error{system.ErrInadmissible}})
				}
			} else {
				cases = append(cases,
					notesSortFailure{"nil context on empty", nil, nil, nil, note.ErrUnset, nil},
					notesSortFailure{"unset tuning on empty", note.Notes{}, nil, &tuning.Tuning{}, note.ErrUnset, nil},
					notesSortFailure{"unset context beats no octave", note.Notes{noOct}, nil, nil, note.ErrUnset, nil},
					notesSortFailure{"inadmissible explicit name", note.Notes{mk(t, cHalf(), 4, ct(0), nil)}, std(), tuning.Standard(), note.ErrInadmissible, []error{system.ErrInadmissible}})
				if method.name == "position In" || method.name == "sound In" {
					cases = append(cases, notesSortFailure{"unset explicit settings", note.Notes{free}, nilSettingsSource(), nil, note.ErrUnsetSettings, []error{settings.ErrUnset}})
				}
			}
			for _, tt := range cases {
				t.Run(tt.name, func(t *testing.T) {
					before := slices.Clone([]note.Note(tt.input))
					reads := stale.reads()
					got, err := method.run(tt.input, tt.src, tt.tu)
					assert.Nil(t, got, "no partial result")
					assertErr(t, err, tt.want, tt.extra...)
					assertPanicErr(t, func() { method.must(tt.input, tt.src, tt.tu) }, tt.want, tt.extra...)
					assertNotesPreserved(t, before, tt.input)
					if tt.want == note.ErrUnset || tt.want == note.ErrNoOctave || tt.want == note.ErrNoSource {
						assert.Equal(t, reads, stale.reads(), "early validation does not read sources")
					}
				})
			}
		})
	}
}

func TestNotesSoundKeyBounds(t *testing.T) {
	t.Parallel()
	for _, method := range notesSortCases() {
		if !method.sound {
			continue
		}
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			// CentsFromReference(C4) is negative; adding the lower cents
			// boundary leaves the supported exact-key range (note 6.10.3).
			n := mk(t, spelling.C(), 4, cents.MustFromInt(-9_000_000), std())
			input := note.Notes{n}
			got, err := method.run(input, std(), tuning.Standard())
			assert.Nil(t, got)
			assertErr(t, err, note.ErrOutOfRange, cents.ErrOutOfRange)
			assertPanicErr(t, func() { method.must(input, std(), tuning.Standard()) }, note.ErrOutOfRange, cents.ErrOutOfRange)
			assertNotesPreserved(t, []note.Note{n}, input)
		})
	}
	for _, aspn := range []int{2000, -2000} {
		name := "mixed tunings frequency overflow"
		if aspn < 0 {
			name = "mixed tunings frequency underflow"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			input := note.Notes{mk(t, spelling.C(), aspn, ct(0), std()), mk(t, spelling.C(), 4, ct(0), set19())}
			before := slices.Clone([]note.Note(input))
			got, err := input.SortedBySound()
			assert.Nil(t, got)
			assertErr(t, err, note.ErrOutOfRange, tuning.ErrOutOfRange)
			assertPanicErr(t, func() { input.MustSortedBySound() }, note.ErrOutOfRange, tuning.ErrOutOfRange)
			assertNotesPreserved(t, before, input)
		})
	}
}
