package measure_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/measure"
)

// Group Т-9, fuzzing of parsing (architecture 4.16; О-22; П-16): for any
// input no panic; on success the text parses back to an equal value; on
// failure the error wraps exactly one package sentinel and its text is at
// most 512 bytes.

// countPackageSentinels returns how many package sentinels err wraps.
func countPackageSentinels(err error) int {
	count := 0

	for _, s := range measureSentinels {
		if errors.Is(err, s.err) {
			count++
		}
	}

	return count
}

// FuzzParse covers Parse and UnmarshalText of a grid.
func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"4/4", "3/4;pickup:1/4", "4/4;3:3/4;4:length:1/2;bars:8", "4/4;5:length:1/4:end", "4/4;5:4/4",
		"C|;2:C", "2/4+3/8;pickup:1/8", "3/4;pickup:2/8", "¢", "6/8[3+3]", "4/4;bars:9223372036854775807",
		"", ";", "4/4;", "Grid(unset)", "4/4;1:3/4", "4/4;0:3/4", "4/4;3:length:1/2;3:3/4", "4/4;pickup:5/4",
		"4/4;3:3/4;bars:2", "4/4;99999999999999999999:3/4", "3/99999999999999999999", "4/4;pickup:0/1",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		g, err := measure.Parse(text)
		if err != nil {
			if n := countPackageSentinels(err); n != 1 {
				t.Fatalf("Parse(%q): the error wraps %d package sentinels: %v", text, n, err)
			}

			if len(err.Error()) > 512 {
				t.Fatalf("Parse(%q): the error text is %d bytes", text, len(err.Error()))
			}

			if g.IsSet() {
				t.Fatalf("Parse(%q): a set grid with an error", text)
			}

			return
		}

		if !g.IsSet() {
			t.Fatalf("Parse(%q): the unset grid without an error", text)
		}

		back, err := measure.Parse(g.String())
		if err != nil || !back.Equal(g) {
			t.Fatalf("Parse(%q): the text %q does not parse back: %v", text, g.String(), err)
		}

		var u measure.Grid
		if err := u.UnmarshalText([]byte(text)); err != nil || !u.Equal(g) {
			t.Fatalf("UnmarshalText(%q) differs from Parse: %v", text, err)
		}
	})
}

// FuzzParseAddress covers ParseAddress and UnmarshalText of an address.
func FuzzParseAddress(f *testing.F) {
	for _, s := range []string{
		"2:1/8", "0:0/1", "2:2/16", "9223372036854775807:3/7", "", "2", "2:", "02:1/8", "-1:0/1",
		"Address(unset)", "99999999999999999999:0/1", "2:1/0", "2:0",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		a, err := measure.ParseAddress(text)
		if err != nil {
			if n := countPackageSentinels(err); n != 1 {
				t.Fatalf("ParseAddress(%q): the error wraps %d package sentinels: %v", text, n, err)
			}

			if len(err.Error()) > 512 {
				t.Fatalf("ParseAddress(%q): the error text is %d bytes", text, len(err.Error()))
			}

			if a != (measure.Address{}) {
				t.Fatalf("ParseAddress(%q): a set address with an error", text)
			}

			return
		}

		back, err := measure.ParseAddress(a.String())
		if err != nil || back != a {
			t.Fatalf("ParseAddress(%q): the text %q does not parse back: %v", text, a.String(), err)
		}

		var u measure.Address
		if err := u.UnmarshalText([]byte(text)); err != nil || u != a {
			t.Fatalf("UnmarshalText(%q) differs from ParseAddress: %v", text, err)
		}
	})
}
