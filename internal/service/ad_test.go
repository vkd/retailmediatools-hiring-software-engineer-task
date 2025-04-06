package service

import (
	"sweng-task/internal/model"
	"testing"

	"go.uber.org/zap"
)

func TestAdService_GetWinningAds_Sorting(t *testing.T) {
	placement := "header"
	category := "toys"
	keyword := "summer"

	lineItemsService := NewLineItemService(zap.NewNop().Sugar())
	adService := NewAdService(lineItemsService, zap.NewNop().Sugar())

	_, err := lineItemsService.Create(model.LineItemCreate{
		Name:         "test_1",
		AdvertiserID: "ad_1",
		Bid:          2,
		Budget:       1000,
		Placement:    placement,
		Categories:   []string{category},
		Keywords:     []string{keyword},
	})
	if err != nil {
		t.Errorf("Create line item: %v", err)
	}
	_, err = lineItemsService.Create(model.LineItemCreate{
		Name:         "test_2",
		AdvertiserID: "ad_2",
		Bid:          1,
		Budget:       1000,
		Placement:    placement,
		Categories:   []string{category},
		Keywords:     []string{keyword},
	})
	if err != nil {
		t.Errorf("Create line item: %v", err)
	}
	_, err = lineItemsService.Create(model.LineItemCreate{
		Name:         "test_3",
		AdvertiserID: "ad_3",
		Bid:          3,
		Budget:       1000,
		Placement:    placement,
		Categories:   []string{category},
		Keywords:     []string{keyword},
	})
	if err != nil {
		t.Errorf("Create line item: %v", err)
	}

	ads, err := adService.GetWinningAds(placement, category, keyword, 2)
	if err != nil {
		t.Errorf("Create line item: %v", err)
	}

	if len(ads) != 2 {
		t.Errorf("Wrong amount of winning ads: %d != 2", len(ads))
	}

	if ads[0].Name != "test_3" {
		t.Errorf("Wrong first winning ad: %v.Name != 'test_3'", ads[0])
	}
	if ads[1].Name != "test_1" {
		t.Errorf("Wrong second winning ad: %v.Name != 'test_1'", ads[1])
	}
}
