package service

import (
	"context"
	"fmt"
	"sweng-task/internal/model"
	"time"

	"go.uber.org/zap"
)

// TrackingService provides operations for tracking
type TrackingService struct {
	inputTrackingEvents       chan model.TrackingEvent
	eventsStorage             TrackingEventsStorage
	eventsStorageWriteTimeout time.Duration

	log *zap.SugaredLogger
}

// TrackingEventsStorage persists tracking events
type TrackingEventsStorage interface {
	Write(context.Context, []model.TrackingEvent) error
}

type TrackingEventsStorageFunc func(context.Context, []model.TrackingEvent) error

func (f TrackingEventsStorageFunc) Write(ctx context.Context, events []model.TrackingEvent) error {
	return f(ctx, events)
}

// NewTrackingService creates a new TrackingService
func NewTrackingService(eventsBufferSize int, trackingEventsStorage TrackingEventsStorage, trackingEventsWriteTimeout time.Duration, log *zap.SugaredLogger) *TrackingService {
	return &TrackingService{
		inputTrackingEvents:       make(chan model.TrackingEvent, eventsBufferSize),
		eventsStorage:             trackingEventsStorage,
		eventsStorageWriteTimeout: trackingEventsWriteTimeout,

		log: log,
	}
}

// RecordAdInteraction records ad interactions.
// Unblocking operation.
func (s *TrackingService) RecordAdInteraction(t model.TrackingEvent) (bool, error) {
	// simple implementation, there are several ways to improvement
	// one of which is to add a timeout to wait
	//
	// TODO: stop accepting events if graceful shutdown signal is received
	select {
	case s.inputTrackingEvents <- t:
		return true, nil
	default:
		return false, nil
	}
}

// TrackingEventsWorker represents the main loop of the worker.
// Buffer will be flushed into the storage in two cases:
// 1. buffer is reached max chunk size 'maxChunkSize'
// 2. events shouldn't stay in the buffer longer than 'flushEvery' duration
func (s *TrackingService) TrackingEventsWorker(ctx context.Context, maxChunkSize int, flushEvery time.Duration) error {
	var isBufferFlushNeeded bool
	buffer := make([]model.TrackingEvent, 0, maxChunkSize)
	ticker := time.NewTicker(flushEvery)

	for {
		switch len(buffer) {
		case 0:
			// no events in the buffer
			select {
			case event := <-s.inputTrackingEvents:
				buffer = append(buffer, event)
			default:
				// check graceful shutdown only if no events in the chan

				select {
				case event := <-s.inputTrackingEvents:
					buffer = append(buffer, event)

				case <-ctx.Done():
					// graceful shutdown
					return nil
				}
			}
			ticker.Reset(flushEvery)

		default:
			// some events present in the buffer
			select {
			case event := <-s.inputTrackingEvents:
				buffer = append(buffer, event)

			case <-ticker.C:
				isBufferFlushNeeded = true

			case <-ctx.Done():
				// graceful shutdown
				isBufferFlushNeeded = true
			}
		}

		if len(buffer) >= maxChunkSize {
			isBufferFlushNeeded = true
		}
		if isBufferFlushNeeded {
			err := s.flushTrackingEventsBuffer(buffer)
			if err != nil {
				// TODO: repeat N times in case of recovery

				s.log.Errorw("Cannot flush events buffer",
					"error", err,
				)
				// TODO: since we cannot flush any events, we cannot accept any new events
				// maybe we should just shut down completely in this case
				return fmt.Errorf("flush buffer: %w", err)
			}

			// TODO: potential optimization by using the sync.Pool
			buffer = make([]model.TrackingEvent, 0, maxChunkSize)
			isBufferFlushNeeded = false
			ticker.Stop() // TODO: additional checks needed, maybe it is safe do not stop ticker each time
		}
	}
}

// flushTrackingEventsBuffer flushes tracking events to the external storage
func (s *TrackingService) flushTrackingEventsBuffer(events []model.TrackingEvent) error {
	ctx, stop := context.WithTimeout(context.Background(), s.eventsStorageWriteTimeout)
	defer stop()

	// TODO: push buffer to the external message queue
	// as example, to Kafka
	return s.eventsStorage.Write(ctx, events)
}
