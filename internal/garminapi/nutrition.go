package garminapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetNutritionDailyFoodLog returns nutrition food log data for a given date.
// Date format: YYYY-MM-DD.
func (c *Client) GetNutritionDailyFoodLog(ctx context.Context, date string) (json.RawMessage, error) {
	path := fmt.Sprintf("/nutrition-service/food/logs/%s", date)
	return c.ConnectAPI(ctx, http.MethodGet, path, nil)
}

// GetNutritionDailyMeals returns nutrition meal summary data for a given date.
// Date format: YYYY-MM-DD.
func (c *Client) GetNutritionDailyMeals(ctx context.Context, date string) (json.RawMessage, error) {
	path := fmt.Sprintf("/nutrition-service/meals/%s", date)
	return c.ConnectAPI(ctx, http.MethodGet, path, nil)
}

// GetNutritionDailySettings returns nutrition settings for a given date.
// Date format: YYYY-MM-DD.
func (c *Client) GetNutritionDailySettings(ctx context.Context, date string) (json.RawMessage, error) {
	path := fmt.Sprintf("/nutrition-service/settings/%s", date)
	return c.ConnectAPI(ctx, http.MethodGet, path, nil)
}
