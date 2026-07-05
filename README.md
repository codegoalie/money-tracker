# Money Tracker

A stateless Go CLI that answers one question: **"Are we OK with our current spending?"**

It pulls current account balances and recurring transactions from the [LunchMoney API](https://lunchmoney.dev), projects cash flow forward in bi-weekly periods over a 6-month horizon, and reports a verdict: **OK** if the projected balance never drops below a configurable floor.

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

The exit code is non-zero when the verdict is NOT OK, so it plays well with cron and scripting (e.g. `moneytracker forecast || ntfy ...`).

## Commands

| Command | Purpose |
|---|---|
| `moneytracker forecast` | Fetch balances + recurring items, project bi-weekly cash flow, print verdict (default) |
| `moneytracker check` | Validate the LunchMoney API token |

## Configuration

- `LUNCHMONEY_TOKEN` (env, required) — LunchMoney API token
- `LUNCHMONEY_API_URL` (env, optional) — API base URL override, used by tests to point at a fake server
- `~/.config/moneytracker/config.toml` (optional) — floor, anchor date, horizon, account include/exclude

The **floor** defaults to 0. The **anchor date** defaults to today; set it to a payday so periods align with paychecks.

## How it works

```
lunchmoney API ──► client ──► domain types ──► forecast engine ──► report ──► terminal
                  (I/O)      (Accounts,       (pure functions)    (render)
                              CashEvents)
```

- **Stateless.** Every run fetches fresh from LunchMoney; no local database.
- **Functional core, imperative shell.** `internal/forecast` is pure domain logic — plain structs in, plain structs out, no HTTP, no clock reads. This is where red/green TDD lives.
- **Thin client.** `internal/lunchmoney` translates API JSON to domain types. Because `GET /v1/recurring_items` returns dated `occurrences`, the forecast engine never implements cadence math — it consumes a flat list of dated cash events.
- **Starting position** = Σ cash-like balances (cash assets, checking/savings) − Σ credit card balances, using `to_base` amounts so multi-currency collapses to the primary currency.

### Layout

```
money-tracker/
├── cmd/moneytracker/       # wiring only
├── internal/config/        # token, API base URL, floor, horizon, anchor date
├── internal/lunchmoney/    # thin HTTP client (GetMe, GetAssets, GetPlaidAccounts, GetRecurringItems)
├── internal/forecast/      # pure domain logic: bucketing, projection, verdict
├── internal/report/        # terminal rendering
└── uat/                    # black-box e2e suite (separate Go module)
    └── twin/               # LunchMoney digital twin (fake HTTP server)
```

## Testing

Two layers, both required to be green on every PR:

1. **Unit tests** (`go test ./...`) — table-driven tests over the pure forecast engine; `httptest` + JSON fixtures for the client.
2. **Green/Green Auto-UAT** (`uat/`) — a separate Go module that never imports `internal/*`. It builds and runs the `moneytracker` binary as a subprocess against a LunchMoney **digital twin** (a fake HTTP server seeded with scenario fixtures: healthy cash flow, floor breach, invalid token) and asserts only on observable behavior — exit codes and output.

   The suite is always green: requirements for unbuilt behavior are codified as **negative assertions** (e.g. "forecast does NOT yet exit 0 against the healthy twin"). When the behavior lands, the negative test starts failing and the assertion is reversed in the same change — embedding the behavior permanently, with no pending or skipped tests. See [docs/Green Green Auto-UAT.md](docs/Green%20Green%20Auto-UAT.md).

The twin means CI needs no real credentials. The full plan lives in [docs/Cash Flow Forecast Plan.md](docs/Cash%20Flow%20Forecast%20Plan.md).
