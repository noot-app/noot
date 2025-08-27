package server

import (
	"github.com/grantbirki/noot/internal/storage"
)

const (
	// API limits
	MaxDaysAllowed  = 7
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
