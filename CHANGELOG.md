# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-02

### Added

- **`courses` command group** — New commands for managing Garmin Connect courses:
  - `courses list` — List all user courses with table/JSON/plain output
  - `courses favorites` — List favorite courses
  - `courses detail <id>` — View full course detail (JSON)
  - `courses send <course-id> <device-id>` — Send a course to a device via the device message API
- **Courses API methods** — `GetCourses`, `GetCourse`, `GetCourseFavorites`, `SendCourseToDevice` in the API client
- **E2E tests for courses** — End-to-end tests against the real Garmin Connect API

### Changed

- Updated Go module dependencies (`golang.org/x/net`)
- Updated CI dependencies (`actions/checkout` 4→6, `actions/setup-go` 5→6, `goreleaser/goreleaser-action` 6→7)

### Docs

- Added Courses section to website (hero badge, feature card, command reference)
- Added courses commands to skill reference (`skill/SKILL.md`)
- Added shell completions card to install section on website

## [0.3.0] - 2026-02-17

### Added

- **`completion` command** — Generate shell completions for bash, zsh, fish, and PowerShell (`gccli completion <shell>`)
- **E2E workflow** — GitHub Actions workflow to run end-to-end tests on push to `main`, daily at 06:00 UTC, and via manual dispatch
- **Dependabot** — Automated dependency update checks via Dependabot configuration

### Fixed

- **Garmin Connect API compatibility** — Adapted to endpoint and payload changes in the Garmin Connect API

### Changed

- Updated outdated Go module dependencies

### Docs

- Added community health files (contributing guide, code of conduct, etc.)
- Added architecture diagram to docs page
- Linked author name to GitHub profile in docs footer

## [0.2.0] - 2026-02-12

### Added

- **`workouts schedule` subcommands** — Restructured `workouts schedule` from a single command into a command group with three subcommands:
  - `workouts schedule add <id> <date>` — Schedule a workout on a calendar date (replaces `workouts schedule <id> <date>`)
  - `workouts schedule list <date>` — List scheduled workouts for a date, with table/JSON/plain output
  - `workouts schedule remove <id> [-f]` — Remove a scheduled workout from the calendar with confirmation prompt
- **Calendar API** — New `GetCalendarWeek` API method to fetch calendar data from Garmin Connect
- **Unschedule API** — New `UnscheduleWorkout` API method to delete scheduled workouts

### Changed

- `workouts schedule` is now a command group instead of a leaf command; use `workouts schedule add` for the previous scheduling behavior

## [0.1.0] - 2026-02-12

Initial release of gccli — a fast, script-friendly CLI for Garmin Connect.

### Added

- **Authentication** — Browser SSO and headless email/password login with MFA support, OAuth1-to-OAuth2 token exchange
- **Secure credential storage** — OS keyring (macOS Keychain, Linux Secret Service, encrypted file fallback), tokens keyed by `gccli:token:<email>`
- **Activities** — List, search, count, view details/splits/weather/HR zones/power zones/exercise sets/gear, download (FIT/GPX/TCX/KML/CSV), upload, create manual entries, rename, retype, delete
- **Health data** — Daily summaries, steps (daily/weekly), heart rate, resting HR, floors, sleep, stress (daily/weekly), respiration, SpO2, HRV, body battery, training readiness/status, fitness age, VO2max/max metrics, lactate threshold, cycling FTP, race predictions, endurance score, hill score, intensity minutes, wellness events
- **Body composition** — Weight tracking, body fat, muscle mass, weigh-in history, blood pressure, FIT file encoding for composition uploads
- **Workouts** — List, view details, download as FIT, upload from JSON, create with pace/HR/power/cadence targets, schedule, delete
- **Devices** — List registered devices, view settings, solar data, alarms, primary/last-used device
- **Gear** — List gear, usage stats, linked activities, defaults per activity type, link/unlink to activities
- **Goals & badges** — Active goals, earned/available/in-progress badges, challenges, personal records
- **Training plans** — Browse and view training plans
- **Profile** — User profile and settings
- **Hydration** — View and log daily water intake
- **Wellness** — Menstrual cycle data, pregnancy summary
- **Output formats** — JSON (`--json`), plain/TSV (`--plain`), and human-friendly tables (default)
- **Multiple accounts** — `--account` flag and `GCCLI_ACCOUNT` env var
- **Resilient networking** — Automatic retry on 429/5xx with exponential backoff, circuit breaker for fault tolerance, auto-refresh OAuth2 tokens on 401
- **CI pipeline** — GitHub Actions for fmt-check, lint, and test
- **Cross-platform builds** — macOS (amd64/arm64) and Linux (amd64/arm64) via goreleaser

[1.0.0]: https://github.com/bpauli/gccli/releases/tag/v1.0.0
[0.3.0]: https://github.com/bpauli/gccli/releases/tag/v0.3.0
[0.2.0]: https://github.com/bpauli/gccli/releases/tag/v0.2.0
[0.1.0]: https://github.com/bpauli/gccli/releases/tag/v0.1.0
