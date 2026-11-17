package template

import "sort"

// Rotation is an entry and the smallest place of its template whose rotation
// equals an identified template.
type Rotation struct {
	Entry Entry
	Place int
}

// Relation describes one related catalog entry. Place is the selected rotation
// of the entry template, or 1 when rotations were not requested. Places is a
// newly allocated caller-owned slice of mismatching places.
type Relation struct {
	Entry  Entry
	Place  int
	Places []int
}

// Count returns the number of mismatching places.
func (r Relation) Count() int {
	return len(r.Places)
}

// Identify returns all entries whose template equals tmpl, in catalog order.
func (c Catalog) Identify(tmpl Template) ([]Entry, error) {
	relations, err := c.Related(tmpl, 0, Adjacent, false)
	if err != nil {
		return nil, err
	}
	result := make([]Entry, len(relations))
	for i, relation := range relations {
		result[i] = relation.Entry
	}
	return result, nil
}

// MustIdentify is Identify without the error result. It panics with the same
// error.
func (c Catalog) MustIdentify(tmpl Template) []Entry {
	return must(c.Identify(tmpl))
}

// IdentifyRotations returns entries for which tmpl is a rotation of the entry
// template. Each entry occurs once with the smallest matching place.
func (c Catalog) IdentifyRotations(tmpl Template) ([]Rotation, error) {
	relations, err := c.Related(tmpl, 0, Adjacent, true)
	if err != nil {
		return nil, err
	}
	result := make([]Rotation, len(relations))
	for i, relation := range relations {
		result[i] = Rotation{Entry: relation.Entry, Place: relation.Place}
	}
	return result, nil
}

// MustIdentifyRotations is IdentifyRotations without the error result. It
// panics with the same error.
func (c Catalog) MustIdentifyRotations(tmpl Template) []Rotation {
	return must(c.IdentifyRotations(tmpl))
}

// Related returns entries of the same kind and length whose templates differ
// from tmpl in at most maxDifferences places under measure. When rotations is
// true, it selects the entry rotation with the fewest mismatches, breaking ties
// by the smallest place. Results are stably ordered by mismatch count and then
// catalog order.
func (c Catalog) Related(tmpl Template, maxDifferences int, measure Measure, rotations bool) ([]Relation, error) {
	if !tmpl.IsValid() {
		return nil, opError("Related", ErrUnset)
	}
	if maxDifferences < 0 || !measure.IsValid() {
		return nil, opError("Related", ErrInvalidArgument)
	}
	result := make([]Relation, 0)
	for _, entry := range c.entries {
		candidate := entry.tmpl
		if candidate.open != tmpl.open || candidate.Len() != tmpl.Len() {
			continue
		}
		if !rotations {
			places, err := candidate.MismatchingPlaces(tmpl, measure)
			if err != nil {
				return nil, err
			}
			if len(places) <= maxDifferences {
				result = append(result, Relation{Entry: entry, Place: 1, Places: places})
			}
			continue
		}

		bestPlace := 0
		var bestPlaces []int
		placesToTry := candidate.Len()
		if candidate.open {
			placesToTry = 1
		}
		for place := 1; place <= placesToTry; place++ {
			rotated, err := candidate.Rotate(place)
			if err != nil {
				return nil, err
			}
			places, err := rotated.MismatchingPlaces(tmpl, measure)
			if err != nil {
				return nil, err
			}
			if bestPlace == 0 || len(places) < len(bestPlaces) {
				bestPlace, bestPlaces = place, places
			}
			if len(bestPlaces) == 0 {
				break
			}
		}
		if bestPlace != 0 && len(bestPlaces) <= maxDifferences {
			result = append(result, Relation{Entry: entry, Place: bestPlace, Places: bestPlaces})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Count() < result[j].Count()
	})
	return result, nil
}

// MustRelated is Related without the error result. It panics with the same
// error.
func (c Catalog) MustRelated(tmpl Template, maxDifferences int, measure Measure, rotations bool) []Relation {
	return must(c.Related(tmpl, maxDifferences, measure, rotations))
}
