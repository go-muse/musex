package ticks

import (
	"math/big"
	"strings"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
)

// StepMethod selects the tempo of each step of a gradual change. The zero
// value and values outside the constants are invalid.
type StepMethod uint8

const (
	StepStart     StepMethod = iota + 1 // speed at the start of the step
	StepEnd                             // speed at the end of the step
	StepEqualTime                       // the step lasts as long as that part of the change
)

var stepMethodNames = [...]string{"", "start", "end", "equal-time"}

// IsValid reports whether s is one of the three methods.
func (s StepMethod) IsValid() bool { return s >= StepStart && s <= StepEqualTime }

// Equal reports whether s and other are the same value.
func (s StepMethod) Equal(other StepMethod) bool { return s == other }

// String returns "start", "end" or "equal-time", or "StepMethod(invalid)".
func (s StepMethod) String() string {
	if !s.IsValid() {
		return "StepMethod(invalid)"
	}
	return stepMethodNames[s]
}

// ParseStepMethod parses the text of String.
func ParseStepMethod(text string) (StepMethod, error) {
	for s := StepStart; s <= StepEqualTime; s++ {
		if stepMethodNames[s] == text {
			return s, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseStepMethod is like ParseStepMethod but panics with its error.
func MustParseStepMethod(text string) StepMethod { return must(ParseStepMethod(text)) }

// HoldPolicy selects how fermata holds of a musex map are exported. The zero
// value and values outside the constants are invalid.
type HoldPolicy uint8

const (
	HoldsRefuse  HoldPolicy = iota + 1 // a hold is an error
	HoldsSkip                          // holds are skipped and reported
	HoldsStretch                       // the segment before a hold is slowed down
)

var holdPolicyNames = [...]string{"", "refuse", "skip", "stretch"}

// IsValid reports whether h is one of the three policies.
func (h HoldPolicy) IsValid() bool { return h >= HoldsRefuse && h <= HoldsStretch }

// Equal reports whether h and other are the same value.
func (h HoldPolicy) Equal(other HoldPolicy) bool { return h == other }

// String returns "refuse", "skip" or "stretch", or "HoldPolicy(invalid)".
func (h HoldPolicy) String() string {
	if !h.IsValid() {
		return "HoldPolicy(invalid)"
	}
	return holdPolicyNames[h]
}

// ParseHoldPolicy parses the text of String.
func ParseHoldPolicy(text string) (HoldPolicy, error) {
	for h := HoldsRefuse; h <= HoldsStretch; h++ {
		if holdPolicyNames[h] == text {
			return h, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseHoldPolicy is like ParseHoldPolicy but panics with its error.
func MustParseHoldPolicy(text string) HoldPolicy { return must(ParseHoldPolicy(text)) }

// ExportOptions configures MIDITempoMapOf: the rounding mode of tempos, an
// optional rounding mode of positions, optional steps of gradual changes and
// their method, steps on MIDI clocks, the hold policy and the stretch length.
// Defaults: positions must be representable, gradual changes are refused and
// holds are refused. The zero value is unset. ExportOptions is comparable:
// a == b reports the same as a.Equal(b).
type ExportOptions struct {
	tempoMode    rounding.Mode
	positionMode rounding.Mode
	step         duration.Duration
	method       StepMethod
	clock        bool
	holds        HoldPolicy
	stretch      duration.Duration
}

// NewExportOptions returns the default options with the rounding mode of
// tempos; an unset or unknown mode is ErrInvalidArgument.
func NewExportOptions(tempoMode rounding.Mode) (ExportOptions, error) {
	if err := checkMode("NewExportOptions", tempoMode); err != nil {
		return ExportOptions{}, err
	}
	return ExportOptions{tempoMode: tempoMode, holds: HoldsRefuse}, nil
}

// MustNewExportOptions is like NewExportOptions but panics with its error.
func MustNewExportOptions(tempoMode rounding.Mode) ExportOptions {
	return must(NewExportOptions(tempoMode))
}

// IsSet reports whether o is set.
func (o ExportOptions) IsSet() bool { return o.tempoMode != 0 }

func (o ExportOptions) check(op string) error {
	if !o.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// TempoMode returns the rounding mode of tempos.
func (o ExportOptions) TempoMode() (rounding.Mode, error) {
	if err := o.check("ExportOptions.TempoMode"); err != nil {
		return 0, err
	}
	return o.tempoMode, nil
}

// MustTempoMode is like TempoMode but panics with its error.
func (o ExportOptions) MustTempoMode() rounding.Mode { return must(o.TempoMode()) }

// PositionMode returns the rounding mode of positions and true, or false if
// positions must be representable.
func (o ExportOptions) PositionMode() (mode rounding.Mode, ok bool, err error) {
	if err := o.check("ExportOptions.PositionMode"); err != nil {
		return 0, false, err
	}
	return o.positionMode, o.positionMode != 0, nil
}

// MustPositionMode is like PositionMode but panics with its error.
func (o ExportOptions) MustPositionMode() (mode rounding.Mode, ok bool) {
	return must2(o.PositionMode())
}

// Steps returns the step and the method of gradual changes and true, or false
// if gradual changes are refused.
func (o ExportOptions) Steps() (step duration.Duration, method StepMethod, ok bool, err error) {
	if err := o.check("ExportOptions.Steps"); err != nil {
		return duration.Duration{}, 0, false, err
	}
	return o.step, o.method, o.method != 0, nil
}

// MustSteps is like Steps but panics with its error.
func (o ExportOptions) MustSteps() (step duration.Duration, method StepMethod, ok bool) {
	return must3(o.Steps())
}

// ClockSteps reports whether every step must start on a MIDI clock.
func (o ExportOptions) ClockSteps() (bool, error) {
	if err := o.check("ExportOptions.ClockSteps"); err != nil {
		return false, err
	}
	return o.clock, nil
}

// MustClockSteps is like ClockSteps but panics with its error.
func (o ExportOptions) MustClockSteps() bool { return must(o.ClockSteps()) }

// Holds returns the hold policy.
func (o ExportOptions) Holds() (HoldPolicy, error) {
	if err := o.check("ExportOptions.Holds"); err != nil {
		return 0, err
	}
	return o.holds, nil
}

// MustHolds is like Holds but panics with its error.
func (o ExportOptions) MustHolds() HoldPolicy { return must(o.Holds()) }

// StretchLength returns the length of the stretched segment before a hold and
// true, only with HoldsStretch.
func (o ExportOptions) StretchLength() (duration.Duration, bool, error) {
	if err := o.check("ExportOptions.StretchLength"); err != nil {
		return duration.Duration{}, false, err
	}
	return o.stretch, o.holds == HoldsStretch, nil
}

// MustStretchLength is like StretchLength but panics with its error.
func (o ExportOptions) MustStretchLength() (duration.Duration, bool) {
	return must2(o.StretchLength())
}

// WithPositionMode returns o with positions rounded by mode.
func (o ExportOptions) WithPositionMode(mode rounding.Mode) (ExportOptions, error) {
	const op = "ExportOptions.WithPositionMode"
	if err := o.check(op); err != nil {
		return ExportOptions{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return ExportOptions{}, err
	}
	o.positionMode = mode
	return o, nil
}

// MustWithPositionMode is like WithPositionMode but panics with its error.
func (o ExportOptions) MustWithPositionMode(mode rounding.Mode) ExportOptions {
	return must(o.WithPositionMode(mode))
}

// WithExactPositions returns o with positions that must be representable.
func (o ExportOptions) WithExactPositions() (ExportOptions, error) {
	if err := o.check("ExportOptions.WithExactPositions"); err != nil {
		return ExportOptions{}, err
	}
	o.positionMode = 0
	return o, nil
}

// MustWithExactPositions is like WithExactPositions but panics with its error.
func (o ExportOptions) MustWithExactPositions() ExportOptions { return must(o.WithExactPositions()) }

// WithSteps returns o with gradual changes exported as steps of the duration
// step by method. An unset step or an invalid method is ErrInvalidArgument.
func (o ExportOptions) WithSteps(step duration.Duration, method StepMethod) (ExportOptions, error) {
	const op = "ExportOptions.WithSteps"
	if err := o.check(op); err != nil {
		return ExportOptions{}, err
	}
	if !step.IsSet() {
		return ExportOptions{}, invalidError(op, "step", duration.ErrUnset)
	}
	if !method.IsValid() {
		return ExportOptions{}, invalidError(op, "invalid step method")
	}
	o.step, o.method = step, method
	return o, nil
}

// MustWithSteps is like WithSteps but panics with its error.
func (o ExportOptions) MustWithSteps(step duration.Duration, method StepMethod) ExportOptions {
	return must(o.WithSteps(step, method))
}

// WithoutSteps returns o with gradual changes refused.
func (o ExportOptions) WithoutSteps() (ExportOptions, error) {
	if err := o.check("ExportOptions.WithoutSteps"); err != nil {
		return ExportOptions{}, err
	}
	o.step, o.method = duration.Duration{}, 0
	return o, nil
}

// MustWithoutSteps is like WithoutSteps but panics with its error.
func (o ExportOptions) MustWithoutSteps() ExportOptions { return must(o.WithoutSteps()) }

// WithClockSteps returns o with steps required on MIDI clocks or not.
func (o ExportOptions) WithClockSteps(onClock bool) (ExportOptions, error) {
	if err := o.check("ExportOptions.WithClockSteps"); err != nil {
		return ExportOptions{}, err
	}
	o.clock = onClock
	return o, nil
}

// MustWithClockSteps is like WithClockSteps but panics with its error.
func (o ExportOptions) MustWithClockSteps(onClock bool) ExportOptions {
	return must(o.WithClockSteps(onClock))
}

func (o ExportOptions) withHolds(op string, holds HoldPolicy, stretch duration.Duration) (ExportOptions, error) {
	if err := o.check(op); err != nil {
		return ExportOptions{}, err
	}
	o.holds, o.stretch = holds, stretch
	return o, nil
}

// WithHoldsRefused returns o with holds refused.
func (o ExportOptions) WithHoldsRefused() (ExportOptions, error) {
	return o.withHolds("ExportOptions.WithHoldsRefused", HoldsRefuse, duration.Duration{})
}

// MustWithHoldsRefused is like WithHoldsRefused but panics with its error.
func (o ExportOptions) MustWithHoldsRefused() ExportOptions { return must(o.WithHoldsRefused()) }

// WithHoldsSkipped returns o with holds skipped and reported.
func (o ExportOptions) WithHoldsSkipped() (ExportOptions, error) {
	return o.withHolds("ExportOptions.WithHoldsSkipped", HoldsSkip, duration.Duration{})
}

// MustWithHoldsSkipped is like WithHoldsSkipped but panics with its error.
func (o ExportOptions) MustWithHoldsSkipped() ExportOptions { return must(o.WithHoldsSkipped()) }

// WithHoldsStretched returns o with the segment of the given length before
// each hold slowed down by the hold. An unset length is ErrInvalidArgument.
func (o ExportOptions) WithHoldsStretched(length duration.Duration) (ExportOptions, error) {
	const op = "ExportOptions.WithHoldsStretched"
	if err := o.check(op); err != nil {
		return ExportOptions{}, err
	}
	if !length.IsSet() {
		return ExportOptions{}, invalidError(op, "stretch length", duration.ErrUnset)
	}
	return o.withHolds(op, HoldsStretch, length)
}

// MustWithHoldsStretched is like WithHoldsStretched but panics with its error.
func (o ExportOptions) MustWithHoldsStretched(length duration.Duration) ExportOptions {
	return must(o.WithHoldsStretched(length))
}

// Equal reports whether o and other have equal options.
func (o ExportOptions) Equal(other ExportOptions) bool { return o == other }

// String returns the canonical text, such as
// "export[tempo=half-up;positions=exact;steps=1/4:equal-time;clock=false;holds=stretch:1/4]",
// or "ExportOptions(unset)".
func (o ExportOptions) String() string {
	if !o.IsSet() {
		return "ExportOptions(unset)"
	}
	var b strings.Builder
	b.WriteString("export[tempo=")
	b.WriteString(modeName(o.tempoMode))
	b.WriteString(";positions=")
	if o.positionMode == 0 {
		b.WriteString("exact")
	} else {
		b.WriteString(modeName(o.positionMode))
	}
	b.WriteString(";steps=")
	if o.method == 0 {
		b.WriteString("none")
	} else {
		b.WriteString(ratText(ratOfDuration(o.step)))
		b.WriteByte(':')
		b.WriteString(o.method.String())
	}
	b.WriteString(";clock=")
	b.WriteString(boolText(o.clock))
	b.WriteString(";holds=")
	b.WriteString(o.holds.String())
	if o.holds == HoldsStretch {
		b.WriteByte(':')
		b.WriteString(ratText(ratOfDuration(o.stretch)))
	}
	b.WriteByte(']')
	return b.String()
}

func boolText(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// MarshalText implements encoding.TextMarshaler.
func (o ExportOptions) MarshalText() ([]byte, error) {
	if err := o.check("ExportOptions.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(o.String()), nil
}

// AppendText implements encoding.TextAppender.
func (o ExportOptions) AppendText(dst []byte) ([]byte, error) {
	if err := o.check("ExportOptions.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, o.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves o unchanged on
// error.
func (o *ExportOptions) UnmarshalText(text []byte) error {
	if o == nil {
		return invalidError("ExportOptions.UnmarshalText", "nil receiver")
	}
	v, err := ParseExportOptions(string(text))
	if err != nil {
		return err
	}
	*o = v
	return nil
}

type rawExport struct {
	tempo, positions       rounding.Mode
	stepNum, stepDen       string
	method                 StepMethod
	clock                  bool
	holds                  HoldPolicy
	stretchNum, stretchDen string
}

func scanExport(text string) (rawExport, bool) {
	body, ok := strings.CutPrefix(text, "export[")
	if !ok {
		return rawExport{}, false
	}
	body, ok = strings.CutSuffix(body, "]")
	if !ok {
		return rawExport{}, false
	}
	fields := strings.Split(body, ";")
	keys := [...]string{"tempo=", "positions=", "steps=", "clock=", "holds="}
	if len(fields) != len(keys) {
		return rawExport{}, false
	}
	values := make([]string, len(keys))
	for i, key := range keys {
		if values[i], ok = strings.CutPrefix(fields[i], key); !ok {
			return rawExport{}, false
		}
	}
	var raw rawExport
	if raw.tempo, ok = parseModeName(values[0]); !ok {
		return rawExport{}, false
	}
	if values[1] != "exact" {
		if raw.positions, ok = parseModeName(values[1]); !ok {
			return rawExport{}, false
		}
	}
	if values[2] != "none" {
		ratio, methodText, found := strings.Cut(values[2], ":")
		if !found {
			return rawExport{}, false
		}
		if raw.stepNum, raw.stepDen, ok = scanRatio(ratio); !ok {
			return rawExport{}, false
		}
		method, err := ParseStepMethod(methodText)
		if err != nil {
			return rawExport{}, false
		}
		raw.method = method
	}
	switch values[3] {
	case "true":
		raw.clock = true
	case "false":
	default:
		return rawExport{}, false
	}
	switch {
	case values[4] == "refuse":
		raw.holds = HoldsRefuse
	case values[4] == "skip":
		raw.holds = HoldsSkip
	case strings.HasPrefix(values[4], "stretch:"):
		raw.holds = HoldsStretch
		if raw.stretchNum, raw.stretchDen, ok = scanRatio(values[4][len("stretch:"):]); !ok {
			return rawExport{}, false
		}
	default:
		return rawExport{}, false
	}
	return raw, true
}

// ParseExportOptions parses the text of String. A zero step or stretch length
// is ErrInvalidArgument.
func ParseExportOptions(text string) (ExportOptions, error) {
	const op = "ParseExportOptions"
	raw, ok := scanExport(text)
	if !ok {
		return ExportOptions{}, grammarError(text)
	}
	var step, stretch *big.Rat
	var err error
	if raw.method != 0 {
		if step, err = ratioValue(text, raw.stepNum, raw.stepDen); err != nil {
			return ExportOptions{}, err
		}
	}
	if raw.holds == HoldsStretch {
		if stretch, err = ratioValue(text, raw.stretchNum, raw.stretchDen); err != nil {
			return ExportOptions{}, err
		}
	}
	o := ExportOptions{tempoMode: raw.tempo, positionMode: raw.positions, clock: raw.clock, holds: raw.holds}
	if raw.method != 0 {
		if step.Sign() == 0 {
			return ExportOptions{}, valueTextError(text, invalidError(op, "zero step"))
		}
		d, err := durationOf(op, step)
		if err != nil {
			return ExportOptions{}, valueTextError(text, err)
		}
		o.step, o.method = d, raw.method
	}
	if raw.holds == HoldsStretch {
		if stretch.Sign() == 0 {
			return ExportOptions{}, valueTextError(text, invalidError(op, "zero stretch length"))
		}
		d, err := durationOf(op, stretch)
		if err != nil {
			return ExportOptions{}, valueTextError(text, err)
		}
		o.stretch = d
	}
	return o, nil
}

// MustParseExportOptions is like ParseExportOptions but panics with its error.
func MustParseExportOptions(text string) ExportOptions { return must(ParseExportOptions(text)) }
