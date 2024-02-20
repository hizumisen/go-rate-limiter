package testutils

import "time"

func NowProvider(t time.Time) func() time.Time {
	return func() time.Time {
		return t
	}
}

func NewTimeAt(hour int) time.Time {
	return time.Date(2000, 1, 1, hour, 0, 0, 0, time.UTC)
}
