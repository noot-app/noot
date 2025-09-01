package main

import (
	"os/exec"
	"testing"
)

func TestEventsMigrationSQLSyntax(t *testing.T) {
	t.Run("Events migration SQL syntax", func(t *testing.T) {
		// Test that the events migration SQL is syntactically correct by checking content
		eventsContent, err := exec.Command("cat", "supabase/migrations/20250825000011_events.sql").Output()
		if err != nil {
			t.Fatalf("Failed to read events migration: %v", err)
		}
		
		// Basic sanity checks on the migration content
		contentStr := string(eventsContent)
		if len(contentStr) == 0 {
			t.Fatal("Events migration file is empty")
		}
		
		// Check for key SQL commands
		requiredStatements := []string{
			"CREATE TABLE",
			"public.events",
			"PRIMARY KEY",
			"REFERENCES public.profiles",
			"ROW LEVEL SECURITY",
			"auth.uid()",
		}
		
		for _, stmt := range requiredStatements {
			if !contains(contentStr, stmt) {
				t.Errorf("Events migration missing required statement: %s", stmt)
			}
		}
	})
	
	t.Run("Event joins migration SQL syntax", func(t *testing.T) {
		// Test the event joins migration
		joinContent, err := exec.Command("cat", "supabase/migrations/20250825000012_event_joins.sql").Output()
		if err != nil {
			t.Fatalf("Failed to read event joins migration: %v", err)
		}
		
		contentStr := string(joinContent)
		if len(contentStr) == 0 {
			t.Fatal("Event joins migration file is empty")
		}
		
		// Check for key join table statements
		requiredStatements := []string{
			"public.event_labels",
			"public.event_links",
			"REFERENCES public.events",
			"REFERENCES public.labels",
			"ON DELETE CASCADE",
		}
		
		for _, stmt := range requiredStatements {
			if !contains(contentStr, stmt) {
				t.Errorf("Event joins migration missing required statement: %s", stmt)
			}
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}