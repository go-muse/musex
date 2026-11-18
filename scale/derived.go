package scale

import (
	"errors"
	"math/big"
	"slices"
	"strings"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// PitchClasses returns the set of name classes in the supplied system.
func (s *Scale) PitchClasses(sys *system.System) (pcset.Set, error) {
	if s == nil {
		return pcset.Set{}, ErrUnset
	}
	if !sys.IsValid() {
		return pcset.Set{}, errors.Join(ErrUnset, system.ErrInvalid)
	}
	values := make([]int, len(s.notes))
	for i, n := range s.notes {
		c, err := sys.Class(n.Spelling())
		if err != nil {
			return pcset.Set{}, err
		}
		values[i] = c.Number()
	}
	return pcset.New(sys.Classes(), values...)
}
func fifthOf(name spelling.Spelling) (*big.Int, bool) {
	num, den := name.Alteration().Fraction()
	if den != 1 {
		return nil, false
	}
	natural := [...]int64{0, 2, 4, -1, 1, 3, 5}
	fifth := new(big.Int).Mul(big.NewInt(int64(num)), big.NewInt(7))
	return fifth.Add(fifth, big.NewInt(natural[name.Letter().Ordinal()])), true
}

// KeySignature recognizes seven consecutive integral fifth coordinates among distinct names.
func (s *Scale) KeySignature() (Signature, bool, error) {
	if s == nil {
		return Signature{}, false, ErrUnset
	}
	names := make([]spelling.Spelling, 0, 7)
	for _, n := range s.notes {
		name := n.Spelling()
		if !slices.ContainsFunc(names, func(x spelling.Spelling) bool { return x.Equal(name) }) {
			names = append(names, name)
			if len(names) > 7 {
				return Signature{}, false, nil
			}
		}
	}
	if len(names) != 7 {
		return Signature{}, false, nil
	}
	coords := make([]*big.Int, 7)
	for i, name := range names {
		var ok bool
		coords[i], ok = fifthOf(name)
		if !ok {
			return Signature{}, false, nil
		}
	}
	slices.SortFunc(coords, func(a, b *big.Int) int { return a.Cmp(b) })
	for i, q := range coords {
		if new(big.Int).Add(coords[0], big.NewInt(int64(i))).Cmp(q) != 0 {
			return Signature{}, false, nil
		}
	}
	index := new(big.Int).Add(coords[0], big.NewInt(1))
	if !index.IsInt64() {
		return Signature{}, false, ErrOutOfRange
	}
	result := Signature{Index: int(index.Int64()), Signs: make([]spelling.Spelling, 0, 7)}
	for _, name := range names {
		if name.Alteration().Sign() != 0 {
			result.Signs = append(result.Signs, name)
		}
	}
	order := [...]int{1, 3, 5, 0, 2, 4, 6}
	slices.SortFunc(result.Signs, func(a, b spelling.Spelling) int { return order[a.Letter().Ordinal()] - order[b.Letter().Ordinal()] })
	if result.Index < 0 {
		slices.Reverse(result.Signs)
	}
	return result, true, nil
}

// String returns a diagnostic list including each note's stored parts.
func (s *Scale) String() string {
	if s == nil {
		return "Scale(unset)"
	}
	values := make([]string, len(s.notes))
	for i, n := range s.notes {
		values[i] = n.String()
	}
	return "Scale[" + strings.Join(values, "; ") + "]"
}

// GoString returns the same diagnostic representation as String.
func (s *Scale) GoString() string { return s.String() }

// Format formats notes in their own settings with explicit cents corrections.
func (s *Scale) Format() (string, error) { return s.format(nil, false) }

// FormatIn formats all notes in explicit settings with cents corrections.
func (s *Scale) FormatIn(src settings.Source) (string, error) { return s.format(src, true) }
func (s *Scale) format(src settings.Source, explicit bool) (string, error) {
	if s == nil {
		return "", ErrUnset
	}
	if explicit && src == nil {
		return "", ErrUnset
	}
	values := make([]string, len(s.notes))
	for i, n := range s.notes {
		var text string
		var err error
		if explicit {
			text, err = n.FormatIn(src)
		} else {
			text, err = n.Format()
		}
		if err != nil {
			return "", err
		}
		values[i] = text + " (" + n.Cents().Text(cents.WithUnit|cents.WithPlus) + ")"
	}
	return "[" + strings.Join(values, "; ") + "]", nil
}
