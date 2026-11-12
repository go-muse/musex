package notation_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func TestNotationConcurrentOperations(t *testing.T) {
	t.Parallel()
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	german := notation.MustBuiltIn(notation.GermanAMID)
	abc := notation.MustBuiltIn(notation.ABCID)
	named := customNotation(t, english.NameRules(), notation.OctaveStyleNumber, "")
	ctx := notation.OctaveContext{Numbering: number.ASPN()}
	p := pitch.MustFromASPN(spelling.CSharp(), 4)
	jobs := []func() error{
		func() error {
			if !notation.MustBuiltIn(notation.EnglishUnicodeID).Equal(english) {
				return fmt.Errorf("built-in changed")
			}
			return nil
		},
		func() error {
			ids := notation.BuiltInIDs()
			ids[0] = "local"
			if notation.BuiltInIDs()[0] != notation.EnglishUnicodeID {
				return fmt.Errorf("built-in IDs are shared")
			}
			return nil
		},
		func() error {
			names := english.NameRules().Names()
			names[0] = spelling.BDoubleSharp()
			if english.NameRules().Names()[0].Equal(names[0]) {
				return fmt.Errorf("names are shared")
			}
			return nil
		},
		func() error {
			if english.MustFormat(spelling.CSharp()) != "C♯" || !english.MustParse("C♯").Equal(spelling.CSharp()) {
				return fmt.Errorf("name round trip changed")
			}
			return nil
		},
		func() error {
			if named.MustFormatPitch(p, system.Traditional(), ctx) != "C♯4" {
				return fmt.Errorf("pitch formatting changed")
			}
			if !named.MustParsePitch("C♯4", system.Traditional(), ctx).Equal(p) {
				return fmt.Errorf("pitch parsing changed")
			}
			return nil
		},
		func() error {
			if notation.MustConvert("cis", german, english) != "C♯" {
				return fmt.Errorf("name conversion changed")
			}
			if notation.MustConvertPitch("cis¹", german, abc, system.Traditional(), notation.OctaveContext{}, notation.OctaveContext{}) != "^C" {
				return fmt.Errorf("pitch conversion changed")
			}
			return nil
		},
		func() error {
			first, err := notation.NewNameRules(basicGeneratedDefinition())
			if err != nil {
				return err
			}
			second, err := notation.NewNameRules(basicGeneratedDefinition())
			if err != nil {
				return err
			}
			if !first.Equal(second) {
				return fmt.Errorf("custom rules changed")
			}
			return nil
		},
		func() error {
			updown := notation.MustBuiltIn(notation.ASCIIUpDownID, 3)
			if updown.MustFormat(nameWithAlteration(letter.C, 1, 3)) != "^C" {
				return fmt.Errorf("updown changed")
			}
			return nil
		},
	}

	const workers = 24
	const repetitions = 40
	errCh := make(chan error, workers*len(jobs))
	var wg sync.WaitGroup
	for range workers {
		for _, job := range jobs {
			wg.Go(func() {
				for range repetitions {
					if err := job(); err != nil {
						errCh <- err
						return
					}
				}
			})
		}
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}
