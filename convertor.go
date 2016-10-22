package xconv

import (
	"fmt"
	"reflect"
	"strings"
)

type Convertor struct {
	timeFormat string

	src                reflect.Value
	fieldRules         map[string]reflect.Value
	fieldRulesUsed     map[string]bool
	fieldRulesMustUsed bool
	convertMap         *convertMapT
	fieldStack         []string
}

func Convert(src, dst interface{}) {
	NewConvertor(src).Apply(dst)
}

func NewConvertor(src interface{}) *Convertor {
	srcVal := reflect.ValueOf(src)
	return &Convertor{
		timeFormat:     TIME_FORMAT,
		src:            srcVal,
		fieldRules:     make(map[string]reflect.Value, 0),
		fieldRulesUsed: make(map[string]bool),
		convertMap:     newConvertMap(),
		fieldStack:     make([]string, 0),
	}
}

func (c *Convertor) FieldRuleMustUsed() *Convertor {
	c.fieldRulesMustUsed = true
	return c
}

func (c *Convertor) Rule(inType, outType interface{}, rule ConvertFuncT) *Convertor {
	if inType == nil || outType == nil {
		return c
	}
	toSlice := func(v interface{}) []interface{} {
		if reflect.TypeOf(v).Kind() != reflect.Slice {
			return []interface{}{v}
		}
		val := reflect.ValueOf(v)
		rs := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			rs[i] = val.Index(i).Interface()
		}
		return rs
	}
	c.convertMap.Set(toSlice(inType), toSlice(outType), rule)
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
		if !dstVal.CanSet() {
			dstVal = reflect.Indirect(dstVal)
			break
		}
		val := reflect.New(dstVal.Type().Elem())
		dstVal.Set(val)
		dstVal = val.Elem()
	}
	return dstVal
}

func (c *Convertor) Apply(dst interface{}) {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		panic("Dst type must be ptr, and not nil.")
	}
	if !reflect.Indirect(c.src).IsValid() {
		return
	}
	dstVal = makeDstVal(dstVal)
	c.apply(c.src, dstVal)
	if c.fieldRulesMustUsed {
		for fieldName, _ := range c.fieldRules {
			if _, ok := c.fieldRulesUsed[fieldName]; !ok {
				panic(fmt.Sprintf("Field \"%s\" is noused", fieldName))
			}
		}
	}
	c.fieldRulesUsed = make(map[string]bool)
}

func (c *Convertor) apply(src, dstVal reflect.Value) {
	srcVal := src
	if srcVal.Kind() == reflect.Ptr {
		srcVal = reflect.Indirect(srcVal)
	}
	if !srcVal.IsValid() {
		return
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
	default:
		c.applyField(src, srcVal, dstVal)
	}
}

func (c *Convertor) applyStruct(src, srcVal, dstVal reflect.Value) {
	if !srcVal.IsValid() {
		dstVal.Set(reflect.ValueOf(nil))
		return
	}
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
			c.fieldRulesUsed[ruleName] = true
			r := fieldRule.Call([]reflect.Value{src})
			fieldVal.Set(r[0])
			continue
		}
		if _, has := srcTyp.FieldByName(fieldName); has {
			c.fieldStack = append(c.fieldStack, fieldName)
			val := srcVal.FieldByName(fieldName)
			if reflect.Indirect(val).IsValid() {
				fieldVal = makeDstVal(fieldVal)
				c.apply(val, fieldVal)
			}
			c.fieldStack = c.fieldStack[0 : len(c.fieldStack)-1]
		} else {
			warning("Field '%s' is not found", ruleName)
		}
	}
}

func (c *Convertor) applyField(src, srcVal, dstVal reflect.Value) {
	if !srcVal.IsValid() {
		dstVal.Set(reflect.ValueOf(nil))
		return
	}
	srcIsIf := srcVal.Kind() == reflect.Interface
	dstIsIf := dstVal.Kind() == reflect.Interface
	if srcIsIf && dstIsIf {
		dstVal.Set(srcVal)
		return
	} else if srcIsIf {
		newSrcVal := reflect.ValueOf(srcVal.Interface())
		c.apply(newSrcVal, dstVal)
		return
	} else if dstIsIf {
		newDstVal := reflect.New(srcVal.Type()).Elem()
		c.apply(srcVal, newDstVal)
		dstVal.Set(newDstVal)
		return
	}
	var skipConvertMap bool
	if f, has := c.convertMap.Get(srcVal, dstVal); has {
		if f == nil {
			skipConvertMap = true
		} else {
			f(c, srcVal, dstVal)
			return
		}
	}
	if dstVal.Kind() != reflect.Struct && srcVal.Type() == dstVal.Type() {
		dstVal.Set(srcVal)
		return
	}
	if !skipConvertMap {
		if f, has := ConvertMap.Get(srcVal, dstVal); has {
			f(c, srcVal, dstVal)
			return
		}
	}
	if dstVal.Kind() == reflect.Struct && srcVal.Kind() == reflect.Struct {
		c.applyStruct(src, srcVal, dstVal)
		return
	}
	warning("'%s' to '%s' convertor is not found", srcVal.Type(), dstVal.Type())
}
