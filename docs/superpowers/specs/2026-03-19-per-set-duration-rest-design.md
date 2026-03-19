# Per-Set Duration & Rest Time for Exercise Sets

## Problem

The `activity exercise-sets set` command currently supports only a global `--rest` flag that inserts uniform REST sets between all ACTIVE sets. This has two gaps:

1. **No ACTIVE set duration** — the Garmin API accepts a `duration` field on ACTIVE sets (how long the set took), but the CLI cannot set it.
2. **No per-set rest time** — the API supports different rest durations between different exercise pairs, but `--rest` applies the same value everywhere.

## Design

### Format Change

Extend the `-e` exercise flag format from:

```
CATEGORY/NAME:reps@weight
```

to:

```
CATEGORY/NAME:reps@weight:dSECS:rSECS
```

where `:dSECS` sets the ACTIVE set duration and `:rSECS` controls the REST set inserted after.

### Parsing Rules

- `:d` and `:r` suffixes are optional and order-independent (`:d30:r60` and `:r60:d30` are equivalent).
- No value after `:d` or `:r` means 0 (e.g., `:d` equals `:d0`).
- If **neither** `:d` nor `:r` is present, the set behaves as today: `duration` is null on the ACTIVE set and no REST set is inserted. This preserves backward compatibility.
- If **either** is present, the missing one defaults to 0. So `:d20` means duration=20, rest=0 (a REST set with duration 0 is inserted). `:r60` means duration=0, rest=60.

### API Mapping

| CLI input | ACTIVE set `duration` | REST set inserted? | REST `duration` |
|---|---|---|---|
| No `:d`/`:r` | `null` | No | N/A |
| `:d20:r60` | `20.0` | Yes | `60.0` |
| `:d20` | `20.0` | Yes | `0.0` |
| `:r60` | `0.0` | Yes | `60.0` |
| `:d:r` | `0.0` | Yes | `0.0` |

REST sets use the same structure as today: `setType: "REST"`, `weight: -1`, `exercises: []`, `repetitionCount: null`.

### Breaking Change

The global `--rest` flag on `ActivityExerciseSetsSetCmd` is removed. Users must specify rest per-exercise using `:rSECS`.

### Examples

```bash
gccli activity exercise-sets set 12345 \
  -e "BENCH_PRESS/BARBELL_BENCH_PRESS:12@20:d30:r60" \
  -e "BENCH_PRESS/BARBELL_BENCH_PRESS:10@25:d25:r45" \
  -e "BENCH_PRESS/BARBELL_BENCH_PRESS:8@30:d20"
```

This creates: ACTIVE(d=30) → REST(60) → ACTIVE(d=25) → REST(45) → ACTIVE(d=20).

Backward-compatible usage (no duration/rest):

```bash
gccli activity exercise-sets set 12345 \
  -e "BENCH_PRESS/BARBELL_BENCH_PRESS:12@20" \
  -e "BENCH_PRESS/BARBELL_BENCH_PRESS:10@25"
```

This creates: ACTIVE → ACTIVE (no REST sets, no duration).

## Files to Change

- `internal/cmd/activity_exercise_sets.go` — remove `Rest` field from command struct, update `parseExerciseSets` and `parseOneExercise` to handle `:d`/`:r` suffixes, return rest info alongside parsed sets.
- `internal/cmd/activity_exercise_sets_test.go` — update existing tests, add new tests for `:d`/`:r` parsing including edge cases (order independence, no-value defaults, backward compatibility).
- `README.md` — update exercise-sets examples.
- `skills/gccli/SKILL.md` — update command reference.
- `docs/index.html` — update feature cards / command reference.

## Testing Strategy

Unit tests covering:
- Valid format parsing: `:d30:r60`, `:r60:d30`, `:d20`, `:r60`, `:d`, `:r`, `:d:r`, no suffixes.
- Invalid formats: `:dabc`, `:r-5`, duplicate `:d`, duplicate `:r`.
- REST set insertion: verify REST sets appear only when `:r` is specified.
- Backward compatibility: exercises without `:d`/`:r` produce same output as before.
- Full command integration test with mock HTTP server.
