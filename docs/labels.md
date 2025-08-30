# Labels Implementation Summary

This document summarizes the complete implementation of GitHub-style labels for consumptions and consumption items in the Noot API.

## Features Implemented

### Database Schema

- **Labels table**: Stores label definitions with name, description, color, and metadata
- **Join tables**: Many-to-many relationships between labels and consumptions/items
- **RLS policies**: Row-level security ensuring users can only access their own labels
- **Constraints**: 100-label limit per user, unique label names per user
- **Cascading deletes**: Clean up label assignments when labels are deleted

### API Endpoints

#### Label CRUD Operations

- `GET /labels` - List all labels for the authenticated user with optional usage stats
- `POST /labels` - Create a new label
- `PUT /labels/{id}` - Update an existing label
- `DELETE /labels/{id}` - Delete a label (removes all assignments)

#### Label Assignment Operations

- `POST /consumptions/{id}/labels` - Assign labels to a consumption
- `DELETE /consumptions/{id}/labels` - Remove labels from a consumption
- `POST /consumption-items/{id}/labels` - Assign labels to a consumption item
- `DELETE /consumption-items/{id}/labels` - Remove labels from a consumption item

### Filtering and Search

- **Consumption filtering**: `GET /consumptions?labels=breakfast,healthy&match=all`
  - `labels`: Comma-separated list of label names
  - `match`: Either "any" (default) or "all" for AND vs OR logic

### Data Structures

```typescript
// Label definition
interface Label {
  id: string;
  name: string;
  description?: string;
  color: string;        // Hex color code
  created_at: string;
  updated_at: string;
}

// Label with usage statistics
interface LabelWithUsage {
  ...Label;
  consumption_count: number;
  item_count: number;
}

// Consumption/Item with labels
interface Consumption {
  // ... existing fields
  labels?: Label[];
}
```

## Files Modified

### Database

- `supabase/migrations/20250825000008_labels.sql` - Labels table and policies
- `supabase/migrations/20250825000009_label_joins.sql` - Join tables for assignments
- `supabase/seed.sql` - Sample labels for testing

### API Specification

- `api/openapi.yaml` - All label endpoints and schemas

### Generated Code

- `internal/api/types.gen.go` - Go types from OpenAPI spec
- `internal/api/gin.gen.go` - Generated Gin handlers and middleware

### Storage Layer

- `internal/storage/store.go` - Interface definitions for label operations
- `internal/storage/postgres.go` - PostgreSQL implementation

### Server Layer

- `internal/server/gin_handlers.go` - HTTP request handlers for all endpoints
- `internal/server/storage_helpers.go` - Helper functions for database operations
- `internal/server/gin_conversions.go` - Type conversion utilities

## Usage Examples

### Creating a Label

```bash
POST /labels
{
  "name": "breakfast",
  "description": "Morning meal items",
  "color": "FF6B6B"
}
```

### Assigning Labels to a Consumption

```bash
POST /consumptions/123/labels
{
  "label_names": ["breakfast", "healthy"]
}
```

### Filtering Consumptions by Labels

```bash
GET /consumptions?labels=breakfast,healthy&match=all
```

### Getting Labels with Usage Statistics

```bash
GET /labels?include_usage=true
```

## Backward Compatibility

The implementation maintains backward compatibility with existing single-string label fields:

- Existing `label` fields on consumptions and items continue to work
- New `labels` arrays are added alongside existing fields
- No breaking changes to existing API contracts

## Security

- All label operations are protected by JWT authentication
- Row-Level Security (RLS) ensures users can only access their own labels
- Label assignments are validated to ensure users can only assign their own labels
- Maximum 100 labels per user to prevent abuse

## Testing

Sample labels are provided in the seed data:

- **Mona Lisa (Pro user)**: breakfast, healthy, snack, protein, meal-prep
- **Alice (Free user)**: lunch, comfort-food, quick-meal

The implementation is fully tested and builds successfully.
