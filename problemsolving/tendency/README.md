# Tendency and resolution

`github.com/go-muse/musex/problemsolving/tendency` answers musical questions under an explicitly chosen scheme. It uses existing modes and specific intervals. The package has no global rule selection and does not infer a tonic, voice context, or a musical style from notes.

## Select a question and a scheme

`Catalog` describes sixteen built-ins and their scope. `Builtin` chooses one by ID. Each answer carries the scheme ID, name, coverage, and explanations whose origins distinguish source rules, product choices, and user rules.

| Question | Operations |
| --- | --- |
| Supports, strength, and labels | `Support`, `Supports`, `CompareSupport`, `Role`, `Roles` |
| Direction and goals | `Tendency`, `Tendencies`, `CheckTransition` |
| Consonance under a selected table | `Classify` |
| Two voices in a mode | `ResolveModal`, `CheckModal` |
| Specific interval movements | `ResolveInterval`, `CheckInterval` |
| Additional constraints | `SelectGoals`, `SelectModal`, `SelectInterval` |
| Compare the same question | `CompareSupports`, `CompareRoles`, `CompareTendencies`, `CompareClassifications`, `CompareModalResolutions`, `CompareIntervalResolutions` |

Every operation returning an error has a `Must` form that panics with that error. Ordinary operations return a zero result on an invalid input. Errors support `errors.Is` with the package sentinels.

```go
m := mode.MustFromCode("aeolian")
s := tendency.MustBuiltin(tendency.AMNeighborTargets)
answer, err := s.Tendency(m, tendency.Place{Degree: 2})
if err != nil {
    return err
}
for _, goal := range answer.Goals {
    fmt.Println(goal.Target.Degree, goal.Target.Cycle, goal.Preferred)
}
```

Here the neighbors are I and III, and III is preferred because its minor second is narrower. The separate `AMTextTargets` table names I as preferred for this same degree. These are independent, visible rules.

## Places and knowledge

A `Place` has an actual degree number and a signed cycle. In a closed mode, `{Degree: 1, Cycle: -1}` differs from `{Degree: 1}`. Open modes accept only cycle zero. A degree number outside `1..m.Len()` is an error; `k+1` never silently becomes I of the next cycle. Geometry uses exact symbolic intervals and checks range after cancellation.

`Known` includes known empty results. `MissingRule`, `MissingContext`, `OutsideScope`, and `UnsupportedTask` describe different limitations, without a Go error. `Partial` preserves known components alongside missing ones. `Allowed`, `Denied`, and `Undetermined` are separate from coverage: a proved violation can deny a candidate even when an independent condition is unknown.

```go
m := mode.MustFromCode("ionian")
s := tendency.MustBuiltin(tendency.SeredaTonal)
answer := s.MustTendency(m, tendency.Place{Degree: 1})
// answer.Knowledge == tendency.MissingContext:
// a neutral relative position requires a companion-voice pair rule.

_, err := s.Tendency(m, tendency.Place{Degree: 0})
if errors.Is(err, tendency.ErrInvalidPlace) {
    // Correct the degree number before asking the musical question.
}
```

A tendency answer has separate `DirectionKnowledge`, `GoalKnowledge`, and `DirectKnowledge`. `Immediate` means a direct move to an exact place. `Named` identifies a named goal without proving a direct move. `Ultimate` identifies a final goal, possibly through another degree. Unknown target cycles remain unknown, with no fabricated movement interval. In melodic minor, the text table's VI → I is `Ultimate` through raised VII; `CheckTransition` remains undetermined for that table.

## Built-in boundaries

AM supports and neighboring targets apply to the six explicitly supported major/minor structures with seven degrees and P8 cycle. AM major/minor degree labels apply only to their corresponding natural forms. `AMTextTargets` covers twelve explicit source rows; it does not supply missing rows by analogy. Sereda tonal and modal directions use `modal.Relative`, with opposite direction mappings, and an explicit product rule realizing direction as a neighboring place. Neutral positions require context; unsupported fractional quality remains outside scope. An open boundary has a known empty direct target set.

The seven consonance tables keep their separate coverage and subtypes, including the AM fourth caveat. Only OMT counterpoint, MT21, and Sereda reduce compound intervals to simple parts, retaining P8 as an octave. Unknown qualities remain unclassified.

`AMModalResolution` preserves voice identity and register. It considers neighboring supports, checks crossing, final support/consonance, seconds and sevenths, parallel fifths/octaves, and the special A4/d5 movements. A simple initial interval through P8 is required. Converging moving voices in unison and uncovered final consonance stay uncertain unless another proved violation rejects the candidate. `AMIntervalResolution` covers the specified ascending basic seconds, sevenths, and tritones; different movement pairs remain different candidates even with the same final interval.

## Finite user rules

`New` accepts support, label, tendency, modal-pair, and interval-movement tables. User IDs begin with `user/`. Each row and option needs at least one explanatory `Basis`; all resulting provenance is `UserRule`. Modal sections apply to the exact current structure, including cycle and openness. Interval-only definitions omit `Mode`.

```go
m := mode.MustFromCode("pentatonic-major")
basis := []tendency.Basis{{
    Rule: "my-neighbor", Explanation: "Use the next degree for this phrase.",
}}
s := tendency.MustNew(tendency.Definition{
    ID: "user/phrase", Name: "Phrase rule", Description: "One explicit direct move.",
    Mode: m,
    Tendencies: []tendency.TendencyRule{{
        Degree: 2, DirectComplete: true, Bases: basis,
        Goals: []tendency.GoalRule{{
            Target: tendency.Target{Kind: tendency.Immediate, Degree: 3, CycleKnown: true},
            Bases: basis,
        }},
    }},
})
answer := s.MustTendency(m, tendency.Place{Degree: 2})
```

A nil section means unsupported task. A present empty section has no known rows. A tendency row's nil `Goals` means unknown goals; a present empty goal list means known absence. `DirectComplete` specifically asserts completeness of immediate goals. Named/ultimate goals may have unknown cycles, with `CycleKnown: false`, `Cycle: 0`, and unset direction. Repeated modal rows explicitly enable joint cycle translation; open modes cannot repeat. Crossing in finite two-voice data requires `AllowCrossing`.

Definitions are copied and canonicalized. Conflicting values, duplicate preferred options, invalid places, contradictory directions, overlapping modal rows, and unrepresentable geometry are rejected. `Definition()` returns an independent canonical user definition. Built-in definitions are not exported as infinite tables.

## Filter, compare, and own results

Filters intersect finite place masks, direction masks, stationary-voice requirements, and optional support rules. Nil masks disable a restriction; present empty masks match nothing. Failed conditions exclude a value; missing information leaves it uncertain. Filtering never upgrades musical uncertainty or creates a preference. `Selection.Original` remains an independent complete report.

Comparisons preserve input scheme order. `Present`, `Absent`, and `Unknown` partition scheme indices for each mentioned value. Unknown does not establish disagreement. `ExactGoal` preserves kind and known cycle; `DegreeGoal` groups mentions only by degree and does not prove identical movements. Repeated scheme IDs are rejected.

Schemes are immutable snapshots, safe for concurrent read calls, and can be copied by assignment. Returned report slices belong to the caller, including separate branches of one report. Independent calls share no mutable report data. Assigning an entire report copies Go slice headers as usual; synchronize concurrent mutation of caller-owned inputs or reports. Successful collection results are non-nil, even when empty; `Definition` preserves nil sections because they carry meaning.
