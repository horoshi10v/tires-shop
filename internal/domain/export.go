package domain

import "context"

// ExportService coordinates data fetching and Google Sheets generation.
type ExportService interface {
	ExportInventory(ctx context.Context) (string, error) // Returns URL of the Google Sheet
	ExportPnL(ctx context.Context) (string, error)       // Returns URL of the Google Sheet
}
