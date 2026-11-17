package degree_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

func TestDegreeComparisonControlExamples(t *testing.T) {
	t.Parallel()
	b := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	boundB := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	flatB := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	flatB.MustLower()
	b4 := degreeTestASPNNote(t, spelling.B(), 4, cents.Zero(), nil)
	b4plus10 := degreeTestASPNNote(t, spelling.B(), 4, cents.MustFromInt(10), nil)
	d := degreeTestNote(t, spelling.D(), cents.Zero(), nil)
	a := degreeTestNote(t, spelling.A(), cents.Zero(), nil)

	tests := []struct {
		name       string
		left       degree.Degree
		right      degree.Degree
		equal      bool
		sameNumber bool
		sameNote   bool
		compare    int
	}{
		{"same B", degree.MustNewWithNote(4, b), degree.MustNewWithNote(4, b), true, true, true, 0},
		{"same B different sources", degree.MustNewWithNote(4, b), degree.MustNewWithNote(4, boundB), true, true, true, 0},
		{"both without notes", degree.MustNew(4), degree.MustNew(4), true, true, false, 0},
		{"without note before B", degree.MustNew(4), degree.MustNewWithNote(4, b), false, true, false, -1},
		{"B after flat B", degree.MustNewWithNote(4, b), degree.MustNewWithNote(4, flatB), false, true, false, 1},
		{"number before note comparison", degree.MustNewWithNote(4, b), degree.MustNewWithNote(5, b), false, false, true, -1},
		{"without octave before octave", degree.MustNewWithNote(4, b), degree.MustNewWithNote(4, b4), false, true, false, -1},
		{"cents ascending", degree.MustNewWithNote(4, b4), degree.MustNewWithNote(4, b4plus10), false, true, false, -1},
		{"D before A by number", degree.MustNewWithNote(2, d), degree.MustNewWithNote(5, a), false, false, false, -1},
		{"number takes precedence", degree.MustNew(5), degree.MustNewWithNote(4, b), false, false, false, 1},
		{"unset first", degree.Degree{}, degree.First(), false, false, false, -1},
		{"two unset", degree.Degree{}, degree.Degree{}, false, false, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.equal, tt.left.Equal(tt.right))
			assert.Equal(t, tt.equal, tt.right.Equal(tt.left))
			assert.Equal(t, tt.sameNumber, tt.left.SameNumber(tt.right))
			assert.Equal(t, tt.sameNumber, tt.right.SameNumber(tt.left))
			assert.Equal(t, tt.sameNote, tt.left.SameNote(tt.right))
			assert.Equal(t, tt.sameNote, tt.right.SameNote(tt.left))
			assert.Equal(t, tt.compare, tt.left.Compare(tt.right))
			assert.Equal(t, -tt.compare, tt.right.Compare(tt.left))
			if tt.left.IsValid() && tt.right.IsValid() {
				assert.Equal(t, tt.equal, tt.left.Compare(tt.right) == 0)
			}
		})
	}
}

func TestDegreeCompareSupportsSorting(t *testing.T) {
	t.Parallel()
	b := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	flatB := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	flatB.MustLower()
	b4 := degreeTestASPNNote(t, spelling.B(), 4, cents.Zero(), nil)
	values := []degree.Degree{
		degree.MustNewWithNote(4, b4),
		degree.MustNew(5),
		degree.MustNewWithNote(4, b),
		{},
		degree.MustNew(4),
		degree.MustNewWithNote(4, flatB),
	}
	slices.SortFunc(values, degree.Degree.Compare)
	expected := []degree.Degree{
		{},
		degree.MustNew(4),
		degree.MustNewWithNote(4, flatB),
		degree.MustNewWithNote(4, b),
		degree.MustNewWithNote(4, b4),
		degree.MustNew(5),
	}
	for i := range expected {
		if !expected[i].IsValid() {
			assertUnsetDegree(t, values[i])
			continue
		}
		assertSameDegree(t, expected[i], values[i])
	}
}
