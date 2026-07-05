# Money Tracker — Implementation Plan (GGA-first)

Derived from [Cash Flow Forecast Plan](<Cash Flow Forecast Plan.md>) and [Green/Green Auto-UAT](<Green Green Auto-UAT.md>). Ordering principle: CI must exist before the first negative test lands, so the green/green invariant is enforceable from PR #1.

## Phase 0 — Walking skeleton + CI (one PR, before any behavior)

The point of this phase is that every later PR lands against a CI gate that's already running the full suite.

- `go mod init`, `cmd/moneytracker/main.go` with a urfave/cli v3 scaffold that registers **no real commands yet**. This matters for GGA: `moneytracker forecast` must exit non-zero today (unknown command), which is exactly what the negative assertions will claim.
- `uat/go.mod` as its own module (no `internal/*` imports possible by construction), plus a `go.work` at the root so editor tooling spans both modules.
- `.github/workflows/ci.yml` triggered on `pull_request` and push to `main`, three jobs:
  - **unit** — root module: `go test -race -shuffle=on ./...`, `go vet`, and a `go mod tidy && git diff --exit-code` step.
  - **uat** — `cd uat && go test -race -count=1 ./...` (`-count=1` because e2e results must never come from the test cache).
  - **lint** — golangci-lint over both modules.
- No secrets anywhere: the twin runs in-process, so UAT in CI needs no credentials or network egress. Pin action versions, least-privilege `permissions:` block, and (in repo settings) require all three checks on PRs.

Commits: `chore: bootstrap Go module and CLI skeleton`, `ci: add PR workflow (unit, uat, lint)`.

## Phase 1 — GGA: digital twin + harness + green negative tests

- **Twin (`uat/twin/`)**: a fake LunchMoney HTTP server implementing the four consumed endpoints (`/v1/me`, `/v1/assets`, `/v1/plaid_accounts`, `/v1/recurring_items`), seeded per-test from scenario fixtures shaped like real API JSON. Three scenarios: healthy (OK), floor breach (NOT OK), invalid token (401).
- **Harness (`uat/*_test.go`)**: `TestMain` builds the real binary once into a temp dir (`go build` against the root module), then each test execs it as a subprocess with `LUNCHMONEY_TOKEN` and `LUNCHMONEY_API_URL` pointed at its twin instance. Asserts only on exit codes and stdout/stderr. Because the harness owns the build, the CI job stays a plain `go test` — local and CI runs are identical.
- **Negative assertions, all green today**: forecast against the healthy twin does NOT exit 0; stdout does NOT contain a `Balance` column or a `Cash position:` line; the breach scenario does NOT exit 1; `check` with a bad token does NOT report token failure; etc. One test per requirement from the plan doc's verdict/output/exit-code spec.
- **Exit code contract** (decided now so assertions are precise): `0` = OK, `1` = NOT OK, `2` = operational error (bad token, network). This lets the breach test assert exit 1 specifically rather than just "non-zero".

Commit: `test(uat): add LunchMoney twin and green negative acceptance tests`.

## Phases 2–5 — Build inward-out, flipping assertions as behaviors land

Each phase is red/green TDD with conventional commits; each behavior that ships flips its UAT negative assertion **in the same PR**, so the suite never has a pending state and CI stays green at every merge.

2. **`internal/forecast`** (pure core): period bucketing from anchor date → running-balance projection → `Verdict`. Table-driven tests: empty events, breach exactly at floor, events on period boundaries, income-only, expense-only.
3. **`internal/lunchmoney`**: thin client with `httptest` + JSON fixtures; `debit_as_negative` sign normalization; base URL injectable (the twin depends on this).
4. **`internal/config`**: env token + API URL, optional TOML for floor/anchor/horizon/account excludes.
5. **`internal/report` + `cmd/moneytracker`**: wire `forecast` (default) and `check`; render the period table; implement the exit-code contract. This is where most negative assertions flip to positive.

## Phase 6 — Real-account smoke test (local only, never CI)

Run with a real `LUNCHMONEY_TOKEN`, cross-check starting position and one period against the LunchMoney UI. Non-intrusive/read-only per the GGA "natural observation" classification.

## Execution notes

- Each phase's coding runs in a subagent (sonnet for the twin/engine/client, haiku for scaffolding and workflow YAML); the main session orchestrates, reviews, and verifies against the UAT suite.
- Optional CI hardening once the core exists: `govulncheck` job, Dependabot with grouped minor/patch updates. Not needed for PR #1.
