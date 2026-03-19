# Per-Set Duration & Rest Time Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow each exercise set to specify its own duration and rest time via `:dSECS` and `:rSECS` suffixes in the `-e` flag format.

**Architecture:** Extend `parseOneExercise` to split on `:` into up to 4 parts and recognize `d`/`r` prefixed suffixes. Return a parsed result struct that includes optional duration and rest values. `parseExerciseSets` uses the rest value to conditionally insert REST sets. Remove the global `--rest` flag.

**Tech Stack:** Go, stdlib testing, Kong CLI framework

**Spec:** `docs/superpowers/specs/2026-03-19-per-set-duration-rest-design.md`

---

### Task 1: Update parseOneExercise to handle `:d` and `:r` suffixes

**Files:**
- Modify: `internal/cmd/activity_exercise_sets.go:97-146`
- Test: `internal/cmd/activity_exercise_sets_test.go`

The function currently splits on `:` with `SplitN(s, ":", 2)`. It needs to split into up to 4 parts and parse `d`/`r` suffixes.

We also need a new return type to carry the rest value alongside the exercise set, since REST sets are created by the caller.

- [ ] **Step 1: Write failing tests for new format parsing**

Add test cases to `TestParseOneExercise_Valid` in `activity_exercise_sets_test.go`. The test struct needs a new `duration` field (pointer to float64) since we now check duration on ACTIVE sets.

Add these cases to the existing table:

```go
{
    input:    "BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:d30:r60",
    category: "BENCH_PRESS",
    name:     "BARBELL_BENCH_PRESS",
    reps:     12,
    weightMg: 20000,
    duration: ptrFloat64(30),
    restSecs: ptrInt(60),
},
{
    input:    "BENCH_PRESS/BARBELL_BENCH_PRESS:10@25:r45:d25",
    category: "BENCH_PRESS",
    name:     "BARBELL_BENCH_PRESS",
    reps:     10,
    weightMg: 25000,
    duration: ptrFloat64(25),
    restSecs: ptrInt(45),
},
{
    input:    "BENCH_PRESS/BARBELL_BENCH_PRESS:8@30:d20",
    category: "BENCH_PRESS",
    name:     "BARBELL_BENCH_PRESS",
    reps:     8,
    weightMg: 30000,
    duration: ptrFloat64(20),
    restSecs: nil,
},
{
    input:    "BENCH_PRESS/BARBELL_BENCH_PRESS:8@30:r60",
    category: "BENCH_PRESS",
    name:     "BARBELL_BENCH_PRESS",
    reps:     8,
    weightMg: 30000,
    duration: ptrFloat64(0),
    restSecs: ptrInt(60),
},
{
    input:    "BENCH_PRESS/BARBELL_BENCH_PRESS:8@30:d:r",
    category: "BENCH_PRESS",
    name:     "BARBELL_BENCH_PRESS",
    reps:     8,
    weightMg: 30000,
    duration: ptrFloat64(0),
    restSecs: ptrInt(0),
},
```

Add helper functions:

```go
func ptrFloat64(v float64) *float64 { return &v }
func ptrInt(v int) *int             { return &v }
```

Update existing test cases to include `duration: nil, restSecs: nil` (backward compat — no `:d`/`:r`).

Update the test assertions to check `set.Duration` and the returned `restSecs` value. Since `parseOneExercise` now returns a `parsedExercise` struct instead of just `exerciseSet`, update the call sites.

Also add invalid format cases to `TestParseOneExercise_Invalid`:

```go
{"BENCH_PRESS/BARBELL:12@20:dabc", "invalid duration"},
{"BENCH_PRESS/BARBELL:12@20:r-5", "invalid rest"},
{"BENCH_PRESS/BARBELL:12@20:d10:d20", "duplicate :d"},
{"BENCH_PRESS/BARBELL:12@20:r10:r20", "duplicate :r"},
{"BENCH_PRESS/BARBELL:12@20:x10", "unknown suffix"},
{"BENCH_PRESS/BARBELL:12@20:d1:r2:extra", "expected format"},
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run 'TestParseOneExercise' -v ./internal/cmd/`
Expected: compilation errors (parsedExercise type does not exist yet)

- [ ] **Step 3: Implement the new parsing logic**

In `activity_exercise_sets.go`:

1. Add a `parsedExercise` struct:

```go
// parsedExercise holds the parsed ACTIVE set and optional rest duration.
type parsedExercise struct {
	set      exerciseSet
	restSecs *int // nil = no REST set; non-nil = insert REST set with this duration
}
```

2. Rewrite `parseOneExercise` to return `parsedExercise`:

```go
func parseOneExercise(s string) (parsedExercise, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 4 {
		return parsedExercise{}, fmt.Errorf("expected format CATEGORY/NAME:reps@weightkg[:dSECS][:rSECS]")
	}

	// Part 1: CATEGORY/NAME
	exParts := strings.SplitN(parts[0], "/", 2)
	if len(exParts) != 2 {
		return parsedExercise{}, fmt.Errorf("expected CATEGORY/NAME before ':'")
	}
	category := strings.ToUpper(exParts[0])
	name := strings.ToUpper(exParts[1])

	// Part 2: reps@weight
	var reps int
	var weightKg float64

	atParts := strings.SplitN(parts[1], "@", 2)
	reps, err := strconv.Atoi(atParts[0])
	if err != nil {
		return parsedExercise{}, fmt.Errorf("invalid rep count %q", atParts[0])
	}
	if len(atParts) == 2 {
		weightStr := strings.TrimSuffix(atParts[1], "kg")
		weightKg, err = strconv.ParseFloat(weightStr, 64)
		if err != nil {
			return parsedExercise{}, fmt.Errorf("invalid weight %q", atParts[1])
		}
	}

	// Parts 3-4: optional :dSECS and :rSECS
	var duration *float64
	var restSecs *int
	for _, suffix := range parts[2:] {
		switch {
		case strings.HasPrefix(suffix, "d"):
			if duration != nil {
				return parsedExercise{}, fmt.Errorf("duplicate :d suffix")
			}
			val := suffix[1:]
			if val == "" {
				d := 0.0
				duration = &d
			} else {
				n, err := strconv.Atoi(val)
				if err != nil || n < 0 {
					return parsedExercise{}, fmt.Errorf("invalid duration value %q", val)
				}
				d := float64(n)
				duration = &d
			}
		case strings.HasPrefix(suffix, "r"):
			if restSecs != nil {
				return parsedExercise{}, fmt.Errorf("duplicate :r suffix")
			}
			val := suffix[1:]
			if val == "" {
				r := 0
				restSecs = &r
			} else {
				n, err := strconv.Atoi(val)
				if err != nil || n < 0 {
					return parsedExercise{}, fmt.Errorf("invalid rest value %q", val)
				}
				restSecs = &n
			}
		default:
			return parsedExercise{}, fmt.Errorf("unknown suffix %q, expected :d or :r", suffix)
		}
	}

	// If :r is present but :d is not, default duration to 0.
	if restSecs != nil && duration == nil {
		d := 0.0
		duration = &d
	}

	weightMg := weightKg * 1000

	return parsedExercise{
		set: exerciseSet{
			Exercises:       []exerciseRef{{Probability: 100, Category: category, Name: name}},
			RepetitionCount: &reps,
			Duration:        duration,
			Weight:          weightMg,
			SetType:         "ACTIVE",
		},
		restSecs: restSecs,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run 'TestParseOneExercise' -v ./internal/cmd/`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/activity_exercise_sets.go internal/cmd/activity_exercise_sets_test.go
git commit -m "feat: add :d and :r suffix parsing to exercise format"
```

---

### Task 2: Update parseExerciseSets and remove global --rest flag

**Files:**
- Modify: `internal/cmd/activity_exercise_sets.go:10-15,45,68-95`
- Test: `internal/cmd/activity_exercise_sets_test.go`

- [ ] **Step 1: Write failing tests for per-set REST insertion**

Replace `TestParseExerciseSets_WithRest` and `TestParseExerciseSets_NoRest` with new tests:

```go
func TestParseExerciseSets_PerSetRest(t *testing.T) {
	exercises := []string{
		"BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:d30:r60",
		"BENCH_PRESS/BARBELL_BENCH_PRESS:10@25:d25:r45",
		"BENCH_PRESS/BARBELL_BENCH_PRESS:8@30:d20",
	}

	sets, err := parseExerciseSets(exercises)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 active + 2 rest = 5 sets total.
	if len(sets) != 5 {
		t.Fatalf("sets count = %d, want 5", len(sets))
	}
	if sets[0].SetType != "ACTIVE" {
		t.Errorf("set[0].setType = %q, want ACTIVE", sets[0].SetType)
	}
	if sets[0].Duration == nil || *sets[0].Duration != 30 {
		t.Errorf("set[0].duration = %v, want 30", sets[0].Duration)
	}
	if sets[1].SetType != "REST" {
		t.Errorf("set[1].setType = %q, want REST", sets[1].SetType)
	}
	if sets[1].Duration == nil || *sets[1].Duration != 60 {
		t.Errorf("set[1].rest duration = %v, want 60", sets[1].Duration)
	}
	if sets[2].SetType != "ACTIVE" {
		t.Errorf("set[2].setType = %q, want ACTIVE", sets[2].SetType)
	}
	if sets[3].SetType != "REST" {
		t.Errorf("set[3].setType = %q, want REST", sets[3].SetType)
	}
	if sets[3].Duration == nil || *sets[3].Duration != 45 {
		t.Errorf("set[3].rest duration = %v, want 45", sets[3].Duration)
	}
	if sets[4].SetType != "ACTIVE" {
		t.Errorf("set[4].setType = %q, want ACTIVE", sets[4].SetType)
	}
	if sets[4].Duration == nil || *sets[4].Duration != 20 {
		t.Errorf("set[4].duration = %v, want 20", sets[4].Duration)
	}
}

func TestParseExerciseSets_NoSuffixes(t *testing.T) {
	exercises := []string{
		"BENCH_PRESS/BARBELL_BENCH_PRESS:12@20",
		"BENCH_PRESS/BARBELL_BENCH_PRESS:10@20",
	}

	sets, err := parseExerciseSets(exercises)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sets) != 2 {
		t.Fatalf("sets count = %d, want 2", len(sets))
	}
	if sets[0].Duration != nil {
		t.Errorf("set[0].duration = %v, want nil", sets[0].Duration)
	}
}

func TestParseExerciseSets_LastExerciseWithRest(t *testing.T) {
	exercises := []string{
		"BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:r60",
	}

	sets, err := parseExerciseSets(exercises)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1 active + 1 rest = 2 (REST inserted even for last exercise).
	if len(sets) != 2 {
		t.Fatalf("sets count = %d, want 2", len(sets))
	}
	if sets[1].SetType != "REST" {
		t.Errorf("set[1].setType = %q, want REST", sets[1].SetType)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run 'TestParseExerciseSets' -v ./internal/cmd/`
Expected: compilation error (`parseExerciseSets` signature changed)

- [ ] **Step 3: Update parseExerciseSets and command struct**

1. Remove `Rest` field from `ActivityExerciseSetsSetCmd`:

```go
type ActivityExerciseSetsSetCmd struct {
	ID       string   `arg:"" help:"Activity ID."`
	Exercise []string `help:"Exercise set in format CATEGORY/NAME:reps@weightkg[:dSECS][:rSECS] (e.g. BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:d30:r60)." short:"e" required:""`
}
```

2. Update `Run` to call `parseExerciseSets(c.Exercise)` (no second arg).

3. Rewrite `parseExerciseSets`:

```go
func parseExerciseSets(exercises []string) ([]exerciseSet, error) {
	var sets []exerciseSet

	for i, ex := range exercises {
		parsed, err := parseOneExercise(ex)
		if err != nil {
			return nil, fmt.Errorf("exercise %d (%q): %w", i+1, ex, err)
		}
		sets = append(sets, parsed.set)

		if parsed.restSecs != nil {
			dur := float64(*parsed.restSecs)
			sets = append(sets, exerciseSet{
				Exercises:       []exerciseRef{},
				RepetitionCount: nil,
				Duration:        &dur,
				Weight:          -1,
				SetType:         "REST",
			})
		}
	}

	return sets, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run 'TestParseExerciseSets|TestParseOneExercise|TestActivityExerciseSets' -v ./internal/cmd/`
Expected: all PASS

- [ ] **Step 5: Update integration test**

Update `TestActivityExerciseSetsSet_Success` to use the new format (remove `Rest` field, use `:r60` suffix instead):

```go
cmd := &ActivityExerciseSetsSetCmd{
    ID: "99999",
    Exercise: []string{
        "BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:r60",
        "BENCH_PRESS/BARBELL_BENCH_PRESS:10@25",
    },
}
```

- [ ] **Step 6: Run full test suite**

Run: `go test -v ./internal/cmd/`
Expected: all PASS

- [ ] **Step 7: Run linter and formatter**

Run: `make fmt && make lint`
Expected: clean

- [ ] **Step 8: Commit**

```bash
git add internal/cmd/activity_exercise_sets.go internal/cmd/activity_exercise_sets_test.go
git commit -m "feat: per-set duration and rest time, remove global --rest flag"
```

---

### Task 3: Update documentation

**Files:**
- Modify: `README.md:269-270`
- Modify: `skills/gccli/SKILL.md:47-48,159,174-184`
- Modify: `docs/index.html:530`

- [ ] **Step 1: Update README.md**

Update the exercise-sets example at line ~270 from:

```
gccli activity exercise-sets set <id> -e "BENCH_PRESS/BARBELL_BENCH_PRESS:12@20" -e "BENCH_PRESS/BARBELL_BENCH_PRESS:10@25" --rest 60
```

to:

```
gccli activity exercise-sets set <id> -e "BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:d30:r60" -e "BENCH_PRESS/BARBELL_BENCH_PRESS:10@25:d25"
```

- [ ] **Step 2: Update skills/gccli/SKILL.md**

Update the command reference at line ~48 from:

```
- Set exercise sets: `gccli activity exercise-sets set <id> -e "CATEGORY/NAME:reps@weightkg" [-e ...] [--rest 60]`
```

to:

```
- Set exercise sets: `gccli activity exercise-sets set <id> -e "CATEGORY/NAME:reps@weightkg[:dSECS][:rSECS]" [-e ...]`
```

Update the strength training workflow section (~lines 159, 174-184) to use the new format with per-set `:d` and `:r` suffixes instead of `--rest`.

- [ ] **Step 3: Update docs/index.html**

Update the command reference at line ~530 to use the new format.

- [ ] **Step 4: Verify formatting**

Read through each changed file to confirm the examples are consistent.

- [ ] **Step 5: Commit**

```bash
git add README.md skills/gccli/SKILL.md docs/index.html
git commit -m "docs: update exercise-sets examples for per-set duration and rest"
```

---

### Task 4: Final verification

- [ ] **Step 1: Run full CI gate**

Run: `make ci`
Expected: fmt-check, lint, and test all pass

- [ ] **Step 2: Manual smoke test**

Run: `make run -- activity exercise-sets set --help`
Expected: help text shows new format with `[:dSECS][:rSECS]`, no `--rest` flag
