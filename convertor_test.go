package xconv

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestSlice(t *testing.T) {
	{
		var src = [3]int{1, 2, 3}
		var dst [4]int32
		Convert(src, &dst)
		debug(src, dst)
		if dst != [4]int32{1, 2, 3, 0} {
			t.Error("convert [3]int to [4]int32 error")
		}
	}
	{
		var src = []int{1, 2, 3}
		var dst []int32
		Convert(src, &dst)
		debug(src, dst)
		if len(dst) != len(src) {
			t.Error("convert []int to []int32 error")
		}
	}
	{
		now := time.Now()
		var src = []time.Time{now, now.AddDate(1, 0, 0), now.AddDate(0, 1, 0)}
		var dst []string
		Convert(src, &dst)
		debug(src, dst)
		if len(dst) != len(src) {
			t.Error("convert []time.Time to []string error")
		}
	}
}

func TestStruct(t *testing.T) {
	{
		type subStruct struct {
			Z string
		}
		type SrcStruct struct {
			Sub subStruct
			A   int32
			B   int32
		}
		type DstStruct struct {
			Sub *subStruct
			A   int
			C   int
		}
		src := SrcStruct{A: 1, B: 2, Sub: subStruct{Z: "sub struct"}}
		var dst1 DstStruct
		Convert(src, &dst1)
		debug(src, dst1)
		if dst1.A != int(src.A) || dst1.C != 0 || dst1.Sub.Z != src.Sub.Z {
			t.Error("convert SrcStruct to DstStruct error")
		}
		var dst2 = new(DstStruct)
		Convert(src, dst2)
		debug(src, dst2)
		if dst2.A != int(src.A) || dst2.C != 0 || dst2.Sub.Z != src.Sub.Z {
			t.Error("convert SrcStruct to DstStruct error")
		}
	}
	{
		type SrcStruct struct {
			A int32
			B float32
		}
		type DstStruct struct {
			A int
			B float64
		}
		src := []SrcStruct{
			SrcStruct{A: 1, B: 0.1},
			SrcStruct{A: 2, B: 0.2},
			SrcStruct{A: 3, B: 0.3},
		}
		var dst1 []DstStruct
		Convert(src, &dst1)
		debug(src, dst1)
		if len(src) != len(dst1) {
			t.Error("convert []SrcStruct to []DstStruct error")
		}
		var dst2 []*DstStruct
		Convert(src, &dst2)
		debug(src, dst2)
		if len(src) != len(dst2) {
			t.Error("convert []SrcStruct to []*DstStruct error")
		}
	}
}

func TestMap(t *testing.T) {
	var src = map[int32]int64{1: 2, 3: 4}
	var dst map[int]int
	Convert(src, &dst)
	debug(src, dst)
	if len(src) != len(dst) {
		t.Error("convert map[int32][int64] to map[int][int] error")
	}
}

func TestInteger(t *testing.T) {
	{
		var src int = 100
		var dst int32
		Convert(src, &dst)
		debug(src, dst)
		if dst != int32(src) {
			t.Error("convert int to int32 error")
		}
	}
	{
		var src int64 = 256
		var dst int8
		Convert(src, &dst)
		debug(src, dst)
		if dst != int8(src) {
			t.Error("convert int64 to int8 error")
		}
	}
}

func TestFloat(t *testing.T) {
	var src float32 = 123.45
	var dst float64
	Convert(src, &dst)
	debug(src, dst)
	if dst != float64(src) {
		t.Error("convert float32 to float64 error")
	}
}

func TestTime(t *testing.T) {
	{
		var src time.Time = time.Now()
		var dst1 int64
		Convert(src, &dst1)
		debug(src, dst1)
		if dst1 != src.Unix() {
			t.Error("convert time to int64 error")
		}
		var dst2 time.Time
		Convert(dst1, &dst2)
		debug(src, dst2)
		if src.Sub(dst2).Seconds() > 1 {
			t.Error("convert int64 to time error")
		}
	}
	{
		var src time.Time = time.Now()
		var dst1 string
		Convert(src, &dst1)
		debug(src, dst1)
		if dst1 != src.Format(TIME_FORMAT) {
			t.Error("convert time to string error")
		}
		var dst2 time.Time
		Convert(dst1, &dst2)
		debug(src, dst2)
		if src.Sub(dst2).Seconds() > 1 {
			t.Error("convert string to time error")
		}
	}
}

func TestReturn(t *testing.T) {
	{
		var src int64 = 128
		var dst1 int
		dst2 := Convert(src, &dst1).(int)
		debug(src, dst1, dst2)
		if dst1 != int(src) && dst2 != dst1 {
			t.Error("convert int64 to int (return) error")
		}
	}
	{
		var src string = "2016-08-17 23:43:12"
		var dst1 time.Time
		dst2 := Convert(src, &dst1).(time.Time)
		debug("\n0:\t%s\n1:\t%s\n2:\t%s", src, dst1, dst2)
		if dst1.Format(TIME_FORMAT) != src || !dst2.Equal(dst1) {
			t.Error("convert string to time (return) error")
		}
	}
}

func TestCustom(t *testing.T) {
	{
		var src int64 = 100
		var dst string
		NewConvertor(src).
			Rule(reflect.Int64, reflect.String,
				func(c *Convertor, src, dst reflect.Value) {
					dst.SetString(strconv.Itoa(int(src.Int())))
				}).
			Apply(&dst)
		debug(src, dst)
		if dst != "100" {
			t.Error("convert int64 to string error")
		}
	}
	{
		type SrcStruct struct {
			A int32
			B int
		}
		type DstStruct struct {
			A string
			C int64
		}
		var src = SrcStruct{A: 1, B: 2}
		var dst DstStruct
		NewConvertor(src).
			Rule(reflect.Int32, StringTypes,
				func(c *Convertor, src, dst reflect.Value) {
					dst.SetString(strconv.Itoa(int(src.Int())))
				}).
			Field("C",
				func(srcObj SrcStruct) int64 {
					return int64(srcObj.B)
				}).
			Apply(&dst)
		debug(src, dst)
		if dst.A != "1" || dst.C != 2 {
			t.Error("convert SrcStruct to DstStruct error")
		}
	}
	{
		ConvertMap.Set(IntTypes, StringTypes,
			func(c *Convertor, src, dst reflect.Value) {
				dst.SetString(strconv.Itoa(int(src.Int())))
			})
		var src int = 100
		var dst string
		Convert(src, &dst)
		debug(src, dst)
		if dst != "100" {
			t.Error("convert integer to string error")
		}
	}
}

func TestConvertor(t *testing.T) {
	{
		type Base struct {
			B int
		}
		type Src struct {
			A    string
			B    int
			C    time.Time
			D    int32
			x    int
			Z    []int
			Same []*Base
		}
		type Dst struct {
			*Base
			A    string
			C    string
			x    int
			E    string
			Z    []int32
			Same []*Base
		}
		dst := Dst{}
		src := &Src{A: "one", B: 1024, C: time.Now(), D: 32, Z: []int{1, 2, 3}, Same: []*Base{&Base{B: 10}}}
		NewConvertor(src).
			Rule(reflect.String, reflect.String,
				func(c *Convertor, src, dst reflect.Value) {
					dst.SetString(src.String() + "xxxxxxxxxxxxx")
				}).
			Field("E",
				func(srcObj *Src) string {
					return strconv.Itoa(int(srcObj.D + 100))
				}).
			Field("Base",
				func(srcObj *Src) *Base {
					return &Base{B: srcObj.B}
				}).
			TimeFormat(DATE_FORMAT).
			Apply(&dst)
		dst.Same[0].B = 1000
		debug(src, dst)
	}
	{
		type Src struct {
			A string
		}
		type Dst struct {
			B string
		}
		src1 := map[int]Src{1: Src{"one"}, 2: Src{"two"}}
		var dst1 map[int32]*Dst
		NewConvertor(src1).Field("B", func(obj Src) string { return obj.A }).Apply(&dst1)
		debug(src1, dst1)

		src2 := map[int]*Src{1: &Src{"one"}, 2: &Src{"two"}}
		var dst2 map[int32]Dst
		NewConvertor(src2).Field("B", func(obj *Src) string { return obj.A }).Apply(&dst2)
		debug(src2, dst2)
	}
	{
		type SrcBase struct {
			A string
		}
		type Src struct {
			Base SrcBase
			A    string
		}
		type DstBase struct {
			B int
		}
		type Dst struct {
			Base *DstBase
			B    string
			DstBase
		}
		src := Src{Base: SrcBase{"one"}, A: "two"}
		var dst *Dst
		dst = new(Dst)
		NewConvertor(src).
			Field("B", func(obj Src) string { return obj.A }).
			Field("Base.B", func(obj SrcBase) int { return 1 }).
			Field("DstBase", func(obj Src) DstBase { return DstBase{B: 3} }).
			Apply(dst)
		debug(src, dst)
	}
}
