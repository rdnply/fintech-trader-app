package format

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type NullInt64 struct {
	sql.NullInt64
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return nil, nil
	}

	return json.Marshal(ni.Int64)
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
	sql.NullFloat64
}

func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return nil, nil
	}
	return json.Marshal(nf.Float64)
}


type NullString struct {
	sql.NullString
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return nil, nil
	}
	return json.Marshal(ns.String)
}

type NullTime struct {
	sql.NullTime
}

func NewTime() NullTime {
	return NullTime{sql.NullTime{time.Now(), true}}
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return nil, nil
	}
	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
	return []byte(val), nil
}

//func (nt NullTime) Value() (driver.Value, error) {
//	if !nt.Valid {
//		return nil, nil
//	}
//
//	return nt.Time, nil
//}
//
//func (nt *NullTime) Scan(value interface{}) error {
//	if !nt.Valid {
//		return nil
//	}
//
//	nt.Time = value.(time.Time)
//
//	return nil
//}
