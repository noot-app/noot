# Events System Implementation Summary

This document summarizes the complete implementation of the Events system for manual health/symptom/activity tracking in the Noot API.

## Features Implemented

### Database Schema

- **Events table**: Stores event definitions with name, category, start/end time, level (0-10), notes, and color
- **Join tables**: Many-to-many relationships between events and labels, plus manual consumption/item links
- **RLS policies**: Row-level security ensuring users can only access their own events
- **Constraints**: 1000-event limit per user, proper data validation
- **Cascading deletes**: Clean up event assignments when events are deleted

### Data Structures

```typescript
// Event definition
interface Event {
  id: string;
  user_id: string;
  name: string;               // Required - event name
  category?: string;          // Optional - e.g., "symptom", "activity", "measurement"
  started_at: string;         // Required - ISO timestamp
  ended_at?: string;          // Optional - ISO timestamp
  level?: number;             // Optional - 0-10 scale for intensity/severity
  note?: string;              // Optional - up to 1000 characters
  color?: string;             // Optional - hex color for UI
  created_at: string;
  updated_at: string;
}

// Event with full details (includes labels and links)
interface EventWithDetails {
  ...Event;
  labels: Label[];           // Assigned labels
  links: EventLink[];        // Manual consumption/item correlations
}

// Manual correlation link
interface EventLink {
  id: string;
  event_id: string;
  consumption_id?: string;    // Links to a consumption
  consumption_item_id?: string; // OR links to a consumption item
  created_at: string;
}
```

### API Endpoints

#### Event CRUD Operations

- `GET /events` - List events with filtering (date, category, level, labels)
- `POST /events` - Create a new event
- `GET /events/{id}` - Get event with labels and links
- `PUT /events/{id}` - Update an existing event
- `DELETE /events/{id}` - Delete event (removes all associations)

#### Label Assignment Operations

- `GET /events/{id}/labels` - Get labels assigned to an event
- `POST /events/{id}/labels` - Assign labels to an event
- `DELETE /events/{id}/labels/{labelId}` - Remove label from event

#### Manual Correlation Operations

- `GET /events/{id}/links` - Get consumption/item links for an event
- `POST /events/{id}/links` - Create manual link between event and consumption/item
- `DELETE /events/{id}/links/{linkId}` - Remove correlation link

### Filtering and Search

- **Date filtering**: `start_date`, `end_date` parameters
- **Category filtering**: Filter by event category
- **Level filtering**: `level_min`, `level_max` for intensity ranges
- **Label filtering**: `labels=label1,label2&match=all|any`
  - `match=any`: Events with at least one of the specified labels
  - `match=all`: Events with all of the specified labels
- **Pagination**: `limit` (max 100) and `offset` parameters

## Usage Examples

### Creating an Event

```bash
POST /events
{
  "name": "Migraine headache",
  "category": "symptom",
  "started_at": "2024-01-15T10:30:00Z",
  "level": 8,
  "note": "Severe headache started after lunch, sensitive to light",
  "color": "#FF6B6B",
  "duration_minutes": 180
}
```

### Assigning Labels to an Event

```bash
POST /events/123/labels
{
  "ids": ["label-123", "label-456"]
}
```

### Creating Manual Correlation

```bash
POST /events/123/links
{
  "consumption_id": "consumption-789"
}
```

### Filtering Events

```bash
# Get symptom events from last week with high severity
GET /events?category=symptom&start_date=2024-01-08T00:00:00Z&level_min=7

# Get events labeled as "trigger-foods" related
GET /events?labels=trigger-foods&match=any
```

## Files Modified

### Database

- `supabase/migrations/20250825000011_events.sql` - Events table and policies
- `supabase/migrations/20250825000012_event_joins.sql` - Join tables for labels and links

### API Specification

- `api/openapi.yaml` - All event endpoints and schemas

### Generated Code

- `internal/api/types.gen.go` - Go types from OpenAPI spec
- `internal/api/gin.gen.go` - Generated Gin handlers and middleware
- `apps/web/src/lib/api/schema.ts` - TypeScript types for frontend

### Storage Layer

- `internal/storage/store.go` - Interface definitions for event operations
- `internal/storage/postgres.go` - PostgreSQL implementation

### Server Layer

- `internal/server/gin_handlers.go` - HTTP request handlers for all endpoints

### Testing

- `internal/storage/events_test.go` - Unit tests for event structs and mock operations
- `migrations_test.go` - Migration syntax validation tests

## Security

- **Row-level security (RLS)** ensures users can only access their own events
- **User limits** prevent abuse (1000 events per user)
- **Input validation** for all fields (length limits, data types, ranges)
- **Proper foreign key constraints** maintain data integrity

## Performance Considerations

- **Indexed queries** on user_id, started_at, category, and level
- **Efficient label filtering** with proper JOIN strategies
- **Pagination** support for large event lists
- **Composite indexes** for common query patterns

## Integration with Existing Systems

- **Labels system**: Full integration with existing GitHub-style labels
- **Consumptions**: Manual correlation linking for trigger analysis
- **User management**: Proper user ownership and permissions
- **API patterns**: Consistent with existing endpoint patterns

## Future Enhancements

- **Automated correlation analysis**: Time-window analysis for trigger identification
- **Event templates**: Common event types with pre-filled data
- **Recurring events**: Support for regular tracking (e.g., daily symptoms)
- **Export functionality**: CSV/JSON export for external analysis
- **Notification system**: Reminders for regular event logging
- **Analytics dashboard**: Visual correlation analysis and trends

## Correlation Analysis Capabilities

The Events system is designed to support comprehensive correlation analysis:

### Time-Window Analysis
- Events can be correlated with consumptions within specified time windows
- Support for both "trigger analysis" (events following meals) and "preparation analysis" (meals following activities)

### Statistical Analysis Ready
- Level field (0-10) enables quantitative analysis
- Timestamp precision allows for accurate temporal correlation
- Label system enables categorical analysis

### Manual Override Support
- Users can create explicit links between events and consumptions
- Supports complex scenarios where automatic correlation might miss connections
- Allows for user knowledge and context to enhance analysis

This implementation provides a solid foundation for advanced nutrition correlation analysis while maintaining simplicity and flexibility for users.