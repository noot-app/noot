package storage

// TableNames defines the database table names for easier migration to other databases
var TableNames = struct {
	Profiles         string
	Consumptions     string
	ConsumptionItems string
	Items            string
	UserGoals        string
	APIKeys          string
}{
	Profiles:         "profiles",
	Consumptions:     "consumptions",
	ConsumptionItems: "consumption_items",
	Items:            "items",
	UserGoals:        "user_goals",
	APIKeys:          "api_keys",
}
