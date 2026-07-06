package lunchmoney

import "context"

// Me describes the authenticated LunchMoney user and budget, as returned by
// GET /v1/me.
type Me struct {
	UserID     int    `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	AccountID  int    `json:"account_id"`
	BudgetName string `json:"budget_name"`
}

// GetMe fetches the authenticated user's profile from GET /v1/me.
func (c *Client) GetMe(ctx context.Context) (*Me, error) {
	var me Me
	if err := c.get("/v1/me", "", &me); err != nil {
		return nil, err
	}
	return &me, nil
}
