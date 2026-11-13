package tuning

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitchclass"
)

// decompose returns q and m such that base+k = n*q+m and 0 <= m < n.
// It avoids evaluating base+k, so every int k is handled without overflow.
func decompose(base, k, n int) (q, m int) {
	q = k / n
	m = k % n
	if m < 0 {
		q--
		m += n
	}
	m += base
	q += m / n
	m %= n
	return q, m
}

func positiveMod(value, modulus int) int {
	result := value % modulus
	if result < 0 {
		result += modulus
	}
	return result
}

func addInt(a, b int) (int, bool) {
	result := a + b
	if b > 0 && result < a || b < 0 && result > a {
		return 0, false
	}
	return result, true
}

func subInt(a, b int) (int, bool) {
	if b > 0 && a < -1<<63+b || b < 0 && a > 1<<63-1+b {
		return 0, false
	}
	return a - b, true
}

func (t *Tuning) referenceClass() pitchclass.Class {
	class, _ := t.sys.PitchClass(t.ref.note)
	return class
}

func (t *Tuning) referenceOffset() int {
	return positiveMod(t.referenceClass().Number()-t.root.Number(), t.sys.Classes())
}

func (t *Tuning) offsetDifference(m, base int) (cents.Cents, error) {
	offset, err := t.temp.Offset(m)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	baseOffset, err := t.temp.Offset(base)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	difference, err := offset.Sub(baseOffset)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	return difference, nil
}

func (t *Tuning) heightFromParts(q, m, base int) (cents.Cents, error) {
	difference, err := t.offsetDifference(m, base)
	if err != nil {
		return cents.Cents{}, err
	}
	octaves, err := cents.Octave().Mul(q)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	height, err := octaves.Add(difference)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	return height, nil
}

func (t *Tuning) heightDecomposition(k int) (q, m, base int) {
	base = t.referenceOffset()
	q, m = decompose(base, k, t.sys.Classes())
	return q, m, base
}

func (t *Tuning) heightAtStep(k int) (cents.Cents, error) {
	q, m, base := t.heightDecomposition(k)
	return t.heightFromParts(q, m, base)
}

func (t *Tuning) heightAtStepWithRoot(k int, root pitchclass.Class) (cents.Cents, error) {
	base := positiveMod(t.referenceClass().Number()-root.Number(), t.sys.Classes())
	q, m := decompose(base, k, t.sys.Classes())
	return t.heightFromParts(q, m, base)
}

func (t *Tuning) heightForClass(classNumber int, root pitchclass.Class) (cents.Cents, error) {
	referenceClass := t.referenceClass().Number()
	k := positiveMod(classNumber-referenceClass, t.sys.Classes())
	return t.heightAtStepWithRoot(k, root)
}

func (t *Tuning) frequencyFromHeight(height cents.Cents) (float64, error) {
	freq, err := height.ApplyTo(t.ref.freq)
	if err != nil {
		return 0, centsError(err)
	}
	return freq, nil
}
