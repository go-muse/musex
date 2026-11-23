package system

import (
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// PaletteGroup is the spellings of one pitch class in a palette. Spellings is
// a caller-owned slice; it is empty when the class has no spelling in the
// palette.
type PaletteGroup struct {
	// Class is the pitch class of the group.
	Class pitchclass.Class
	// Spellings are the spellings of the class, in the order of the palette.
	Spellings []spelling.Spelling
}

// Palette returns, for each class of s in class order, the spellings the rules
// produce, in rule order and without repeats: a spelling equal to one already
// in the group is not added again. The selected policy is preserved in the
// result order. For example, Palette(Flats(), Sharps()) on the traditional
// system gives 12 groups and 17 distinct spellings, the class of C♯/D♭
// offering D♭ before C♯.
//
// With no rules every group is empty. Errors are those of [System.Spell] for
// the first class and rule that fail; an unset rule gives an error that
// matches ErrInvalid. The result is a new slice that the caller owns.
func (s *System) Palette(rules ...SpellingRule) ([]PaletteGroup, error) {
	if !s.IsValid() {
		return nil, errorf(ErrInvalid, "unset system")
	}
	for i, r := range rules {
		if !r.IsValid() {
			return nil, errorf(ErrInvalid, "unset spelling rule %d", i)
		}
	}

	groups := make([]PaletteGroup, 0, s.n)
	for k := 0; k < s.n; k++ {
		c := pitchclass.MustNew(k, s.n)
		group := PaletteGroup{Class: c}
		for _, r := range rules {
			sp, err := s.Spell(c, r)
			if err != nil {
				return nil, err
			}
			if !slices.ContainsFunc(group.Spellings, func(x spelling.Spelling) bool { return x.Equal(sp) }) {
				group.Spellings = append(group.Spellings, sp)
			}
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// MustPalette is like Palette but panics on error.
func (s *System) MustPalette(rules ...SpellingRule) []PaletteGroup {
	return must(s.Palette(rules...))
}

// PaletteWithin returns, for each class of s in class order, every admissible
// spelling of the class with |a| ≤ limit, in the order of [System.Spellings].
// A class without such spellings gives an empty group. With a limit of one
// sharps and flats the class of C offers C and B♯, and the class of B offers
// C♭ and B.
//
// Errors, in this order: ErrInvalid when s is unset; ErrLimit when limit is
// negative. The result is a new slice that the caller owns.
func (s *System) PaletteWithin(limit alteration.Alteration) ([]PaletteGroup, error) {
	if !s.IsValid() {
		return nil, errorf(ErrInvalid, "unset system")
	}
	if limit.Sign() < 0 {
		return nil, errorf(ErrLimit, "limit %v", limit)
	}

	groups := make([]PaletteGroup, 0, s.n)
	for k := 0; k < s.n; k++ {
		c := pitchclass.MustNew(k, s.n)
		group := PaletteGroup{Class: c}
		seq, err := s.Spellings(c, limit)
		if err != nil {
			return nil, err
		}
		for sp := range seq {
			group.Spellings = append(group.Spellings, sp)
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// MustPaletteWithin is like PaletteWithin but panics on error.
func (s *System) MustPaletteWithin(limit alteration.Alteration) []PaletteGroup {
	return must(s.PaletteWithin(limit))
}
