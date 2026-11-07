package setclass

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/setclass/catalog"
)

func wrapCatalogError(err error) error {
	if errors.Is(err, catalog.ErrInvalidCardinality) || errors.Is(err, catalog.ErrNoCarterCatalog) {
		return fmt.Errorf("%w (%w): %w", ErrInvalidCardinality, ErrNotation, err)
	}
	return fmt.Errorf("%w (%w): %w", ErrNotation, ErrInvalidValue, err)
}

// FromForteClass constructs the mod-12 TnIn SetClass corresponding to fc in O(1) time.
func FromForteClass(fc catalog.ForteClass) (SetClass, error) {
	if !fc.IsValid() {
		return SetClass{}, fmt.Errorf("%w (%w): invalid ForteClass", ErrInvalidValue, catalog.ErrInvalidValue)
	}
	return SetClass{
		kind:  TnIn,
		prime: pcset.MustFromBitmask(fc.PrimeMask(), 12),
	}, nil
}

// MustFromForteClass is like [FromForteClass] but panics on error.
func MustFromForteClass(fc catalog.ForteClass) SetClass {
	sc, err := FromForteClass(fc)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromTnType constructs the mod-12 Tn SetClass corresponding to tt in O(1) time.
func FromTnType(tt catalog.TnType) (SetClass, error) {
	if !tt.IsValid() {
		return SetClass{}, fmt.Errorf("%w (%w): invalid TnType", ErrInvalidValue, catalog.ErrInvalidValue)
	}
	return SetClass{
		kind:  Tn,
		prime: pcset.MustFromBitmask(tt.PrimeMask(), 12),
	}, nil
}

// MustFromTnType is like [FromTnType] but panics on error.
func MustFromTnType(tt catalog.TnType) SetClass {
	sc, err := FromTnType(tt)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromCarterClass constructs the mod-12 TnIn SetClass corresponding to cc in O(1) time.
func FromCarterClass(cc catalog.CarterClass) (SetClass, error) {
	if !cc.IsValid() {
		return SetClass{}, fmt.Errorf("%w (%w): invalid CarterClass", ErrInvalidValue, catalog.ErrInvalidValue)
	}
	return FromForteClass(cc.ForteClass())
}

// MustFromCarterClass is like [FromCarterClass] but panics on error.
func MustFromCarterClass(cc catalog.CarterClass) SetClass {
	sc, err := FromCarterClass(cc)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromCarterNumber constructs the mod-12 TnIn SetClass for Elliott Carter's (cardinality, number) entry.
func FromCarterNumber(cardinality, number int) (SetClass, error) {
	cc, err := catalog.CarterByNumber(cardinality, number)
	if err != nil {
		return SetClass{}, wrapCatalogError(err)
	}
	return FromForteClass(cc.ForteClass())
}

// MustFromCarterNumber is like [FromCarterNumber] but panics on error.
func MustFromCarterNumber(cardinality, number int) SetClass {
	sc, err := FromCarterNumber(cardinality, number)
	if err != nil {
		panic(err)
	}
	return sc
}

func fromCatalogCoordinates(cardinality, ordinal int, suffix catalog.Suffix, kind Kind) (SetClass, error) {
	if !kind.IsValid() {
		return SetClass{}, fmt.Errorf("%w: %v", ErrInvalidKind, kind)
	}
	var (
		mask uint64
		err  error
	)
	if kind == Tn {
		var tt catalog.TnType
		tt, err = catalog.TnTypeByNumber(cardinality, ordinal, suffix)
		mask = tt.PrimeMask()
	} else {
		var fc catalog.ForteClass
		fc, err = catalog.ForteClassByNumber(cardinality, ordinal)
		mask = fc.PrimeMask()
	}
	if err != nil {
		return SetClass{}, wrapCatalogError(err)
	}
	return FromSet(pcset.MustFromBitmask(mask, 12), kind)
}

// FromForteNumber constructs the mod-12 SetClass of kind from Forte's (cardinality, ordinal) coordinates.
// For Kind == Tn on a chiral Forte class, FromForteNumber returns an error wrapping [ErrNotation]
// and [catalog.ErrInvalidSuffix] because an explicit A/B suffix is required (use [FromTnTypeNumber]).
func FromForteNumber(cardinality, ordinal int, kind Kind) (SetClass, error) {
	return fromCatalogCoordinates(cardinality, ordinal, catalog.SuffixNone, kind)
}

// MustFromForteNumber is like [FromForteNumber] but panics on error.
func MustFromForteNumber(cardinality, ordinal int, kind Kind) SetClass {
	sc, err := FromForteNumber(cardinality, ordinal, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromTnTypeNumber constructs the mod-12 Tn SetClass from (cardinality, ordinal, suffix).
func FromTnTypeNumber(cardinality, ordinal int, suffix catalog.Suffix) (SetClass, error) {
	return fromCatalogCoordinates(cardinality, ordinal, suffix, Tn)
}

// MustFromTnTypeNumber is like [FromTnTypeNumber] but panics on error.
func MustFromTnTypeNumber(cardinality, ordinal int, suffix catalog.Suffix) SetClass {
	sc, err := FromTnTypeNumber(cardinality, ordinal, suffix)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromCatalogName parses a mod-12 catalog label into a SetClass of the specified kind:
//   - for Kind == Tn: strict TnType notation ("3-1", "3-11A", "3-11B", "4-Z15A");
//   - for Kind == TnIn: strict ForteClass notation ("0-1", "3-11", "4-Z15", "6-Z29");
//   - for Kind == TTO: either a single ForteClass label ("6-1", "6-32", "4-Z15") or a composite
//     Morris label "<c>-<d1>/<d2>" ("6-1/32", "4-Z15/Z29", "6-Z24/Z39") where d1 < d2 and the two
//     Forte classes are M_5 partners.
func FromCatalogName(s string, kind Kind) (SetClass, error) {
	switch kind {
	case Tn:
		tt, err := catalog.ParseTnType(s)
		if err != nil {
			return SetClass{}, wrapCatalogError(err)
		}
		return FromTnType(tt)
	case TnIn:
		fc, err := catalog.ParseForteClass(s)
		if err != nil {
			return SetClass{}, wrapCatalogError(err)
		}
		return FromForteClass(fc)
	default:
		before, after, hasSlash := strings.Cut(s, "/")
		if !hasSlash {
			fc, err := catalog.ParseForteClass(s)
			if err != nil {
				return SetClass{}, wrapCatalogError(err)
			}
			return FromSet(pcset.MustFromBitmask(fc.PrimeMask(), 12), kind)
		}
		fc1, err1 := catalog.ParseForteClass(before)
		cPrefix, _, _ := strings.Cut(before, "-")
		fc2, err2 := catalog.ParseForteClass(cPrefix + "-" + after)
		if err1 != nil || err2 != nil || fc1.Ordinal() >= fc2.Ordinal() || fc1.M5Partner() != fc2 {
			return SetClass{}, fmt.Errorf("%w (%w): invalid composite TTO label %q", ErrNotation, catalog.ErrInvalidNotation, s)
		}
		return FromSet(pcset.MustFromBitmask(fc1.PrimeMask(), 12), kind)
	}
}

// MustFromCatalogName is like [FromCatalogName] but panics on error.
func MustFromCatalogName(s string, kind Kind) SetClass {
	sc, err := FromCatalogName(s, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromForteNotation is a synonym for [FromCatalogName].
func FromForteNotation(s string, kind Kind) (SetClass, error) {
	return FromCatalogName(s, kind)
}

// MustFromForteNotation is like [FromForteNotation] but panics on error.
func MustFromForteNotation(s string, kind Kind) SetClass {
	sc, err := FromForteNotation(s, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromTnTypeNotation parses a strict mod-12 TnType notation string into a Tn SetClass.
func FromTnTypeNotation(s string) (SetClass, error) {
	return FromCatalogName(s, Tn)
}

// MustFromTnTypeNotation is like [FromTnTypeNotation] but panics on error.
func MustFromTnTypeNotation(s string) SetClass {
	sc, err := FromTnTypeNotation(s)
	if err != nil {
		panic(err)
	}
	return sc
}

// ForteClass returns the [catalog.ForteClass] corresponding to sc.ToTnIn() at N = 12 in O(1) time.
// It works for all three Kinds (Tn, TnIn, and TTO). For N != 12, it returns [ErrNotModulus12].
func (sc SetClass) ForteClass() (catalog.ForteClass, error) {
	if !sc.IsValid() {
		return catalog.ForteClass{}, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return catalog.ForteClass{}, fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	mask := sc.prime.MustBitmask()
	if sc.kind == Tn {
		return catalog.MustTnTypeByPrimeMask(mask).ForteClass(), nil
	}
	return catalog.MustForteClassByPrimeMask(mask), nil
}

// MustForteClass is like [SetClass.ForteClass] but panics on error.
func (sc SetClass) MustForteClass() catalog.ForteClass {
	fc, err := sc.ForteClass()
	if err != nil {
		panic(err)
	}
	return fc
}

// MustForteClass returns sc.MustForteClass().
func MustForteClass(sc SetClass) catalog.ForteClass {
	return sc.MustForteClass()
}

// TnType returns the [catalog.TnType] corresponding to sc at N = 12 in O(1) time.
// It is defined only for Kind == Tn (for Kind == TnIn or TTO, it returns [ErrKindMismatch];
// use [SetClass.CatalogTnTypes] to obtain all constituent TnTypes).
func (sc SetClass) TnType() (catalog.TnType, error) {
	if !sc.IsValid() {
		return catalog.TnType{}, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return catalog.TnType{}, fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	if sc.kind != Tn {
		return catalog.TnType{}, fmt.Errorf("%w: TnType requires Kind == Tn, got %v", ErrKindMismatch, sc.kind)
	}
	return catalog.MustTnTypeByPrimeMask(sc.prime.MustBitmask()), nil
}

// MustTnType is like [SetClass.TnType] but panics on error.
func (sc SetClass) MustTnType() catalog.TnType {
	tt, err := sc.TnType()
	if err != nil {
		panic(err)
	}
	return tt
}

// CatalogTnTypes returns the slice of constituent [catalog.TnType] entries of sc at N = 12,
// in the exact order of [SetClass.TnClasses].
func (sc SetClass) CatalogTnTypes() ([]catalog.TnType, error) {
	if !sc.IsValid() {
		return nil, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return nil, fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	tcs := sc.TnClasses()
	out := make([]catalog.TnType, len(tcs))
	for i, tc := range tcs {
		out[i] = catalog.MustTnTypeByPrimeMask(tc.prime.MustBitmask())
	}
	return out, nil
}

// MustCatalogTnTypes is like [SetClass.CatalogTnTypes] but panics on error.
func (sc SetClass) MustCatalogTnTypes() []catalog.TnType {
	list, err := sc.CatalogTnTypes()
	if err != nil {
		panic(err)
	}
	return list
}

// CarterClass returns Elliott Carter's [catalog.CarterClass] for sc at N = 12 (for cardinalities 2, 3, 4, 6, 7).
// For other cardinalities at N = 12, it returns (CarterClass{}, false, nil).
func (sc SetClass) CarterClass() (catalog.CarterClass, bool, error) {
	fc, err := sc.ForteClass()
	if err != nil {
		return catalog.CarterClass{}, false, err
	}
	cc, ok := fc.Carter()
	return cc, ok, nil
}

// MustCarterClass is like [SetClass.CarterClass] but panics on error.
func (sc SetClass) MustCarterClass() (catalog.CarterClass, bool) {
	cc, ok, err := sc.CarterClass()
	if err != nil {
		panic(err)
	}
	return cc, ok
}

// CarterNumber returns Elliott Carter's catalog number for sc at N = 12 (for cardinalities 2, 3, 4, 6, 7).
// For other cardinalities at N = 12, it returns (0, false, nil).
func (sc SetClass) CarterNumber() (int, bool, error) {
	fc, err := sc.ForteClass()
	if err != nil {
		return 0, false, err
	}
	num, ok := fc.CarterNumber()
	return num, ok, nil
}

// MustCarterNumber is like [SetClass.CarterNumber] but panics on error.
func (sc SetClass) MustCarterNumber() (int, bool) {
	num, ok, err := sc.CarterNumber()
	if err != nil {
		panic(err)
	}
	return num, ok
}

// ZPair returns the [catalog.ZPair] of sc.ForteClass() at N = 12 if it is a Z-set,
// or (ZPair{}, false, nil) otherwise.
func (sc SetClass) ZPair() (catalog.ZPair, bool, error) {
	fc, err := sc.ForteClass()
	if err != nil {
		return catalog.ZPair{}, false, err
	}
	zp, ok := fc.ZPair()
	return zp, ok, nil
}

// MustZPair is like [SetClass.ZPair] but panics on error.
func (sc SetClass) MustZPair() (catalog.ZPair, bool) {
	zp, ok, err := sc.ZPair()
	if err != nil {
		panic(err)
	}
	return zp, ok
}

// ZPartner returns the unique Forte Z-partner SetClass for a mod-12 TnIn SetClass in O(1) time.
// For N != 12, it returns [ErrNotModulus12]. For N == 12 and Kind != TnIn, it returns [ErrKindMismatch].
// If sc has no Z-partner, it returns (SetClass{}, false, nil).
func (sc SetClass) ZPartner() (SetClass, bool, error) {
	if !sc.IsValid() {
		return SetClass{}, false, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return SetClass{}, false, fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	if sc.kind != TnIn {
		return SetClass{}, false, fmt.Errorf("%w: ZPartner requires Kind == TnIn, got %v", ErrKindMismatch, sc.kind)
	}
	fc := catalog.MustForteClassByPrimeMask(sc.prime.MustBitmask())
	zp, ok := fc.ZPartner()
	if !ok {
		return SetClass{}, false, nil
	}
	return MustFromForteClass(zp), true, nil
}

// MustZPartner is like [SetClass.ZPartner] but panics on error.
func (sc SetClass) MustZPartner() (SetClass, bool) {
	zp, ok, err := sc.ZPartner()
	if err != nil {
		panic(err)
	}
	return zp, ok
}

// CatalogName returns the canonical mod-12 catalog label of sc:
//   - for Kind == Tn: the TnType name ("3-11A", "3-11B", "3-1");
//   - for Kind == TnIn: the ForteClass name ("3-11", "4-Z15");
//   - for Kind == TTO: "<c>-<d>" for M_5-invariant classes or Morris's composite label
//     "<c>-<d1>/<d2>" with d1 < d2 for M_5-paired classes ("6-1/32", "4-Z15/Z29", "6-Z24/Z39").
func (sc SetClass) CatalogName() (string, error) {
	if !sc.IsValid() {
		return "", fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return "", fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	mask := sc.prime.MustBitmask()
	switch sc.kind {
	case Tn:
		return catalog.MustTnTypeByPrimeMask(mask).String(), nil
	case TnIn:
		return catalog.MustForteClassByPrimeMask(mask).String(), nil
	default: // TTO
		fc := catalog.MustForteClassByPrimeMask(mask)
		m5 := fc.M5Partner()
		if fc == m5 {
			return fc.String(), nil
		}
		if m5.Ordinal() < fc.Ordinal() {
			fc, m5 = m5, fc
		}
		_, d2, _ := strings.Cut(m5.String(), "-")
		return fc.String() + "/" + d2, nil
	}
}

// MustCatalogName is like [SetClass.CatalogName] but panics on error.
func (sc SetClass) MustCatalogName() string {
	name, err := sc.CatalogName()
	if err != nil {
		panic(err)
	}
	return name
}
