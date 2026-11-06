package cents_test

import (
	"fmt"
	"math/big"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// appendValue is a value for the AppendText checks; literal is its
// canonical text when the criteria give it ("" otherwise: the oracle of the
// content is String of the same value, checked by K-23).
type appendValue struct {
	name    string
	v       cents.Cents
	literal string
}

// appendValues returns values with texts of different lengths: n = 1
// ("0"), short texts, the ratio of MK (53 bytes), the value with all 11
// bases to the power 1/2 (94 bytes), all values of S and a long product of
// powers (n > 255).
func appendValues(t testing.TB) []appendValue {
	t.Helper()

	set := setS(t)
	literals := map[string]string{
		"0":       "0",
		"<700>":   "700",
		"<-13.7>": "-13.7",
		"R(3:2)":  "3:2",
		"Q":       "5^(1/4)",
		"E(1,19)": "1200/19 cents",
		// 3^53:2^84, 26 + 1 + 26 bytes.
		"MK": "19383245667680019896796723:19342813113834066795298816",
	}

	vs := make([]appendValue, 0, len(set)+2)
	used := 0
	for _, a := range set {
		literal, ok := literals[a.name]
		if ok {
			used++
		}

		vs = append(vs, appendValue{a.name, a.v, literal})
	}

	// Every literal names a value of S: a renamed value would otherwise
	// silently lose its literal check (N-8).
	require.Equal(t, len(literals), used, "literals of names missing from S")

	half := make([]cents.Power, 0, len(allBases()))
	for _, p := range allBases() {
		half = append(half, cents.Power{Base: p, Num: big.NewInt(1), Den: big.NewInt(2)})
	}

	vs = append(vs, appendValue{
		"all bases 1/2",
		noErr(cents.FromComposition(nil, nil, half...))(t),
		"2^(1/2)*3^(1/2)*5^(1/2)*7^(1/2)*11^(1/2)*13^(1/2)*17^(1/2)*19^(1/2)*23^(1/2)*29^(1/2)*31^(1/2)",
	})

	vs = append(vs, appendValue{"long power", longPowerValue(t), ""})

	return vs
}

// longPowerValue returns the value with the 10 odd primes p_i to the
// powers 1/(10^29 + i), i = 1...10: denominators below the record limit
// 10^30, a tiny value in range, and a canonical text longer than 255 bytes
// (its length is taken from String, not from a literal).
func longPowerValue(t testing.TB) cents.Cents {
	t.Helper()

	primes := oddPrimes()
	ps := make([]cents.Power, 0, len(primes))
	for i, p := range primes {
		den := new(big.Int).Add(pow10(29), big.NewInt(int64(i+1)))
		ps = append(ps, cents.Power{Base: p, Num: big.NewInt(1), Den: den})
	}

	return noErr(cents.FromComposition(nil, nil, ps...))(t, "long power value")
}

// appendBuffer describes the buffer b of one AppendText case.
type appendBuffer struct {
	name   string
	prefix string // b[:len(b)]
	isNil  bool   // b == nil
	spare  int    // cap(b) - len(b)
	mark   byte   // the content of the spare capacity
}

// build returns a new buffer: nil, or a slice made with an explicit
// capacity (the capacity of []byte("...") is not defined by the language),
// with the prefix and the mark in all its spare capacity.
func (a appendBuffer) build() []byte {
	if a.isNil {
		return nil
	}

	b := make([]byte, len(a.prefix), len(a.prefix)+a.spare)
	copy(b, a.prefix)

	spare := b[len(b):cap(b)]
	for i := range spare {
		spare[i] = a.mark
	}

	return b
}

// Marks of the spare capacity: 0xAA is a byte of no canonical text (not
// ASCII), 0x00 checks that the result does not depend on the content (A7).
const (
	markAA byte = 0xAA
	mark00 byte = 0x00
)

// appendBuffers returns the buffer cases A1...A7 of the implementation plan
// for a text of n bytes.
func appendBuffers(n int) []appendBuffer {
	bufs := []appendBuffer{
		{name: "A1 nil", isNil: true},
		{name: "A2 empty without capacity"},
		{name: "A3 prefix without spare capacity", prefix: "x="},
	}

	for _, prefix := range []string{"", "x="} {
		label := "no prefix"
		if prefix != "" {
			label = "prefix " + prefix
		}

		for _, mark := range []byte{markAA, mark00} {
			// A4: partial spare capacity (the regression of the task 3.8).
			var partial []int
			for _, k := range []int{1, n - 1, n / 2} {
				if k >= 1 && k < n && !slices.Contains(partial, k) {
					partial = append(partial, k)
				}
			}

			for _, k := range partial {
				bufs = append(bufs, appendBuffer{name: fmt.Sprintf("A4 partial, %s, spare %d, mark %02x", label, k, mark), prefix: prefix, spare: k, mark: mark})
			}

			bufs = append(bufs,
				appendBuffer{name: fmt.Sprintf("A5 exact room, %s, mark %02x", label, mark), prefix: prefix, spare: n, mark: mark},
				appendBuffer{name: fmt.Sprintf("A6 more room, %s, mark %02x", label, mark), prefix: prefix, spare: n + 8, mark: mark},
			)
		}
	}

	return bufs
}

// KR-C16, K-23, G-11, K-33: AppendText behaves as one append(b,
// c.String()...): the result is b[:len(b)] followed by the canonical text;
// with enough spare capacity the text is written into the array of b and
// the bytes past it are untouched; otherwise the result is a new array and
// the array of b is not changed at all; b[:len(b)] and the value never
// change; the error is nil.
func TestAppendTextBuffer(t *testing.T) {
	t.Parallel()

	for _, a := range appendValues(t) {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			s := a.v.String()
			require.NotEmpty(t, s, "a canonical text is never empty")

			if a.literal != "" {
				assert.Equal(t, a.literal, s, "canonical text")
			}

			if a.name == "long power" {
				require.Greater(t, len(s), 255, "the long value has more than 255 bytes")
			}

			for _, buf := range appendBuffers(len(s)) {
				t.Run(buf.name, func(t *testing.T) {
					t.Parallel()

					checkAppendText(t, a.v, s, buf)
				})
			}
		})
	}
}

// checkAppendText runs one case of TestAppendTextBuffer.
func checkAppendText(t *testing.T, v cents.Cents, s string, buf appendBuffer) {
	t.Helper()

	n := len(s)
	b := buf.build()
	before := slices.Clone(b[:cap(b)])
	original := v

	result, err := v.AppendText(b)
	require.NoError(t, err)

	// A8: the content of b, the value and the error in every case.
	assert.Equal(t, buf.prefix, string(b), "b[:len(b)] is unchanged")
	assert.True(t, original == v, "the value is unchanged")

	// A1, A2: the text alone in a new slice that is not nil.
	assert.NotNil(t, result, "the result is not nil")
	require.Len(t, result, len(buf.prefix)+n)
	assert.Equal(t, buf.prefix+s, string(result))

	if cap(b)-len(b) >= n {
		// A5, A6: the same array, the text in b[len(b):len(b)+n], the
		// bytes past it untouched.
		assert.Equal(t, cap(b), cap(result), "the result uses the array of b")
		assert.Same(t, &b[:cap(b)][0], &result[:cap(result)][0], "the result uses the array of b")
		assert.Equal(t, buf.prefix+s, string(b[:len(b)+n]), "the text is in the spare capacity of b")
		assert.Equal(t, before[len(b)+n:], b[len(b)+n:cap(b)], "the bytes past the text are untouched")

		return
	}

	// A1...A4: the array of b is not changed at all, also with partial
	// spare capacity; the result is a new array.
	assert.Equal(t, before, b[:cap(b)], "the array of b is unchanged")

	result[0] ^= 0xFF
	assert.Equal(t, before, b[:cap(b)], "the result is a new array")
}

// KR-C16: a chain of AppendText calls on one buffer with enough capacity
// stays in the array of the buffer (the example of the architecture).
func TestAppendTextChain(t *testing.T) {
	t.Parallel()

	buf := make([]byte, 0, 64)
	first := &buf[:1][0]

	buf, err := noErr(cents.FromRatio(3, 2))(t).AppendText(buf)
	require.NoError(t, err)

	buf = append(buf, ' ')

	buf, err = cents.SyntonicComma().AppendText(buf)
	require.NoError(t, err)

	assert.Equal(t, "3:2 81:80", string(buf))
	assert.Equal(t, 64, cap(buf), "the capacity is the same")
	assert.Same(t, first, &buf[0], "the same array")
}

// KR-C16, C-23: AppendText(nil) gives the same text as MarshalText.
func TestAppendTextMatchesMarshalText(t *testing.T) {
	t.Parallel()

	for _, a := range appendValues(t) {
		appended, err := a.v.AppendText(nil)
		require.NoError(t, err, a.name)

		marshaled, err := a.v.MarshalText()
		require.NoError(t, err, a.name)

		assert.Equal(t, string(marshaled), string(appended), a.name)
	}
}

// KR-C16, K-33: a buffer whose spare capacity (1 byte) is too small for any
// text of the values (n >= 2) is not written by AppendText, so workers may
// share it without synchronization; run under the race detector (make test,
// CI). The results are compared after wg.Wait in the test goroutine; for
// each worker only the first error or difference is reported (N-9).
func TestAppendTextSharedBufferWithoutRoom(t *testing.T) {
	t.Parallel()

	var vs []appendValue
	for _, a := range appendValues(t) {
		if len(a.v.String()) >= 2 {
			vs = append(vs, a)
		}
	}

	require.NotEmpty(t, vs)

	shared := appendBuffer{prefix: "x=", spare: 1, mark: markAA}.build()
	before := slices.Clone(shared[:cap(shared)])

	workers := min(runtime.GOMAXPROCS(0)*4, 32)
	got := make([][]string, workers)
	errs := make([][]error, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			<-start

			for _, a := range vs {
				out, err := a.v.AppendText(shared)
				got[i] = append(got[i], string(out))
				errs[i] = append(errs[i], err)
			}
		})
	}

	close(start)
	wg.Wait()

	assert.Equal(t, before, shared[:cap(shared)], "the shared array is unchanged")

	for i := range workers {
		require.Len(t, got[i], len(vs), "worker %d", i)

		for j, a := range vs {
			if !assert.NoError(t, errs[i][j], "worker %d, %s", i, a.name) {
				break
			}

			if want := "x=" + a.v.String(); got[i][j] != want {
				assert.Fail(t, "AppendText result differs", "worker %d, %s: want %q, got %q", i, a.name, want, got[i][j])

				break
			}
		}
	}
}
