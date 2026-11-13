package settings

import "reflect"

// Source gives the current set of settings whole.
//
// A set is itself a Source that returns itself; an instrument is another
// Source. A reader, such as a note, calls Settings at the moment of an
// operation and works with the returned snapshot; another call may return
// another set.
//
// Requirements for implementations:
//   - Settings is safe for concurrent calls;
//   - it returns either the previous or the new set whole, never a mixture of
//     the parts of two sets;
//   - it does not change the source and does not block for a long time;
//   - the returned set is immutable;
//   - a source without a set returns the unset set (nil), with which every
//     operation reports ErrUnset;
//   - the dynamic type of an implementation is comparable, preferably a
//     pointer to the owning object, so that SameSource can tell sources apart.
//
// == on Source values follows the rules of Go: for *Settings it is pointer
// identity, so two equal sets, such as two results of Standard, differ by ==,
// and == panics for an incomparable dynamic type. Use SameSource to decide
// whether two values are one source.
type Source interface {
	// Settings returns the current set of settings whole.
	Settings() *Settings
}

var _ Source = (*Settings)(nil)

// SameSource reports whether a and b are one source. It never calls Settings
// and never panics:
//   - a nil interface is no source: the result is false, also for two nil
//     interfaces;
//   - two *Settings (including the nil pointer and a pointer to the zero
//     value) are one source when they are Equal: a set returns only itself,
//     so equal sets cannot be told apart as sources, and two unset sets are
//     one source;
//   - values of different dynamic types are different sources;
//   - values of the same incomparable dynamic type are different sources;
//   - values of the same comparable dynamic type are one source when they are
//     == (for pointers, the same object).
//
// The result is symmetric. It does not allocate for two *Settings and for
// implementations that are pointers.
func SameSource(a, b Source) bool {
	if a == nil || b == nil {
		return false
	}
	sa, okA := a.(*Settings)
	sb, okB := b.(*Settings)
	if okA || okB {
		return okA && okB && sa.Equal(sb)
	}
	ta := reflect.TypeOf(a)
	if ta != reflect.TypeOf(b) {
		return false
	}
	if ta.Kind() != reflect.Pointer &&
		(!reflect.ValueOf(a).Comparable() || !reflect.ValueOf(b).Comparable()) {
		return false
	}
	return a == b
}
