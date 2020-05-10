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
	V sql.NullInt64
}

func NewNullInt64(n int64) *NullInt64 {
	return &NullInt64{V: sql.NullInt64{Int64: n, Valid: true}}
}

func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{V: sql.NullInt64{Int64: i.Int64, Valid: false}}
	} else {
		*ni = NullInt64{V: sql.NullInt64{Int64: i.Int64, Valid: true}}
	}

	return nil
}

func (ni NullInt64) Value() (driver.Value, error) {
	if !ni.V.Valid {
		return nil, nil
	}

	return ni.V.Int64, nil
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.V.Valid {
		return nil, nil
	}

	return json.Marshal(ni.V.Int64)
}

func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	var x *int64
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}

	if x != nil {
		ni.V.Valid = true
		ni.V.Int64 = *x
	} else {
		ni.V.Valid = false
	}

	return nil
}

//func NewNullInt64(n int64) *NullInt64 {
//	return &NullInt64{V: &sql.NullInt64{Int64: n, Valid: true}}
//}
//
//func (ni *NullInt64) Scan(value interface{}) error {
//	var i sql.NullInt64
//	if err := i.Scan(value); err != nil {
//		return err
//	}
//
//	if reflect.TypeOf(value) == nil {
//		*ni = NullInt64{V: &sql.NullInt64{Int64: i.Int64, Valid: false}}
//	} else {
//		*ni = NullInt64{V: &sql.NullInt64{Int64: i.Int64, Valid: true}}
//	}
//
//	return nil
//}
//
//func (ni NullInt64) Value() (driver.Value, error) {
//	if !ni.V.Valid {
//		return nil, nil
//	}
//
//	return ni.V.Int64, nil
//}
//
//func (ni *NullInt64) MarshalJSON() ([]byte, error) {
//	if !ni.V.Valid {
//		return nil, nil
//	}
//
//	return json.Marshal(ni.V.Int64)
//}
//
//func (ni *NullInt64) UnmarshalJSON(b []byte) error {
//	var x *int64
//	if err := json.Unmarshal(b, &x); err != nil {
//		return err
//	}
//
//	if x != nil {
//		ni.V.Valid = true
//		ni.V.Int64 = *x
//	} else {
//		ni.V.Valid = false
//	}
//
//	return nil
//}

//func (ni *NullInt64) UnmarshalJSON(b []byte) error {
//	err := json.Unmarshal(b, &ni.V.Int64)
//	ni.V.Valid = err == nil
//
//	return err
//}

type NullFloat64 struct {
	V sql.NullFloat64
}

func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{V: sql.NullFloat64{Float64: f.Float64, Valid: false}}
	} else {
		*nf = NullFloat64{V: sql.NullFloat64{Float64: f.Float64, Valid: true}}
	}

	return nil
}

func (nf NullFloat64) Value() (driver.Value, error) {
	if !nf.V.Valid {
		return nil, nil
	}

	return nf.V.Float64, nil
}

func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.V.Valid {
		return nil, nil
	}

	return json.Marshal(nf.V.Float64)
}

func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &nf.V.Float64)
	nf.V.Valid = err == nil

	return err
}

type NullString struct {
	V sql.NullString
}

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ns = NullString{V: sql.NullString{String: s.String, Valid: false}}
	} else {
		*ns = NullString{V: sql.NullString{String: s.String, Valid: true}}
	}

	return nil
}

func (ns NullString) Value() (driver.Value, error) {
	if !ns.V.Valid {
		return nil, nil
	}

	return ns.V.String, nil
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.V.Valid {
		return json.Marshal("\"\"")
	}

	return json.Marshal(ns.V.String)
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.V.String)
	ns.V.Valid = err == nil
	return err
}

type NullTime struct {
	V sql.NullTime
}

func NewNullTime() *NullTime {
	return &NullTime{V: sql.NullTime{Time: time.Now(), Valid: true}}
}

func (nt *NullTime) Scan(value interface{}) error {
	var t sql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nt = NullTime{V: sql.NullTime{Time: t.Time, Valid: false}}
	} else {
		*nt = NullTime{V: sql.NullTime{Time: t.Time, Valid: true}}
	}

	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.V.Valid {
		return nil, nil
	}

	return nt.V.Time, nil
}


func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.V.Valid {
		return nil, nil
	}

	t := nt.V.Time
	b := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return []byte(`"` + b + `"`), nil
}

func (nt *NullTime) UnmarshalJSON(b []byte) error {
	s := string(b)

	t, err := time.Parse(`"` + time.RFC3339 + `"`, s)
	if err != nil {
		nt.V.Valid = false
		return err
	}

	nt.V.Time = t
	nt.V.Valid = true

	return nil
}