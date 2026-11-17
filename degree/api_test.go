package degree_test

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDegreeConstantsAndZeroValue(t *testing.T) {
	t.Parallel()
	assert.Equal(t, math.MaxInt, degree.MaxNumber)
	assert.Equal(t, 3999, degree.MaxRomanNumber)
	assertUnsetDegree(t, degree.Degree{})
	var d degree.Degree
	assertUnsetDegree(t, d)
	var stringer fmt.Stringer = degree.MustNew(1)
	assert.Equal(t, "Degree(1, no note)", stringer.String())
}

func TestDegreeNilPointerMutators(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	identity := func(n note.Note) (note.Note, error) { return n, nil }
	tests := []struct {
		name string
		call func(*degree.Degree) error
		must func(*degree.Degree)
	}{
		{
			name: "SetNote",
			call: func(d *degree.Degree) error { return d.SetNote(n) },
			must: func(d *degree.Degree) { d.MustSetNote(n) },
		},
		{
			name: "RemoveNote",
			call: func(d *degree.Degree) error { return d.RemoveNote() },
			must: func(d *degree.Degree) { d.MustRemoveNote() },
		},
		{
			name: "TransformNote",
			call: func(d *degree.Degree) error { return d.TransformNote(identity) },
			must: func(d *degree.Degree) { d.MustTransformNote(identity) },
		},
		{
			name: "SetNumber",
			call: func(d *degree.Degree) error { return d.SetNumber(5) },
			must: func(d *degree.Degree) { d.MustSetNumber(5) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var d *degree.Degree
			err := tt.call(d)
			assertDegreeError(t, err, degree.ErrUnset)
			_ = assertMustPanic(t, degree.ErrUnset, func() { tt.must(d) })
		})
	}
}

func TestDegreeNilPointerReadersPanic(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	other := degree.MustNewWithNote(2, n)
	tests := []struct {
		name string
		call func(*degree.Degree)
	}{
		{"IsValid", func(d *degree.Degree) { d.IsValid() }},
		{"Number", func(d *degree.Degree) { d.Number() }},
		{"HasNote", func(d *degree.Degree) { d.HasNote() }},
		{"Note", func(d *degree.Degree) { d.Note() }},
		{"Equal", func(d *degree.Degree) { d.Equal(other) }},
		{"SameNumber", func(d *degree.Degree) { d.SameNumber(other) }},
		{"SameNote", func(d *degree.Degree) { d.SameNote(other) }},
		{"Compare", func(d *degree.Degree) { d.Compare(other) }},
		{"Decimal", func(d *degree.Degree) { _, _ = d.Decimal() }},
		{"Roman", func(d *degree.Degree) { _, _ = d.Roman() }},
		{"String", func(d *degree.Degree) { _ = d.String() }},
		{"InCircle", func(d *degree.Degree) { _, _ = d.InCircle(7) }},
		{"Next", func(d *degree.Degree) { _, _, _ = d.Next(7) }},
		{"Previous", func(d *degree.Degree) { _, _, _ = d.Previous(7) }},
		{"Shift", func(d *degree.Degree) { _, _, _ = d.Shift(7, 1) }},
		{"DistanceUp", func(d *degree.Degree) { _, _ = d.DistanceUp(other, 7) }},
		{"DistanceDown", func(d *degree.Degree) { _, _ = d.DistanceDown(other, 7) }},
		{"ShortestDistance", func(d *degree.Degree) { _, _ = d.ShortestDistance(other, 7) }},
		{"Normalize", func(d *degree.Degree) { _, _, _ = d.Normalize(7) }},
		{"Denormalize", func(d *degree.Degree) { _, _ = d.Denormalize(7, 0) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var d *degree.Degree
			assert.Panics(t, func() { tt.call(d) })
		})
	}
}

func TestDegreeAddressableMutationAndPointerChains(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	d := new(degree.MustNew(4))
	returned := d.MustSetNote(n)
	assert.Same(t, d, returned)
	assertDegreeState(t, *d, 4, true, n)
	returned = returned.MustSetNumber(5).MustRemoveNote()
	assert.Same(t, d, returned)
	assertDegreeState(t, *d, 5, false, note.Note{})
}

func TestDegreeInterfaceComparisonPanics(t *testing.T) {
	t.Parallel()
	a := degree.MustNew(1)
	b := degree.MustNew(1)
	assert.Panics(t, func() { _ = any(a) == any(b) })
}

type degreeCompileProbeImporter struct {
	importer   types.Importer
	loadErrors []error
}

func newDegreeCompileProbeImporter(fset *token.FileSet) *degreeCompileProbeImporter {
	loader := &degreeCompileProbeImporter{}
	loader.importer = importer.ForCompiler(fset, "gc", loader.lookup)
	return loader
}

func (l *degreeCompileProbeImporter) lookup(path string) (io.ReadCloser, error) {
	cmd := exec.Command("go", "list", "-export", "-f", "{{.Export}}", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("resolve export data for %s: %w", path, err)
		l.loadErrors = append(l.loadErrors, err)
		return nil, err
	}
	exportPath := strings.TrimSpace(string(output))
	if exportPath == "" {
		err = fmt.Errorf("resolve export data for %s: empty export path", path)
		l.loadErrors = append(l.loadErrors, err)
		return nil, err
	}
	file, err := os.Open(exportPath)
	if err != nil {
		err = fmt.Errorf("open export data for %s: %w", path, err)
		l.loadErrors = append(l.loadErrors, err)
		return nil, err
	}
	return file, nil
}

func typeCheckDegreeProbe(t *testing.T, source string) ([]error, error) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "probe.go", source, 0)
	require.NoError(t, err)
	loader := newDegreeCompileProbeImporter(fset)
	config := types.Config{
		Importer: loader.importer,
		Error:    func(error) {},
	}
	_, err = config.Check("example.com/probe", fset, []*ast.File{file}, nil)
	return loader.loadErrors, err
}

func TestDegreeCompileTimePositiveProbes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "read function result",
			source: `package probe
import "github.com/go-muse/musex/degree"
var _ = degree.MustNew(4).IsValid()`,
		},
		{
			name: "read map element",
			source: `package probe
import "github.com/go-muse/musex/degree"
var m map[int]degree.Degree
var k int
var _ = m[k].IsValid()`,
		},
		{
			name: "mutate through new function result",
			source: `package probe
import (
    "github.com/go-muse/musex/degree"
    "github.com/go-muse/musex/note"
)
var n note.Note
var _ = new(degree.MustNew(4)).MustSetNote(n)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			loadErrors, err := typeCheckDegreeProbe(t, tt.source)
			require.Empty(t, loadErrors)
			require.NoError(t, err)
		})
	}
}

func TestDegreeCompileTimeNegativeProbes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "built-in equality",
			source: `package probe
import "github.com/go-muse/musex/degree"
var a, b degree.Degree
var _ = degree.MustNew(4).IsValid()
var _ = a == b`,
		},
		{
			name: "degree as map key",
			source: `package probe
import "github.com/go-muse/musex/degree"
var _ = degree.MustNew(4).IsValid()
var _ map[degree.Degree]int`,
		},
		{
			name: "mutate function result",
			source: `package probe
import (
    "github.com/go-muse/musex/degree"
    "github.com/go-muse/musex/note"
)
var n note.Note
var _ = degree.MustNew(4).IsValid()
var _ = degree.MustNew(4).MustSetNote(n)`,
		},
		{
			name: "mutate map element",
			source: `package probe
import (
    "github.com/go-muse/musex/degree"
    "github.com/go-muse/musex/note"
)
var m map[int]degree.Degree
var k int
var n note.Note
var _ = m[k].IsValid()
var _ = m[k].MustSetNote(n)`,
		},
		{
			name: "struct containing degree is not comparable",
			source: `package probe
import "github.com/go-muse/musex/degree"
type holder struct { d degree.Degree }
var a, b holder
var _ = degree.MustNew(4).IsValid()
var _ = a == b`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			loadErrors, err := typeCheckDegreeProbe(t, tt.source)
			require.Empty(t, loadErrors, "a package loading failure would make this negative probe vacuous")
			assert.Error(t, err)
		})
	}
}
