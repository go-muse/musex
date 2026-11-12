package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombinatorialFormsEvenWithoutForms(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		row  pcseq.Row
	}{
		{name: "N10", row: mustRow(t, 10, 0, 1, 2, 3, 6, 4, 5, 7, 8, 9)},
		{name: "N12", row: mustRow(t, 12, 0, 1, 2, 3, 4, 7, 5, 6, 8, 9, 10, 11)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			forms, err := tt.row.CombinatorialForms()
			require.NoError(t, err)
			assert.NotNil(t, forms)
			assert.Empty(t, forms)
		})
	}
}
