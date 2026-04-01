package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
)

func (r *LotRepo) TrackAnalyticsEvent(ctx context.Context, req domain.TrackLotAnalyticsEventRequest, userAgent string) error {
	event := models.LotAnalyticsEvent{
		LotID:     req.LotID,
		EventType: string(req.EventType),
		Source:    string(req.Source),
		SessionID: strings.TrimSpace(req.SessionID),
		UserAgent: strings.TrimSpace(userAgent),
	}

	if err := r.db.WithContext(ctx).Create(&event).Error; err != nil {
		return fmt.Errorf("failed to store lot analytics event: %w", err)
	}

	return nil
}
