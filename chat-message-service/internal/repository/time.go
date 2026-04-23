package repository

import "time"

func toUtcTime(t time.Time) *time.Time {
	utc := t.UTC()
	return &utc
}

func toUtcTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}
