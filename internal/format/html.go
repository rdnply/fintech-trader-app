package format

import (
	"fmt"
	"strconv"
)

func PrintNullInt64(n *NullInt64) string {
	if n == nil || !n.V.Valid {
		return ""
	}

	return fmt.Sprintf("%d", n.V.Int64)
}

func PrintNullFloat64(n *NullFloat64) string {
	if n == nil || !n.V.Valid {
		return ""
	}

	return fmt.Sprintf("%2f", n.V.Float64)
}

func PrintNullString(n *NullString) string {
	if n == nil || !n.V.Valid {
		return ""
	}

	return n.V.String
}

func PrintNullTime(n *NullTime) string {
	if n == nil || !n.V.Valid {
		return ""
	}

	const layout = "2006-01-02T15:04:05Z"

	return n.V.Time.Format(layout)
}

func JoinNullInt(s string , n *NullInt64) string {
	if n == nil || !n.V.Valid {
		return ""
	}

	return s +  strconv.FormatInt(n.V.Int64, 64)
}

