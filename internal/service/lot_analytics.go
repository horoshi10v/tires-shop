package service

import (
	"context"
	"log/slog"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

func (s *lotService) TrackLotAnalyticsEvent(ctx context.Context, req domain.TrackLotAnalyticsEventRequest, userAgent string) error {
	s.logger.Debug(
		"tracking lot analytics event",
		slog.String("lot_id", req.LotID.String()),
		slog.String("event_type", string(req.EventType)),
		slog.String("source", string(req.Source)),
	)

	return s.repo.TrackAnalyticsEvent(ctx, req, userAgent)
}
