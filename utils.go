package xconv

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func _log(prefix string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		var format []string
		if len(args) == 1 {
			format = []string{"%s"}
		} else if val := reflect.Indirect(reflect.ValueOf(args[0])); val.Kind() == reflect.String {
			format, args = []string{val.String()}, args[1:len(args)]
		} else {
			format = []string{""}
			for i, _ := range args {
				format = append(format, fmt.Sprintf("%d:\t", i)+"%s")
			}
		}
		msg = fmt.Sprintf(strings.Join(format, "\n"), args...)
	}
	if prefix != "" {
		prefix = fmt.Sprintf("%s:", prefix)
	}
	log.Println(fmt.Sprintf("GoConv:%s%s", prefix, msg))
}

func debug(args ...interface{}) {
	_log("DEBUG", args...)
}

func warning(args ...interface{}) {
	_log("WARNING", args...)
}

func isFuncValid(fType reflect.Type, inTypes, outTypes []interface{}) bool {
	isValid := func(t reflect.Type, required interface{}) bool {
		if required == nil {
			return true
		}
		if typ, ok := required.(reflect.Type); ok {
			if typ.Kind() == reflect.Interface {
				return t.Implements(typ)
			}
			return typ == t
		}
		if kind, ok := required.(reflect.Kind); ok {
			return kind == t.Kind()
		}
		return true
	}

	if fType.Kind() != reflect.Func {
		return false
	}
	if inTypes != nil {
		if fType.NumIn() != len(inTypes) {
			return false
		}
		for i, t := range inTypes {
			if !isValid(fType.In(i), t) {
				return false
			}
		}
	}
	if outTypes != nil {
		if fType.NumOut() != len(outTypes) {
			return false
		}
		for i, t := range outTypes {
			if !isValid(fType.Out(i), t) {
				return false
			}
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
