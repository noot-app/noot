package server

import (
	"github.com/grantbirki/noot/internal/storage"
)

const (
	// API limits - updated to match tier-based requirements
	MaxDaysAllowed  = 365 // Pro tier maximum days
	MaxDaysFreeTier = 7   // Free tier maximum days
	MaxConsumptions = 50
	DefaultDays     = 7

	// HTTP headers and content types
	ContentTypeJSON = "application/json"
)

// User configuration aliases for consistency
const (
	SubscriptionTierFree = storage.SubscriptionTierFree
	SubscriptionTierPro  = storage.SubscriptionTierPro
)
