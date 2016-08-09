package goconv

import (
	"reflect"
	"strings"
)

type Convertor struct {
	timeFormat string

	src        reflect.Value
	fieldRules map[string]reflect.Value
	convertMap *convertMapT
	fieldStack []string
}

func Convert(src, dst interface{}) {
	NewConvertor(src).Apply(dst)
}

func NewConvertor(src interface{}) *Convertor {
	srcVal := reflect.ValueOf(src)
	return &Convertor{
		timeFormat: TIME_FORMAT,
		src:        srcVal,
		fieldRules: make(map[string]reflect.Value, 0),
		convertMap: newConvertMap(),
		fieldStack: make([]string, 0),
	}
}

func (c *Convertor) Rule(inType, outType interface{}, rule ConvertFuncT) *Convertor {
	c.convertMap.Set([]interface{}{inType}, []interface{}{outType}, rule)
	return c
}

func (c *Convertor) Rules(inTypes, outTypes []interface{}, rule ConvertFuncT) *Convertor {
	c.convertMap.Set(inTypes, outTypes, rule)
	return c
}

func (c *Convertor) Field(fieldName string, convertorF interface{}) *Convertor {
	cVal := reflect.ValueOf(convertorF)
	if !isFuncValid(cVal.Type(), []interface{}{nil}, []interface{}{nil}) {
		panic("Field convertor function invalid")
	}
	c.fieldRules[fieldName] = cVal
	return c
}

func (c *Convertor) TimeFormat(format string) *Convertor {
	if format == "" {
		format = TIME_FORMAT
	}
	c.timeFormat = format
	return c
}

func makeDstVal(dstVal reflect.Value) reflect.Value {
	for dstVal.Kind() == reflect.Ptr {
		val := reflect.New(dstVal.Type().Elem())
		dstVal.Set(val)
		dstVal = val.Elem()
	}
	return dstVal
}

func (c *Convertor) Apply(dst interface{}) {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() == reflect.Ptr {
		dstVal = reflect.Indirect(dstVal)
	}
	c.apply(c.src, dstVal)
}

func (c *Convertor) apply(src, dstVal reflect.Value) {
	srcVal := src
	if srcVal.Type().Kind() == reflect.Ptr {
		srcVal = reflect.Indirect(srcVal)
	}
	dstVal = makeDstVal(dstVal)
	switch k := dstVal.Kind(); k {
	case reflect.Slice, reflect.Array:
		val := dstVal
		if k == reflect.Slice {
			val = reflect.MakeSlice(dstVal.Type(), srcVal.Len(), srcVal.Cap())
			defer dstVal.Set(val)
		}
		for idx := 0; idx < min(val.Len(), srcVal.Len()); idx++ {
			c.apply(srcVal.Index(idx), val.Index(idx))
		}
	case reflect.Map:
		val := reflect.MakeMap(dstVal.Type())
		dstType := val.Type()
		for _, keyVal := range srcVal.MapKeys() {
			valueVal := srcVal.MapIndex(keyVal)
			dstKeyVal := reflect.New(dstType.Key()).Elem()
			c.apply(keyVal, dstKeyVal)
			dstValueVal := reflect.New(dstType.Elem()).Elem()
			c.apply(valueVal, dstValueVal)
			val.SetMapIndex(dstKeyVal, dstValueVal)
		}
		dstVal.Set(val)
	default:
		c.applyField(src, srcVal, dstVal)
	}
}

func (c *Convertor) applyStruct(src, srcVal, dstVal reflect.Value) {
	dstTyp := dstVal.Type()
	srcTyp := srcVal.Type()
	for idx := 0; idx < dstTyp.NumField(); idx++ {
		fieldTyp := dstTyp.FieldByIndex([]int{idx})
		fieldVal := dstVal.FieldByIndex([]int{idx})
		fieldName := fieldTyp.Name
		if !fieldVal.CanSet() {
			warning("Field '%s' can not set", fieldName)
			continue
		}
		ruleName := strings.Join(append(c.fieldStack, fieldName), ".")
		if fieldRule, ok := c.fieldRules[ruleName]; ok {
			r := fieldRule.Call([]reflect.Value{src})
			fieldVal.Set(r[0])
			continue
		}
		fieldVal = makeDstVal(fieldVal)
		if _, has := srcTyp.FieldByName(fieldName); has {
			c.fieldStack = append(c.fieldStack, fieldName)
			c.apply(srcVal.FieldByName(fieldName), fieldVal)
			c.fieldStack = c.fieldStack[0 : len(c.fieldStack)-1]
		} else {
			warning("Field '%s' is not found", ruleName)
		}
	}
}

func (c *Convertor) applyField(src, srcVal, dstVal reflect.Value) {
	if f, has := c.convertMap.Get(srcVal, dstVal); has {
		f(c, srcVal, dstVal)
		return
	}
	if srcVal.Type() == dstVal.Type() {
		dstVal.Set(srcVal)
		return
	}
	if f, has := ConvertMap.Get(srcVal, dstVal); has {
		f(c, srcVal, dstVal)
		return
	}
	if dstVal.Kind() == reflect.Struct {
		c.applyStruct(src, srcVal, dstVal)
		return
	}
	warning("'%s' to '%s' convertor is not found", srcVal.Type(), dstVal.Type())
}
