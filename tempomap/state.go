package tempomap

import "github.com/go-muse/musex/tempo"

// State is the exact current and main tempo outside a gradual segment.
type State struct {
	current   tempo.Tempo
	main      tempo.Tempo
	deviating bool
	set       bool
}

// IsSet reports whether s is set.
func (s State) IsSet() bool { return s.set }

// Current returns the current exact tempo.
func (s State) Current() (tempo.Tempo, error) {
	if !s.IsSet() {
		return tempo.Tempo{}, simpleError(ErrUnset, "State.Current", "receiver")
	}
	return s.current, nil
}

// MustCurrent is like Current but panics with its error.
func (s State) MustCurrent() tempo.Tempo { return must(s.Current()) }

// Main returns the main tempo to which a tempo returns.
func (s State) Main() (tempo.Tempo, error) {
	if !s.IsSet() {
		return tempo.Tempo{}, simpleError(ErrUnset, "State.Main", "receiver")
	}
	return s.main, nil
}

// MustMain is like Main but panics with its error.
func (s State) MustMain() tempo.Tempo { return must(s.Main()) }

// Deviating reports whether the current tempo deviates from the main tempo.
func (s State) Deviating() (bool, error) {
	if !s.IsSet() {
		return false, simpleError(ErrUnset, "State.Deviating", "receiver")
	}
	return s.deviating, nil
}

// MustDeviating is like Deviating but panics with its error.
func (s State) MustDeviating() bool { return must(s.Deviating()) }

// Equal reports whether s and other have the same state.
func (s State) Equal(other State) bool { return s == other }

// String returns a diagnostic representation, not a serialization format.
func (s State) String() string {
	if !s.IsSet() {
		return "State(unset)"
	}
	return "State{current:" + s.current.String() + ",main:" + s.main.String() + ",deviating:" + boolString(s.deviating) + "}"
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
