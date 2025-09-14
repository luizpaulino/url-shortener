package urls

import "time"

type URLItem struct {
	ShortCode string    `json:"shortCode"`
	LongURL   string    `json:"longURL"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"createdAt"`
}
