package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventOperations(t *testing.T) {
	// These are unit tests that don't require a database connection
	// They test the struct and interface definitions

	t.Run("Event struct creation", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(-1 * time.Hour)
		endTime := now
		level := 7
		note := "Test event note"
		color := "#FF6B6B"
		eventTypeID := "event-type-123"

		event := &Event{
			ID:          "test-id",
			UserID:      "user-123",
			Name:        "Headache",
			EventTypeID: &eventTypeID,
			StartedAt:   startTime,
			EndedAt:     &endTime,
			Level:       &level,
			Note:        &note,
			Color:       &color,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		assert.Equal(t, "test-id", event.ID)
		assert.Equal(t, "user-123", event.UserID)
		assert.Equal(t, "Headache", event.Name)
		assert.Equal(t, "event-type-123", *event.EventTypeID)
		assert.Equal(t, startTime, event.StartedAt)
		assert.Equal(t, endTime, *event.EndedAt)
		assert.Equal(t, 7, *event.Level)
		assert.Equal(t, "Test event note", *event.Note)
		assert.Equal(t, "#FF6B6B", *event.Color)
	})

	t.Run("EventLink struct creation", func(t *testing.T) {
		now := time.Now()
		consumptionID := "consumption-123"

		link := &EventLink{
			ID:            "link-id",
			EventID:       "event-123",
			ConsumptionID: &consumptionID,
			CreatedAt:     now,
		}

		assert.Equal(t, "link-id", link.ID)
		assert.Equal(t, "event-123", link.EventID)
		assert.Equal(t, "consumption-123", *link.ConsumptionID)
		assert.Nil(t, link.ConsumptionItemID)
		assert.Equal(t, now, link.CreatedAt)
	})

	t.Run("EventListOptions creation", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		eventTypeID := "activity-type-123"
		levelMin := 3
		levelMax := 8

		options := EventListOptions{
			Limit:       100,
			Offset:      0,
			StartDate:   &startDate,
			EndDate:     &endDate,
			EventTypeID: &eventTypeID,
			LevelMin:    &levelMin,
			LevelMax:    &levelMax,
			Labels:      []string{"exercise", "cardio"},
			MatchAll:    true,
		}

		assert.Equal(t, 100, options.Limit)
		assert.Equal(t, 0, options.Offset)
		assert.Equal(t, startDate, *options.StartDate)
		assert.Equal(t, endDate, *options.EndDate)
		assert.Equal(t, "activity-type-123", *options.EventTypeID)
		assert.Equal(t, 3, *options.LevelMin)
		assert.Equal(t, 8, *options.LevelMax)
		assert.Len(t, options.Labels, 2)
		assert.Contains(t, options.Labels, "exercise")
		assert.Contains(t, options.Labels, "cardio")
		assert.True(t, options.MatchAll)
	})
}

// MockEventStore is a simple mock implementation for testing
type MockEventStore struct {
	events []*Event
	links  []*EventLink
}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{
		events: make([]*Event, 0),
		links:  make([]*EventLink, 0),
	}
}

func (m *MockEventStore) CreateEvent(ctx context.Context, event *Event) error {
	event.ID = "mock-event-id"
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventStore) GetEvent(ctx context.Context, userID, id string) (*Event, error) {
	for _, event := range m.events {
		if event.ID == id && event.UserID == userID {
			return event, nil
		}
	}
	return nil, nil
}

func (m *MockEventStore) ListEvents(ctx context.Context, userID string, options EventListOptions) ([]*Event, error) {
	var result []*Event
	for _, event := range m.events {
		if event.UserID == userID {
			result = append(result, event)
		}
	}
	return result, nil
}

func TestMockEventStore(t *testing.T) {
	t.Run("Mock event store operations", func(t *testing.T) {
		store := NewMockEventStore()
		ctx := context.Background()

		// Create an event
		event := &Event{
			UserID:    "user-123",
			Name:      "Test Event",
			StartedAt: time.Now(),
		}

		err := store.CreateEvent(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, "mock-event-id", event.ID)
		assert.NotZero(t, event.CreatedAt)

		// Get the event
		retrieved, err := store.GetEvent(ctx, "user-123", "mock-event-id")
		require.NoError(t, err)
		assert.Equal(t, "Test Event", retrieved.Name)

		// List events
		events, err := store.ListEvents(ctx, "user-123", EventListOptions{})
		require.NoError(t, err)
		assert.Len(t, events, 1)
	})
}
