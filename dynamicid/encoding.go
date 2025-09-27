package dynamicid

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	stringReflectType = reflect.TypeOf(string(""))
	intReflectType    = reflect.TypeOf(int(0))
	uintReflectType   = reflect.TypeOf(uint(0))
	boolReflectType   = reflect.TypeOf(bool(false))
)

// UnmarshallCSV record into a struct in-place
func UnmarshallCSV(data string, v any) error {
	r := csv.NewReader(strings.NewReader(data))
	record, err := r.Read()
	if err != nil {
		return err
	}
	s := reflect.ValueOf(v).Elem()
	t := s.Type()
	if s.NumField() != len(record) {
		return &ErrFieldMismatch{s.NumField(), len(record)}
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if t.Field(i).Tag.Get("cid") != "-" {
			switch f.Type() {
			case stringReflectType:
				f.SetString(record[i])
			case intReflectType:
				v, err := strconv.ParseInt(record[i], 10, 0)
				if err != nil {
					return err
				}
				f.SetInt(v)
			case uintReflectType:
				v, err := strconv.ParseUint(record[i], 10, 0)
				if err != nil {
					return err
				}
				f.SetUint(v)
			case boolReflectType:
				switch record[i] {
				case "0":
					f.SetBool(false)
				case "1":
					f.SetBool(true)
				default:
					return &ErrUnreadable{"boolean", record[i]}
				}
			default:
				return &ErrUnsupportedType{Type: f.Type().String()}
			}
		}
	}
	return nil
}

// MarshallCSV from a struct
func MarshallCSV(v any) string {
	s := reflect.ValueOf(v)
	r := make([]string, 0)
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Type() {
		case stringReflectType:
			r = append(r, f.String())
		case intReflectType:
			r = append(r, strconv.FormatInt(f.Int(), 10))
		case uintReflectType:
			r = append(r, strconv.FormatUint(f.Uint(), 10))
		case boolReflectType:
			if f.Bool() {
				r = append(r, "1")
			} else {
				r = append(r, "0")
			}
		}
	}
	b := new(strings.Builder)
	w := csv.NewWriter(b)
	w.Write(r)
	w.Flush()
	return b.String()
}

type ErrFieldMismatch struct {
	Expected, Found int
}

func (e *ErrFieldMismatch) Error() string {
	return fmt.Sprintf("CSV line fields mismatch. Expected %d found %d", e.Expected, e.Found)
}

type ErrUnreadable struct {
	Format, Found string
}

func (e *ErrUnreadable) Error() string {
	return fmt.Sprintf("Unreadable value as %s. Found %s", e.Format, e.Found)
}

type ErrUnsupportedType struct {
	Type string
}

func (e *ErrUnsupportedType) Error() string {
	return "Unsupported type: " + e.Type
}
