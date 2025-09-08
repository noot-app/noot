package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
)

// DateRangeParams represents parsed date range parameters from a request
type DateRangeParams struct {
	StartTime time.Time
	EndTime   time.Time
	Days      int
}

// getCurrentUser retrieves the current authenticated user from context
func getCurrentUser(c *gin.Context) (*storage.User, error) {
	user := GetAuthenticatedUser(c)
	if user != nil {
		return user, nil
	}

	// No fallback to default user - authentication is required
	return nil, NewAppError("Authentication required", http.StatusUnauthorized, nil)
}

// parseDateRangeParams parses date range parameters from HTTP request
// Supports both direct date range (start/end) and simplified timeWindow approach
func parseDateRangeParams(r *http.Request) (*DateRangeParams, error) {
	startParam := r.URL.Query().Get("start")
	endParam := r.URL.Query().Get("end")

	var startTime, endTime time.Time
	var days int

	if startParam != "" && endParam != "" {
		// Direct date range provided by client (already in UTC)
		var err error
		startTime, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			return nil, NewAppError("Invalid start date format", http.StatusBadRequest, err)
		}

		endTime, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			return nil, NewAppError("Invalid end date format", http.StatusBadRequest, err)
		}

		// Calculate days for subscription validation
		days = int(endTime.Sub(startTime).Hours()/24) + 1
	} else {
		// Use simplified approach - check for days parameter
		days = DefaultDays // default to week view

		dayStr := r.URL.Query().Get("days")
		if dayStr != "" {
			if parsedDays, err := strconv.Atoi(dayStr); err == nil && parsedDays > 0 {
				days = parsedDays
			}
		}

		// Calculate date range using server UTC time
		endTime = time.Now().UTC()
		startTime = endTime.AddDate(0, 0, -days+1)

		// Set times to beginning/end of day for proper date range queries
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)
	}

	return &DateRangeParams{
		StartTime: startTime,
		EndTime:   endTime,
		Days:      days,
	}, nil
}

// validateSubscriptionAccess checks if user has access to the requested number of days
func validateSubscriptionAccess(user *storage.User, days int) error {
	// Free tier users are limited to 7 days maximum
	if strings.ToLower(user.SubscriptionTier) != SubscriptionTierPro {
		if days > MaxDaysFreeTier {
			return NewAppError("Free tier limited to 7 days maximum. Upgrade to Pro for access to up to 365 days.", http.StatusForbidden, nil)
		}
	} else {
		// Pro tier users are limited to 365 days maximum
		if days > MaxDaysAllowed {
			return NewAppError("Maximum date range is 365 days", http.StatusForbidden, nil)
		}
	}
	return nil
}

// applyDaysLimit applies the maximum days limit for performance
func applyDaysLimit(days int) int {
	if days > MaxDaysAllowed {
		return MaxDaysAllowed
	}
	return days
}

// QuantityInfo represents extracted quantity information from item names
type QuantityInfo struct {
	CleanName  string  // Name with quantity expressions removed
	Multiplier float64 // Quantity multiplier (0.5 for "half", 2.0 for "2 cans", etc.)
}

// ExtractQuantityFromName extracts quantity expressions from item names (exported for testing)
func ExtractQuantityFromName(name string) QuantityInfo {
	return extractQuantityFromName(name)
}

// extractQuantityFromName extracts quantity expressions from item names
func extractQuantityFromName(name string) QuantityInfo {
	// Normalize spaces first
	lowerName := strings.ToLower(strings.TrimSpace(name))
	// Replace multiple spaces with single spaces
	lowerName = strings.Join(strings.Fields(lowerName), " ")

	// Define quantity patterns and their multipliers
	patterns := []struct {
		pattern    string
		multiplier float64
	}{
		// Fractional quantities
		{"half", 0.5},
		{"1/2", 0.5},
		{"quarter", 0.25},
		{"1/4", 0.25},
		{"third", 0.33},
		{"1/3", 0.33},
		{"two thirds", 0.67},
		{"2/3", 0.67},

		// Size descriptors (approximate multipliers based on common food sizes)
		{"small", 0.75},
		{"mini", 0.5},
		{"tiny", 0.4},
		{"medium", 1.0},
		{"large", 1.3},
		{"big", 1.4},
		{"extra large", 1.6},
		{"xl", 1.6},
		{"jumbo", 1.8},

		// Portion descriptors
		{"handful", 0.6},  // Roughly a small portion
		{"pinch", 0.05},   // Very small amount
		{"dash", 0.08},    // Small amount
		{"sprinkle", 0.1}, // Small amount
		{"few", 0.3},      // A few of something
		{"couple", 0.2},   // A couple/pair
		{"some", 0.5},     // Some amount
		{"bit", 0.3},      // A bit of something
		{"little", 0.3},   // A little bit
		{"bunch", 0.8},    // A bunch of something

		// Article + quantity combinations
		{"a handful", 0.6},
		{"a pinch", 0.05},
		{"a dash", 0.08},
		{"a sprinkle", 0.1},
		{"a few", 0.3},
		{"a couple", 0.2},
		{"a bit", 0.3},
		{"a little", 0.3},
		{"a bunch", 0.8},

		// Article + size combinations
		{"a small", 0.75},
		{"a large", 1.3},
		{"a big", 1.4},

		// Slicing/cutting descriptors
		{"slices of", 1.0},   // "slices of" pattern
		{"slice of", 1.0},    // "slice of" pattern

		// Numeric quantities
		{"two", 2.0},
		{"three", 3.0},
		{"four", 4.0},
		{"five", 5.0},
		{"six", 6.0},
		{"seven", 7.0},
		{"eight", 8.0},
		{"nine", 9.0},
		{"ten", 10.0},
	}

	cleanName := lowerName
	multiplier := 1.0

	// Check for numeric prefixes first (1 can, 2 cans, etc.)
	for i := 2; i <= 10; i++ {
		numPrefix := fmt.Sprintf("%d ", i)
		if strings.HasPrefix(lowerName, numPrefix) {
			multiplier = float64(i)
			cleanName = strings.TrimPrefix(lowerName, numPrefix)
			break
		}
	}

	// If no numeric prefix found, check for word patterns
	if multiplier == 1.0 {
		for _, p := range patterns {
			// Look for pattern at the beginning of the name
			if strings.HasPrefix(lowerName, p.pattern+" ") {
				multiplier = p.multiplier
				cleanName = strings.TrimPrefix(lowerName, p.pattern+" ")
				break
			}
		}
	}

	// Additional cleanup for compound patterns
	// Handle remaining slice/piece descriptors after initial quantity extraction
	additionalDescriptors := []string{
		"slices of ", "slice of ", "pieces of ", "piece of ",
		"strips of ", "strip of ", "chunks of ", "chunk of ",
	}
	
	for _, desc := range additionalDescriptors {
		if strings.HasPrefix(cleanName, desc) {
			cleanName = strings.TrimPrefix(cleanName, desc)
			break
		}
	}

	// Clean up container words and plural forms after quantity extraction
	containerWords := []string{
		"cans ", "can ", "bottles ", "bottle ", "cups ", "cup ",
		"glasses ", "glass ", "bowls ", "bowl ", "servings ", "serving ",
		"portions ", "portion ", "pieces ", "piece ", "pints ", "pint ",
		"containers ", "container ",
	}

	// Remove container words at the beginning
	for _, cw := range containerWords {
		if strings.HasPrefix(cleanName, cw) {
			cleanName = strings.TrimPrefix(cleanName, cw)
			break
		}
	}

	// Handle "of" patterns (e.g., "cans of soda", "bottle of water")
	for _, cw := range containerWords {
		ofPattern := cw + "of "
		if strings.HasPrefix(cleanName, ofPattern) {
			cleanName = strings.TrimPrefix(cleanName, ofPattern)
			break
		}
	}

	// Clean up remaining "of " at the start (in case container word was removed first)
	cleanName = strings.TrimPrefix(cleanName, "of ")

	// Remove trailing container words (e.g., "water bottle" -> "water")
	trailingContainers := []string{
		" cans", " can", " bottles", " bottle", " cups", " cup",
		" glasses", " glass", " bowls", " bowl", " servings", " serving",
		" portions", " portion", " pieces", " piece", " pints", " pint",
		" containers", " container",
	}
	for _, tc := range trailingContainers {
		if strings.HasSuffix(cleanName, tc) {
			cleanName = strings.TrimSuffix(cleanName, tc)
			break
		}
	}

	return QuantityInfo{
		CleanName:  strings.TrimSpace(cleanName),
		Multiplier: multiplier,
	}
}

// normalizeItemName normalizes item names for consistent matching
func normalizeItemName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// normalizeItemNameForCache normalizes item names for caching by removing brand information
// This prevents cache fragmentation due to brand names appearing in different positions
func normalizeItemNameForCache(name string, brand *string) string {
	normalized := normalizeItemName(name)

	// If no brand provided, return as-is
	if brand == nil || strings.TrimSpace(*brand) == "" {
		return normalized
	}

	brandLower := strings.ToLower(strings.TrimSpace(*brand))

	// First, try to remove the complete brand phrase
	cleanName := normalized

	// Remove complete brand phrase from beginning (e.g., "ben jerry vanilla ice cream" -> "vanilla ice cream")
	if strings.HasPrefix(cleanName, brandLower+" ") {
		cleanName = strings.TrimPrefix(cleanName, brandLower+" ")
	} else if strings.HasSuffix(cleanName, " "+brandLower) {
		// Remove complete brand phrase from end (e.g., "vanilla ice cream ben jerry" -> "vanilla ice cream")
		cleanName = strings.TrimSuffix(cleanName, " "+brandLower)
	} else if cleanName == brandLower {
		// Remove standalone brand phrase (e.g., just "ben jerry" -> "")
		cleanName = ""
	} else {
		// Fallback: try removing individual brand words for partial matches
		brandWords := strings.Fields(brandLower)
		for _, brandWord := range brandWords {
			// Remove brand word from beginning (e.g., "ben vanilla ice cream" -> "vanilla ice cream")
			if strings.HasPrefix(cleanName, brandWord+" ") {
				cleanName = strings.TrimPrefix(cleanName, brandWord+" ")
				break // Only remove first match to avoid over-processing
			}
			// Remove brand word from end (e.g., "vanilla ice cream jerry" -> "vanilla ice cream")
			if strings.HasSuffix(cleanName, " "+brandWord) {
				cleanName = strings.TrimSuffix(cleanName, " "+brandWord)
				break // Only remove first match to avoid over-processing
			}
		}
	}

	// Handle case where brand appears multiple times - remove from end if still present
	if strings.Contains(cleanName, " "+brandLower) {
		cleanName = strings.TrimSuffix(cleanName, " "+brandLower)
	}

	// Fallback to original if we removed everything (shouldn't happen in practice)
	if cleanName == "" {
		return normalized
	}

	return cleanName
}

// normalizeItemNameWithQuantity normalizes item names and extracts quantity information
func normalizeItemNameWithQuantity(name string) (normalizedName string, quantityInfo QuantityInfo) {
	quantityInfo = extractQuantityFromName(name)
	normalizedName = normalizeItemName(quantityInfo.CleanName)
	return normalizedName, quantityInfo
}

// normalizeItemNameWithQuantityForCache normalizes item names, extracts quantity, and removes brand for cache keys
func normalizeItemNameWithQuantityForCache(name string, brand *string) (normalizedName string, quantityInfo QuantityInfo) {
	quantityInfo = extractQuantityFromName(name)
	normalizedName = normalizeItemNameForCache(quantityInfo.CleanName, brand)
	return normalizedName, quantityInfo
}

// NormalizeItemNameWithQuantity normalizes item names and extracts quantity information (exported for testing)
func NormalizeItemNameWithQuantity(name string) (normalizedName string, quantityInfo QuantityInfo) {
	return normalizeItemNameWithQuantity(name)
}

// generateCanonicalFoodName creates a canonical base food name for deduplication
// This strips all quantity descriptors, containers, and brand info to get the core food item
func generateCanonicalFoodName(name string, brand *string) string {
	// First extract quantity info to get clean name
	quantityInfo := extractQuantityFromName(name)
	cleanName := quantityInfo.CleanName
	
	// Special case: if the clean name is empty after quantity extraction, return empty
	if strings.TrimSpace(cleanName) == "" {
		return ""
	}
	
	// Handle brand removal manually for canonical names (don't use normalizeItemNameForCache fallback)
	canonicalName := strings.ToLower(strings.TrimSpace(cleanName))
	
	if brand != nil && strings.TrimSpace(*brand) != "" {
		brandLower := strings.ToLower(strings.TrimSpace(*brand))
		
		// If the cleaned name is exactly the brand, return empty (no food content)
		if canonicalName == brandLower {
			return ""
		}
		
		// Remove brand from the canonical name using the same logic as normalizeItemNameForCache
		// but without the fallback to original
		if strings.HasPrefix(canonicalName, brandLower+" ") {
			canonicalName = strings.TrimPrefix(canonicalName, brandLower+" ")
		} else if strings.HasSuffix(canonicalName, " "+brandLower) {
			canonicalName = strings.TrimSuffix(canonicalName, " "+brandLower)
		} else {
			// Try removing individual brand words
			brandWords := strings.Fields(brandLower)
			for _, brandWord := range brandWords {
				if strings.HasPrefix(canonicalName, brandWord+" ") {
					canonicalName = strings.TrimPrefix(canonicalName, brandWord+" ")
					break
				}
				if strings.HasSuffix(canonicalName, " "+brandWord) {
					canonicalName = strings.TrimSuffix(canonicalName, " "+brandWord)
					break
				}
			}
		}
		
		// Final cleanup for multiple brand occurrences
		if strings.Contains(canonicalName, " "+brandLower) {
			canonicalName = strings.TrimSuffix(canonicalName, " "+brandLower)
		}
	}
	
	// Additional normalization for canonical names
	canonicalName = normalizeCanonicalName(canonicalName)
	
	return canonicalName
}

// normalizeCanonicalName applies additional normalization for consistent canonical food names
func normalizeCanonicalName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	
	// Remove common descriptors that don't change the base food (handle multiple iterations)
	descriptors := []string{
		"fresh ", "organic ", "raw ", "cooked ", "baked ", "grilled ", "fried ",
		"steamed ", "boiled ", "roasted ", "sliced ", "diced ", "chopped ",
		"whole ", "ground ", "crushed ", "minced ", "shredded ", "grated ",
		"frozen ", "canned ", "dried ", "dehydrated ", "pickled ", "salted ",
		"unsalted ", "seasoned ", "spiced ", "sweet ", "sour ", "bitter ",
		"ripe ", "unripe ", "green ", "red ", "yellow ", "white ", "black ",
	}
	
	// Keep removing descriptors until no more are found
	changed := true
	for changed {
		changed = false
		for _, desc := range descriptors {
			if strings.HasPrefix(normalized, desc) {
				normalized = strings.TrimPrefix(normalized, desc)
				normalized = strings.TrimSpace(normalized)
				changed = true
				break
			}
		}
	}
	
	// Handle plural/singular normalization for common foods
	pluralToSingular := map[string]string{
		"carrots":    "carrot",
		"apples":     "apple", 
		"bananas":    "banana",
		"oranges":    "orange",
		"tomatoes":   "tomato",
		"potatoes":   "potato",
		"onions":     "onion",
		"eggs":       "egg",
		"cookies":    "cookie",
		"crackers":   "cracker",
		"chips":      "chip",
		"strawberries": "strawberry",
		"blueberries": "blueberry",
		"grapes":     "grape",
		"nuts":       "nut",
		"almonds":    "almond",
		"walnuts":    "walnut",
		"berries":    "berry",
	}
	
	// Check for exact plural matches
	if singular, exists := pluralToSingular[normalized]; exists {
		normalized = singular
	}
	
	return strings.TrimSpace(normalized)
}

// GenerateCanonicalFoodName exported function for testing
func GenerateCanonicalFoodName(name string, brand *string) string {
	return generateCanonicalFoodName(name, brand)
}

// getBrandOrEmpty returns the brand string or empty string if nil
func getBrandOrEmpty(brand *string) string {
	if brand == nil {
		return ""
	}
	return *brand
}

// generateBrandAwareVariations generates brand-aware name variations using the LLM-parsed brand
// This leverages the fact that our LLM already extracts brand information during parsing
func generateBrandAwareVariations(itemName string, brandFromLLM *string) []string {
	lowerName := strings.ToLower(strings.TrimSpace(itemName))
	variations := []string{lowerName} // Always include original

	// If we don't have an LLM-extracted brand, only return the original
	if brandFromLLM == nil || strings.TrimSpace(*brandFromLLM) == "" {
		return variations
	}

	brand := strings.ToLower(strings.TrimSpace(*brandFromLLM))
	brandWords := strings.Fields(brand)

	// Generate brand + item variations
	// Example: brand="noosa", itemName="vanilla yogurt" -> ["vanilla yogurt", "noosa vanilla yogurt", "vanilla yogurt noosa"]

	// Variation 1: Brand + Item Name (if not already present)
	brandPlusItem := brand + " " + lowerName
	if brandPlusItem != lowerName && !contains(variations, brandPlusItem) {
		variations = append(variations, brandPlusItem)
	}

	// Variation 2: Item Name + Brand (if not already present)
	itemPlusBrand := lowerName + " " + brand
	if itemPlusBrand != lowerName && !contains(variations, itemPlusBrand) {
		variations = append(variations, itemPlusBrand)
	}

	// Variation 3: Handle cases where brand might already be in the item name
	// Try removing the brand from the item name to get the base product name
	for _, brandWord := range brandWords {
		// Remove brand word from beginning
		if strings.HasPrefix(lowerName, brandWord+" ") {
			baseProduct := strings.TrimPrefix(lowerName, brandWord+" ")
			if baseProduct != lowerName && !contains(variations, baseProduct) {
				variations = append(variations, baseProduct)
			}
		}

		// Remove brand word from end
		if strings.HasSuffix(lowerName, " "+brandWord) {
			baseProduct := strings.TrimSuffix(lowerName, " "+brandWord)
			if baseProduct != lowerName && !contains(variations, baseProduct) {
				variations = append(variations, baseProduct)
			}
		}
	}

	// Variation 4: For multi-word brands, try different orderings
	if len(brandWords) > 1 {
		// Reverse brand word order + item
		reversedBrand := strings.Join(reverseStringSlice(brandWords), " ")
		reversedBrandPlusItem := reversedBrand + " " + lowerName
		if !contains(variations, reversedBrandPlusItem) {
			variations = append(variations, reversedBrandPlusItem)
		}
	}

	return variations
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to reverse a string slice
func reverseStringSlice(slice []string) []string {
	reversed := make([]string, len(slice))
	for i, s := range slice {
		reversed[len(slice)-1-i] = s
	}
	return reversed
}
