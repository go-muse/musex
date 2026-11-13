package settings

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// Settings is an immutable set of settings: a tuning (and through it a
// musical system) together with the designations of written pitches — a
// notation, an octave numbering and a catalog of octave names.
//
// A Settings is used through a pointer. The nil pointer and a pointer to the
// zero value are the unset set: IsValid reports false, the readers return
// unset parts, and the other operations return an error that matches
// ErrUnset. Methods on the unset set do not panic.
//
// The set owns its own copies of the tuning and of the catalog: they are
// copied when a set is created and when they are read, so no caller can
// change a set through them. Values of Settings are not comparable; use
// Equal. == on pointers is identity.
type Settings struct {
	// tuning is the set's own copy; nil only in the unset set. It is never
	// given out.
	tuning *tuning.Tuning
	// notation is defined in a set Settings.
	notation notation.Notation
	// numbering is defined in a set Settings.
	numbering number.Numbering
	// catalog is the set's own copy, never nil in a set Settings and never
	// given out.
	catalog *naming.Catalog
}

// standardTemplate holds the standard settings. It is built when the package
// is initialized and only read afterward; Standard returns copies of it.
var standardTemplate = Settings{
	tuning:    tuning.Standard(),
	notation:  notation.MustBuiltIn(notation.EnglishUnicodeID),
	numbering: number.ASPN(),
	catalog:   copyCatalog(naming.English()),
}

// New returns a set of settings from a tuning, a notation, an octave
// numbering and a catalog of octave names. The system of the set is the
// system of the tuning.
//
// Failures, in this order: an error that matches ErrUnset when t is unset or
// n or num is undefined; an error that matches ErrInconsistent (and
// notation.ErrIncompatibleSystem) when n is an ups-and-downs notation whose
// chromatic unit differs from the chromatic unit of the system of t.
//
// The catalog is never unset: nil, a pointer to the zero value and a catalog
// without names all mean the empty catalog. New copies t and *c; changing them
// afterward does not change the set. On error New returns nil.
func New(t *tuning.Tuning, n notation.Notation, num number.Numbering, c *naming.Catalog) (*Settings, error) {
	const op = "new"
	if !t.IsValid() {
		return nil, unsetError(op, "tuning")
	}
	if !n.IsDefined() {
		return nil, unsetError(op, "notation")
	}
	if !num.IsDefined() {
		return nil, unsetError(op, "numbering")
	}
	if err := checkConsistent(op, n, t.System()); err != nil {
		return nil, err
	}
	return &Settings{
		tuning:    copyTuning(t),
		notation:  n,
		numbering: num,
		catalog:   copyCatalog(c),
	}, nil
}

// MustNew is like New but panics with its error on failure.
func MustNew(t *tuning.Tuning, n notation.Notation, num number.Numbering, c *naming.Catalog) *Settings {
	return must(New(t, n, num, c))
}

// Standard returns the standard settings: the traditional system, the
// standard tuning of package tuning (12-EDO, root C, A4 at 440 Hz), the
// english-unicode notation, the ASPN numbering and the English catalog of
// octave names.
//
// Every call returns a new pointer to an equal set, so assigning through the
// pointer changes only that copy. Standard allocates only the new set.
func Standard() *Settings {
	s := standardTemplate
	return &s
}

// IsValid reports whether s is set. It is false for nil and for a pointer to
// the zero value.
func (s *Settings) IsValid() bool {
	return s != nil && s.tuning != nil
}

// System returns the musical system of the set: the system of its tuning.
// The system is shared and immutable, as in package system; it must not be
// overwritten through the pointer. The unset set returns nil.
func (s *Settings) System() *system.System {
	if !s.IsValid() {
		return nil
	}
	return s.tuning.System()
}

// Tuning returns a new tuning equal to the tuning of the set; the caller owns
// it. The unset set returns nil.
func (s *Settings) Tuning() *tuning.Tuning {
	if !s.IsValid() {
		return nil
	}
	return copyTuning(s.tuning)
}

// Notation returns the notation of the set. The copy shares only immutable
// tables. The unset set returns the undefined notation.
func (s *Settings) Notation() notation.Notation {
	if !s.IsValid() {
		return notation.Notation{}
	}
	return s.notation
}

// Numbering returns the octave numbering of the set. The unset set returns
// the undefined numbering.
func (s *Settings) Numbering() number.Numbering {
	if !s.IsValid() {
		return number.Numbering{}
	}
	return s.numbering
}

// Catalog returns a new, non-nil catalog equal to the catalog of the set;
// for the empty catalog it is a catalog without names. The caller owns the
// pointer. The unset set returns nil.
func (s *Settings) Catalog() *naming.Catalog {
	if !s.IsValid() {
		return nil
	}
	return copyCatalog(s.catalog)
}

// OctaveContext returns the octave context of the set for direct calls of
// package notation: its numbering and a new copy of its catalog. The unset
// set returns the zero context.
func (s *Settings) OctaveContext() notation.OctaveContext {
	if !s.IsValid() {
		return notation.OctaveContext{}
	}
	return notation.OctaveContext{Numbering: s.numbering, Catalog: copyCatalog(s.catalog)}
}

// Equal reports whether s and other have equal tunings, notations,
// numberings and catalogs, each compared by its own Equal. Two unset sets are
// equal; an unset and a set one are not. Pointer identity does not matter.
func (s *Settings) Equal(other *Settings) bool {
	sv, ov := s.IsValid(), other.IsValid()
	if !sv || !ov {
		return sv == ov
	}
	return s.tuning.Equal(other.tuning) &&
		s.notation.Equal(other.notation) &&
		s.numbering.Equal(other.numbering) &&
		s.catalog.Equal(other.catalog)
}

// String returns a debugging form that names all five parts: the system,
// the tuning, the notation with its octave style, the numbering and the
// catalog. It is not a notation, has no parser and does not affect
// equality. The unset set is "Settings(unset)".
func (s *Settings) String() string {
	if !s.IsValid() {
		return "Settings(unset)"
	}
	var b strings.Builder
	b.WriteString("Settings(system: ")
	b.WriteString(debugName(s.tuning.System().Name()))
	b.WriteString(", tuning: ")
	b.WriteString(s.tuning.String())
	b.WriteString(", notation: ")
	b.WriteString(debugName(s.notation.Name()))
	b.WriteString(" (octave style: ")
	b.WriteString(s.notation.OctaveStyle().String())
	b.WriteString("), numbering: ")
	b.WriteString(s.numbering.String())
	b.WriteString(", catalog: ")
	writeCatalog(&b, s.catalog)
	b.WriteString(")")
	return b.String()
}

// Settings returns s itself, so that a set is its own Source. It returns the
// receiver also for the unset set.
func (s *Settings) Settings() *Settings {
	return s
}

// WithTuning returns a new set with the tuning t and the other parts of s.
// Failures, in this order: ErrUnset when s or t is unset; ErrInconsistent
// when the notation of s does not fit the system of t. s does not change.
func (s *Settings) WithTuning(t *tuning.Tuning) (*Settings, error) {
	const op = "with tuning"
	if !s.IsValid() {
		return nil, unsetError(op, "settings")
	}
	if !t.IsValid() {
		return nil, unsetError(op, "tuning")
	}
	if err := checkConsistent(op, s.notation, t.System()); err != nil {
		return nil, err
	}
	r := *s
	r.tuning = copyTuning(t)
	return &r, nil
}

// MustWithTuning is like WithTuning but panics with its error on failure.
func (s *Settings) MustWithTuning(t *tuning.Tuning) *Settings {
	return must(s.WithTuning(t))
}

// WithSystem returns a new set whose tuning is translated to the system sys
// by tuning.Tuning.WithSystem: the same temperament, root number and
// reference. The designations do not change.
//
// Failures, in this order: ErrUnset when s or sys is unset; the error of
// tuning.Tuning.WithSystem with its cause from package tuning
// (tuning.ErrClassCount when sys has another number of classes,
// tuning.ErrInadmissible when the reference note is not admissible in sys);
// ErrInconsistent when the notation of s does not fit sys. To change the
// number of classes, replace the whole tuning with WithTuning.
func (s *Settings) WithSystem(sys *system.System) (*Settings, error) {
	const op = "with system"
	if !s.IsValid() {
		return nil, unsetError(op, "settings")
	}
	if !sys.IsValid() {
		return nil, unsetError(op, "system")
	}
	t, err := s.tuning.WithSystem(sys)
	if err != nil {
		return nil, wrapError(op, err)
	}
	if err := checkConsistent(op, s.notation, sys); err != nil {
		return nil, err
	}
	r := *s
	r.tuning = t
	return &r, nil
}

// MustWithSystem is like WithSystem but panics with its error on failure.
func (s *Settings) MustWithSystem(sys *system.System) *Settings {
	return must(s.WithSystem(sys))
}

// WithNotation returns a new set with the notation n and the other parts of
// s. Failures, in this order: ErrUnset when s is unset or n is undefined;
// ErrInconsistent when n does not fit the system of s.
func (s *Settings) WithNotation(n notation.Notation) (*Settings, error) {
	const op = "with notation"
	if !s.IsValid() {
		return nil, unsetError(op, "settings")
	}
	if !n.IsDefined() {
		return nil, unsetError(op, "notation")
	}
	if err := checkConsistent(op, n, s.tuning.System()); err != nil {
		return nil, err
	}
	r := *s
	r.notation = n
	return &r, nil
}

// MustWithNotation is like WithNotation but panics with its error on failure.
func (s *Settings) MustWithNotation(n notation.Notation) *Settings {
	return must(s.WithNotation(n))
}

// WithNumbering returns a new set with the octave numbering num and the other
// parts of s. It fails with ErrUnset when s is unset or num is undefined.
func (s *Settings) WithNumbering(num number.Numbering) (*Settings, error) {
	const op = "with numbering"
	if !s.IsValid() {
		return nil, unsetError(op, "settings")
	}
	if !num.IsDefined() {
		return nil, unsetError(op, "numbering")
	}
	r := *s
	r.numbering = num
	return &r, nil
}

// MustWithNumbering is like WithNumbering but panics with its error on
// failure.
func (s *Settings) MustWithNumbering(num number.Numbering) *Settings {
	return must(s.WithNumbering(num))
}

// WithCatalog returns a new set with a copy of the catalog c and the other
// parts of s; nil is the empty catalog. It fails with ErrUnset when s is
// unset.
func (s *Settings) WithCatalog(c *naming.Catalog) (*Settings, error) {
	if !s.IsValid() {
		return nil, unsetError("with catalog", "settings")
	}
	r := *s
	r.catalog = copyCatalog(c)
	return &r, nil
}

// MustWithCatalog is like WithCatalog but panics with its error on failure.
func (s *Settings) MustWithCatalog(c *naming.Catalog) *Settings {
	return must(s.WithCatalog(c))
}

// WithDesignations returns a new set with the notation n, the numbering num
// and a copy of the catalog c, and the tuning of s. Failures, in this order:
// ErrUnset when s is unset or n or num is undefined; ErrInconsistent when n
// does not fit the system of s.
func (s *Settings) WithDesignations(n notation.Notation, num number.Numbering, c *naming.Catalog) (*Settings, error) {
	const op = "with designations"
	if !s.IsValid() {
		return nil, unsetError(op, "settings")
	}
	if !n.IsDefined() {
		return nil, unsetError(op, "notation")
	}
	if !num.IsDefined() {
		return nil, unsetError(op, "numbering")
	}
	if err := checkConsistent(op, n, s.tuning.System()); err != nil {
		return nil, err
	}
	r := *s
	r.notation = n
	r.numbering = num
	r.catalog = copyCatalog(c)
	return &r, nil
}

// MustWithDesignations is like WithDesignations but panics with its error on
// failure.
func (s *Settings) MustWithDesignations(n notation.Notation, num number.Numbering, c *naming.Catalog) *Settings {
	return must(s.WithDesignations(n, num, c))
}

// checkConsistent reports ErrInconsistent when n is an ups-and-downs notation
// whose chromatic unit differs from the chromatic unit of sys.
func checkConsistent(op string, n notation.Notation, sys *system.System) error {
	if cu, ok := n.ChromaticUnit(); ok && cu != sys.Chromatic() {
		return inconsistentError(op, cu, sys)
	}
	return nil
}

// copyTuning returns a new tuning equal to the valid tuning t. A copy of the
// value is a full, independent copy because tuning.Tuning is immutable and
// has no derived fields, caches or synchronization state, and the data it
// shares (the system and the temperament table) are immutable too. If
// package tuning ever gains state that must not be copied, build the copy
// with tuning.New from the parts of t instead.
func copyTuning(t *tuning.Tuning) *tuning.Tuning {
	cp := *t
	return &cp
}

// copyCatalog returns a new, non-nil catalog equal to c; nil gives a catalog
// without names. A catalog is never changed through its methods, so a copy
// of the value is independent of later assignments to *c.
func copyCatalog(c *naming.Catalog) *naming.Catalog {
	if c == nil {
		return &naming.Catalog{}
	}
	cp := *c
	return &cp
}

// debugName returns name, or "unnamed" for the empty name.
func debugName(name string) string {
	if name == "" {
		return "unnamed"
	}
	return name
}

// writeCatalog writes the named octaves of c in ascending order, or "empty".
func writeCatalog(b *strings.Builder, c *naming.Catalog) {
	first := true
	for aspn, name := range c.All() {
		if first {
			b.WriteString("{")
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(aspn))
		b.WriteString(": ")
		b.WriteString(strconv.Quote(name))
	}
	if first {
		b.WriteString("empty")
		return
	}
	b.WriteString("}")
}

// must returns v or panics with err.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
