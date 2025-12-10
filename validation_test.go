package aqm

import (
	"testing"

	"github.com/google/uuid"
)

func TestIsRequired(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"valid string", "test", true},
		{"string with spaces", "  test  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRequired(tt.value)
			if result != tt.expected {
				t.Errorf("IsRequired(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsRequiredUUID(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name     string
		value    uuid.UUID
		expected bool
	}{
		{"nil UUID", NilUUID, false},
		{"valid UUID", validUUID, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRequiredUUID(tt.value)
			if result != tt.expected {
				t.Errorf("IsRequiredUUID(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		min      int
		expected bool
	}{
		{"empty string, min 0", "", 0, true},
		{"empty string, min 1", "", 1, false},
		{"exact length", "test", 4, true},
		{"longer than min", "testing", 4, true},
		{"shorter than min", "ab", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinLength(tt.value, tt.min)
			if result != tt.expected {
				t.Errorf("MinLength(%q, %d) = %v, want %v", tt.value, tt.min, result, tt.expected)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		max      int
		expected bool
	}{
		{"empty string, max 0", "", 0, true},
		{"empty string, max 10", "", 10, true},
		{"exact length", "test", 4, true},
		{"shorter than max", "ab", 4, true},
		{"longer than max", "testing", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxLength(tt.value, tt.max)
			if result != tt.expected {
				t.Errorf("MaxLength(%q, %d) = %v, want %v", tt.value, tt.max, result, tt.expected)
			}
		})
	}
}

func TestIsEmail(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty string", "", true},
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"valid email with plus", "user+tag@example.com", true},
		{"invalid - no @", "testexample.com", false},
		{"invalid - no domain", "test@", false},
		{"invalid - no user", "@example.com", false},
		{"invalid - spaces", "test @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmail(tt.value)
			if result != tt.expected {
				t.Errorf("IsEmail(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestMinValueInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		expected bool
	}{
		{"equal to min", 5, 5, true},
		{"greater than min", 10, 5, true},
		{"less than min", 3, 5, false},
		{"negative values", -5, -10, true},
		{"zero", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinValueInt(tt.value, tt.min)
			if result != tt.expected {
				t.Errorf("MinValueInt(%d, %d) = %v, want %v", tt.value, tt.min, result, tt.expected)
			}
		})
	}
}

func TestMaxValueInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		max      int
		expected bool
	}{
		{"equal to max", 5, 5, true},
		{"less than max", 3, 5, true},
		{"greater than max", 10, 5, false},
		{"negative values", -10, -5, true},
		{"zero", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxValueInt(tt.value, tt.max)
			if result != tt.expected {
				t.Errorf("MaxValueInt(%d, %d) = %v, want %v", tt.value, tt.max, result, tt.expected)
			}
		})
	}
}

func TestIsInList(t *testing.T) {
	t.Run("string values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    string
			list     []string
			expected bool
		}{
			{"value in list", "apple", []string{"apple", "banana", "cherry"}, true},
			{"value not in list", "orange", []string{"apple", "banana", "cherry"}, false},
			{"empty list", "apple", []string{}, false},
			{"value is empty string in list", "", []string{"", "apple"}, true},
			{"value is empty string not in list", "", []string{"apple", "banana"}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsInList(tt.value, tt.list)
				if result != tt.expected {
					t.Errorf("IsInList(%q, %v) = %v, want %v", tt.value, tt.list, result, tt.expected)
				}
			})
		}
	})

	t.Run("int values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    int
			list     []int
			expected bool
		}{
			{"value in list", 5, []int{1, 3, 5, 7, 9}, true},
			{"value not in list", 4, []int{1, 3, 5, 7, 9}, false},
			{"empty list", 5, []int{}, false},
			{"zero in list", 0, []int{0, 1, 2}, true},
			{"negative value in list", -5, []int{-10, -5, 0, 5}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsInList(tt.value, tt.list)
				if result != tt.expected {
					t.Errorf("IsInList(%d, %v) = %v, want %v", tt.value, tt.list, result, tt.expected)
				}
			})
		}
	})

	t.Run("UUID values", func(t *testing.T) {
		uuid1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		uuid2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		uuid3 := uuid.MustParse("00000000-0000-0000-0000-000000000003")
		uuid4 := uuid.MustParse("00000000-0000-0000-0000-000000000004")

		tests := []struct {
			name     string
			value    uuid.UUID
			list     []uuid.UUID
			expected bool
		}{
			{"UUID in list", uuid2, []uuid.UUID{uuid1, uuid2, uuid3}, true},
			{"UUID not in list", uuid4, []uuid.UUID{uuid1, uuid2, uuid3}, false},
			{"empty list", uuid1, []uuid.UUID{}, false},
			{"nil UUID in list", NilUUID, []uuid.UUID{NilUUID, uuid1}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsInList(tt.value, tt.list)
				if result != tt.expected {
					t.Errorf("IsInList(%v, %v) = %v, want %v", tt.value, tt.list, result, tt.expected)
				}
			})
		}
	})
}
