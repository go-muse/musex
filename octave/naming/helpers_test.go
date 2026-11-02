package naming_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
)

// The oracles below are the built-in catalogs as literal tables typed from the
// acceptance criteria K-1, K-2 and K-3 of the requirements. They do not depend
// on the package under test. Every call returns a new map, so no test state is
// shared. Strings are in NFC; Cyrillic io (U+0451) and sharp s (U+00DF) are
// written as escapes to rule out decomposed forms.

// russianOracle returns the Russian catalog of K-1: ASPN -1...9.
func russianOracle() map[int]string {
	return map[int]string{
		-1: "субсубконтроктава",
		0:  "субконтроктава",
		1:  "контроктава",
		2:  "большая",
		3:  "малая",
		4:  "первая",
		5:  "вторая",
		6:  "третья",
		7:  "четв\u0451ртая",
		8:  "пятая",
		9:  "шестая",
	}
}

// englishOracle returns the English catalog of K-2: ASPN 0...7.
func englishOracle() map[int]string {
	return map[int]string{
		0: "subcontra",
		1: "contra",
		2: "great",
		3: "small",
		4: "one-line",
		5: "two-line",
		6: "three-line",
		7: "four-line",
	}
}

// germanOracle returns the German catalog of K-3: ASPN 0...7.
func germanOracle() map[int]string {
	return map[int]string{
		0: "Subkontraoktave",
		1: "Kontraoktave",
		2: "gro\u00dfe",
		3: "kleine",
		4: "eingestrichene",
		5: "zweigestrichene",
		6: "dreigestrichene",
		7: "viergestrichene",
	}
}

// builtin is a built-in catalog together with its oracle.
type builtin struct {
	name    string
	catalog func() *naming.Catalog
	oracle  func() map[int]string
}

// builtins returns the three built-in catalogs of the package.
func builtins() []builtin {
	return []builtin{
		{name: "Russian", catalog: naming.Russian, oracle: russianOracle},
		{name: "English", catalog: naming.English, oracle: englishOracle},
		{name: "German", catalog: naming.German, oracle: germanOracle},
	}
}

// namedCatalog is a test case that makes a new catalog on every call.
type namedCatalog struct {
	name    string
	catalog func() *naming.Catalog
}

// customPairs returns the pairs of the custom catalog of K-7. The names are
// arbitrary test values, not built-in names.
func customPairs() map[int]string {
	return map[int]string{4: "middle", -3: "deep", 12: "high"}
}

// numberRange returns the octave numbers -20...20 of K-4 in ascending order.
func numberRange() []int {
	numbers := make([]int, 0, 41)
	for n := -20; n <= 20; n++ {
		numbers = append(numbers, n)
	}

	return numbers
}

// edgeNumbers returns the extreme int values (int is 64-bit) and ±1000 of K-4.
func edgeNumbers() []int {
	return []int{math.MinInt, math.MinInt + 1, -1000, 1000, math.MaxInt - 1, math.MaxInt}
}

// probeNumbers returns numberRange followed by edgeNumbers.
func probeNumbers() []int {
	return append(numberRange(), edgeNumbers()...)
}

// inverse returns the "name -> number" map of an oracle; oracle names are
// distinct.
func inverse(oracle map[int]string) map[string]int {
	names := make(map[string]int, len(oracle))
	for n, name := range oracle {
		names[name] = n
	}

	return names
}

// recoverPanic calls f and reports whether it panicked and with what value.
func recoverPanic(f func()) (recovered any, panicked bool) {
	defer func() {
		recovered = recover()
	}()

	panicked = true
	f()
	panicked = false

	return nil, panicked
}

// assertPanicsWithErrorIs checks that f panics with an error value that
// matches target and does not match other (errors.Is). The text of the error
// is not part of the contract and is not checked.
func assertPanicsWithErrorIs(t *testing.T, target, other error, f func(), msgAndArgs ...any) bool {
	t.Helper()

	recovered, panicked := recoverPanic(f)
	if !panicked {
		return assert.Fail(t, "expected a panic", msgAndArgs...)
	}

	err, ok := recovered.(error)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("panic value %T (%v) is not an error", recovered, recovered), msgAndArgs...)
	}

	return assert.ErrorIs(t, err, target, msgAndArgs...) && assert.NotErrorIs(t, err, other, msgAndArgs...)
}
