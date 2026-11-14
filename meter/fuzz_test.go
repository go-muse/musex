package meter_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/meter"
)

// FuzzParse covers О-19 and architecture 4.15: Parse does not panic; on
// success the value is set and its text parses back to it; an error wraps
// exactly one package sentinel and its text is at most 512 bytes.
func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"3/4", "C", "C|", "¢", "(3+2+3)/8", "2/4+3/8", "5/8[2+3]", "6/8[3+3]", "C[3+1]",
		"", " 3/4", "3 /4", "03/4", "0/4", "3/0", "c", "(3)/8", "(2+3)/8[2+3]", "2/4+C",
		"2/4 + 3/8", "5/8[]", "5/8[2+3", "Meter(unset)", "1025/8", "5/8[2+2]",
		"3/99999999999999999999", "99999999999999999999/4", "1024/8+1/4",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		m, err := meter.Parse(text)
		if err != nil {
			count := 0

			for _, s := range meterSentinels {
				if errors.Is(err, s.err) {
					count++
				}
			}

			if count != 1 {
				t.Fatalf("Parse(%q): the error wraps %d package sentinels: %v", text, count, err)
			}

			if len(err.Error()) > 512 {
				t.Fatalf("Parse(%q): the error text is %d bytes", text, len(err.Error()))
			}

			if m != (meter.Meter{}) {
				t.Fatalf("Parse(%q): a value on error: %v", text, m)
			}

			return
		}

		if !m.IsSet() {
			t.Fatalf("Parse(%q) returned an unset meter", text)
		}

		back, err := meter.Parse(m.String())
		if err != nil || back != m {
			t.Fatalf("Parse(%q) = %v; Parse(%q) = %v, %v", text, m, m.String(), back, err)
		}
	})
}
