package storage

// TableNames defines the database table names for easier migration to other databases
var TableNames = struct {
	Users            string
	Consumptions     string
	ConsumptionItems string
	ItemsCache       string
	ItemAliases      string
	UserGoals        string
	UserBiometrics   string
}{
	Users:            "users",
	Consumptions:     "consumptions",
	ConsumptionItems: "consumption_items",
	ItemsCache:       "items_cache",
	ItemAliases:      "item_aliases",
	UserGoals:        "user_goals",
	UserBiometrics:   "user_biometrics",
}

// GetDropTableOrder returns tables in reverse dependency order for safe dropping
func GetDropTableOrder() []string {
	return []string{
		TableNames.ItemAliases,
		TableNames.ConsumptionItems,
		TableNames.ItemsCache,
		TableNames.UserGoals,
		TableNames.UserBiometrics,
		TableNames.Consumptions,
		TableNames.Users,
	}
}
