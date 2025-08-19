package storage

// TableNames defines the database table names for easier migration to other databases
var TableNames = struct {
	Users       string
	Meals       string
	ItemsCache  string
	ItemAliases string
}{
	Users:       "users",
	Meals:       "meals",
	ItemsCache:  "items_cache",
	ItemAliases: "item_aliases",
}

// GetDropTableOrder returns tables in reverse dependency order for safe dropping
func GetDropTableOrder() []string {
	return []string{
		TableNames.ItemAliases,
		TableNames.ItemsCache,
		TableNames.Meals,
		TableNames.Users,
	}
}
