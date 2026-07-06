package lunchmoney

import (
	"context"
	"fmt"
	"strconv"
)

// PlaidAccount is a bank/credit account synced via Plaid, as returned by
// GET /v1/plaid_accounts.
//
// Type distinguishes account kinds (e.g. "credit", "depository"). Balance is
// parsed from the API's string representation into a float64; ToBase is
// already numeric in the API response and is passed straight through.
type PlaidAccount struct {
	ID       int
	Name     string
	Type     string
	Balance  float64
	ToBase   float64
	Currency string
}

// plaidAccountWire mirrors the wire format of a single element of the
// "plaid_accounts" array in the GET /v1/plaid_accounts response, where
// Balance is a JSON string.
type plaidAccountWire struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Balance  string  `json:"balance"`
	ToBase   float64 `json:"to_base"`
	Currency string  `json:"currency"`
}

// plaidAccountsResponse mirrors the top-level object returned by
// GET /v1/plaid_accounts, which wraps the array under a "plaid_accounts" key.
type plaidAccountsResponse struct {
	PlaidAccounts []plaidAccountWire `json:"plaid_accounts"`
}

// GetPlaidAccounts fetches all Plaid-synced accounts from
// GET /v1/plaid_accounts.
func (c *Client) GetPlaidAccounts(ctx context.Context) ([]PlaidAccount, error) {
	var resp plaidAccountsResponse
	if err := c.get("/v1/plaid_accounts", "", &resp); err != nil {
		return nil, err
	}

	accounts := make([]PlaidAccount, 0, len(resp.PlaidAccounts))
	for _, w := range resp.PlaidAccounts {
		balance, err := strconv.ParseFloat(w.Balance, 64)
		if err != nil {
			return nil, fmt.Errorf("lunchmoney: parsing balance %q for plaid account %d (%s): %w", w.Balance, w.ID, w.Name, err)
		}
		accounts = append(accounts, PlaidAccount{
			ID:       w.ID,
			Name:     w.Name,
			Type:     w.Type,
			Balance:  balance,
			ToBase:   w.ToBase,
			Currency: w.Currency,
		})
	}
	return accounts, nil
}
