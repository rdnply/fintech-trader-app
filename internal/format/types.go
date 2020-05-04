package format

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type NullInt64 struct {
	V *sql.NullInt64
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.V.Valid {
		return nil, nil
	}

	return json.Marshal(ni.V.Int64)
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

func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.V.Valid {
		return nil, nil
	}

	return json.Marshal(nf.V.Float64)
}


type NullString struct {
	V *sql.NullString
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.V.Valid {
		return json.Marshal("\"\"")
	}
	return json.Marshal(ns.V.String)
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

type NullTime struct {
	V *sql.NullTime
}

func NewTime() *NullTime {
	return &NullTime{&sql.NullTime{time.Now(), true}}
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.V.Valid {
		return nil, nil
	}
	val := fmt.Sprintf("\"%s\"", nt.V.Time.Format(time.RFC3339))
	return []byte(val), nil
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
