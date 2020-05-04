package format

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type NullInt64 struct {
	Value sql.NullInt64
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Value.Valid {
		return []byte("0"), nil
	}

	return json.Marshal(ni.Value.Int64)
}

func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}


	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{Value: sql.NullInt64{i.Int64, false}}
	} else {
		*ni = NullInt64{Value: sql.NullInt64{i.Int64, true}}
	}

	return nil
}

type NullBool struct {
	sql.NullBool
}

func (nb *NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return nil, nil
	}

	return json.Marshal(nb.Bool)
}

type NullFloat64 struct {
	Value sql.NullFloat64
}

func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{Value: sql.NullFloat64{f.Float64, false}}
	} else {
		*nf = NullFloat64{Value: sql.NullFloat64{f.Float64, true}}
	}

	return nil
}

func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Value.Valid {
		return []byte("0"), nil
	}

	return json.Marshal(nf.Value.Float64)
}


type NullString struct {
	Value sql.NullString
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Value.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.Value.String)
}

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ns = NullString{Value:sql.NullString{s.String, false}}
	} else {
		*ns = NullString{Value:sql.NullString{s.String, true}}
	}

	return nil
}

type NullTime struct {
	Value sql.NullTime
}

func NewTime() NullTime {
	return NullTime{sql.NullTime{time.Now(), true}}
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Value.Valid {
		return []byte("null"), nil
	}
	val := fmt.Sprintf("\"%s\"", nt.Value.Time.Format(time.RFC3339))
	return []byte(val), nil
}

func (nt *NullTime) Scan(value interface{}) error {
	var t sql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nt = NullTime{Value: sql.NullTime{t.Time, false}}
	} else {
		*nt = NullTime{Value: sql.NullTime{t.Time, true}}
	}

	return nil
}
