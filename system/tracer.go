package system

import (
	"bytes"
	"fmt"
	"reflect"
)

// CpuTracer shows a pretty print of the CPU state. s must be a pointer to a
// struct with exported fields.
func CpuTracer(s interface{}) string {
	buf := new(bytes.Buffer)

	ptr := reflect.ValueOf(s)
	if kind := ptr.Kind(); kind != reflect.Ptr {
		panic(fmt.Sprintf("CpuTracer: expected a struct pointer, got %v", kind))
	}
	v := ptr.Elem()
	if kind := v.Kind(); kind != reflect.Struct {
		panic(fmt.Sprintf("CpuTracer: expected a struct pointer, got pointer to %v", kind))
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		if f.PkgPath != "" {
			// Unexported.
			continue
		}
		name := f.Name
		if f.Tag != "" {
			name = string(f.Tag)
		}
		switch v.Field(i).Kind() {
		case reflect.Int, reflect.Uint8:
			fmt.Fprintf(buf, "%v: %02x\n", name, v.Field(i).Interface())
		case reflect.Uint16:
			fmt.Fprintf(buf, "%v: %04x\n", name, v.Field(i).Interface())
		case reflect.Slice:
			s := v.Field(i).Interface()
			if bs, ok := s.([]byte); ok {
				fmt.Fprintf(buf, "%v: ", name)
				for i, v := range bs {
					fmt.Fprintf(buf, "%v[%0x]=%02x", name, i, v)
					if i < len(bs)-1 {
						fmt.Fprintf(buf, ", ")
					}
				}
			} else {
				fmt.Fprintf(buf, "%v: % 02x", name, v.Field(i).Interface())
			}
			fmt.Fprintln(buf)
		default:
			fmt.Fprintf(buf, "%v: %v\n", name, v.Field(i).Interface())
		}

	}
	return buf.String()
}
