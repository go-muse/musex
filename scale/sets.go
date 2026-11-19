package scale

import "github.com/go-muse/musex/note"

// partition assigns each occurrence to its first representative.
func partition(p prepared, ns note.Notes) (note.Notes, []int, error) {
	representatives := make(note.Notes, 0)
	classes := make([]int, len(ns))
	for i, n := range ns {
		found := -1
		for j, r := range representatives {
			same, err := p.equal(r, n)
			if err != nil {
				return nil, nil, err
			}
			if same {
				found = j
				break
			}
		}
		if found < 0 {
			found = len(representatives)
			representatives = append(representatives, n)
		}
		classes[i] = found
	}
	return representatives, classes, nil
}
func (s *Scale) contents(other *Scale, by []Comparison) (note.Notes, []int, []int, error) {
	if s == nil || other == nil {
		return nil, nil, nil, ErrUnset
	}
	p, err := prepare(by, s.notes, other.notes)
	if err != nil {
		return nil, nil, nil, err
	}
	ns := append(s.Notes(), other.notes...)
	reps, classes, err := partition(p, ns)
	if err != nil {
		return nil, nil, nil, err
	}
	left, right := make([]int, len(reps)), make([]int, len(reps))
	for _, c := range classes[:len(s.notes)] {
		left[c]++
	}
	for _, c := range classes[len(s.notes):] {
		right[c]++
	}
	return reps, left, right, nil
}

// SameSet compares distinct content while ignoring order and multiplicity.
func (s *Scale) SameSet(other *Scale, by ...Comparison) (bool, error) {
	_, left, right, err := s.contents(other, by)
	if err != nil {
		return false, err
	}
	for i, n := range left {
		if (n > 0) != (right[i] > 0) {
			return false, nil
		}
	}
	return true, nil
}

// SameMultiset compares content and multiplicity while ignoring order.
func (s *Scale) SameMultiset(other *Scale, by ...Comparison) (bool, error) {
	_, left, right, err := s.contents(other, by)
	if err != nil {
		return false, err
	}
	for i, n := range left {
		if n != right[i] {
			return false, nil
		}
	}
	return true, nil
}

// IsSubsetOf reports whether every distinct note occurs in other.
func (s *Scale) IsSubsetOf(other *Scale, by ...Comparison) (bool, error) {
	_, left, right, err := s.contents(other, by)
	if err != nil {
		return false, err
	}
	for i, n := range left {
		if n > 0 && right[i] == 0 {
			return false, nil
		}
	}
	return true, nil
}
func (s *Scale) combine(other *Scale, kind uint8, by []Comparison) (*Scale, error) {
	reps, left, right, err := s.contents(other, by)
	if err != nil {
		return nil, err
	}
	result := make(note.Notes, 0)
	for i, n := range reps {
		include := kind == 0 && left[i] > 0 && right[i] > 0 || kind == 1 && left[i] > 0 && right[i] == 0 || kind == 2
		if include {
			result = append(result, n)
		}
	}
	return New(result...)
}

// Intersection returns first representatives from the receiver shared with other.
func (s *Scale) Intersection(other *Scale, by ...Comparison) (*Scale, error) {
	return s.combine(other, 0, by)
}

// Difference returns first representatives from the receiver absent from other.
func (s *Scale) Difference(other *Scale, by ...Comparison) (*Scale, error) {
	return s.combine(other, 1, by)
}

// Union returns first representatives from the receiver followed by new ones from other.
func (s *Scale) Union(other *Scale, by ...Comparison) (*Scale, error) { return s.combine(other, 2, by) }

// Unique returns first representatives in their original order.
func (s *Scale) Unique(by ...Comparison) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	p, err := prepare(by, s.notes)
	if err != nil {
		return nil, err
	}
	reps, _, err := partition(p, s.notes)
	if err != nil {
		return nil, err
	}
	return New(reps...)
}

// RotationStarts returns every receiver start index whose rotation equals other.
func (s *Scale) RotationStarts(other *Scale, by ...Comparison) ([]int, error) {
	if s == nil || other == nil {
		return nil, ErrUnset
	}
	p, err := prepare(by, s.notes, other.notes)
	if err != nil {
		return nil, err
	}
	result := make([]int, 0)
	if len(s.notes) != len(other.notes) {
		return result, nil
	}
	for start := range s.notes {
		match := true
		for i, n := range other.notes {
			same, err := p.equal(s.notes[wrap(start, i, len(s.notes))], n)
			if err != nil {
				return nil, err
			}
			if !same {
				match = false
				break
			}
		}
		if match {
			result = append(result, start)
		}
	}
	return result, nil
}

// IsRotationOf reports whether the scales are circular rotations, including two empty lists.
func (s *Scale) IsRotationOf(other *Scale, by ...Comparison) (bool, error) {
	starts, err := s.RotationStarts(other, by...)
	if err != nil {
		return false, err
	}
	return len(starts) > 0 || s.Len() == 0 && other.Len() == 0, nil
}
