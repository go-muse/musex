package pcseq

import (
	"strconv"
	"strings"
)

func formatCanonical(modulus int, elements string) string {
	var builder strings.Builder
	builder.WriteByte('<')
	for i := range len(elements) / 2 {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(strconv.Itoa(numberAt(elements, i)))
	}
	builder.WriteString("> mod ")
	builder.WriteString(strconv.Itoa(modulus))
	return builder.String()
}

// String returns the canonical sequence notation or Sequence(invalid).
func (s Sequence) String() string {
	if !s.IsValid() {
		return "Sequence(invalid)"
	}
	return formatCanonical(s.modulus, s.elements)
}

// String returns the canonical row notation or Row(invalid).
func (r Row) String() string {
	if !r.IsValid() {
		return "Row(invalid)"
	}
	return formatCanonical(r.modulus, r.elements)
}
