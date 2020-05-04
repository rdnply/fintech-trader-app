package format

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type Day struct {
	sql.NullTime
}

const DateLayout = "2006-01-02"

func (b *Day) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		return nil
	}

	const layout = `"` + DateLayout + `"`

	t, err := time.Parse(layout, s)
	if err != nil {
		return fmt.Errorf("can't parse date string: %v", err)
	}

	b.Time = t

	return nil
}

func (b *Day) MarshalJSON() ([]byte, error) {
	s := b.Time.Format(DateLayout)

	return []byte(s), nil
}

func (b Day) Value() (driver.Value, error) {
	return b.Time, nil
}

func (b *Day) Scan(value interface{}) error {
	t := value.(time.Time)
	str := t.Format(DateLayout)

	t, err := time.Parse(DateLayout, str)
	if err != nil {
		return fmt.Errorf("can't parse date in scanner for database: %v", err)
	}

	b.Time = t

	return nil
}
