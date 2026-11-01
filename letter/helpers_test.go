package letter_test

import (
	"github.com/go-muse/musex/letter"
)

// oracleCount is the number of letters in the seven-letter set, defined
// independently of the package under test.
const oracleCount = 7

// oracle returns the seven letters in canonical order C D E F G A B.
// The ordinal of a letter is its index in this array. The array is returned
// by value, so every caller gets its own copy and no test state is shared.
func oracle() [oracleCount]letter.Letter {
	return [oracleCount]letter.Letter{
		letter.C, letter.D, letter.E, letter.F, letter.G, letter.A, letter.B,
	}
}

// oracleNames returns the expected string forms in canonical order.
func oracleNames() [oracleCount]string {
	return [oracleCount]string{"C", "D", "E", "F", "G", "A", "B"}
}

// mod7 returns the mathematical (non-negative) remainder of x modulo 7.
// It does not overflow for any int value, including math.MinInt.
func mod7(x int) int {
	return ((x % oracleCount) + oracleCount) % oracleCount
}

// expectedShift returns the letter at place (ordinal + steps) mod 7.
// The remainder of steps is taken before the addition, so the intermediate
// values stay within a few units and never overflow.
func expectedShift(ordinal, steps int) letter.Letter {
	return oracle()[mod7(ordinal+mod7(steps))]
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
