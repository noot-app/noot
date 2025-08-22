# Adding New Nutrients to Noot

This guide walks through the complete process of adding a new nutrient to the noot application, using **DHA (Docosahexaenoic acid)** as a concrete example. The process involves changes across multiple layers of the application stack.

## Overview

Adding a nutrient requires updates to:
- Database schema (SQL migrations)
- Go backend types and logic
- OpenAI nutrition prompt (on OpenAI portal)
- API schema definitions
- Frontend UI components
- Goals/DRI reference data

## Prerequisites

- Go development environment
- Node.js for frontend development
- SQLite for database operations
- Understanding of the noot architecture

## Step-by-Step Process

### 1. Database Layer

#### 1.1 Create SQL Migration

Create a new migration file `internal/storage/migrations/007_add_dha_nutrient.sql`:

```sql
-- Migration 007: Add DHA (Docosahexaenoic acid) nutrient tracking
-- DHA is an omega-3 fatty acid important for brain and heart health

-- Add DHA column to consumptions table for tracking totals
ALTER TABLE consumptions ADD COLUMN dha_mg REAL NOT NULL DEFAULT 0;

-- Add DHA column to items_cache table for per-100g nutrition data
ALTER TABLE items_cache ADD COLUMN dha_mg_per_100g REAL NOT NULL DEFAULT 0;
```

**Pattern**: Use format `{nutrient_name}_{unit}` for column names. Always add to both tables:
- `consumptions`: Total amount consumed
- `items_cache`: Amount per 100g for calculation

#### 1.2 Verify Migration

After creating the migration file, run the database migration:

```bash
# Migration will run automatically when the server starts, or manually trigger
go run cmd/noot/main.go migrate
```

### 2. Backend Go Types

#### 2.1 Update CompleteNutrient Struct

Add DHA field to `internal/server/types.go`:

```go
type CompleteNutrient struct {
    // ... existing fields ...
    
    // Omega-3 fatty acids
    DHAmg float64 `json:"dha_mg"` // Docosahexaenoic acid in milligrams
    
    // ... rest of existing fields ...
}
```

**Pattern**: Use `{NutrientName}{Unit}` format, proper JSON tags, and descriptive comments.

#### 2.2 Update Storage Structs

Add DHA fields to both `Consumption` and `Item` structs in `internal/storage/store.go`:

```go
// Consumption struct - add to micronutrient section
type Consumption struct {
    // ... existing fields ...
    DHAmg float64 `json:"dha_mg"`
    // ... rest of fields ...
}

// Item struct - add to nutrition section  
type Item struct {
    // ... existing fields ...
    DHAmgPer100g float64 `json:"dha_mg_per_100g"`
    // ... rest of fields ...
}

// NutritionSummary struct - add total field
type NutritionSummary struct {
    // ... existing fields ...
    TotalDHAmg float64 `json:"total_dha_mg"`
    // ... rest of fields ...
}
```

#### 2.3 Update Unit Conversion Logic

If needed, update `internal/server/unit_converter.go` for any specific unit handling.

#### 2.4 Update Nutrition Service

Update aggregation logic in `internal/server/nutrition_service.go` to include DHA in consumption totals and summary calculations.

### 3. AI Provider Integration

#### 3.1 Update OpenAI System Prompt

The nutrition system prompt is now managed directly on OpenAI's portal rather than in the Go code. The application references the prompt using environment variables:

- `OPENAI_NUTRITION_PROMPT_ID`: The ID of the prompt on OpenAI's portal
- `OPENAI_NUTRITION_PROMPT_VERSION`: The version of the prompt to use

To add a new nutrient like DHA:

1. **Access OpenAI Portal**: Log into the OpenAI portal where the nutrition prompt is hosted
2. **Update JSON Structure**: Add the new nutrient to the expected JSON response structure:

```json
{
  "nutrients": {
    "calories": number,
    "protein_g": number,
    // ... existing nutrients ...
    "chloride_mg": number,
    "dha_mg": number
  }
}
```

3. **Add Reference Data**: Include accuracy reference data in the prompt instructions:

```text
DHA CONTENT: Include DHA (docosahexaenoic acid) content in milligrams. Focus on:
- Fatty fish (salmon, mackerel, sardines, tuna): 500-2000mg per 100g
- Fish oil supplements: Very high DHA content
- Algae-based foods: Moderate DHA for vegetarian sources  
- Most plant foods: 0mg DHA (ALA omega-3 instead)
- Fortified foods: Check product specifications

For DHA specifically:
- Salmon (farmed): ~1800mg per 100g
- Salmon (wild): ~1400mg per 100g  
- Sardines: ~1400mg per 100g
- Mackerel: ~1600mg per 100g
- Tuna: ~300-1200mg per 100g (varies by species)
- Most non-marine foods: 0mg
```

4. **Update Version**: Create a new version of the prompt and update the `OPENAI_NUTRITION_PROMPT_VERSION` environment variable

**Pattern**: The Go code automatically uses the prompt from OpenAI's portal - no code changes needed for prompt updates.

### 4. API Schema Updates

#### 4.1 Update OpenAPI Schema

Add DHA to relevant schemas in `api/openapi.yaml`:

```yaml
components:
  schemas:
    CompleteNutrient:
      type: object
      properties:
        # ... existing properties ...
        dha_mg:
          type: number
          format: float
          description: "Docosahexaenoic acid (DHA) in milligrams"
        # ... rest of properties ...

    NutritionSummary:
      type: object  
      properties:
        # ... existing properties ...
        total_dha_mg:
          type: number
          format: float
          description: "Total DHA consumed in milligrams"
```

#### 4.2 Regenerate API Types

After updating the OpenAPI schema:

```bash
./script/generate-types
```

This regenerates the Go API types from the OpenAPI specification.

### 5. Frontend Components

#### 5.1 Update Goals Component

Add DHA to the nutrients list in `apps/web/src/lib/components/Goals.svelte`:

```javascript
// All nutrients to display - add DHA to appropriate category
const keyNutrients = [
    // Essential macronutrients
    "calories", "protein_g", "total_carbs_g", "total_fat_g", "saturated_fat_g", "trans_fat_g", 
    
    // Omega-3 fatty acids
    "dha_mg",
    
    // ... rest of existing nutrients
];
```

**Pattern**: Add to the `keyNutrients` array in the appropriate category. The component will automatically handle display formatting and progress calculation.

### 6. Goals & DRI Integration

#### 6.1 Add to DRI Data (if available)

If DHA has established DRI values, add to `internal/goals/data/rda_ai.json`:

```json
{
  "nutrients": {
    "dha_mg": {
      "male": {
        "19-30": 250,
        "31-50": 250,
        "51-70": 250,
        "70+": 250
      },
      "female": {
        "19-30": 250,  
        "31-50": 250,
        "51-70": 250,
        "70+": 250
      },
      "unit": "mg",
      "note": "AI values; combined EPA+DHA recommendation"
    }
  }
}
```

#### 6.2 Add to FDA Daily Values (if available)

If DHA has FDA Daily Values, add to `internal/goals/data/dv.json`:

```json
{
  "fdaDailyValues": {
    "dha_mg": { 
      "value": 250, 
      "unit": "mg",
      "note": "Combined EPA+DHA recommendation"
    }
  }
}
```

**Note**: As of 2024, DHA does not have official FDA Daily Values, but this shows the pattern for when they exist.

## Testing Your Implementation

### 1. Unit Tests

Create tests for the new nutrient calculations:

```go
func TestDHACalculation(t *testing.T) {
    // Test DHA calculation logic
    item := Item{
        DHAmgPer100g: 1800, // Salmon DHA content
    }
    
    result := calculateNutrientForWeight(item, 100) // 100g serving
    assert.Equal(t, 1800.0, result.DHAmg)
}
```

### 2. Database Migration Tests

Verify migrations work correctly:

```bash
# Run full test suite to verify migrations
./script/test

# Look for migration-specific test output:
# "✓ Applied migration 007_add_dha_nutrient.sql"
```

### 3. Integration Tests  

Test database operations with new nutrient:

```bash
# Tests should cover:
# - Creating consumptions with DHA values
# - Retrieving nutrition summaries including DHA totals
# - Item cache operations with DHA per-100g values
./script/test -run TestSQLiteStore
```

### 4. API Schema Validation

Ensure OpenAPI types are correctly generated:

```bash
# Regenerate and verify types
./script/generate-types

# Check that generated types include new nutrient
grep -r "dha_mg" internal/api/types.gen.go
```

### 5. End-to-End Testing Checklist

Use this checklist to verify complete functionality:

- [ ] **Database Migration**: Migration applies without errors
- [ ] **Consumption Creation**: Can create consumption records with DHA values
- [ ] **Item Storage**: Can store and retrieve items with DHA per-100g data  
- [ ] **Nutrition Summaries**: DHA appears in aggregated nutrition totals
- [ ] **API Responses**: DHA fields present in JSON responses
- [ ] **Frontend Display**: DHA appears in Goals component with proper formatting
- [ ] **OpenAI Integration**: AI can parse and return DHA values for fish/seafood
- [ ] **Data Consistency**: DHA values calculated correctly across all layers

### 6. Manual Testing Scenarios

#### Scenario 1: Record DHA-rich meal
```
Test Input: "I ate 6 ounces of grilled salmon"
Expected: DHA value should be ~306mg (170g × 1800mg/100g)
Verify: Check consumption response includes dha_mg field
```

#### Scenario 2: View nutrition summary
```
Test: Create multiple fish consumptions over several days
Expected: Total DHA should sum correctly in daily/weekly summaries  
Verify: Check nutrition summary API includes total_dha_mg
```

#### Scenario 3: Goals component display
```  
Test: Open Goals component with DHA consumption data
Expected: DHA appears in nutrient list with progress bar
Verify: Units display as "mg", progress calculates if targets exist
```

### 7. Error Scenarios

Test edge cases and error handling:

- [ ] **Zero DHA foods**: Non-marine foods should return 0 DHA
- [ ] **Invalid units**: System handles unit conversion correctly
- [ ] **Missing data**: Graceful handling when DHA data unavailable
- [ ] **Large values**: High DHA values (supplements) don't break display

### 8. Performance Testing

For production deployments:

- [ ] **Migration speed**: Large database migrations complete in reasonable time
- [ ] **Query performance**: Adding DHA columns doesn't slow down nutrition queries
- [ ] **Frontend rendering**: Goals component renders efficiently with additional nutrient

### 9. Automated Test Commands

Run these commands to validate the complete implementation:

```bash
# Full test suite
./script/test

# Specific storage tests (includes migration testing)
./script/test -run TestSQLiteStore

# API type generation
./script/generate-types

# Lint code changes
./script/lint

# Build application (verifies no compilation errors)  
./script/build --single-target
```

### 10. Frontend Testing

If you have the frontend development environment:

```bash
# Start development server
cd apps/web
npm run dev

# Test scenarios:
# 1. Navigate to Goals/Progress page
# 2. Verify DHA appears in nutrient list
# 3. Create test consumption with salmon
# 4. Verify DHA progress updates
```

### Success Criteria

✅ **All tests pass**: No failing unit or integration tests  
✅ **Clean migration**: Database migration applies without errors  
✅ **API consistency**: DHA appears in all relevant API responses  
✅ **Frontend integration**: Goals component displays DHA correctly  
✅ **Data accuracy**: DHA calculations match expected values for known foods  
✅ **Documentation**: Process is clearly documented for future nutrients

### Troubleshooting Common Issues

| Issue | Solution |
|-------|----------|
| Migration fails with "column exists" | Check table names in `tables.go`, ensure Reset() drops correct tables |
| API types don't include DHA | Run `./script/generate-types` after updating OpenAPI schema |
| Frontend doesn't show DHA | Verify nutrient added to `keyNutrients` array in Goals.svelte |
| Zero DHA values for fish | Update OpenAI nutrition prompt on portal to include DHA reference data |
| Tests fail after adding nutrient | Verify all structs updated consistently across types.go and store.go |

## Automation Script

To streamline future nutrient additions, use the helper script:

```bash
./script/add-nutrient
```

This interactive script will:
- Prompt for nutrient name, unit, and metadata
- Generate SQL migration files
- Update Go structs with proper formatting  
- Update OpenAPI schema
- Create a PR-ready changeset

## Common Patterns

### Naming Conventions
- **Database columns**: `{nutrient}_{unit}` (e.g., `dha_mg`)
- **Go struct fields**: `{Nutrient}{Unit}` (e.g., `DHAmg`)
- **JSON fields**: `{nutrient}_{unit}` (e.g., `"dha_mg"`)
- **API schema**: Same as JSON fields

### Units
- Use standard nutrition units: `mg`, `mcg`, `g`, `kcal`
- Be consistent across all layers
- Include unit in field names for clarity

### Default Values  
- Always use `DEFAULT 0` in SQL migrations
- This ensures backwards compatibility with existing records

### Categories
Common nutrient categories for organization:
- **Macronutrients**: calories, protein, carbs, fat, fiber
- **Vitamins**: fat-soluble (A,D,E,K) and water-soluble (B-complex, C)
- **Minerals**: major (Ca, P, Mg, K) and trace (Fe, Zn, Cu, Se)
- **Fatty Acids**: saturated, trans, omega-3, omega-6
- **Other**: cholesterol, choline, caffeine

## Troubleshooting

### Common Issues

1. **Migration fails**: Ensure column names don't conflict with SQLite reserved words
2. **API types don't update**: Run `./script/generate-types` after OpenAPI changes
3. **Frontend doesn't show nutrient**: Check it's added to `keyNutrients` array
4. **Zero values in responses**: Update OpenAI nutrition prompt on portal with reference data

### Validation Checklist

- [ ] Migration file created and follows naming convention
- [ ] Both database tables updated (`consumptions` and `items_cache`)
- [ ] Go structs updated in `types.go` and `store.go`
- [ ] OpenAI nutrition prompt updated on portal with nutrient and reference data
- [ ] OpenAPI schema updated with nutrient fields
- [ ] API types regenerated successfully
- [ ] Frontend component includes nutrient in `keyNutrients`  
- [ ] Goals data added (if DRI/DV values exist)
- [ ] All tests pass
- [ ] End-to-end functionality verified

## Advanced Considerations

### Performance Impact
- New columns add minimal storage overhead
- Consider indexing for frequently queried nutrients
- Aggregation queries scale linearly with nutrient count

### Data Sources for Reference Values
- **USDA FoodData Central**: Comprehensive food composition database
- **NIH Office of Dietary Supplements**: Authoritative nutrient information  
- **FDA Nutrition Facts**: Labeling requirements and Daily Values
- **Scientific literature**: For newly researched nutrients

### Future Enhancements
- **Dynamic nutrient registry**: Configuration-driven nutrient system
- **Bulk import tools**: Automated data import from authoritative sources
- **Custom nutrient types**: User-defined nutrients for specialized tracking

## Example: Complete DHA Implementation

Here's what the complete DHA implementation looks like across all files:

### Migration (`007_add_dha_nutrient.sql`):
```sql
ALTER TABLE consumptions ADD COLUMN dha_mg REAL NOT NULL DEFAULT 0;
ALTER TABLE items_cache ADD COLUMN dha_mg_per_100g REAL NOT NULL DEFAULT 0;
```

### Go Types (`types.go`):
```go
type CompleteNutrient struct {
    // ... existing fields ...
    DHAmg float64 `json:"dha_mg"`
}
```

### OpenAPI Schema:
```yaml
dha_mg:
  type: number
  format: float
  description: "Docosahexaenoic acid (DHA) in milligrams"
```

### Frontend (`Goals.svelte`):
```javascript  
const keyNutrients = [
    // ... existing nutrients ...
    "dha_mg",
];
```

This systematic approach ensures complete integration of new nutrients across the entire application stack while maintaining data consistency and user experience quality.

## Quick Reference

### Adding a New Nutrient (Summary)

1. **Run the helper script**: `./script/add-nutrient`
2. **Test the implementation**: `./script/test`
3. **Update OpenAI nutrition prompt** on portal with reference data if needed
4. **Add DRI/DV values** to goals data if available
5. **Test end-to-end functionality**

### File Checklist

When adding a nutrient manually, ensure these files are updated:

- [ ] `internal/storage/migrations/###_add_[nutrient]_nutrient.sql`
- [ ] `internal/server/types.go` (CompleteNutrient struct)
- [ ] `internal/storage/store.go` (Consumption, Item, NutritionSummary structs)
- [ ] OpenAI nutrition prompt on portal (JSON structure and reference data)
- [ ] `api/openapi.yaml` (CompleteNutrient and NutritionSummary schemas)
- [ ] Generated API types (via `./script/generate-types`)
- [ ] `apps/web/src/lib/components/Goals.svelte` (keyNutrients array)
- [ ] `internal/goals/data/*.json` (optional, if DRI/DV values exist)

### Testing Commands

```bash
# Full test suite
./script/test

# Generate API types  
./script/generate-types

# Build application
./script/build --single-target

# Lint code
./script/lint
```

The process is designed to be reliable, repeatable, and maintainable for the long-term evolution of noot's nutrition tracking capabilities.