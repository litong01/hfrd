package utils

import (
	"fmt"
	"reflect"
	"bytes"
)

const DELIMITER = "."

func PrettyStructString(i interface{}) string {
	params := flat(i)
	var buffer bytes.Buffer
	for i := range params {
		buffer.WriteString("\n\t")
		buffer.WriteString(params[i])
	}
	return buffer.String()
}

func PrettyMapString(m map[string]string) string {
	var buffer bytes.Buffer
	for k, v := range m {
		buffer.WriteString("\n\t")
		buffer.WriteString(k + ": " + v)
	}
	return buffer.String()
}

func flat(i interface{}) []string {
	var res []string
	flatten("", &res, reflect.ValueOf(i))
	return res
}

func flatten(k string, m *[]string, v reflect.Value) {
	delimiter := DELIMITER
	if k == "" {
		delimiter = ""
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			*m = append(*m, fmt.Sprintf("%s =", k))
			return
		}
		flatten(k, m, v.Elem())
	case reflect.Struct:
		if x, ok := v.Interface().(fmt.Stringer); ok {
			*m = append(*m, fmt.Sprintf("%s = %v", k, x))
			return
		}

		for i := 0; i < v.NumField(); i++ {
			flatten(k+delimiter+v.Type().Field(i).Name, m, v.Field(i))
		}
	case reflect.String:
		// It is useful to quote string values
		*m = append(*m, fmt.Sprintf("%s = \"%s\"", k, v))
	default:
		*m = append(*m, fmt.Sprintf("%s = %v", k, v))
	}
}
