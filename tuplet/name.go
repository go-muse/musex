package tuplet

// englishName returns the English name of a tuplet with actual count n.
func englishName(n int) (string, bool) {
	switch n {
	case 2:
		return "duplet", true
	case 3:
		return "triplet", true
	case 4:
		return "quadruplet", true
	case 5:
		return "quintuplet", true
	case 6:
		return "sextuplet", true
	case 7:
		return "septuplet", true
	case 8:
		return "octuplet", true
	case 9:
		return "nonuplet", true
	case 10:
		return "decuplet", true
	default:
		return "", false
	}
}

// russianName returns the Russian name of a tuplet with actual count n.
func russianName(n int) (string, bool) {
	switch n {
	case 2:
		return "дуоль", true
	case 3:
		return "триоль", true
	case 4:
		return "квартоль", true
	case 5:
		return "квинтоль", true
	case 6:
		return "секстоль", true
	case 7:
		return "септоль", true
	case 8:
		return "октоль", true
	case 9:
		return "нонемоль", true
	case 10:
		return "децимоль", true
	default:
		return "", false
	}
}

// countByName returns the actual count named by text: an English name, an
// alternative English form, or a Russian name, in lower case.
func countByName(text string) (int, bool) {
	switch text {
	case "duplet", "дуоль":
		return 2, true
	case "triplet", "триоль":
		return 3, true
	case "quadruplet", "quartole", "квартоль":
		return 4, true
	case "quintuplet", "quintolet", "pentuplet", "квинтоль":
		return 5, true
	case "sextuplet", "sextolet", "sestole", "секстоль":
		return 6, true
	case "septuplet", "septolet", "septimole", "септоль":
		return 7, true
	case "octuplet", "октоль":
		return 8, true
	case "nonuplet", "нонемоль":
		return 9, true
	case "decuplet", "децимоль":
		return 10, true
	default:
		return 0, false
	}
}

// named returns the named tuplet with actual count n in 2…10: the default
// tuplet for Binary division or, if there is none, for Ternary division.
func named(n int) Tuplet {
	normal, ok := defaultNormal(n, Binary)
	if !ok {
		normal, _ = defaultNormal(n, Ternary)
	}
	return Tuplet{actual: n, normal: normal}
}

// Name returns the English name of t by its actual count alone: "duplet",
// "triplet", "quadruplet", "quintuplet", "sextuplet", "septuplet",
// "octuplet", "nonuplet", or "decuplet" for 2 through 10, so 5:3 is also a
// quintuplet. For other actual counts it returns ("", false, nil). It returns
// an error wrapping ErrUnset if t is unset.
func (t Tuplet) Name() (name string, ok bool, err error) {
	if !t.IsSet() {
		return "", false, unsetError("receiver")
	}
	name, ok = englishName(t.actual)
	return name, ok, nil
}

// MustName is like Name but panics with the returned error.
func (t Tuplet) MustName() (name string, ok bool) {
	name, ok, err := t.Name()
	if err != nil {
		panic(err)
	}
	return name, ok
}

// RussianName returns the Russian name of t by its actual count alone:
// "дуоль", "триоль", "квартоль", "квинтоль", "секстоль", "септоль", "октоль",
// "нонемоль", or "децимоль" for 2 through 10. For other actual counts it
// returns ("", false, nil). The forms "дуплет" and "триплет" are not used. It
// returns an error wrapping ErrUnset if t is unset.
func (t Tuplet) RussianName() (name string, ok bool, err error) {
	if !t.IsSet() {
		return "", false, unsetError("receiver")
	}
	name, ok = russianName(t.actual)
	return name, ok, nil
}

// MustRussianName is like RussianName but panics with the returned error.
func (t Tuplet) MustRussianName() (name string, ok bool) {
	name, ok, err := t.RussianName()
	if err != nil {
		panic(err)
	}
	return name, ok
}
