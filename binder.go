package easyreq

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

func stringBinder(val string, typ reflect.Type) reflect.Value {
	return reflect.ValueOf(val)
}

func uintBinder(val string, typ reflect.Type) reflect.Value {
	if len(val) == 0 {
		return reflect.Zero(typ)
	}

	uintValue, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetUint(uintValue)
	return pValue.Elem()
}

func intBinder(val string, typ reflect.Type) reflect.Value {
	if len(val) == 0 {
		return reflect.Zero(typ)
	}
	intValue, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetInt(intValue)
	return pValue.Elem()
}

func floatBinder(val string, typ reflect.Type) reflect.Value {
	if len(val) == 0 {
		return reflect.Zero(typ)
	}
	floatValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetFloat(floatValue)
	return pValue.Elem()
}

func boolBinder(val string, typ reflect.Type) reflect.Value {
	v := strings.TrimSpace(strings.ToLower(val))
	switch v {
	case "true":
		return reflect.ValueOf(true)
	}
	// Return false by default.
	return reflect.ValueOf(false)
}

func timeBinder(val string, typ reflect.Type) reflect.Value {
	for _, f := range TimeFormats {
		if f == "" {
			continue
		}

		if strings.Contains(f, "07") || strings.Contains(f, "MST") {
			if r, err := time.Parse(f, val); err == nil {
				return reflect.ValueOf(r)
			}
		} else {
			if r, err := time.ParseInLocation(f, val, time.Local); err == nil {
				return reflect.ValueOf(r)
			}
		}
	}

	if unixInt, err := strconv.ParseInt(val, 10, 64); err == nil {
		return reflect.ValueOf(time.Unix(unixInt, 0))
	}

	return reflect.Zero(typ)
}

func pointerBinder(val string, typ reflect.Type) reflect.Value {
	if len(val) == 0 {
		return reflect.Zero(typ)
	}

	v := BindValue(val, typ.Elem())
	p := reflect.New(v.Type()).Elem()
	p.Set(v)
	return p.Addr()
}

func sliceBinder(vals []string, typ reflect.Type) reflect.Value {
	slices := reflect.MakeSlice(typ, 0, len(vals))
	for i := 0; i < len(vals); i++ {
		val := BindValue(vals[i], typ.Elem())
		slices = reflect.Append(slices, val.Convert(typ.Elem()))
	}

	return slices
}

const (
	// DefaultDateFormat day
	DefaultDateFormat = "2006-01-02"
	// DefaultDatetimeFormat minute
	DefaultDatetimeFormat = "2006-01-02 15:0"
	// DefaultDatetimeFormatSecond second
	DefaultDatetimeFormatSecond = "2006-01-02 15:04:05"
)

// BindValue string to specified type
func BindValue(val string, typ reflect.Type) reflect.Value {
	binder, ok := TypeBinders[typ]
	if !ok {
		binder, ok = KindBinders[typ.Kind()]
		if !ok {
			// WARN.Println("no binder for type:", typ)
			// TODO slice | struct
			return reflect.Zero(typ)
		}
	}

	return binder(val, typ)
}

type binder func(string, reflect.Type) reflect.Value

var (
	// TimeFormats supported time formats, also support unix time and time.RFC3339.
	TimeFormats []string

	// TypeBinders bind type
	TypeBinders = make(map[reflect.Type]binder)

	// KindBinders bind kind
	KindBinders = make(map[reflect.Kind]binder)
)

func init() {
	KindBinders[reflect.Int] = intBinder
	KindBinders[reflect.Int8] = intBinder
	KindBinders[reflect.Int16] = intBinder
	KindBinders[reflect.Int32] = intBinder
	KindBinders[reflect.Int64] = intBinder

	KindBinders[reflect.Uint] = uintBinder
	KindBinders[reflect.Uint8] = uintBinder
	KindBinders[reflect.Uint16] = uintBinder
	KindBinders[reflect.Uint32] = uintBinder
	KindBinders[reflect.Uint64] = uintBinder

	KindBinders[reflect.Float32] = floatBinder
	KindBinders[reflect.Float64] = floatBinder

	KindBinders[reflect.String] = stringBinder
	KindBinders[reflect.Bool] = boolBinder
	KindBinders[reflect.Ptr] = pointerBinder

	TypeBinders[reflect.TypeOf(time.Time{})] = timeBinder

	TimeFormats = append(TimeFormats, DefaultDateFormat, DefaultDatetimeFormat, DefaultDatetimeFormatSecond, time.RFC3339)
}
