package naming

// The data of the built-in catalogs. They are created once, through the same
// checks as any custom catalog, and are only read afterward. Every name is a
// UTF-8 string in Unicode normalization form NFC.
var (
	russian = MustNew(map[int]string{
		-1: "субсубконтроктава",
		0:  "субконтроктава",
		1:  "контроктава",
		2:  "большая",
		3:  "малая",
		4:  "первая",
		5:  "вторая",
		6:  "третья",
		7:  "четвёртая",
		8:  "пятая",
		9:  "шестая",
	})

	english = MustNew(map[int]string{
		0: "subcontra",
		1: "contra",
		2: "great",
		3: "small",
		4: "one-line",
		5: "two-line",
		6: "three-line",
		7: "four-line",
	})

	german = MustNew(map[int]string{
		0: "Subkontraoktave",
		1: "Kontraoktave",
		2: "große",
		3: "kleine",
		4: "eingestrichene",
		5: "zweigestrichene",
		6: "dreigestrichene",
		7: "viergestrichene",
	})
)

// Russian returns the built-in catalog of Russian octave names. It names
// the octaves with ASPN numbers -1 through 9; other octaves have the empty
// name.
//
// The names are lowercase Russian words in NFC form; the name for ASPN 7
// is spelled with the letter yo (U+0451). The names for ASPN -1 (the octave
// below the subcontra octave) and 9 (the sixth octave) are an extension of
// musex: they follow the pattern of the other names but were not found in
// the literature.
//
// Each call returns a new *Catalog that the caller owns; catalogs from
// different calls are equal by [Catalog.Equal] and share read-only data.
func Russian() *Catalog {
	return new(*russian)
}

// English returns the built-in catalog of English octave names. It names
// the octaves with ASPN numbers 0 through 7: "subcontra", "contra", "great",
// "small", "one-line", "two-line", "three-line" and "four-line"; other
// octaves have the empty name.
//
// Each call returns a new *Catalog that the caller owns; catalogs from
// different calls are equal by [Catalog.Equal] and share read-only data.
func English() *Catalog {
	return new(*english)
}

// German returns the built-in catalog of German octave names. It names the
// octaves with ASPN numbers 0 through 7: "Subkontraoktave", "Kontraoktave",
// "große", "kleine", "eingestrichene", "zweigestrichene", "dreigestrichene"
// and "viergestrichene"; other octaves have the empty name.
//
// The nouns are capitalized and the adjectives are lowercase; "große" is
// spelled with ß (U+00DF).
//
// Each call returns a new *Catalog that the caller owns; catalogs from
// different calls are equal by [Catalog.Equal] and share read-only data.
func German() *Catalog {
	return new(*german)
}
