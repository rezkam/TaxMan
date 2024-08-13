package utils

import "time"

// DateOnly returns a time.Time with the time set to 00:00:00
// This is useful for comparing dates without worrying about the time component
func DateOnly(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
