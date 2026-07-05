# Money Tracker — Cash Flow Forecast CLI

## Context

Build a system that answers **"Are we OK with our current spending?"** using real data from the LunchMoney API. It fetches current account balances and recurring transactions, projects cash flow forward in bi-weekly periods over a 6-month horizon, and reports a verdict: OK if the projected cash balance never drops below a configurable floor.

Decisions made with the user:
- **Form:** CLI tool, run on demand (`moneytracker forecast`)
- **Stack:** Go
- **State:** Stateless — every run fetches fresh from LunchMoney; no local DB
- **Horizon / verdict:** 6 months (~13 bi-weekly periods); "OK" = projected balance ≥ floor in every period

## LunchMoney API surface (verified against lunchmoney.dev)

Base `https://dev.lunchmoney.app`, auth via `Authorization: Bearer $LUNCHMONEY_TOKEN`.

| Endpoint | Use |
|---|---|
| `GET /v1/me` | Token validation (`moneytracker check` / startup) |
| `GET /v1/assets` | Manually-managed account balances (`balance`, `type_name`, `to_base`) |
| `GET /v1/plaid_accounts` | Synced account balances (`balance`, `type`, `to_base`) |
| `GET /v1/recurring_items?start_date&end_date` | Recurring items **with `occurrences`** — LunchMoney expands each item into dated occurrences within the range (`amount`, `is_income`, `granularity`, `quantity`, `to_base`) |

Key simplification: because `recurring_items` returns dated `occurrences`, the forecast engine never implements cadence math. It consumes a flat list of dated cash events.

## System shape

```
lunchmoney API ──► client ──► domain types ──► forecast engine ──► report ──► terminal
                  (I/O)      (Accounts,       (pure functions)    (render)
                              CashEvents)
```

```
money-tracker/
├── cmd/moneytracker/main.go      # wiring only
├── internal/config/              # token (env LUNCHMONEY_TOKEN), API base URL
│                                 #   (env LUNCHMONEY_API_URL, for the digital twin),
│                                 #   floor, horizon, account include/exclude, anchor date
├── internal/lunchmoney/          # thin HTTP client: GetMe, GetAssets,
│                                 #   GetPlaidAccounts, GetRecurringItems
├── internal/forecast/            # PURE domain logic (the heart, fully TDD'd):
│                                 #   - net liquid position from accounts
│                                 #   - occurrences → []CashEvent
│                                 #   - bi-weekly bucketing from anchor date
│                                 #   - running balance projection
│                                 #   - Verdict{OK bool, MinBalance, BreachPeriod}
├── internal/report/              # terminal rendering of verdict + period table
└── uat/                          # Green/Green Auto-UAT — separate Go module,
    ├── go.mod                    #   black-box: no imports of internal/*
    ├── twin/                     #   LunchMoney digital twin (fake HTTP server)
    └── *_test.go                 #   e2e tests driving the built binary
```

Principles:
- **Functional core, imperative shell.** `internal/forecast` takes plain structs in, returns plain structs out — no HTTP, no clock reads (current time passed in). This is where red/green TDD lives.
- **Client is a thin translator** from API JSON to domain types, tested with `httptest` + JSON fixtures.
- Standard library only where possible; `urfave/cli` (v3) for the CLI command structure.

## Domain model

- **Starting position** = Σ cash-like balances (assets `type_name: cash`, plaid `checking`/`savings`) − Σ credit card balances. Account selection overridable in config; use `to_base` amounts so multi-currency collapses to primary currency.
- **CashEvent** `{Date, Amount, Payee, IsIncome}` — one per recurring-item occurrence in the horizon.
- **Period** = bi-weekly bucket anchored at a configurable `anchor_date` (default: today; user can set to payday so periods align with paychecks). Each period gets `Income`, `Expenses`, `Net`, `ProjectedEndBalance`.
- **Verdict**: OK iff every period's projected end balance ≥ `floor` (config, default 0). Report also shows the minimum point and, if breached, which period fails.

## Green/Green Auto-UAT (per docs/Green Green Auto-UAT.md)

A separate workload that exercises the SUT externally — black-box, no knowledge of internals — so it can't cheat or couple to one implementation.

- **Harness (`uat/`)**: its own Go module that never imports `internal/*`. Tests build/exec the `moneytracker` binary as a subprocess and assert only on observable behavior: exit codes, stdout/stderr.
- **LunchMoney digital twin (`uat/twin/`)**: a fake HTTP server implementing the four endpoints we consume (`/v1/me`, `/v1/assets`, `/v1/plaid_accounts`, `/v1/recurring_items`), seeded per-test with scenario fixtures shaped like real API responses. Scenarios: healthy cash flow (OK), floor breach (NOT OK), invalid token (401). The harness owns starting the twin and pointing the SUT at it via `LUNCHMONEY_API_URL`.
- **Green/green workflow**: no pending/skipped tests, ever. Requirements are codified up front as **negative assertions** that pass today — e.g., "`moneytracker forecast` against the healthy twin does NOT exit 0", "stdout does NOT contain a `Balance` column", "the NOT-OK scenario does NOT exit 1". As each behavior is implemented the negative test starts failing; reverse the assertion in the same change, embedding the behavior permanently.
- **CI**: every PR runs the full suite — unit tests plus the UAT suite (twin-backed, so no real credentials needed in CI).

## Implementation order (each step: red/green TDD, Conventional Commit)

1. `git init`, `go mod init`, project skeleton.
2. **GGA first**: build the LunchMoney digital twin and the `uat/` harness; codify the requirements above as green negative tests (forecast output shape, OK/NOT-OK exit codes, token validation).
3. `internal/forecast`: bucketing — given anchor + horizon, produce period boundaries. Then projection — events + starting balance → per-period nets and running balance. Then verdict.
4. `internal/lunchmoney`: client with `httptest` fixtures for assets, plaid_accounts, recurring_items (including `occurrences` parsing — note the API's `debit_as_negative` param to normalize signs). Base URL configurable so the twin can stand in.
5. `internal/config`: env token + API base URL + optional `~/.config/moneytracker/config.toml` (floor, anchor, horizon, account excludes).
6. `internal/report` + `cmd/moneytracker`: wire it together; `forecast` (default) and `check` commands; non-zero exit code when NOT OK (cron/scripting friendly). As each behavior lands, reverse the corresponding UAT negative assertion.
7. Run against the real LunchMoney account to validate.

## Sample output (target)

```
$ moneytracker forecast
Cash position: $8,420  (3 accounts, less $1,230 credit)

Period          Income   Expenses      Net   Balance
Jul 05–Jul 18   +3,100     -2,650     +450     8,870
Jul 19–Aug 01   +3,100     -3,890     -790     8,080
...
Jan 03–Jan 16   +3,100     -2,650     +450    11,240

✔ OK — balance stays above $2,000 floor (min: $7,610 on Aug 29)
```

## Verification

- **Auto-UAT (primary)**: the `uat/` suite is always green — run it after every step. Negative assertions before a behavior exists, positive after; the twin-backed scenarios (OK / floor breach / bad token) prove exit codes and output externally, with no access to internals.
- `go test ./...` green throughout (forecast engine has table-driven tests covering: empty events, breach exactly at floor, events on period boundaries, income-only, expense-only).
- Final smoke test (non-intrusive, read-only): run `moneytracker forecast` with a real `LUNCHMONEY_TOKEN`; cross-check the starting position against the LunchMoney web UI and spot-check one period's events against the recurring items page.

## Future growth (not in scope now)

The stateless CLI core makes both later forms cheap: a scheduled report is `cron + moneytracker forecast | ntfy`; a dashboard wraps `internal/forecast` output in a web view. Local SQLite history (forecast-vs-actual) can be added behind `internal/forecast` without touching it.
