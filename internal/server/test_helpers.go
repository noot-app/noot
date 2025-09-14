package server

import (
	"context"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// NullStore implements the full storage.Store interface with no-op methods for testing
type NullStore struct{}

// User operations
func (n *NullStore) CreateUser(ctx context.Context, user *storage.User) error { return nil }
func (n *NullStore) GetUser(ctx context.Context, id string) (*storage.User, error) { return nil, nil }
func (n *NullStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) { return nil, nil }
func (n *NullStore) UpdateUser(ctx context.Context, user *storage.User) error { return nil }

// Consumption operations
func (n *NullStore) CreateConsumption(ctx context.Context, consumption *storage.Consumption) (*storage.Consumption, error) { return nil, nil }
func (n *NullStore) GetConsumption(ctx context.Context, id string) (*storage.Consumption, error) { return nil, nil }
func (n *NullStore) GetConsumptionForUser(ctx context.Context, userID, id string) (*storage.Consumption, error) { return nil, nil }
func (n *NullStore) GetPublicConsumption(ctx context.Context, id string) (*storage.Consumption, error) { return nil, nil }
func (n *NullStore) UpdateConsumption(ctx context.Context, consumption *storage.Consumption) error { return nil }
func (n *NullStore) DeleteConsumption(ctx context.Context, id string) error { return nil }
func (n *NullStore) GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*storage.Consumption, error) { return nil, nil }
func (n *NullStore) GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*storage.Consumption, error) { return nil, nil }
func (n *NullStore) GetConsumptionsByUserDateRange(ctx context.Context, userID string, start, end time.Time, limit, offset int) ([]*storage.Consumption, error) { return nil, nil }
func (n *NullStore) GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*storage.NutritionSummary, error) { return nil, nil }

// User favorites operations
func (n *NullStore) CreateFavorite(ctx context.Context, userID, consumptionID string) (*storage.UserFavorite, error) { return nil, nil }
func (n *NullStore) DeleteFavorite(ctx context.Context, userID, consumptionID string) error { return nil }
func (n *NullStore) GetUserFavorites(ctx context.Context, userID string) ([]*storage.UserFavoriteWithConsumption, error) { return nil, nil }
func (n *NullStore) IsFavorited(ctx context.Context, userID, consumptionID string) (bool, error) { return false, nil }

// ConsumptionItem operations
func (n *NullStore) CreateConsumptionItem(ctx context.Context, item *storage.ConsumptionItem) error { return nil }
func (n *NullStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*storage.ConsumptionItem, error) { return nil, nil }
func (n *NullStore) UpdateConsumptionItem(ctx context.Context, item *storage.ConsumptionItem) error { return nil }
func (n *NullStore) DeleteConsumptionItem(ctx context.Context, id string) error { return nil }
func (n *NullStore) DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error { return nil }

// Item operations
func (n *NullStore) CreateItem(ctx context.Context, item *storage.Item) error { return nil }
func (n *NullStore) GetItem(ctx context.Context, id string) (*storage.Item, error) { return nil, nil }
func (n *NullStore) GetItemByName(ctx context.Context, canonicalName, brand string) (*storage.Item, error) { return nil, nil }
func (n *NullStore) UpdateItem(ctx context.Context, item *storage.Item) error { return nil }
func (n *NullStore) GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*storage.Item, error) { return nil, nil }

// User goal operations
func (n *NullStore) UpsertUserGoal(ctx context.Context, goal *storage.UserGoal) error { return nil }
func (n *NullStore) GetUserGoal(ctx context.Context, userID, name string) (*storage.UserGoal, error) { return nil, nil }
func (n *NullStore) GetUserGoalByID(ctx context.Context, userID, goalID string) (*storage.UserGoal, error) { return nil, nil }
func (n *NullStore) GetUserGoals(ctx context.Context, userID string) ([]*storage.UserGoal, error) { return nil, nil }
func (n *NullStore) GetUserGoalIDByName(ctx context.Context, userID, name string) (*string, error) { return nil, nil }
func (n *NullStore) DeleteUserGoal(ctx context.Context, userID, name string) error { return nil }
func (n *NullStore) SetActiveGoal(ctx context.Context, userID, goalID string) error { return nil }
func (n *NullStore) ClearActiveGoal(ctx context.Context, userID string) error { return nil }
func (n *NullStore) GetActiveGoalID(ctx context.Context, userID string) (*string, error) { return nil, nil }

// User biometrics operations
func (n *NullStore) UpsertUserBiometrics(ctx context.Context, biometrics *storage.UserBiometrics) error { return nil }
func (n *NullStore) GetUserBiometrics(ctx context.Context, userID string) (*storage.UserBiometrics, error) { return nil, nil }
func (n *NullStore) DeleteUserBiometrics(ctx context.Context, userID string) error { return nil }

// Label operations
func (n *NullStore) CreateLabel(ctx context.Context, label *storage.Label) error { return nil }
func (n *NullStore) UpdateLabel(ctx context.Context, label *storage.Label) error { return nil }
func (n *NullStore) DeleteLabel(ctx context.Context, userID, id string) error { return nil }
func (n *NullStore) GetLabel(ctx context.Context, userID, id string) (*storage.Label, error) { return nil, nil }
func (n *NullStore) ListLabels(ctx context.Context, userID string) ([]*storage.LabelWithUsage, error) { return nil, nil }

// Label assignment operations
func (n *NullStore) ListConsumptionLabels(ctx context.Context, userID, consumptionID string) ([]*storage.Label, error) { return nil, nil }
func (n *NullStore) AssignConsumptionLabels(ctx context.Context, userID, consumptionID string, labelIDs []string) error { return nil }
func (n *NullStore) UnassignConsumptionLabel(ctx context.Context, userID, consumptionID, labelID string) error { return nil }
func (n *NullStore) ListConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string) ([]*storage.Label, error) { return nil, nil }
func (n *NullStore) AssignConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string, labelIDs []string) error { return nil }
func (n *NullStore) UnassignConsumptionItemLabel(ctx context.Context, userID, consumptionItemID, labelID string) error { return nil }

// Filtering operations
func (n *NullStore) GetConsumptionsByLabels(ctx context.Context, userID string, labelNames []string, matchAll bool, limit, offset int) ([]*storage.Consumption, error) { return nil, nil }

// Event operations
func (n *NullStore) CreateEvent(ctx context.Context, event *storage.Event) error { return nil }
func (n *NullStore) UpdateEvent(ctx context.Context, event *storage.Event) error { return nil }
func (n *NullStore) DeleteEvent(ctx context.Context, userID, id string) error { return nil }
func (n *NullStore) GetEvent(ctx context.Context, userID, id string) (*storage.Event, error) { return nil, nil }
func (n *NullStore) ListEvents(ctx context.Context, userID string, options storage.EventListOptions) ([]*storage.Event, error) { return nil, nil }

// Event type operations
func (n *NullStore) CreateEventType(ctx context.Context, eventType *storage.EventType) error { return nil }
func (n *NullStore) UpdateEventType(ctx context.Context, eventType *storage.EventType) error { return nil }
func (n *NullStore) DeleteEventType(ctx context.Context, userID, id string) error { return nil }
func (n *NullStore) GetEventType(ctx context.Context, userID, id string) (*storage.EventType, error) { return nil, nil }
func (n *NullStore) ListEventTypes(ctx context.Context, userID string) ([]*storage.EventType, error) { return nil, nil }
func (n *NullStore) GetEventTypeCounts(ctx context.Context, userID string) (map[string]int, error) { return nil, nil }

// Event label assignment operations
func (n *NullStore) ListEventLabels(ctx context.Context, userID, eventID string) ([]*storage.Label, error) { return nil, nil }
func (n *NullStore) AssignEventLabels(ctx context.Context, userID, eventID string, labelIDs []string) error { return nil }
func (n *NullStore) UnassignEventLabel(ctx context.Context, userID, eventID, labelID string) error { return nil }

// Event link operations
func (n *NullStore) CreateEventLink(ctx context.Context, link *storage.EventLink) error { return nil }
func (n *NullStore) DeleteEventLink(ctx context.Context, userID, linkID string) error { return nil }
func (n *NullStore) ListEventLinks(ctx context.Context, userID, eventID string) ([]*storage.EventLink, error) { return nil, nil }

// API key operations
func (n *NullStore) CreateAPIKey(ctx context.Context, apiKey *storage.APIKey) error { return nil }
func (n *NullStore) GetAPIKey(ctx context.Context, userID, id string) (*storage.APIKey, error) { return nil, nil }
func (n *NullStore) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*storage.APIKey, error) { return nil, nil }
func (n *NullStore) ListAPIKeys(ctx context.Context, userID string) ([]*storage.APIKey, error) { return nil, nil }
func (n *NullStore) UpdateAPIKey(ctx context.Context, apiKey *storage.APIKey) error { return nil }
func (n *NullStore) RevokeAPIKey(ctx context.Context, userID, id string) error { return nil }
func (n *NullStore) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error { return nil }

// Database lifecycle
func (n *NullStore) Close() error { return nil }