package pcseq_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/pcseq"
)

func ExampleSequence() {
	s := pcseq.MustNew(12, 7, 11, 9, 6, 7)
	transposed := s.MustTranspose(2)
	retrograde := s.MustRetrograde()
	fmt.Println(s)
	fmt.Println(transposed)
	fmt.Println(retrograde)
	// Output:
	// <7, 11, 9, 6, 7> mod 12
	// <9, 1, 11, 8, 9> mod 12
	// <7, 6, 9, 11, 7> mod 12
}

func ExampleRow_Form() {
	row := pcseq.MustNewRow(12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	prime := row.MustForm(pcseq.MustNewFormLabel(pcseq.P, 0, 12))
	fmt.Println(prime)
	// Output:
	// <0, 4, 8, 10, 2, 9, 11, 7, 3, 1, 5, 6> mod 12
}

func ExampleRow_Multiply() {
	row := pcseq.MustNewRow(4, 0, 1, 2, 3)
	sequence := row.MustMultiply(2)
	fmt.Println(sequence)
	fmt.Println(sequence.IsRow())
	// Output:
	// <0, 2, 0, 2> mod 4
	// false
}

func ExampleParse() {
	s, err := pcseq.Parse("<0, 1, 0> mod 12")
	fmt.Println(s, err)
	_, err = pcseq.ParseRow("<0, 0> mod 2")
	fmt.Println(errors.Is(err, pcseq.ErrNotRow))
	// Output:
	// <0, 1, 0> mod 12 <nil>
	// true
}
