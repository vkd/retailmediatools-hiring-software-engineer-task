package service

import (
	"fmt"
	"sort"
	"sweng-task/internal/model"

	"go.uber.org/zap"
)

// AdService provides operations for ads
type AdService struct {
	lineItemsService *LineItemService
	log              *zap.SugaredLogger
}

// NewAdService creates a new AdService
func NewAdService(lineItemsService *LineItemService, log *zap.SugaredLogger) *AdService {
	return &AdService{
		lineItemsService: lineItemsService,
		log:              log,
	}
}

// GetWinningAds returns winning ads
func (s *AdService) GetWinningAds(placement string, category string, keyword string, limit int) ([]model.Ad, error) {
	// for better optimization we can add sorting inside of this method
	items, err := s.lineItemsService.FindMatchingLineItems(placement, category, keyword)
	if err != nil {
		return nil, fmt.Errorf("find matching line items: %w", err)
	}

	// reversed sorting - from biggest bid to lowest
	// it is based on the assumption that we would like to prioritise ads with higher bid
	sort.Slice(items, func(i, j int) bool { return items[i].Bid >= items[j].Bid })

	if len(items) > limit {
		items = items[:limit]
	}

	ads := make([]model.Ad, len(items))
	for i, item := range items {
		ads[i] = model.Ad{
			ID:           item.ID,
			Name:         item.Name,
			AdvertiserID: item.AdvertiserID,
			Bid:          item.Bid,
			Placement:    item.Placement,
			ServeURL:     "", // TODO: add serve URL
		}
	}
	return ads, nil
}
