package tuning

import (
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Root selects a temperament root either by pitch class or by note name.
// The zero value is unset. Root is comparable, and == agrees with Equal for
// every value, including the zero value. It is safe for concurrent use.
type Root struct {
	class pitchclass.Class
	name  spelling.Spelling
}

// RootFromClass returns a root specified by a pitch class. An unset class
// produces an unset root. The class modulus is checked when the root is used.
func RootFromClass(c pitchclass.Class) Root {
	if !c.IsValid() {
		return Root{}
	}
	return Root{class: c}
}

// RootFromName returns a root specified by a note name. An unset name produces
// an unset root. The system translates the name to a class when the root is used.
func RootFromName(name spelling.Spelling) Root {
	if !name.IsValid() {
		return Root{}
	}
	return Root{name: name}
}

// IsValid reports whether r has exactly one specified form: class or name.
func (r Root) IsValid() bool {
	return r.class.IsValid() != r.name.IsValid()
}

// Class returns the class form and whether it is present.
func (r Root) Class() (pitchclass.Class, bool) {
	return r.class, r.class.IsValid()
}

// Name returns the name form and whether it is present.
func (r Root) Name() (spelling.Spelling, bool) {
	return r.name, r.name.IsValid()
}

// Equal reports whether r and other use the same form with the same value.
// Two unset roots are equal; an unset and a set root are not.
func (r Root) Equal(other Root) bool {
	return r == other
}

// String returns a debug representation of the specified form, or "Root(unset)".
func (r Root) String() string {
	if class, ok := r.Class(); ok {
		return "Root(" + class.String() + ")"
	}
	if name, ok := r.Name(); ok {
		return "Root(" + name.String() + ")"
	}
	return "Root(unset)"
}

func resolveRoot(sys *system.System, root Root) (pitchclass.Class, error) {
	if !root.IsValid() {
		return pitchclass.Class{}, ErrUnset
	}
	if class, ok := root.Class(); ok {
		if class.Modulus() != sys.Classes() {
			return pitchclass.Class{}, invalidArgument()
		}
		return class, nil
	}
	name, _ := root.Name()
	class, err := sys.Class(name)
	if err != nil {
		return pitchclass.Class{}, systemDistanceError(err)
	}
	return class, nil
}
