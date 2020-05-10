package format

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

type Day struct {
	V sql.NullTime
}

const DateLayout = "2006-01-02"

func (d *Day) UnmarshalJSON(b []byte) error {
	s := string(b)

	const layout = `"` + DateLayout + `"`

	t, err := time.Parse(layout, s)
	if err != nil {
		d.V.Valid = false
		return err
	}

	d.V.Time = t
	d.V.Valid = true

	return nil
}

func (d *Day) MarshalJSON() ([]byte, error) {
	if !d.V.Valid {
		return nil, nil
	}

	t := d.V.Time
	f := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())

	return []byte(f), nil
}

func (d Day) Value() (driver.Value, error) {
	if !d.V.Valid {
		return nil, nil
	}

	return d.V.Time, nil
}

func (d *Day) Scan(value interface{}) error {
	var t sql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*d = Day(NullTime{V: sql.NullTime{Time: t.Time, Valid: false}})
	} else {
		*d = Day(NullTime{V: sql.NullTime{Time: t.Time, Valid: true}})
	}

	return nil
}
