package lunchmoney

import (
	"context"
	"fmt"
	"strconv"
)

// Asset is a manually-tracked or synced LunchMoney asset (e.g. a checking or
// savings account), as returned by GET /v1/assets.
//
// TypeName distinguishes cash-like accounts ("checking", "savings") from
// other asset types. Balance is parsed from the API's string representation
// into a float64; ToBase is already numeric in the API response and is
// passed straight through.
type Asset struct {
	ID       int
	Name     string
	TypeName string
	Balance  float64
	ToBase   float64
	Currency string
}

// assetWire mirrors the wire format of a single element of the "assets"
// array in the GET /v1/assets response, where Balance is a JSON string.
type assetWire struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	TypeName string  `json:"type_name"`
	Balance  string  `json:"balance"`
	ToBase   float64 `json:"to_base"`
	Currency string  `json:"currency"`
}

// assetsResponse mirrors the top-level object returned by GET /v1/assets,
// which wraps the array under an "assets" key.
type assetsResponse struct {
	Assets []assetWire `json:"assets"`
}

// GetAssets fetches all manually-tracked/synced assets from GET /v1/assets.
func (c *Client) GetAssets(ctx context.Context) ([]Asset, error) {
	var resp assetsResponse
	if err := c.get(ctx, "/v1/assets", "", &resp); err != nil {
		return nil, err
	}

	assets := make([]Asset, 0, len(resp.Assets))
	for _, w := range resp.Assets {
		balance, err := strconv.ParseFloat(w.Balance, 64)
		if err != nil {
			return nil, fmt.Errorf("lunchmoney: parsing balance %q for asset %d (%s): %w", w.Balance, w.ID, w.Name, err)
		}
		assets = append(assets, Asset{
			ID:       w.ID,
			Name:     w.Name,
			TypeName: w.TypeName,
			Balance:  balance,
			ToBase:   w.ToBase,
			Currency: w.Currency,
		})
	}
	return assets, nil
}
