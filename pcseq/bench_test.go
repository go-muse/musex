package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"
)

func BenchmarkTransform(b *testing.B) {
	s, err := pcseq.New(12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	if err != nil {
		b.Fatal(err)
	}
	op, err := pitchclass.NewOperator(5, 2, 12)
	if err != nil {
		b.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		call func() (pcseq.Sequence, error)
	}{
		{name: "Transpose", call: func() (pcseq.Sequence, error) { return s.Transpose(2) }},
		{name: "Invert", call: func() (pcseq.Sequence, error) { return s.Invert(2) }},
		{name: "Multiply", call: func() (pcseq.Sequence, error) { return s.Multiply(2) }},
		{name: "Apply", call: func() (pcseq.Sequence, error) { return s.Apply(op) }},
		{name: "Retrograde", call: func() (pcseq.Sequence, error) { return s.Retrograde() }},
		{name: "RetrogradeInversion", call: func() (pcseq.Sequence, error) { return s.RetrogradeInversion(2) }},
		{name: "Rotate", call: func() (pcseq.Sequence, error) { return s.Rotate(3) }},
	} {
		b.Run(tt.name, func(b *testing.B) {
			var result pcseq.Sequence
			var err error
			for b.Loop() {
				result, err = tt.call()
			}
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		})
	}
}

func BenchmarkMatrix(b *testing.B) {
	row, err := pcseq.NewRow(12, 0, 11, 3, 7, 8, 4, 2, 6, 5, 1, 9, 10)
	if err != nil {
		b.Fatal(err)
	}
	matrix, err := pcseq.NewMatrix(row)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("Cells", func(b *testing.B) {
		var result [][]pitchclass.Class
		var err error
		for b.Loop() {
			result, err = matrix.Cells()
		}
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	})
}

func BenchmarkRowClass(b *testing.B) {
	row, err := pcseq.NewRow(12, 0, 11, 3, 7, 8, 4, 2, 6, 5, 1, 9, 10)
	if err != nil {
		b.Fatal(err)
	}
	for _, kind := range []pcseq.RowClassKind{pcseq.TIR, pcseq.TIRRotations, pcseq.TTOR} {
		b.Run(kind.String(), func(b *testing.B) {
			var result pcseq.RowClass
			var err error
			for b.Loop() {
				result, err = pcseq.NewRowClass(row, kind)
			}
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		})
	}
}

func BenchmarkDerivation(b *testing.B) {
	row, err := pcseq.NewRow(12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	if err != nil {
		b.Fatal(err)
	}
	var result [][]pcseq.DerivationOperator
	for b.Loop() {
		result, err = row.Derive(3)
	}
	if err != nil {
		b.Fatal(err)
	}
	_ = result
}
