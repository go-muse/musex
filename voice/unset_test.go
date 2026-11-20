package voice_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/require"
)

// lineOp is a method of *voice.Line that returns an error, with its Must
// variant; the arguments are valid, so only the receiver can fail.
type lineOp struct {
	name string
	call func(l *voice.Line) (any, error)
	must func(l *voice.Line) any
}

func anyOf[T any](v T, err error) (any, error) { return v, err }

func anyOfTwo[A, B any](a A, b B, err error) (any, error) { return pair[A, B]{A: a, B: b}, err }

// lineOps lists every method of *voice.Line that returns an error.
func lineOps() []lineOp {
	ref := func() voice.Reference { return voice.PresetReference(soprano()) }
	sys := traditional()
	m := map60()
	third := interval.MajorThird()
	return []lineOp{
		{"Len", func(l *voice.Line) (any, error) { return anyOf(l.Len()) }, func(l *voice.Line) any { return l.MustLen() }},
		{"IsEmpty", func(l *voice.Line) (any, error) { return anyOf(l.IsEmpty()) }, func(l *voice.Line) any { return l.MustIsEmpty() }},
		{"SoundCount", func(l *voice.Line) (any, error) { return anyOf(l.SoundCount()) }, func(l *voice.Line) any { return l.MustSoundCount() }},
		{"RestCount", func(l *voice.Line) (any, error) { return anyOf(l.RestCount()) }, func(l *voice.Line) any { return l.MustRestCount() }},
		{"Element", func(l *voice.Line) (any, error) { return anyOf(l.Element(0)) }, func(l *voice.Line) any { return l.MustElement(0) }},
		{"Elements", func(l *voice.Line) (any, error) { return anyOf(l.Elements()) }, func(l *voice.Line) any { return l.MustElements() }},
		{"Notes", func(l *voice.Line) (any, error) { return anyOf(l.Notes()) }, func(l *voice.Line) any { return l.MustNotes() }},
		{"Name", func(l *voice.Line) (any, error) { return anyOf(l.Name()) }, func(l *voice.Line) any { return l.MustName() }},
		{"Placement", func(l *voice.Line) (any, error) { return anyOf(l.Placement()) }, func(l *voice.Line) any { return l.MustPlacement() }},
		{"Thematic", func(l *voice.Line) (any, error) { return anyOf(l.Thematic()) }, func(l *voice.Line) any { return l.MustThematic() }},
		{"Importance", func(l *voice.Line) (any, error) { return anyOf(l.Importance()) }, func(l *voice.Line) any { return l.MustImportance() }},
		{"StartOf", func(l *voice.Line) (any, error) { return anyOf(l.StartOf(0)) }, func(l *voice.Line) any { return l.MustStartOf(0) }},
		{"EndOf", func(l *voice.Line) (any, error) { return anyOf(l.EndOf(0)) }, func(l *voice.Line) any { return l.MustEndOf(0) }},
		{"Starts", func(l *voice.Line) (any, error) { return anyOfTwo(l.Starts()) }, func(l *voice.Line) any { return mustTwo(l.MustStarts()) }},
		{"Length", func(l *voice.Line) (any, error) { return anyOf(l.Length()) }, func(l *voice.Line) any { return l.MustLength() }},
		{"End", func(l *voice.Line) (any, error) { return anyOf(l.End()) }, func(l *voice.Line) any { return l.MustEnd() }},
		{"At", func(l *voice.Line) (any, error) { return anyOfTwo(l.At(pos(0, 1))) }, func(l *voice.Line) any { return mustTwo(l.MustAt(pos(0, 1))) }},
		{"SoundAt", func(l *voice.Line) (any, error) { return anyOfTwo(l.SoundAt(pos(0, 1))) },
			func(l *voice.Line) any { return mustTwo(l.MustSoundAt(pos(0, 1))) }},
		{"Span", func(l *voice.Line) (any, error) { return anyOf(l.Span(pos(0, 1), pos(1, 1))) },
			func(l *voice.Line) any { return l.MustSpan(pos(0, 1), pos(1, 1)) }},
		{"Sounds", func(l *voice.Line) (any, error) { return anyOf(l.Sounds()) }, func(l *voice.Line) any { return l.MustSounds() }},
		{"Fermatas", func(l *voice.Line) (any, error) { return anyOf(l.Fermatas()) }, func(l *voice.Line) any { return l.MustFermatas() }},
		{"FermatasFrom", func(l *voice.Line) (any, error) { return anyOf(l.FermatasFrom(pos(1, 1))) },
			func(l *voice.Line) any { return l.MustFermatasFrom(pos(1, 1)) }},
		{"SoundSeconds", func(l *voice.Line) (any, error) { return anyOf(l.SoundSeconds(m, 3, rounding.HalfUp)) },
			func(l *voice.Line) any { return l.MustSoundSeconds(m, 3, rounding.HalfUp) }},
		{"SoundSecondsFrom", func(l *voice.Line) (any, error) { return anyOf(l.SoundSecondsFrom(m, pos(1, 1), 3, rounding.HalfUp)) },
			func(l *voice.Line) any { return l.MustSoundSecondsFrom(m, pos(1, 1), 3, rounding.HalfUp) }},
		{"ElementSeconds", func(l *voice.Line) (any, error) { return anyOf(l.ElementSeconds(0, m, 3, rounding.HalfUp)) },
			func(l *voice.Line) any { return l.MustElementSeconds(0, m, 3, rounding.HalfUp) }},
		{"ElementSecondsFrom", func(l *voice.Line) (any, error) {
			return anyOf(l.ElementSecondsFrom(0, m, pos(1, 1), 3, rounding.HalfUp))
		}, func(l *voice.Line) any { return l.MustElementSecondsFrom(0, m, pos(1, 1), 3, rounding.HalfUp) }},
		{"Intervals", func(l *voice.Line) (any, error) { return anyOf(l.Intervals()) }, func(l *voice.Line) any { return l.MustIntervals() }},
		{"IntervalsIn", func(l *voice.Line) (any, error) { return anyOf(l.IntervalsIn(sys)) }, func(l *voice.Line) any { return l.MustIntervalsIn(sys) }},
		{"Range", func(l *voice.Line) (any, error) { return anyOf(l.Range()) }, func(l *voice.Line) any { return l.MustRange() }},
		{"RangeIn", func(l *voice.Line) (any, error) { return anyOf(l.RangeIn(sys)) }, func(l *voice.Line) any { return l.MustRangeIn(sys) }},
		{"Extremes", func(l *voice.Line) (any, error) { return anyOfTwo(l.Extremes()) }, func(l *voice.Line) any { return mustTwo(l.MustExtremes()) }},
		{"ExtremesIn", func(l *voice.Line) (any, error) { return anyOfTwo(l.ExtremesIn(sys)) },
			func(l *voice.Line) any { return mustTwo(l.MustExtremesIn(sys)) }},
		{"Ambitus", func(l *voice.Line) (any, error) { return anyOf(l.Ambitus()) }, func(l *voice.Line) any { return l.MustAmbitus() }},
		{"AmbitusIn", func(l *voice.Line) (any, error) { return anyOf(l.AmbitusIn(sys)) }, func(l *voice.Line) any { return l.MustAmbitusIn(sys) }},
		{"Locate", func(l *voice.Line) (any, error) { return anyOf(l.Locate(ref())) }, func(l *voice.Line) any { return l.MustLocate(ref()) }},
		{"LocateIn", func(l *voice.Line) (any, error) { return anyOf(l.LocateIn(sys, ref())) },
			func(l *voice.Line) any { return l.MustLocateIn(sys, ref()) }},
		{"Within", func(l *voice.Line) (any, error) { return anyOfTwo(l.Within(ref())) }, func(l *voice.Line) any { return mustTwo(l.MustWithin(ref())) }},
		{"WithinIn", func(l *voice.Line) (any, error) { return anyOfTwo(l.WithinIn(sys, ref())) },
			func(l *voice.Line) any { return mustTwo(l.MustWithinIn(sys, ref())) }},
		{"Tessitura", func(l *voice.Line) (any, error) { return anyOf(l.Tessitura(ref())) }, func(l *voice.Line) any { return l.MustTessitura(ref()) }},
		{"TessituraIn", func(l *voice.Line) (any, error) { return anyOf(l.TessituraIn(sys, ref())) },
			func(l *voice.Line) any { return l.MustTessituraIn(sys, ref()) }},
		{"TessituraBy", func(l *voice.Line) (any, error) {
			return anyOf(l.TessituraBy(ref(), voice.ByCount, voice.DefaultZones()))
		},
			func(l *voice.Line) any { return l.MustTessituraBy(ref(), voice.ByCount, voice.DefaultZones()) }},
		{"TessituraByIn", func(l *voice.Line) (any, error) {
			return anyOf(l.TessituraByIn(sys, ref(), voice.ByCount, voice.DefaultZones()))
		}, func(l *voice.Line) any { return l.MustTessituraByIn(sys, ref(), voice.ByCount, voice.DefaultZones()) }},
		{"Transpose", func(l *voice.Line) (any, error) { return anyOf(l.Transpose(third)) }, func(l *voice.Line) any { return l.MustTranspose(third) }},
		{"TransposeIn", func(l *voice.Line) (any, error) { return anyOf(l.TransposeIn(sys, third)) },
			func(l *voice.Line) any { return l.MustTransposeIn(sys, third) }},
		{"Double", func(l *voice.Line) (any, error) { return anyOf(l.Double(third)) }, func(l *voice.Line) any { return l.MustDouble(third) }},
		{"DoubleIn", func(l *voice.Line) (any, error) { return anyOf(l.DoubleIn(sys, third)) },
			func(l *voice.Line) any { return l.MustDoubleIn(sys, third) }},
		{"Imitate", func(l *voice.Line) (any, error) { return anyOf(l.Imitate(third)) }, func(l *voice.Line) any { return l.MustImitate(third) }},
		{"ImitateIn", func(l *voice.Line) (any, error) { return anyOf(l.ImitateIn(sys, third)) },
			func(l *voice.Line) any { return l.MustImitateIn(sys, third) }},
		{"ImitateAfter", func(l *voice.Line) (any, error) { return anyOf(l.ImitateAfter(third, q())) },
			func(l *voice.Line) any { return l.MustImitateAfter(third, q()) }},
		{"ImitateAfterIn", func(l *voice.Line) (any, error) { return anyOf(l.ImitateAfterIn(sys, third, q())) },
			func(l *voice.Line) any { return l.MustImitateAfterIn(sys, third, q()) }},
		{"Invert", func(l *voice.Line) (any, error) { return anyOf(l.Invert(c4())) }, func(l *voice.Line) any { return l.MustInvert(c4()) }},
		{"InvertIn", func(l *voice.Line) (any, error) { return anyOf(l.InvertIn(sys, c4())) },
			func(l *voice.Line) any { return l.MustInvertIn(sys, c4()) }},
		{"InvertPitch", func(l *voice.Line) (any, error) { return anyOf(l.InvertPitch(pt(spelling.C(), 4))) },
			func(l *voice.Line) any { return l.MustInvertPitch(pt(spelling.C(), 4)) }},
		{"InvertPitchIn", func(l *voice.Line) (any, error) { return anyOf(l.InvertPitchIn(sys, pt(spelling.C(), 4))) },
			func(l *voice.Line) any { return l.MustInvertPitchIn(sys, pt(spelling.C(), 4)) }},
		{"Retrograde", func(l *voice.Line) (any, error) { return anyOf(l.Retrograde(voice.RetrogradeBoth)) },
			func(l *voice.Line) any { return l.MustRetrograde(voice.RetrogradeBoth) }},
		{"Append", func(l *voice.Line) (any, error) { return anyOf(l.Append(rst(q()))) }, func(l *voice.Line) any { return l.MustAppend(rst(q())) }},
		{"Insert", func(l *voice.Line) (any, error) { return anyOf(l.Insert(0, rst(q()))) },
			func(l *voice.Line) any { return l.MustInsert(0, rst(q())) }},
		{"Replace", func(l *voice.Line) (any, error) { return anyOf(l.Replace(0, rst(q()))) },
			func(l *voice.Line) any { return l.MustReplace(0, rst(q())) }},
		{"Remove", func(l *voice.Line) (any, error) { return anyOf(l.Remove(0)) }, func(l *voice.Line) any { return l.MustRemove(0) }},
		{"RemoveRange", func(l *voice.Line) (any, error) { return anyOf(l.RemoveRange(0, 0)) },
			func(l *voice.Line) any { return l.MustRemoveRange(0, 0) }},
		{"Slice", func(l *voice.Line) (any, error) { return anyOf(l.Slice(0, 0)) }, func(l *voice.Line) any { return l.MustSlice(0, 0) }},
		{"Split", func(l *voice.Line) (any, error) { return anyOfTwo(l.Split(0)) }, func(l *voice.Line) any { return mustTwo(l.MustSplit(0)) }},
		{"WithName", func(l *voice.Line) (any, error) { return anyOf(l.WithName("x")) }, func(l *voice.Line) any { return l.MustWithName("x") }},
		{"WithPlacement", func(l *voice.Line) (any, error) { return anyOf(l.WithPlacement(voice.Bass)) },
			func(l *voice.Line) any { return l.MustWithPlacement(voice.Bass) }},
		{"WithThematic", func(l *voice.Line) (any, error) { return anyOf(l.WithThematic(voice.Pedal)) },
			func(l *voice.Line) any { return l.MustWithThematic(voice.Pedal) }},
		{"WithImportance", func(l *voice.Line) (any, error) { return anyOf(l.WithImportance(voice.Filler)) },
			func(l *voice.Line) any { return l.MustWithImportance(voice.Filler) }},
		{"WithNote", func(l *voice.Line) (any, error) { return anyOf(l.WithNote(0, d4())) },
			func(l *voice.Line) any { return l.MustWithNote(0, d4()) }},
		{"WithDuration", func(l *voice.Line) (any, error) { return anyOf(l.WithDuration(0, h())) },
			func(l *voice.Line) any { return l.MustWithDuration(0, h()) }},
		{"WithFermata", func(l *voice.Line) (any, error) { return anyOf(l.WithFermata(0, true)) },
			func(l *voice.Line) any { return l.MustWithFermata(0, true) }},
		{"Events", func(l *voice.Line) (any, error) { return anyOf(l.Events()) }, func(l *voice.Line) any { return l.MustEvents() }},
	}
}

// Architecture 6.1, П-1: every method of *voice.Line that returns an error
// fails with Е-1 alone for the nil pointer and for the zero Line, before its
// arguments are checked; its Must variant panics with that error.
func TestUnsetReceiver(t *testing.T) {
	t.Parallel()
	receivers := map[string]func() *voice.Line{
		"nil":  func() *voice.Line { return nil },
		"zero": func() *voice.Line { return &voice.Line{} },
	}
	for receiver, build := range receivers {
		for _, op := range lineOps() {
			t.Run(receiver+" "+op.name, func(t *testing.T) {
				t.Parallel()
				l := build()
				assertFailure(t, fails(voice.ErrUnset).only(),
					func() (any, error) { return op.call(l) },
					func() any { return op.must(l) })
			})
		}
	}
}

// The same methods succeed on a set line with these arguments: the table
// above exercises the receiver check only.
func TestLineOpsSucceedOnSetLine(t *testing.T) {
	t.Parallel()
	for _, op := range lineOps() {
		t.Run(op.name, func(t *testing.T) {
			t.Parallel()
			l := line(snd(c4(), q()), rst(q()), snd(e4(), h()))
			_, err := op.call(l)
			require.NoError(t, err)
		})
	}
}
