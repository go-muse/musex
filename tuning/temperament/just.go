package temperament

var (
	presetJust5Limit = MustFromRatios("5-Limit Just Intonation", []Ratio{
		{1, 1}, {16, 15}, {9, 8}, {6, 5}, {5, 4}, {4, 3},
		{45, 32}, {3, 2}, {8, 5}, {5, 3}, {9, 5}, {15, 8},
	})

	presetJust5LimitSymmetric1 = MustFromRatios("5-Limit Just Intonation (Symmetric I)", []Ratio{
		{1, 1}, {16, 15}, {9, 8}, {6, 5}, {5, 4}, {4, 3},
		{45, 32}, {3, 2}, {8, 5}, {5, 3}, {16, 9}, {15, 8},
	})

	presetJust5LimitSymmetric2 = MustFromRatios("5-Limit Just Intonation (Symmetric II)", []Ratio{
		{1, 1}, {16, 15}, {10, 9}, {6, 5}, {5, 4}, {4, 3},
		{45, 32}, {3, 2}, {8, 5}, {5, 3}, {9, 5}, {15, 8},
	})

	presetJust7Limit = MustFromRatios("7-Limit Just Intonation", []Ratio{
		{1, 1}, {15, 14}, {8, 7}, {6, 5}, {5, 4}, {4, 3},
		{7, 5}, {3, 2}, {8, 5}, {5, 3}, {7, 4}, {15, 8},
	})
)

// Just5Limit returns the classical asymmetric 5-limit just intonation
// ("5-Limit Just Intonation", N = 12, D=9/8, F#=45/32, Bb=9/5).
func Just5Limit() Temperament {
	return presetJust5Limit
}

// Just is an alias for [Just5Limit] ("5-Limit Just Intonation", N = 12).
func Just() Temperament {
	return presetJust5Limit
}

// Just5LimitSymmetric1 returns the first inversionally symmetric 5-limit just
// intonation ("5-Limit Just Intonation (Symmetric I)", N = 12, D=9/8,
// F#=45/32, Bb=16/9).
func Just5LimitSymmetric1() Temperament {
	return presetJust5LimitSymmetric1
}

// Just5LimitSymmetric2 returns the second inversionally symmetric 5-limit just
// intonation ("5-Limit Just Intonation (Symmetric II)", N = 12, D=10/9,
// F#=45/32, Bb=9/5).
func Just5LimitSymmetric2() Temperament {
	return presetJust5LimitSymmetric2
}

// Just7Limit returns a 12-tone 7-limit just intonation
// ("7-Limit Just Intonation", N = 12: 1/1, 15/14, 8/7, 6/5, 5/4, 4/3, 7/5,
// 3/2, 8/5, 5/3, 7/4, 15/8). The table coincides at all 12 positions with the
// scale smithgw_hahn12 of the Scala scale archive; it is one of many 12-tone
// 7-limit tunings, which differ in particular in the tritone.
func Just7Limit() Temperament {
	return presetJust7Limit
}
