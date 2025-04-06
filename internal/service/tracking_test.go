package service

import (
	"context"
	"sweng-task/internal/model"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestTrackingService_TrackingEventsWorker(t *testing.T) {
	ctx, stop := context.WithCancel(t.Context())
	defer stop()

	pause := make(chan struct{})
	wait := make(chan struct{})
	bufferSize := 5
	chunkSize := 2
	trackingEventStorageWrite := make(chan struct{}, bufferSize+chunkSize)

	var eventsPersisted int

	tService := NewTrackingService(bufferSize, TrackingEventsStorageFunc(func(ctx context.Context, events []model.TrackingEvent) error {
		trackingEventStorageWrite <- struct{}{}

		// wait to be unpaused
		<-pause

		eventsPersisted += len(events)

		return nil
	}), time.Second, zap.NewNop().Sugar())

	// start worker
	go func() {
		err := tService.TrackingEventsWorker(ctx, chunkSize, time.Second)
		if err != nil {
			t.Errorf("Tracking events worker stopped with an error: %v", err)
		}
		close(wait)
	}()

	// write events
	for range bufferSize {
		ok, err := tService.RecordAdInteraction(model.TrackingEvent{})
		if !ok || err != nil {
			t.Errorf("Ad interaction doesn't not recorded: %v, %v", ok, err)
		}
	}
	<-trackingEventStorageWrite
	for range chunkSize {
		ok, err := tService.RecordAdInteraction(model.TrackingEvent{})
		if !ok || err != nil {
			t.Errorf("Ad interaction doesn't not recorded: %v, %v", ok, err)
		}
	}

	// buffer is full
	ok, err := tService.RecordAdInteraction(model.TrackingEvent{})
	if err != nil {
		t.Errorf("Ad interaction doesn't not recorded: %v, %v", ok, err)
	}
	if ok {
		t.Errorf("Buffer must be full")
	}

	close(pause) // unpause storage writer
	stop()
	<-wait

	if eventsPersisted != bufferSize+chunkSize {
		t.Errorf("Wrong number of events: %d (persisted) != %d (sent)", eventsPersisted, bufferSize+chunkSize)
	}
}
