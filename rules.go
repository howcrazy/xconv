package goconv

import (
	"fmt"
	"reflect"
	"time"
)

const (
	DATE_FORMAT = "2006-01-02"
	TIME_FORMAT = "2006-01-02 15:04:05"
)

type ConvertFuncT func(*Convertor, reflect.Value, reflect.Value)

var ConvertMap *convertMapT

func init() {
	ConvertMap = newConvertMap()
	// INTEGER
	intTypes := []interface{}{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64}
	ConvertMap.Set(intTypes, intTypes,
		func(c *Convertor, src, dst reflect.Value) { dst.SetInt(src.Int()) })
	// FLOAT
	floatTypes := []interface{}{reflect.Float32, reflect.Float64}
	ConvertMap.Set(floatTypes, floatTypes,
		func(c *Convertor, src, dst reflect.Value) { dst.SetFloat(src.Float()) })
	// STRING
	stringTypes := []interface{}{reflect.String}
	ConvertMap.Set(stringTypes, stringTypes,
		func(c *Convertor, src, dst reflect.Value) { dst.SetString(src.String()) })
	// TIME
	timeIntTypes := []interface{}{reflect.Int, reflect.Int32, reflect.Int64}
	timeTypes := []interface{}{new(time.Time)}
	ConvertMap.Set(timeIntTypes, timeTypes,
		func(c *Convertor, src, dst reflect.Value) { dst.Set(reflect.ValueOf(time.Unix(src.Int(), 0))) })
	ConvertMap.Set(timeTypes, timeIntTypes,
		func(c *Convertor, src, dst reflect.Value) { dst.SetInt(src.Interface().(time.Time).Unix()) })
	ConvertMap.Set(stringTypes, timeTypes,
		func(c *Convertor, src, dst reflect.Value) {
			t, _ := time.Parse(c.timeFormat, src.String())
			dst.Set(reflect.ValueOf(t))
		})
	ConvertMap.Set(timeTypes, stringTypes,
		func(c *Convertor, src, dst reflect.Value) {
			dst.SetString(src.Interface().(time.Time).Format(c.timeFormat))
		})
}

type convertMapT struct {
	cmap map[string]ConvertFuncT
}

func newConvertMap() *convertMapT {
	return &convertMapT{cmap: make(map[string]ConvertFuncT)}
}

func (cm *convertMapT) Set(inTypes, outTypes []interface{}, convertorF ConvertFuncT) {
	for _, inType := range inTypes {
		for _, outType := range outTypes {
			key := cm.key(inType, outType)
			cm.cmap[key] = convertorF
		}
	}
}

func (cm *convertMapT) Get(inVal, outVal reflect.Value) (f ConvertFuncT, has bool) {
	inType, outType := inVal.Type(), outVal.Type()
	for _, in := range []interface{}{inType, inType.Kind()} {
		if k, kok := in.(reflect.Kind); kok {
			if k == reflect.Struct {
				continue
			}
		}
		for _, out := range []interface{}{outType, outType.Kind()} {
			if k, kok := out.(reflect.Kind); kok {
				if k == reflect.Struct {
					continue
				}
			}
			key := cm.key(in, out)
			f, has = cm.cmap[key]
			if has {
				return
			}
		}
	}
	return
}

func (cm *convertMapT) typeName(typ interface{}) string {
	if t, ok := typ.(reflect.Kind); ok {
		return fmt.Sprintf("KIND-%s", t)
	}
	if t, ok := typ.(reflect.Type); ok {
		return fmt.Sprintf("TYPE-%s", t)
	}
	if t, ok := typ.(reflect.Value); ok {
		return fmt.Sprintf("TYPE-%s", t.Type())
	}
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	return fmt.Sprintf("TYPE-%s", t)
}

func (cm *convertMapT) key(inType, outType interface{}) string {
	return fmt.Sprintf("%s:%s", cm.typeName(inType), cm.typeName(outType))
}
