package server

import (
	"strings"
)

// strPtrOrNil returns a trimmed string pointer or nil if empty
func strPtrOrNil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

// float64Ptr returns a pointer to the float64 value
func float64Ptr(f float64) *float64 {
	return &f
}
