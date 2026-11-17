# Template

A mode template is a lightweight record that answers one question: how many places a mode has, and how far apart those places are. It stores named intervals between neighboring places, in order. It is not a mode itself: it has no degrees, tonic, note names, musical system, or compositional analysis.

A template can be built from adjacent intervals, from intervals measured from its first place, or from note names read upward in order. Building from names uses only their written intervals and needs no musical system. The names are inputs, not data retained by the template.

## Closed and open templates

A **closed** template repeats. With k places it has k adjacent intervals; the last one leads from place k back to place 1 in the next cycle. Their sum is the **cycle**. The cycle need not be an octave: it can be shorter, wider, or otherwise different. A template is called octave-based only when this sum is a perfect octave.

An **open** template does not repeat. With k places it has k-1 adjacent intervals and no closing interval or cycle. It ends after its last place. Open and closed templates remain different even when their existing adjacent intervals are written identically.

## Places and virtual places

Places are simply numbered 1, 2, 3, and so on. The interval of place 1 from itself is a perfect unison; every later place has the sum of the adjacent intervals before it. The interval between any two places has a direction and is independent of a tonic.

A closed template also has virtual places beyond place k. Place k+1 is place 1 raised by one cycle, and the same pattern continues through any positive place number. An open template has only its actual places. Rotating a closed template chooses another place as its beginning; an open template has only the trivial rotation from place 1.

Templates can be compared for exact equality or placed in a conventional sorting order. Neither operation chooses a musical reference standard. For two templates of the same kind and length, mismatching places can be counted in two ways: by intervals measured from the first place, including the cycle, or by adjacent intervals, including the closing interval.

## Catalogs and relations

A catalog entry adds a code and an ordered list of names to a template. The built-in catalog contains 81 entries in the accepted source order. Each built-in template and entry is also available directly. Custom catalogs are ordinary immutable values; there is no global registry.

Entries can be found by exact code or by name. Name search normalizes letter case, musical signs, dashes, Latin diacritics, Russian ё, and whitespace while preserving the original names. One name may belong to several entries, and one template may appear under several codes.

Identification finds equal templates or their rotations. Related-entry search goes further: it finds templates differing in no more than a chosen number of places, using either mismatch measure, with or without rotations. Each result reports the entry, selected rotation, and the mismatching places themselves.

Every template has a short text form made from adjacent interval notation. An open template ends with a vertical bar. Russian interval abbreviations are available when every interval has one. The text form supports storage, exchange, and JSON through Go's standard text interfaces.

Templates, entries, and catalogs are immutable values safe for concurrent reading. Input and output slices are copied according to the API, so changing a caller's slice does not change a stored value.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/template) · [Go examples](../../template/example_test.go) · [Roadmap](../../ROADMAP.md)
