package tempomap

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
)

// SecondsResult is an exact or correctly rounded seconds result.
type SecondsResult struct {
	value       tempo.Seconds
	approximate bool
	places      int
	bound       tempo.Seconds
}

func exactSecondsResult(value tempo.Seconds) SecondsResult {
	return SecondsResult{value: value, bound: tempo.ZeroSeconds()}
}

func approximateSecondsResult(value tempo.Seconds, places int, bound tempo.Seconds) SecondsResult {
	return SecondsResult{value: value, approximate: true, places: places, bound: bound}
}

// IsSet reports whether r is set.
func (r SecondsResult) IsSet() bool { return r.value.IsSet() && r.bound.IsSet() }

// Value returns the exact or rounded seconds.
func (r SecondsResult) Value() (tempo.Seconds, error) {
	if !r.IsSet() {
		return tempo.Seconds{}, simpleError(ErrUnset, "SecondsResult.Value", "receiver")
	}
	return r.value, nil
}

// MustValue is like Value but panics with its error.
func (r SecondsResult) MustValue() tempo.Seconds { return must(r.Value()) }

// Approximate reports whether the value is rounded from an irrational expression.
func (r SecondsResult) Approximate() (bool, error) {
	if !r.IsSet() {
		return false, simpleError(ErrUnset, "SecondsResult.Approximate", "receiver")
	}
	return r.approximate, nil
}

// MustApproximate is like Approximate but panics with its error.
func (r SecondsResult) MustApproximate() bool { return must(r.Approximate()) }

// Places returns the decimal precision and true for an approximate result.
func (r SecondsResult) Places() (int, bool, error) {
	if !r.IsSet() {
		return 0, false, simpleError(ErrUnset, "SecondsResult.Places", "receiver")
	}
	if !r.approximate {
		return 0, false, nil
	}
	return r.places, true, nil
}

// MustPlaces is like Places but panics with its error.
func (r SecondsResult) MustPlaces() (int, bool) { return must2(r.Places()) }

// ErrorBound returns zero for an exact result and the guaranteed rounding bound otherwise.
func (r SecondsResult) ErrorBound() (tempo.Seconds, error) {
	if !r.IsSet() {
		return tempo.Seconds{}, simpleError(ErrUnset, "SecondsResult.ErrorBound", "receiver")
	}
	return r.bound, nil
}

// MustErrorBound is like ErrorBound but panics with its error.
func (r SecondsResult) MustErrorBound() tempo.Seconds { return must(r.ErrorBound()) }

// Equal reports whether r and other are equal in value and precision metadata.
func (r SecondsResult) Equal(other SecondsResult) bool { return r == other }

func scaledResultText(value *big.Rat, places int) string {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	numerator := new(big.Int).Mul(value.Num(), scale)
	numerator.Quo(numerator, value.Denom())
	return numerator.String() + "/" + scale.String()
}

// String returns the canonical result text or SecondsResult(unset).
func (r SecondsResult) String() string {
	if !r.IsSet() {
		return "SecondsResult(unset)"
	}
	if !r.approximate {
		return "s:" + r.value.String()
	}
	value, _ := r.value.MustFraction().BigRat()
	return "s~" + scaledResultText(value, r.places) + "@" + strconv.Itoa(r.places) + ":" + r.bound.String()
}

// MarshalText implements encoding.TextMarshaler.
func (r SecondsResult) MarshalText() ([]byte, error) {
	if !r.IsSet() {
		return nil, simpleError(ErrUnset, "SecondsResult.MarshalText", "receiver")
	}
	return []byte(r.String()), nil
}

// AppendText implements encoding.TextAppender.
func (r SecondsResult) AppendText(dst []byte) ([]byte, error) {
	if !r.IsSet() {
		return dst, simpleError(ErrUnset, "SecondsResult.AppendText", "receiver")
	}
	return append(dst, r.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler and leaves r unchanged on error.
func (r *SecondsResult) UnmarshalText(text []byte) error {
	if r == nil {
		return simpleError(ErrInvalidArgument, "SecondsResult.UnmarshalText", "nil receiver")
	}
	v, err := ParseSecondsResult(string(text))
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// PositionResult is an exact or correctly rounded position result.
type PositionResult struct {
	value       duration.Position
	approximate bool
	places      int
	bound       duration.Position
}

func exactPositionResult(value duration.Position) PositionResult {
	return PositionResult{value: value, bound: duration.Origin()}
}

func approximatePositionResult(value duration.Position, places int, bound duration.Position) PositionResult {
	return PositionResult{value: value, approximate: true, places: places, bound: bound}
}

// IsSet reports whether r is set.
func (r PositionResult) IsSet() bool { return r.value.IsSet() && r.bound.IsSet() }

// Value returns the exact or rounded position.
func (r PositionResult) Value() (duration.Position, error) {
	if !r.IsSet() {
		return duration.Position{}, simpleError(ErrUnset, "PositionResult.Value", "receiver")
	}
	return r.value, nil
}

// MustValue is like Value but panics with its error.
func (r PositionResult) MustValue() duration.Position { return must(r.Value()) }

// Approximate reports whether the value is rounded from an irrational expression.
func (r PositionResult) Approximate() (bool, error) {
	if !r.IsSet() {
		return false, simpleError(ErrUnset, "PositionResult.Approximate", "receiver")
	}
	return r.approximate, nil
}

// MustApproximate is like Approximate but panics with its error.
func (r PositionResult) MustApproximate() bool { return must(r.Approximate()) }

// Places returns the decimal precision and true for an approximate result.
func (r PositionResult) Places() (int, bool, error) {
	if !r.IsSet() {
		return 0, false, simpleError(ErrUnset, "PositionResult.Places", "receiver")
	}
	if !r.approximate {
		return 0, false, nil
	}
	return r.places, true, nil
}

// MustPlaces is like Places but panics with its error.
func (r PositionResult) MustPlaces() (int, bool) { return must2(r.Places()) }

// ErrorBound returns zero for an exact result and the guaranteed rounding bound otherwise.
func (r PositionResult) ErrorBound() (duration.Position, error) {
	if !r.IsSet() {
		return duration.Position{}, simpleError(ErrUnset, "PositionResult.ErrorBound", "receiver")
	}
	return r.bound, nil
}

// MustErrorBound is like ErrorBound but panics with its error.
func (r PositionResult) MustErrorBound() duration.Position { return must(r.ErrorBound()) }

// Equal reports whether r and other are equal in value and precision metadata.
func (r PositionResult) Equal(other PositionResult) bool { return r == other }

// String returns the canonical result text or PositionResult(unset).
func (r PositionResult) String() string {
	if !r.IsSet() {
		return "PositionResult(unset)"
	}
	if !r.approximate {
		return "p:" + r.value.String()
	}
	value, _ := r.value.MustFraction().BigRat()
	return "p~" + scaledResultText(value, r.places) + "@" + strconv.Itoa(r.places) + ":" + r.bound.String()
}

// MarshalText implements encoding.TextMarshaler.
func (r PositionResult) MarshalText() ([]byte, error) {
	if !r.IsSet() {
		return nil, simpleError(ErrUnset, "PositionResult.MarshalText", "receiver")
	}
	return []byte(r.String()), nil
}

// AppendText implements encoding.TextAppender.
func (r PositionResult) AppendText(dst []byte) ([]byte, error) {
	if !r.IsSet() {
		return dst, simpleError(ErrUnset, "PositionResult.AppendText", "receiver")
	}
	return append(dst, r.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler and leaves r unchanged on error.
func (r *PositionResult) UnmarshalText(text []byte) error {
	if r == nil {
		return simpleError(ErrInvalidArgument, "PositionResult.UnmarshalText", "nil receiver")
	}
	v, err := ParsePositionResult(string(text))
	if err != nil {
		return err
	}
	*r = v
	return nil
}

func parsePlaces(text, full string) (int, error) {
	if text == "" || text != "0" && (text[0] < '1' || text[0] > '9') {
		return 0, invalidTextError(full)
	}
	for i := range len(text) {
		if text[i] < '0' || text[i] > '9' {
			return 0, invalidTextError(full)
		}
	}
	places, err := strconv.Atoi(text)
	if err != nil || places > MaxPlaces {
		return 0, simpleError(ErrInvalidArgument, "ParseResult", "decimal places outside 0..1000")
	}
	return places, nil
}

func validResultNode(value *big.Rat, places int) bool {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	scaled := new(big.Rat).Mul(value, new(big.Rat).SetInt(scale))
	return scaled.IsInt()
}

func resultBound(places int, directed bool) (*big.Rat, error) {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	den := new(big.Int).Set(scale)
	if !directed {
		den.Mul(den, big.NewInt(2))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), den), nil
}

func validResultBound(bound *big.Rat, places int) bool {
	directed, _ := resultBound(places, true)
	nearest, _ := resultBound(places, false)
	return bound.Cmp(directed) == 0 || bound.Cmp(nearest) == 0
}

// ParseSecondsResult parses an exact or approximate seconds result.
func ParseSecondsResult(text string) (SecondsResult, error) {
	if strings.HasPrefix(text, "s:") {
		v, err := tempo.ParseSeconds(text[2:])
		if err != nil {
			return SecondsResult{}, nestedTextError(text, err)
		}
		return exactSecondsResult(v), nil
	}
	if !strings.HasPrefix(text, "s~") {
		return SecondsResult{}, invalidTextError(text)
	}
	valueText, rest, ok := strings.Cut(text[2:], "@")
	if !ok {
		return SecondsResult{}, invalidTextError(text)
	}
	placesText, boundText, ok := strings.Cut(rest, ":")
	if !ok || strings.Contains(boundText, ":") {
		return SecondsResult{}, invalidTextError(text)
	}
	places, err := parsePlaces(placesText, text)
	if err != nil {
		return SecondsResult{}, err
	}
	value, err := tempo.ParseSeconds(valueText)
	if err != nil {
		return SecondsResult{}, nestedTextError(text, err)
	}
	bound, err := tempo.ParseSeconds(boundText)
	if err != nil {
		return SecondsResult{}, nestedTextError(text, err)
	}
	valueRat, _ := value.MustFraction().BigRat()
	boundRat, _ := bound.MustFraction().BigRat()
	if !validResultNode(valueRat, places) || !validResultBound(boundRat, places) {
		return SecondsResult{}, invalidTextError(text)
	}
	return approximateSecondsResult(value, places, bound), nil
}

// MustParseSecondsResult is like ParseSecondsResult but panics with its error.
func MustParseSecondsResult(text string) SecondsResult { return must(ParseSecondsResult(text)) }

// ParsePositionResult parses an exact or approximate position result.
func ParsePositionResult(text string) (PositionResult, error) {
	if strings.HasPrefix(text, "p:") {
		v, err := duration.ParsePosition(text[2:])
		if err != nil {
			return PositionResult{}, nestedTextError(text, err)
		}
		return exactPositionResult(v), nil
	}
	if !strings.HasPrefix(text, "p~") {
		return PositionResult{}, invalidTextError(text)
	}
	valueText, rest, ok := strings.Cut(text[2:], "@")
	if !ok {
		return PositionResult{}, invalidTextError(text)
	}
	placesText, boundText, ok := strings.Cut(rest, ":")
	if !ok || strings.Contains(boundText, ":") {
		return PositionResult{}, invalidTextError(text)
	}
	places, err := parsePlaces(placesText, text)
	if err != nil {
		return PositionResult{}, err
	}
	value, err := duration.ParsePosition(valueText)
	if err != nil {
		return PositionResult{}, nestedTextError(text, err)
	}
	bound, err := duration.ParsePosition(boundText)
	if err != nil {
		return PositionResult{}, nestedTextError(text, err)
	}
	valueRat, _ := value.MustFraction().BigRat()
	boundRat, _ := bound.MustFraction().BigRat()
	if !validResultNode(valueRat, places) || !validResultBound(boundRat, places) {
		return PositionResult{}, invalidTextError(text)
	}
	return approximatePositionResult(value, places, bound), nil
}

// MustParsePositionResult is like ParsePositionResult but panics with its error.
func MustParsePositionResult(text string) PositionResult { return must(ParsePositionResult(text)) }
