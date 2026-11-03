package octave_test

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
)

// K-19, C-19, NF-2: String is "ASPN <ASPN number>" for an octave and
// "Octave(invalid)" for the invalid value; one octave has one string however
// it was obtained.
func TestString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		o    octave.Octave
		want string
	}{
		{name: "O(4)", o: oct(4), want: "ASPN 4"},
		{name: "O(3, FB)", o: octave.MustFromNumber(3, fb()), want: "ASPN 4"},
		{name: "P(первая, Ru)", o: octave.MustFromName("первая", naming.Russian()), want: "ASPN 4"},
		{name: "O(-1)", o: oct(-1), want: "ASPN -1"},
		{name: "O(0)", o: oct(0), want: "ASPN 0"},
		{name: "O(MIN)", o: oct(minInt), want: "ASPN -9223372036854775808"},
		{name: "O(MAX)", o: oct(maxInt), want: "ASPN 9223372036854775807"},
		{name: "Z", o: octave.Octave{}, want: "Octave(invalid)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got string
			assert.NotPanics(t, func() { got = tc.o.String() })
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("usual range", func(t *testing.T) {
		t.Parallel()

		for _, a := range span(-20, 20) {
			assert.Equal(t, "ASPN "+strconv.Itoa(a), oct(a).String(), "O(%d).String()", a)
		}
	})
}

// K-19, K-24: the string of the invalid value differs from the string of
// every octave, in particular O(0) and O(4); strings contain no names and no
// note notation such as "C4".
func TestStringIsDistinctAndPlain(t *testing.T) {
	t.Parallel()

	invalid := octave.Octave{}.String()
	forbidden := []string{"C4", "c4", "первая", "one-line", "eingestrichene", "middle"}

	for _, a := range append(span(-20, 20), edgeASPN()...) {
		s := oct(a).String()
		assert.NotEqual(t, invalid, s, "String of O(%d) equals the string of Z", a)
		for _, f := range forbidden {
			assert.NotContains(t, s, f, "String of O(%d)", a)
		}
	}
	for _, f := range forbidden {
		assert.NotContains(t, invalid, f, "String of Z")
	}
}

// K-19, C-19, NF-2: GoString is a Go expression of the public API that
// creates an equal value.
func TestGoString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		o    octave.Octave
		want string
	}{
		{name: "O(4)", o: oct(4), want: "octave.FromASPN(4)"},
		{name: "O(3, FB)", o: octave.MustFromNumber(3, fb()), want: "octave.FromASPN(4)"},
		{name: "O(-1)", o: oct(-1), want: "octave.FromASPN(-1)"},
		{name: "O(MIN)", o: oct(minInt), want: "octave.FromASPN(-9223372036854775808)"},
		{name: "O(MAX)", o: oct(maxInt), want: "octave.FromASPN(9223372036854775807)"},
		{name: "Z", o: octave.Octave{}, want: "octave.Octave{}"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got string
			assert.NotPanics(t, func() { got = tc.o.GoString() })
			assert.Equal(t, tc.want, got)
		})
	}
}

// K-19, NF-2: the fmt verbs covered by the string forms (architecture,
// "Which fmt verbs use the string forms"): %v, %+v, %s, Sprint and Println
// give String; %q gives the quoted String; %x and %X give the hexadecimal
// form of String; %#v gives GoString. Other verbs are not part of the
// contract and are not checked.
func TestFmtVerbs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		o    octave.Octave
	}{
		{name: "O(4)", o: oct(4)},
		{name: "O(-1)", o: oct(-1)},
		{name: "O(MIN)", o: oct(minInt)},
		{name: "O(MAX)", o: oct(maxInt)},
		{name: "Z", o: octave.Octave{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.o.String()
			verbs := []struct {
				format string
				want   string
			}{
				{format: "%v", want: s},
				{format: "%+v", want: s},
				{format: "%s", want: s},
				{format: "%q", want: strconv.Quote(s)},
				{format: "%x", want: hex.EncodeToString([]byte(s))},
				{format: "%X", want: strings.ToUpper(hex.EncodeToString([]byte(s)))},
				{format: "%#v", want: tc.o.GoString()},
			}
			for _, v := range verbs {
				assert.Equal(t, v.want, fmt.Sprintf(v.format, tc.o), "verb %s", v.format)
			}
			assert.Equal(t, s, fmt.Sprint(tc.o), "Sprint")
			assert.Equal(t, s+"\n", fmt.Sprintln(tc.o), "Sprintln")
		})
	}

	t.Run("default output shows no fields", func(t *testing.T) {
		t.Parallel()

		out := fmt.Sprintf("%v %+v %#v", oct(4), oct(4), oct(4))
		assert.Equal(t, "ASPN 4 ASPN 4 octave.FromASPN(4)", out)
		assert.False(t, strings.ContainsAny(out, "{}"), "no struct braces: %s", out)
	})
}
