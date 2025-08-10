package service

import (
	"testing"
	"time"
)

// TestParseMonthYear - тестирует функцию parseMonthYear
func TestParseMonthYear(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth time.Month
		wantErr   bool
	}{
		{"valid MM-YYYY", "01-2025", 2025, time.January, false},
		{"valid YYYY-MM", "2025-12", 2025, time.December, false},
		{"empty string", "", 0, 0, true},
		{"invalid format", "2025/01", 0, 0, true},
		{"invalid month", "13-2025", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMonthYear(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMonthYear() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Year() != tt.wantYear || got.Month() != tt.wantMonth {
					t.Errorf("parseMonthYear() = %v, want %v-%v", got, tt.wantYear, tt.wantMonth)
				}
			}
		})
	}
}

// TestMonthsInclusive - тестирует функцию monthsInclusive
func TestMonthsInclusive(t *testing.T) {
	tests := []struct {
		name string
		a    time.Time
		b    time.Time
		want int
	}{
		{"same month", date(2025, 1), date(2025, 1), 1},
		{"two months", date(2025, 1), date(2025, 2), 2},
		{"one year", date(2025, 1), date(2026, 1), 13},
		{"cross year", date(2025, 11), date(2026, 2), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := monthsInclusive(tt.a, tt.b); got != tt.want {
				t.Errorf("monthsInclusive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMaxDate - тестирует функцию maxDate
func TestMaxDate(t *testing.T) {
	a := date(2025, 1)
	b := date(2025, 5)

	if got := maxDate(a, b); !got.Equal(b) {
		t.Errorf("maxDate() = %v, want %v", got, b)
	}

	if got := maxDate(b, a); !got.Equal(b) {
		t.Errorf("maxDate() = %v, want %v", got, b)
	}
}

func date(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}
