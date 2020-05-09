package format

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type NullInt64 struct {
	V *sql.NullInt64
}


func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{V: &sql.NullInt64{i.Int64, false}}
	} else {
		*ni = NullInt64{V: &sql.NullInt64{i.Int64, true}}
	}

	return nil
}

func (nt NullInt64) Value() (driver.Value, error) {
	if !nt.V.Valid {
		return nil, nil
	}

	return nt.V.Int64, nil
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.V.Valid {
		return nil, nil
	}

	return json.Marshal(ni.V.Int64)
}

type NullFloat64 struct {
	V *sql.NullFloat64
}

func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{V: &sql.NullFloat64{f.Float64, false}}
	} else {
		*nf = NullFloat64{V: &sql.NullFloat64{f.Float64, true}}
	}

	return nil
}

func (nt NullFloat64) Value() (driver.Value, error) {
	if !nt.V.Valid {
		return nil, nil
	}

	return nt.V.Float64, nil
}

func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.V.Valid {
		return nil, nil
	}

	return json.Marshal(nf.V.Float64)
}


type NullString struct {
	V *sql.NullString
}


func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ns = NullString{V: &sql.NullString{s.String, false}}
	} else {
		*ns = NullString{V: &sql.NullString{s.String, true}}
	}

	return nil
}

func (nt NullString) Value() (driver.Value, error) {
	if !nt.V.Valid {
		return nil, nil
	}

	return nt.V.String, nil
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.V.Valid {
		return json.Marshal("\"\"")
	}
	return json.Marshal(ns.V.String)
}

type NullTime struct {
	V *sql.NullTime
}

func NewTime() *NullTime {
	return &NullTime{&sql.NullTime{time.Now(), true}}
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.V.Valid 	{
		return nil, nil
	}

	t := nt.V.Time
	b := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return []byte(`"` + b + `"`), nil
}


func (nt *NullTime) Scan(value interface{}) error {
	var t sql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nt = NullTime{V: &sql.NullTime{t.Time, false}}
	} else {
		*nt = NullTime{V: &sql.NullTime{t.Time, true}}
	}

	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.V.Valid {
		return nil, nil
	}

	return nt.V.Time, nil
}

