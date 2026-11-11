package pcseq

import (
	"cmp"
	"strings"

	"github.com/go-muse/musex/pitchclass"
)

func newSequenceUnchecked(modulus int, elements string) Sequence {
	return Sequence{modulus: modulus, elements: elements}
}

func newRowUnchecked(modulus int, elements string) Row {
	return Row{modulus: modulus, elements: elements}
}

func encodeNumbers(numbers []int) string {
	buf := make([]byte, 2*len(numbers))
	for i, number := range numbers {
		putNumber(buf, i, number)
	}
	return string(buf)
}

func encodeClasses(classes []pitchclass.Class) string {
	buf := make([]byte, 2*len(classes))
	for i, class := range classes {
		putNumber(buf, i, class.Number())
	}
	return string(buf)
}

func putNumber(buf []byte, index, number int) {
	buf[2*index] = byte(uint16(number) >> 8)
	buf[2*index+1] = byte(number)
}

func numberAt(elements string, index int) int {
	return int(elements[2*index])<<8 | int(elements[2*index+1])
}

func classAt(modulus int, elements string, index int) pitchclass.Class {
	return pitchclass.MustNew(numberAt(elements, index), modulus)
}

func compareEncoded(aModulus int, a string, bModulus int, b string) int {
	if result := cmp.Compare(aModulus, bModulus); result != 0 {
		return result
	}
	if result := cmp.Compare(len(a), len(b)); result != 0 {
		return result
	}
	return strings.Compare(a, b)
}

func sequenceIsValid(s Sequence) bool {
	return s.modulus >= 1 && s.modulus <= MaxModulus && len(s.elements)%2 == 0
}

func rowIsValid(r Row) bool {
	if !sequenceIsValid(r.Sequence()) || r.Len() != r.modulus {
		return false
	}
	seen := make([]bool, r.modulus)
	for i := range r.Len() {
		number := numberAt(r.elements, i)
		if number >= r.modulus || seen[number] {
			return false
		}
		seen[number] = true
	}
	return true
}
