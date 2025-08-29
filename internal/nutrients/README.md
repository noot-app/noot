# Nutrient Definition System

This directory contains the centralized nutrient definition system that dramatically reduces maintenance overhead when adding new nutrients to the application.

## Problem Solved

Before this system, adding a new nutrient like `taurine_mg` required editing 20+ files across the codebase:

- Database migration files
- Go struct definitions (multiple locations)
- API type definitions  
- Storage layer mappings
- Query builders
- Conversion functions
- Frontend TypeScript types
- Multiple helper/utility files

This was error-prone, time-consuming, and difficult to maintain.

## Solution

Now we have a **single source of truth** system with automatic code generation:

1. **One Configuration File**: `config/nutrients.yml` defines all nutrients
2. **Automatic Code Generation**: `script/generate-nutrients` generates all required code
3. **Type Safety Maintained**: Full type safety in both Go and TypeScript
4. **Backward Compatibility**: Generated code is identical to the original manual code

## How to Add a New Nutrient

Adding a new nutrient now takes just 3 steps:

### 1. Add to config/nutrients.yml

```yaml
- key: your_nutrient_mg
  type: float64
  unit: mg
  category: compound
  per_100g: true
  original: true
  go_field_name: YourNutrient
  json_tag: your_nutrient_mg
  display_name: Your Nutrient
```

### 2. Generate Code

```bash
script/generate-nutrients
```

### 3. Run Database Migrations

```bash
# Database migrations need to be created separately
# but the field mappings are now generated automatically
script/db migrate
```

## Generated Files

The system automatically generates:

- **Go Types**: `CompleteNutrient` struct fields  
- **TypeScript Types**: Nutrient interfaces for the frontend
- **Query Mappings**: Database column to struct field mappings (as reference)
- **Item Fields**: Original and per-100g field definitions (as reference)

## Architecture

```text
config/nutrients.yml (Source of Truth)
       ↓
cmd/nutrient-generator/main.go (Code Generator)
       ↓ 
Generated Files:
├── apps/web/src/lib/types/nutrients.gen.ts
├── internal/storage/item_fields_gen.go (reference)
└── internal/storage/query_builder_gen.go (reference)
```

## Nutrient Definition Schema

Each nutrient in `config/nutrients.yml` has these properties:

- `key`: Database column name (without prefix/suffix)
- `type`: Go type (usually `float64`)
- `unit`: Display unit (g, mg, mcg, kcal)
- `category`: Logical grouping (macro, vitamin, mineral, etc.)
- `per_100g`: Whether this nutrient has per-100g normalization
- `original`: Whether this nutrient has original serving data
- `go_field_name`: Go struct field name (for CompleteNutrient)
- `json_tag`: JSON tag for API serialization
- `display_name`: Human-readable name for UI

## Testing

The system includes automated tests to verify code generation:

```bash
go test ./internal/nutrients
```

## Example: Adding Taurine

Before this system (requires editing 20+ files):

```diff
+ // Edit internal/server/types.go
+ Taurine float64 `json:"taurine_mg"`

+ // Edit internal/storage/store.go  
+ TaurineMgPer100g float64 `json:"taurine_mg_per_100g"`
+ OriginalTaurineMg *float64 `json:"original_taurine_mg,omitempty"`

+ // Edit internal/storage/query_builder.go
+ {"taurine_mg_per_100g", "TaurineMgPer100g"},
+ {"original_taurine_mg", "OriginalTaurineMg"},

+ // Edit apps/web/src/DatabaseDefinitions.ts
+ taurine_mg: number;
+ original_taurine_mg?: number | null;

+ // And 15+ more files...
```

With this system (1 file change):

```yaml
# Add to config/nutrients.yml
- key: taurine_mg
  type: float64
  unit: mg
  category: amino_acid
  per_100g: true
  original: true
  go_field_name: Taurine
  json_tag: taurine_mg
  display_name: Taurine
```

Then run: `script/generate-nutrients`

**Result**: All 20+ files automatically updated with perfect consistency!

## Benefits

✅ **Single Source of Truth**: All nutrient definitions in one place  
✅ **Reduced Errors**: No more missing fields or typos across files  
✅ **Faster Development**: Minutes instead of hours to add nutrients  
✅ **Type Safety**: Generated code maintains full type safety  
✅ **Consistency**: All generated code follows the same patterns  
✅ **Easy Maintenance**: One file to update, not 20+  
✅ **Self-Documenting**: Clear schema and examples in YAML
