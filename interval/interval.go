package interval

import (
	"math"

	"github.com/go-muse/musex/alteration"
)

// MaxSize is the largest interval size (number of letter steps plus one).
const MaxSize = math.MaxInt

// Interval is a named interval: a direction, a size, and a quality.
//
// The size n >= 1 is the number of letters the interval spans, counting both
// ends: 1 for a unison, 3 for a third, 8 for an octave, 10 for a tenth.
// The quality is stored as the deviation d from the basic kind of the size,
// an alteration value: the basic kind is perfect for the perfect group and
// major for the major-minor group (see Group), and d says by how many
// chromatic units the interval is wider. The direction is up or down;
// the perfect unison has no direction of its own and is always up.
//
// The zero value is the unset interval: it means that no interval is given,
// it is not the perfect unison, and Interval.IsValid reports false for it.
// Intervals are created by New, NewMultiple, FromDeviation, the ready-made
// intervals such as MajorThird, Predefined, Parse, ParseName, and the results
// of operations. Interval has only unexported fields, so a literal of another
// package can create only the unset interval.
//
// Interval is a comparable value type of fixed size: it can be copied freely,
// a copy shares no data with the original, and it can be used as a map key.
// For set intervals == and Interval.Equal agree; the unset interval is equal
// to itself under == but not under Equal. Methods have value receivers and
// never change the receiver or the arguments, except Interval.UnmarshalText,
// which writes into its receiver on success.
type Interval struct {
	size int
	dev  alteration.Alteration
	down bool
}

// newInterval returns the interval (n, d) in the given direction in canonical
// form: a unison that is not wider than perfect is ascending. The caller
// ensures n >= 1 and, for n = 1, d >= 0.
func newInterval(n int, d alteration.Alteration, down bool) Interval {
	if n == 1 && d.Sign() <= 0 {
		down = false
	}
	return Interval{size: n, dev: d, down: down}
}

// simpleSizeOf returns the size of the simple part of an interval of size
// n >= 1: 1 for a unison, otherwise 2 to 8.
func simpleSizeOf(n int) int {
	if n == 1 {
		return 1
	}
	return (n-2)%7 + 2
}

// perfectSize reports whether the size n >= 1 belongs to the perfect group.
func perfectSize(n int) bool {
	switch simpleSizeOf(n) {
	case 1, 4, 5, 8:
		return true
	default:
		return false
	}
}

// basicOffset returns the deviation of the diminished interval of
// multiplicity 1 below the basic kind: 1 in the perfect group and 2 in the
// major-minor group. The minor kind is 1 below the major kind.
func basicOffset(perfect bool) int {
	if perfect {
		return 1
	}
	return 2
}

// neutralDeviation is the deviation of the neutral quality: -1/2.
var neutralDeviation = alteration.MustFromFraction(-1, 2)

// qualityOf returns the kind and the multiplicity of the quality with the
// deviation d in the group given by perfect.
func qualityOf(perfect bool, d alteration.Alteration) (Quality, int) {
	if !d.IsInteger() {
		if !perfect && d == neutralDeviation {
			return Neutral, 0
		}
		return Fractional, 0
	}
	k, _ := d.Fraction()
	switch {
	case k == 0 && perfect:
		return Perfect, 0
	case k == 0:
		return Major, 0
	case k == -1 && !perfect:
		return Minor, 0
	case k > 0:
		return Augmented, k
	case perfect:
		return Diminished, -k
	default:
		return Diminished, -1 - k
	}
}

// New returns the ascending interval of the given quality and size n, with
// multiplicity 1 for augmented and diminished intervals: perfect and major
// have deviation 0, minor -1, neutral -1/2, augmented +1, and diminished -1
// in the perfect group or -2 in the major-minor group.
//
// Errors, in this order: ErrUnset for QualityUnset; ErrInvalidArgument for
// n < 1, a quality outside the constants or Fractional, a quality that does
// not fit the group of n (perfect for the major-minor group; major, minor,
// or neutral for the perfect group), and a diminished unison. For augmented
// and diminished qualities New(q, n) equals NewMultiple(1, q, n). Use Down or
// WithDirection for a descending interval.
func New(q Quality, size int) (Interval, error) {
	const op = "New"
	if q == QualityUnset {
		return Interval{}, detailError(op, ErrUnset, "quality")
	}
	if size < 1 {
		return Interval{}, detailError(op, ErrInvalidArgument, "size less than 1")
	}
	perfect := perfectSize(size)
	var d int
	switch q {
	case Perfect:
		if !perfect {
			return Interval{}, detailError(op, ErrInvalidArgument, "perfect quality of a major-minor size")
		}
	case Major, Minor, Neutral:
		if perfect {
			return Interval{}, detailError(op, ErrInvalidArgument, q.String()+" quality of a perfect size")
		}
		switch q {
		case Minor:
			d = -1
		case Neutral:
			return Interval{size: size, dev: neutralDeviation}, nil
		}
	case Augmented, Diminished:
		return NewMultiple(1, q, size)
	default:
		return Interval{}, detailError(op, ErrInvalidArgument, "quality "+q.String())
	}
	return Interval{size: size, dev: alteration.MustFromInt(d)}, nil
}

// MustNew is like New but panics if New returns an error. The panic value is
// the error returned by New.
func MustNew(q Quality, size int) Interval {
	return must(New(q, size))
}

// NewMultiple returns the ascending augmented or diminished interval of size
// n and multiplicity k >= 1: the deviation is k for augmented and -k in the
// perfect group or -1-k in the major-minor group for diminished.
//
// Errors, in this order: ErrUnset for QualityUnset; ErrInvalidArgument for
// n < 1, k < 1, a quality other than Augmented and Diminished, and a
// diminished unison; ErrOutOfRange, together with alteration.ErrOutOfRange,
// when the deviation is outside the range of package alteration.
func NewMultiple(multiplicity int, q Quality, size int) (Interval, error) {
	const op = "NewMultiple"
	if q == QualityUnset {
		return Interval{}, detailError(op, ErrUnset, "quality")
	}
	if size < 1 {
		return Interval{}, detailError(op, ErrInvalidArgument, "size less than 1")
	}
	if multiplicity < 1 {
		return Interval{}, detailError(op, ErrInvalidArgument, "multiplicity less than 1")
	}
	if q != Augmented && q != Diminished {
		return Interval{}, detailError(op, ErrInvalidArgument, "quality "+q.String()+" with a multiplicity")
	}
	if q == Diminished && size == 1 {
		return Interval{}, detailError(op, ErrInvalidArgument, "diminished unison")
	}
	d := multiplicity
	if q == Diminished {
		// -multiplicity - (offset - 1) never overflows: multiplicity >= 1.
		d = -multiplicity - (basicOffset(perfectSize(size)) - 1)
	}
	a, err := alteration.FromInt(d)
	if err != nil {
		return Interval{}, rangeError(op, err)
	}
	return Interval{size: size, dev: a}, nil
}

// MustNewMultiple is like NewMultiple but panics if NewMultiple returns an
// error. The panic value is the error returned by NewMultiple.
func MustNewMultiple(multiplicity int, q Quality, size int) Interval {
	return must(NewMultiple(multiplicity, q, size))
}

// FromDeviation returns the ascending interval of size n whose quality
// deviates by d from the basic kind of the size (perfect or major). Any
// deviation is accepted, including fractional values, which give fractional
// qualities: FromDeviation(3, -1/2) is the neutral third.
//
// It returns ErrInvalidArgument for n < 1 and for a unison with a negative
// deviation, which would be a diminished unison.
func FromDeviation(size int, d alteration.Alteration) (Interval, error) {
	const op = "FromDeviation"
	if size < 1 {
		return Interval{}, detailError(op, ErrInvalidArgument, "size less than 1")
	}
	if size == 1 && d.Sign() < 0 {
		return Interval{}, detailError(op, ErrInvalidArgument, "diminished unison")
	}
	return Interval{size: size, dev: d}, nil
}

// MustFromDeviation is like FromDeviation but panics if FromDeviation returns
// an error. The panic value is the error returned by FromDeviation.
func MustFromDeviation(size int, d alteration.Alteration) Interval {
	return must(FromDeviation(size, d))
}

// IsValid reports whether i is set. It returns false only for the zero
// Interval.
func (i Interval) IsValid() bool {
	return i.size > 0
}

// Direction returns the direction of i: Up or Down. The perfect unison is Up.
// The unset interval returns DirectionUnset.
func (i Interval) Direction() Direction {
	switch {
	case !i.IsValid():
		return DirectionUnset
	case i.down:
		return Down
	default:
		return Up
	}
}

// IsAscending reports whether i is set and ascending. The perfect unison is
// ascending.
func (i Interval) IsAscending() bool {
	return i.IsValid() && !i.down
}

// IsDescending reports whether i is set and descending.
func (i Interval) IsDescending() bool {
	return i.IsValid() && i.down
}

// Size returns the size n of i: the number of letters it spans, counting both
// ends. The unset interval returns 0.
func (i Interval) Size() int {
	return i.size
}

// Deviation returns the deviation d of the quality of i from the basic kind of
// its size: 0 for perfect and major, -1 for minor, +1 for augmented, and so
// on. The unset interval returns the natural; the result makes sense only
// together with Interval.IsValid.
func (i Interval) Deviation() alteration.Alteration {
	return i.dev
}

// Quality returns the kind of the quality of i, given by its deviation and
// the group of its size. Whole deviations give Perfect, Major, Minor,
// Augmented, or Diminished; the deviation -1/2 in the major-minor group gives
// Neutral, and other fractional deviations give Fractional. The unset
// interval returns QualityUnset.
func (i Interval) Quality() Quality {
	if !i.IsValid() {
		return QualityUnset
	}
	q, _ := qualityOf(perfectSize(i.size), i.dev)
	return q
}

// Multiplicity returns the multiplicity k >= 1 of an augmented or diminished
// interval: 2 for a doubly augmented one. It returns 0 for other qualities and
// for the unset interval.
func (i Interval) Multiplicity() int {
	if !i.IsValid() {
		return 0
	}
	_, k := qualityOf(perfectSize(i.size), i.dev)
	return k
}

// IsIntegerQuality reports whether i is set and its deviation is a whole
// number.
func (i Interval) IsIntegerQuality() bool {
	return i.IsValid() && i.dev.IsInteger()
}

// Group returns the group of the size of i, given by its simple part:
// PerfectGroup for 1, 4, 5, and 8, MajorMinorGroup for 2, 3, 6, and 7.
// The unset interval returns GroupUnset.
func (i Interval) Group() Group {
	switch {
	case !i.IsValid():
		return GroupUnset
	case perfectSize(i.size):
		return PerfectGroup
	default:
		return MajorMinorGroup
	}
}

// SimpleSize returns the size of the simple part of i: 1 for a unison,
// otherwise ((n - 2) mod 7) + 2, from 2 to 8. An octave is its own simple
// part: P15 has the simple size 8. The unset interval returns 0.
func (i Interval) SimpleSize() int {
	if !i.IsValid() {
		return 0
	}
	return simpleSizeOf(i.size)
}

// Octaves returns the number of octaves of i beyond its simple part:
// (n - SimpleSize) / 7. The unset interval returns 0.
func (i Interval) Octaves() int {
	if !i.IsValid() {
		return 0
	}
	return (i.size - simpleSizeOf(i.size)) / 7
}

// Simple returns the simple part of i: the interval of the size SimpleSize
// with the same quality and direction. i.MustSimple().MustAddOctaves(
// i.Octaves()) equals i.
//
// It returns ErrUnset for the unset interval.
func (i Interval) Simple() (Interval, error) {
	if !i.IsValid() {
		return Interval{}, opError("Simple", ErrUnset)
	}
	return newInterval(simpleSizeOf(i.size), i.dev, i.down), nil
}

// MustSimple is like Simple but panics if Simple returns an error. The panic
// value is the error returned by Simple.
func (i Interval) MustSimple() Interval {
	return must(i.Simple())
}

// IsPerfectUnison reports whether i is the perfect unison.
func (i Interval) IsPerfectUnison() bool {
	return i.size == 1 && i.dev.IsNatural()
}

// IsSimple reports whether i is set and simple: n <= 8.
func (i Interval) IsSimple() bool {
	return i.IsValid() && i.size <= 8
}

// IsCompound reports whether i is compound: n > 8.
func (i Interval) IsCompound() bool {
	return i.size > 8
}

// IsBeyondDoubleOctave reports whether i is wider than a fifteenth: n > 15.
func (i Interval) IsBeyondDoubleOctave() bool {
	return i.size > 15
}

// isTritoneShape reports whether the interval of simple size s with the
// deviation d is an augmented fourth or a diminished fifth.
func isTritoneShape(s int, d alteration.Alteration) bool {
	return (s == 4 && d == alteration.Sharp()) || (s == 5 && d == alteration.Flat())
}

// IsBasic reports whether i is a basic interval: its quality is perfect,
// major, or minor, or its simple part is an augmented fourth or a diminished
// fifth. For a compound interval the simple part decides.
func (i Interval) IsBasic() bool {
	if !i.IsValid() {
		return false
	}
	switch i.Quality() {
	case Perfect, Major, Minor:
		return true
	}
	return isTritoneShape(simpleSizeOf(i.size), i.dev)
}

// IsDerived reports whether i is set and not basic.
func (i Interval) IsDerived() bool {
	return i.IsValid() && !i.IsBasic()
}

// IsDiatonic reports whether i is diatonic: the same as basic (perfect,
// major, and minor intervals and tritones, by the simple part).
func (i Interval) IsDiatonic() bool {
	return i.IsBasic()
}

// IsChromatic reports whether i has a whole quality and is not diatonic.
// An interval with a fractional quality is neither diatonic nor chromatic.
func (i Interval) IsChromatic() bool {
	return i.IsIntegerQuality() && !i.IsDiatonic()
}

// IsNarrow reports whether i is set and narrow: n <= 4.
func (i Interval) IsNarrow() bool {
	return i.IsValid() && i.size <= 4
}

// IsWide reports whether i is wide: 5 <= n <= 8. A compound interval is
// neither narrow nor wide.
func (i Interval) IsWide() bool {
	return i.size >= 5 && i.size <= 8
}

// IsTritone reports whether i is a simple augmented fourth or a simple
// diminished fifth, in either direction. Compound intervals are not tritones.
func (i Interval) IsTritone() bool {
	return i.size <= 8 && isTritoneShape(i.size, i.dev)
}

// WithDirection returns i in the direction d. The perfect unison stays
// ascending in any direction.
//
// Errors, in this order: ErrUnset for the unset interval; ErrNoDirection when
// d is neither Up nor Down.
func (i Interval) WithDirection(d Direction) (Interval, error) {
	const op = "WithDirection"
	if !i.IsValid() {
		return Interval{}, opError(op, ErrUnset)
	}
	if !d.isSet() {
		return Interval{}, opError(op, ErrNoDirection)
	}
	return newInterval(i.size, i.dev, d == Down), nil
}

// MustWithDirection is like WithDirection but panics if WithDirection returns
// an error. The panic value is the error returned by WithDirection.
func (i Interval) MustWithDirection(d Direction) Interval {
	return must(i.WithDirection(d))
}

// Up returns i ascending. It returns ErrUnset for the unset interval.
func (i Interval) Up() (Interval, error) {
	if !i.IsValid() {
		return Interval{}, opError("Up", ErrUnset)
	}
	return newInterval(i.size, i.dev, false), nil
}

// MustUp is like Up but panics if Up returns an error. The panic value is the
// error returned by Up.
func (i Interval) MustUp() Interval {
	return must(i.Up())
}

// Down returns i descending; the perfect unison stays ascending. It returns
// ErrUnset for the unset interval.
func (i Interval) Down() (Interval, error) {
	if !i.IsValid() {
		return Interval{}, opError("Down", ErrUnset)
	}
	return newInterval(i.size, i.dev, true), nil
}

// MustDown is like Down but panics if Down returns an error. The panic value
// is the error returned by Down.
func (i Interval) MustDown() Interval {
	return must(i.Down())
}

// Opposite returns the interval with the same size and quality in the other
// direction. The opposite of the perfect unison is the perfect unison.
// It returns ErrUnset for the unset interval.
func (i Interval) Opposite() (Interval, error) {
	if !i.IsValid() {
		return Interval{}, opError("Opposite", ErrUnset)
	}
	return i.opposite(), nil
}

// opposite returns the opposite of the set interval i.
func (i Interval) opposite() Interval {
	return newInterval(i.size, i.dev, !i.down)
}

// MustOpposite is like Opposite but panics if Opposite returns an error. The
// panic value is the error returned by Opposite.
func (i Interval) MustOpposite() Interval {
	return must(i.Opposite())
}

// sign returns +1 for an ascending and -1 for a descending set interval.
func (i Interval) sign() int {
	if i.down {
		return -1
	}
	return 1
}

// Equal reports whether i and other are both set and have the same direction,
// size, and deviation. A3 and d4, A4 and d5, M3 and the descending M3 are not
// equal. The unset interval is equal to no interval, not even to itself.
func (i Interval) Equal(other Interval) bool {
	return i.IsValid() && i == other
}

// SameSize reports whether i and other are both set and have the same size.
func (i Interval) SameSize(other Interval) bool {
	return i.IsValid() && other.IsValid() && i.size == other.size
}

// SameQuality reports whether i and other are both set and have the same kind
// and multiplicity of quality. Two fractional qualities are the same only in
// the same group and with the same deviation.
func (i Interval) SameQuality(other Interval) bool {
	if !i.IsValid() || !other.IsValid() {
		return false
	}
	pi, po := perfectSize(i.size), perfectSize(other.size)
	qi, ki := qualityOf(pi, i.dev)
	qo, ko := qualityOf(po, other.dev)
	if qi != qo || ki != ko {
		return false
	}
	if qi == Fractional {
		return pi == po && i.dev == other.dev
	}
	return true
}

// SameDirection reports whether i and other are both set and have the same
// direction. The perfect unison is ascending.
func (i Interval) SameDirection(other Interval) bool {
	return i.IsValid() && other.IsValid() && i.down == other.down
}

// EqualUndirected reports whether i and other are both set and have the same
// size and deviation, in any directions.
func (i Interval) EqualUndirected(other Interval) bool {
	return i.IsValid() && other.IsValid() && i.size == other.size && i.dev == other.dev
}

// Compare compares i and other in an agreed order for sorting and stable
// output: -1 if i sorts before other, 0 if they are equal, +1 otherwise.
// The unset interval sorts first, and two unset intervals compare as 0.
// Set intervals are ordered by the signed number of letter steps (negative
// when descending), then by the signed deviation. In ascending order:
// descending M3, descending m3, descending A1, P1, A1, d2, m2, M2.
// Compare returns 0 for set intervals exactly when Equal reports true. It can
// be used with slices.SortFunc as interval.Interval.Compare.
func (i Interval) Compare(other Interval) int {
	switch {
	case !i.IsValid() && !other.IsValid():
		return 0
	case !i.IsValid():
		return -1
	case !other.IsValid():
		return 1
	}
	ti, to := i.sign()*(i.size-1), other.sign()*(other.size-1)
	switch {
	case ti < to:
		return -1
	case ti > to:
		return 1
	}
	di, do := i.signedDeviation(), other.signedDeviation()
	return di.Compare(do)
}

// signedDeviation returns the deviation of the set interval i, negated when i
// is descending.
func (i Interval) signedDeviation() alteration.Alteration {
	if !i.down {
		return i.dev
	}
	// Negation stays in range: the range of package alteration is symmetric.
	return alteration.Natural().MustSub(i.dev)
}
