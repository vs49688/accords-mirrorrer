package goutils

import (
	"database/sql"
	"time"
)

func SQLString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func SQLInt64(val int64) sql.NullInt64 {
	return sql.NullInt64{Int64: val, Valid: val != 0}
}

func SQLTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}
