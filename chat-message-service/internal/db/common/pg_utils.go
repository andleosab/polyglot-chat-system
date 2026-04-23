package common

import "github.com/jackc/pgx/v5/pgtype"

// StringToPgText safely converts an optional *string pointer to a pgtype.Text struct.
func StringToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{
			Valid: false,
		}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

// You could add similar functions here, e.g.:
// func Int32ToPgInt4(i *int32) pgtype.Int4 { ... }
// func TimeToPgTimestamptz(t *time.Time) pgtype.Timestamptz { ... }
