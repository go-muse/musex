package system

import (
	"iter"

	"github.com/go-muse/musex/letter"
)

// Built-in systems, created once at package initialization.
var (
	traditional = MustNew(Definition{
		Name:        "traditional",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 2, 4, 5, 7, 9, 11},
		Chromatic:   1,
		Alterations: Integers(),
	})

	positions19 = MustNew(Definition{
		Name:        "19-position",
		Classes:     19,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 3, 6, 8, 11, 14, 17},
		Chromatic:   1,
		Alterations: Integers(),
	})

	quarterTones24 = MustNew(Definition{
		Name:        "24-quarter-tone",
		Classes:     24,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 4, 8, 10, 14, 18, 22},
		Chromatic:   2,
		Alterations: MustMultiples(2),
	})

	builtins = [...]*System{traditional, positions19, quarterTones24}
)

// Traditional returns the traditional system, named "traditional": N = 12,
// boundary letter C, offsets C 0 D 2 E 4 F 5 G 7 A 9 B 11, c = 1 and the
// integer alteration values. It is regular with the steps 2 and 1. Every call
// returns the same immutable *System.
func Traditional() *System {
	return traditional
}

// Positions19 returns the 19-position system, named "19-position": N = 19,
// boundary letter C, offsets C 0 D 3 E 6 F 8 G 11 A 14 B 17, c = 1 and the
// integer alteration values. It is regular with the steps 3 and 2. Every call
// returns the same immutable *System.
func Positions19() *System {
	return positions19
}

// QuarterTones24 returns the 24-class quarter-tone system, named
// "24-quarter-tone": N = 24, boundary letter C, offsets C 0 D 4 E 8 F 10 G 14
// A 18 B 22, c = 2 and the alteration values that are multiples of 1/2. It is
// regular with the steps 4 and 2. Every call returns the same immutable
// *System.
func QuarterTones24() *System {
	return quarterTones24
}

// Builtins returns all built-in systems in the order Traditional, Positions19,
// QuarterTones24. New built-in systems may be added after these; the order of
// the existing ones is kept.
func Builtins() iter.Seq[*System] {
	return func(yield func(*System) bool) {
		for _, s := range builtins {
			if !yield(s) {
				return
			}
		}
	}
}

// LookupBuiltin returns the built-in system with exactly the name name (case
// sensitive, without normalization) and true, or nil and false.
func LookupBuiltin(name string) (*System, bool) {
	for _, s := range builtins {
		if s.name == name {
			return s, true
		}
	}

	return nil, false
}
